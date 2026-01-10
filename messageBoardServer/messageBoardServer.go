package messageboardserver

import (
	protobufRazpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"
	"PS_projekt/storage"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	randMu              sync.Mutex
	reconfigModeCh      chan struct{}
	reconfigMode        atomic.Bool
	topicErr            = fmt.Errorf("topic does not exist")
	userExistsErr       = fmt.Errorf("user does not exist")
	messageExistsErr    = fmt.Errorf("message does not exist")
	emptyTopicErr       = fmt.Errorf("topic can not be empty")
	emptyUsernameErr    = fmt.Errorf("username can not be empty")
	clientStopedErr     = fmt.Errorf("the next client has stoped responding")
	FailErr             = fmt.Errorf("sending data has failed")
	DataNotAvailableErr = fmt.Errorf("data not availabale")
)

// generate uint32 uids
func GenerateRand32[T comparable](s *storage.LockableMap[int64, T]) int64 {
	for {
		// generate id under randMu
		randMu.Lock()
		id := rand.Uint32()
		randMu.Unlock()
		// check uniqueness
		_, ok := s.GetValByKey(int64(id))
		if !ok {
			return int64(id)
		}
		// collision; try again
	}
}

type MessageBoardServer struct {
	protobufRazpravljalnica.MessageBoardServer
	Id           int64
	Version      atomic.Int64
	UserStorage  *storage.LockableMap[int64, *UserData]
	TopicStorage *storage.LockableMap[int64, *TopicData]
	//map of [3]{messageId , userId , topicId}
	MessageStorage *storage.LockableMap[int64, *MessageData]
	//map to track used message ids
	ConnNext   *grpc.ClientConn
	ConnPrev   *grpc.ClientConn
	ConnTail   *grpc.ClientConn
	ClientCurr protobufRazpravljalnica.MessageBoardClient
	ClientNext protobufRazpravljalnica.MessageBoardClient
	ClientPrev protobufRazpravljalnica.MessageBoardClient
	ClientTail protobufRazpravljalnica.MessageBoardClient
	// subscription support
	subMu       sync.Mutex
	subscribers map[int]chan *protobufRazpravljalnica.MessageEvent
	nextSubId   int
	seq         int64 // ADJUST ONLY USING ATOMIC
}

// initialize the storage
func NewMessageBoardServer(id int64) *MessageBoardServer {
	userStorage := storage.NewLockableMap[int64, *UserData]()
	topicStorage := storage.NewLockableMap[int64, *TopicData]()
	messageStorage := storage.NewLockableMap[int64, *MessageData]()
	return &MessageBoardServer{protobufRazpravljalnica.UnimplementedMessageBoardServer{}, id, atomic.Int64{}, userStorage, topicStorage, messageStorage, nil, nil, nil, nil, nil, nil, nil, sync.Mutex{}, make(map[int]chan *protobufRazpravljalnica.MessageEvent), 0, 0}
}

// generates random user id and adds to map
// returns error if the user name consists of spaces
func (server *MessageBoardServer) CreateUser(ctx context.Context, in *protobufRazpravljalnica.CreateUserRequest) (*protobufRazpravljalnica.User, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, emptyTopicErr
	}

	if reconfigMode.Load() {

		<-reconfigModeCh
	}
	if server.ClientPrev == nil {
		in.Version = server.GetVersion()
	}

	id := GenerateRand32(server.UserStorage)
	user := &protobufRazpravljalnica.User{Id: id, Name: in.Name}
	userData := &UserData{User: user, Dirty: true}
	server.UserStorage.Put(int64(id), userData)
	if server.ClientNext == nil {
		fmt.Println("the data has arrived at the tail")
		userData.Dirty = false
		server.UserStorage.Put(int64(id), userData)
		return user, nil
	}
	if fail() {
		fmt.Println("sending data to the next node has failed")
		return nil, fmt.Errorf("sending data has failed")
	}
	fmt.Println("the data has been forwarded")
	fmt.Println("next server is ", server.ClientNext)
	if user, err := server.ClientNext.CreateUser(ctx, in); status.Code(err) == codes.Unavailable {
		//the other node had been disconnected from the chain
		server.handleUnavailableNode()
		server.ClientNext = nil
		return nil, clientStopedErr
	} else if err != nil {
		//some other error
		fmt.Println(err)
		return nil, err
	} else {
		//successful addition of the user
		//received clean data
		userData.Dirty = false
		server.UserStorage.Put(int64(id), userData)
		return user, nil
	}
}

func (server *MessageBoardServer) CreateTopic(ctx context.Context, in *protobufRazpravljalnica.CreateTopicRequest) (*protobufRazpravljalnica.Topic, error) {
	fmt.Println("data saved in ", server.Id)
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, emptyTopicErr
	}
	if reconfigMode.Load() {

		<-reconfigModeCh
	}

	if server.ClientPrev == nil {
		in.Version = server.GetVersion()
	}

	id := GenerateRand32(server.TopicStorage)
	topic := &protobufRazpravljalnica.Topic{Name: name, Id: id}
	topicData := &TopicData{Topic: topic, Dirty: true}
	if server.ClientNext == nil {
		fmt.Println("the data has arrived at the tail")
		topicData.Dirty = false
		fmt.Println("data is not dirty")
		server.TopicStorage.Put(id, topicData)
		return topic, nil
	}

	server.TopicStorage.Put(id, topicData)
	if fail() {
		return nil, FailErr
	}

	if topic, err := server.ClientNext.CreateTopic(ctx, in); status.Code(err) == codes.Unavailable {
		server.handleUnavailableNode()
		return nil, clientStopedErr
	} else if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		topicData.Dirty = false
		server.TopicStorage.Put(id, topicData)
		fmt.Println("data has been uploaded successfully")
		return topic, nil
	}
}

func (server *MessageBoardServer) PostMessage(ctx context.Context, in *protobufRazpravljalnica.PostMessageRequest) (*protobufRazpravljalnica.Message, error) {
	topicId, userId := in.TopicId, in.UserId
	if _, ok := server.UserStorage.GetValByKey(userId); !ok {
		return nil, userExistsErr
	}
	if _, ok := server.TopicStorage.GetValByKey(topicId); !ok {
		return nil, topicErr
	}

	if reconfigMode.Load() {

		<-reconfigModeCh
	}
	if server.ClientPrev == nil {
		in.Version = server.GetVersion()
	}
	messageId := GenerateRand32(server.MessageStorage)
	message := &protobufRazpravljalnica.Message{Id: messageId, TopicId: topicId, UserId: userId, Text: in.Text, CreatedAt: timestamppb.Now()}
	messageData := &MessageData{Message: message, Dirty: true}
	server.MessageStorage.Put(messageId, messageData)
	if server.ClientNext == nil {
		return message, nil
	}
	if fail() {
		return nil, FailErr
	}
	if msg, err := server.PostMessage(ctx, in); err == nil {
		messageData.Dirty = false
		server.MessageStorage.Put(messageId, messageData)
		// publish event
		seq := atomic.AddInt64(&server.seq, 1)
		event := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_POST, Message: message, EventAt: timestamppb.Now()}
		server.publishEvent(event)
		return msg, nil
	} else if status.Code(err) == codes.Unavailable {
		server.handleUnavailableNode()
		return nil, clientStopedErr
	} else {
		return nil, err
	}

}

func (server *MessageBoardServer) UpdateMessage(ctx context.Context, in *protobufRazpravljalnica.UpdateMessageRequest) (*protobufRazpravljalnica.Message, error) {
	if _, ok := server.UserStorage.GetValByKey(in.UserId); !ok {
		return nil, userExistsErr
	}
	if _, ok := server.TopicStorage.GetValByKey(in.TopicId); !ok {
		return nil, topicErr
	}
	msgData, ok := server.MessageStorage.GetValByKey(in.MessageId)
	if !ok {
		return nil, messageExistsErr
	}

	if reconfigMode.Load() {

		<-reconfigModeCh
	}

	if server.ClientPrev == nil {
		in.Version = server.GetVersion()
	}

	msgData.Text = in.Text
	msgData.Dirty = true
	server.MessageStorage.Put(in.MessageId, msgData)

	seq := atomic.AddInt64(&server.seq, 1)
	event := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_UPDATE, Message: msgData.Message, EventAt: timestamppb.Now()}
	server.publishEvent(event)
	if server.ClientNext == nil {
		return msgData.Message, nil
	}
	if msg, err := server.ClientNext.UpdateMessage(ctx, in); err == nil {
		msgData.Dirty = false

		server.MessageStorage.Put(in.MessageId, msgData)
		return msg, nil
	} else if status.Code(err) == codes.Unavailable {
		server.handleUnavailableNode()
		return nil, clientStopedErr
	} else {
		return nil, err

	}
}

func (server *MessageBoardServer) DeleteMessage(ctx context.Context, in *protobufRazpravljalnica.DeleteMessageRequest) (*emptypb.Empty, error) {
	// prepare event for delete (include message id)
	if msgData, ok := server.MessageStorage.GetValByKey(in.MessageId); ok {
		seq := atomic.AddInt64(&server.seq, 1)
		event := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_DELETE, Message: msgData.Message, EventAt: timestamppb.Now()}
		server.publishEvent(event)
	}

	if reconfigMode.Load() {
		<-reconfigModeCh
	}
	server.MessageStorage.Delete(in.MessageId)
	if server.ClientNext == nil {
		return &emptypb.Empty{}, nil
	} else if empty, err := server.ClientNext.DeleteMessage(ctx, in); err == nil {
		return empty, nil
	} else if status.Code(err) == codes.Unavailable {
		server.handleUnavailableNode()
		return nil, clientStopedErr
	} else {
		return nil, err
	}
}

func (server *MessageBoardServer) ListTopics(ctx context.Context, empty *emptypb.Empty) (*protobufRazpravljalnica.ListTopicsResponse, error) {
	topicsData := server.TopicStorage.GetAllValues()
	topics := make([]*protobufRazpravljalnica.Topic, len(topicsData))
	for i, topic := range topicsData {
		topics[i] = topic.Topic
	}

	listTopicResponse := &protobufRazpravljalnica.ListTopicsResponse{Topics: topics}
	return listTopicResponse, nil
}

func (server *MessageBoardServer) LikeMessage(ctx context.Context, in *protobufRazpravljalnica.LikeMessageRequest) (*protobufRazpravljalnica.Message, error) {
	if _, ok := server.UserStorage.GetValByKey(in.UserId); !ok {
		return nil, userExistsErr
	}
	if _, ok := server.TopicStorage.GetValByKey(in.TopicId); !ok {
		return nil, topicErr
	}

	message, ok := server.MessageStorage.GetValByKey(in.MessageId)
	if !ok {
		return nil, messageExistsErr
	}

	if reconfigMode.Load() {

		<-reconfigModeCh
	}
	message.Likes++
	server.MessageStorage.Put(in.MessageId, message)

	seq := atomic.AddInt64(&server.seq, 1)
	event := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_LIKE, Message: message.Message, EventAt: timestamppb.Now()}
	server.publishEvent(event)
	if server.ClientNext == nil {
		return message.Message, nil
	} else if message, err := server.ClientNext.LikeMessage(ctx, in); err == nil {
		return message, nil
	} else if codes.Unavailable == status.Code(err) {
		server.handleUnavailableNode()
		return nil, clientStopedErr
	} else {
		return nil, err
	}

}

func (server *MessageBoardServer) GetMessages(ctx context.Context, in *protobufRazpravljalnica.GetMessagesRequest) (*protobufRazpravljalnica.GetMessagesResponse, error) {
	messagesData := server.MessageStorage.GetAllValues()
	messages := make([]*protobufRazpravljalnica.Message, len(messagesData))
	for i, message := range messagesData {
		messages[i] = message.Message
	}
	getMessagesResponse := protobufRazpravljalnica.GetMessagesResponse{Messages: messages}
	return &getMessagesResponse, nil
}

func fail() bool {
	p := rand.Float32()
	if p < 0.25 {
		return true
	}
	return false
}

func (server *MessageBoardServer) handleUnavailableNode() {
	if server.ConnNext != nil {
		defer server.ConnNext.Close()
	}
	server.ClientNext = nil
}

func (server *MessageBoardServer) GetVersion() int64 {
	currentVersion := server.Version.Load()
	server.Version.Add(1)
	return currentVersion
}

func (server *MessageBoardServer) publishEvent(ev *protobufRazpravljalnica.MessageEvent) {
	server.subMu.Lock()
	subs := len(server.subscribers)
	server.subMu.Unlock()

	// debug: show there are subscribers
	if subs == 0 {
		// no subscribers; nothing to do
		return
	}

	server.subMu.Lock()
	defer server.subMu.Unlock()
	for id, eventChan := range server.subscribers {
		select {
		case eventChan <- ev:
			// delivered
			fmt.Println("publishEvent: delivered to subscriber", id, "seq", ev.SequenceNumber)
		default:
			// subscriber is unreachable (likely a better way to do this but i can't find it)
			fmt.Println("publishEvent: drop for subscriber", id)
			_ = id
		}
	}
}

func (server *MessageBoardServer) ReadMessage(ctx context.Context, in *protobufRazpravljalnica.ReadMessageRequest) (*protobufRazpravljalnica.MessageData, error) {
	if msgData, ok := server.MessageStorage.GetValByKey(in.Id); !ok {
		return nil, DataNotAvailableErr
	} else {
		resp := &protobufRazpravljalnica.MessageData{Msg: msgData.Message, Dirty: msgData.IsDirty()}
		return resp, nil
	}
}

func (server *MessageBoardServer) ReadUser(ctx context.Context, in *protobufRazpravljalnica.ReadUserRequest) (*protobufRazpravljalnica.UserData, error) {
	fmt.Println("id: ", in.Id)
	if userData, ok := server.UserStorage.GetValByKey(in.Id); !ok {
		return nil, DataNotAvailableErr
	} else {
		resp := &protobufRazpravljalnica.UserData{User: userData.User, Dirty: userData.IsDirty()}
		return resp, nil
	}
}

func (server *MessageBoardServer) ReadTopic(ctx context.Context, in *protobufRazpravljalnica.ReadTopicRequest) (*protobufRazpravljalnica.TopicData, error) {
	fmt.Println("id: ", in.Id)
	if topicData, ok := server.TopicStorage.GetValByKey(in.Id); !ok {
		return nil, DataNotAvailableErr
	} else {
		resp := &protobufRazpravljalnica.TopicData{Topic: topicData.Topic, Dirty: topicData.IsDirty()}
		return resp, nil
	}
}

//	func (server *MessageBoardServer) ListTopicsFromInner(ctx context.Context, empty *emptypb.Empty) (*protobufRazpravljalnica.ListTopicsResponse, error) {
//		topicsData := server.TopicStorage.GetAllValues()
//		topicsResponse := []*protobufRazpravljalnica.Topic{}
//		for _, td := range topicsData {
//			if !td.Dirty {
//				fmt.Println("current version already clean")
//				topicsResponse = append(topicsResponse, td.Topic)
//			} else if server.ClientTail != nil {
//				// ctx1, cancel := context.withtimeout(context.background(), time.second)
//				fmt.Println("data is dirty reading from tail ", td.Topic.Name)
//				// defer cancel()
//				// if topic, err := server.ClientTail.ReadTopic(ctx1, &protobufRazpravljalnica.ReadTopicRequest{Id: td.Topic.Id}); err == nil {
//				// 	fmt.Println("clean version succ retreived from the tail")
//				// 	topicsResponse = append(topicsResponse, topic)
//				// }
//
//			} else {
//				panic("this is not supposed to be ever called")
//			}
//		}
//		listTopicResponse := &protobufRazpravljalnica.ListTopicsResponse{Topics: topicsResponse}
//		return listTopicResponse, nil
//	}
//
//	func (server *MessageBoardServer) ListMessagessFromInner(ctx context.Context, empty *emptypb.Empty) (*protobufRazpravljalnica.ListMessagesResponse, error) {
//		messagesData := server.TopicStorage.GetAllValues()
//		messagesResponse := []*protobufRazpravljalnica.Message{}
//		for _, msgData := range messagesData {
//			if !msgData.Dirty {
//				fmt.Println("current version already clean")
//				messagesData = append(messagesData, msgData)
//			} else if server.ClientTail != nil {
//				ctx1, cancel := context.WithTimeout(context.Background(), time.Second)
//				defer cancel()
//				if msg, err := server.ClientTail.ReadMessage(ctx1, &protobufRazpravljalnica.ReadMessageRequest{Id: msgData.Id}); err != nil {
//					return nil, err
//				} else {
//					fmt.Println("clean version succ retreived from the tail")
//					messagesResponse = append(messagesResponse, msg)
//				}
//
//			} else {
//				panic("this is not supposed to be ever called")
//			}
//		}
//		listTopicResponse := &protobufRazpravljalnica.ListMessagesResponse{Messages: messagesResponse}
//		return listTopicResponse, nil
//	}
func (server *MessageBoardServer) SignalNewTail(ctx context.Context, in *protobufRazpravljalnica.SyncTailsRequest) (*protobufRazpravljalnica.SyncTailsACK, error) {
	fmt.Println("master node has signaled that a new tail has been added")
	conn, err := grpc.NewClient(in.NewTail.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer conn.Close()
	fmt.Println("data transfer has been opened")
	reconfigMode.Store(true)
	reconfigModeCh = make(chan struct{})
	grpcClient := protobufRazpravljalnica.NewMessageBoardClient(conn)
	ctx2, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	topicsData := server.TopicStorage.GetAllValues()
	messagesData := server.MessageStorage.GetAllValues()
	usersData := server.UserStorage.GetAllValues()
	users := []*protobufRazpravljalnica.User{}
	messages := []*protobufRazpravljalnica.Message{}
	topics := []*protobufRazpravljalnica.Topic{}
	for _, topicData := range topicsData {
		topic := convertDataToTopic(topicData)
		topics = append(topics, topic)
	}
	for _, messageData := range messagesData {
		message := convertDataToMessage(messageData)
		messages = append(messages, message)
	}
	for _, userData := range usersData {
		user := convertDataToUser(userData)
		users = append(users, user)
	}

	if _, err = grpcClient.TransferData(ctx2, &protobufRazpravljalnica.TransferDataRequest{Users: users, Messages: messages, Topics: topics}); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 2)
	if reconfigMode.CompareAndSwap(true, false) {
		close(reconfigModeCh)
		fmt.Println(reconfigMode.Load())
	}
	fmt.Println("data transfer has ended")

	return &protobufRazpravljalnica.SyncTailsACK{Succ: true}, nil

}
