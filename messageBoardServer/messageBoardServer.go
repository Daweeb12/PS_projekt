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

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	randMu           sync.Mutex
	topicErr         = fmt.Errorf("topic does not exist")
	userExistsErr    = fmt.Errorf("user does not exist")
	messageExistsErr = fmt.Errorf("message does not exist")
	emptyTopicErr    = fmt.Errorf("topic can not be empty")
	emptyUsernameErr = fmt.Errorf("username can not be empty")
	clientStopedErr  = fmt.Errorf("the next client has stoped responding")
	failErr          = fmt.Errorf("sending data has failed")
)

// generate uint32 uids
func GenerateRand32[T comparable](s *storage.LockableMap[int64, T]) int64 {
	for {
		// generate id under randMu
		randMu.Lock()
		id := rand.Uint32()
		randMu.Unlock()
		fmt.Println("GenerateRand32: generated id", id)
		// check uniqueness
		_, ok := s.GetValByKey(int64(id))
		fmt.Println("GenerateRand32: uniqueness check ok=", ok)
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
<<<<<<< HEAD
	userStorage := storage.NewLockableMap[int64, *UserData]()
	topicStorage := storage.NewLockableMap[int64, *TopicData]()
	messageStorage := storage.NewLockableMap[int64, *MessageData]()
	return &MessageBoardServer{protobufRazpravljalnica.UnimplementedMessageBoardServer{}, id, atomic.Int64{}, userStorage, topicStorage, messageStorage, nil, nil, nil, nil, nil, nil, nil}
=======
	userStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.User]()
	topicStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.Topic]()
	messageStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.Message]()
	return &MessageBoardServer{protobufRazpravljalnica.UnimplementedMessageBoardServer{}, id, userStorage, topicStorage, messageStorage, nil, nil, nil, nil, nil, nil, nil, sync.Mutex{}, make(map[int]chan *protobufRazpravljalnica.MessageEvent), 0, 0}
>>>>>>> origin/craq
}

// generates random user id and adds to map
// returns error if the user name consists of spaces
func (server *MessageBoardServer) CreateUser(ctx context.Context, in *protobufRazpravljalnica.CreateUserRequest) (*protobufRazpravljalnica.User, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, emptyTopicErr
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
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, emptyTopicErr
	}

	if server.ClientPrev == nil {
		in.Version = server.GetVersion()
	}

	id := GenerateRand32(server.TopicStorage)
	topic := &protobufRazpravljalnica.Topic{Name: name, Id: id}
	topicData := &TopicData{Topic: topic, Dirty: true}
	server.TopicStorage.Put(id, topicData)
	if server.ClientNext == nil {
		fmt.Println("the data has arrived at the tail")
		topicData.Dirty = false
		server.TopicStorage.Put(id, topicData)
		return topic, nil
	}
	if fail() {
		return nil, failErr
	}

	if user, err := server.ClientNext.CreateTopic(ctx, in); status.Code(err) == codes.Unavailable {
		server.handleUnavailableNode()
		return nil, clientStopedErr
	} else if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		topicData.Dirty = false
		server.TopicStorage.Put(id, topicData)
		return user, nil
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
	if server.ClientPrev == nil {
		in.Version = server.GetVersion()
	}
	messageId := GenerateRand32(server.MessageStorage)
<<<<<<< HEAD
	randMu.Unlock()
	message := &protobufRazpravljalnica.Message{Id: messageId, TopicId: topicId, UserId: userId, Text: in.Text}
	messageData := &MessageData{Message: message, Dirty: true}
	server.MessageStorage.Put(messageId, messageData)
	if server.ClientNext == nil {
		return message, nil
	}
	if fail() {
		return nil, failErr
	}
	if msg, err := server.PostMessage(ctx, in); err == nil {
		messageData.Dirty = false
		server.MessageStorage.Put(messageId, messageData)
		return msg, nil
	} else if status.Code(err) == codes.Unavailable {
		server.handleUnavailableNode()
		return nil, clientStopedErr
	} else {
		return nil, err
	}
=======
	message := &protobufRazpravljalnica.Message{Id: messageId, TopicId: topicId, UserId: userId, Text: in.Text, CreatedAt: timestamppb.Now()}
	server.MessageStorage.Put(messageId, message)

	// publish event
	seq := atomic.AddInt64(&server.seq, 1)
	event := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_POST, Message: message, EventAt: timestamppb.Now()}
	server.publishEvent(event)

	return &protobufRazpravljalnica.Message{Id: messageId, TopicId: topicId, UserId: userId, Text: in.Text, CreatedAt: message.CreatedAt}, nil
>>>>>>> origin/craq
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

	if server.ClientPrev == nil {
		in.Version = server.GetVersion()
	}

	msgData.Text = in.Text
	msgData.Dirty = true
	server.MessageStorage.Put(in.MessageId, msgData)
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
<<<<<<< HEAD
		return nil, err
=======
		msg.Text = in.Text
		server.MessageStorage.Put(in.MessageId, msg)
		seq := atomic.AddInt64(&server.seq, 1)
		event := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_UPDATE, Message: msg, EventAt: timestamppb.Now()}
		server.publishEvent(event)
		return msg, nil
>>>>>>> origin/craq
	}
}

func (server *MessageBoardServer) DeleteMessage(ctx context.Context, in *protobufRazpravljalnica.DeleteMessageRequest) (*emptypb.Empty, error) {
	// prepare event for delete (include message id)
	if msg, ok := server.MessageStorage.GetValByKey(in.MessageId); ok {
		seq := atomic.AddInt64(&server.seq, 1)
		event := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_DELETE, Message: msg, EventAt: timestamppb.Now()}
		server.publishEvent(event)
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

	message.Likes++
	server.MessageStorage.Put(in.MessageId, message)
<<<<<<< HEAD
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
=======
	seq := atomic.AddInt64(&server.seq, 1)
	event := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_LIKE, Message: message, EventAt: timestamppb.Now()}
	server.publishEvent(event)

	return message, nil
>>>>>>> origin/craq
}

// send type MessageEvent to all subs
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

// stream MessageEvent updates to matching requsted topics.
func (server *MessageBoardServer) SubscribeTopic(req *protobufRazpravljalnica.SubscribeTopicRequest, stream protobufRazpravljalnica.MessageBoard_SubscribeTopicServer) error {
	eventChan := make(chan *protobufRazpravljalnica.MessageEvent, 64)
	server.subMu.Lock()
	id := server.nextSubId
	server.nextSubId++
	server.subscribers[id] = eventChan
	server.subMu.Unlock()
	fmt.Println("SubscribeTopic: registered subscriber", id)

	// vile thing i found on stack overflow
	defer func() {
		server.subMu.Lock()
		delete(server.subscribers, id)
		close(eventChan)
		server.subMu.Unlock()
	}()

	// send existing messages from FromMessageId
	if len(req.TopicId) > 0 {
		for _, m := range server.MessageStorage.GetAllValues() {
			for _, topicId := range req.TopicId {
				if m.TopicId == topicId && m.Id > req.FromMessageId {
					seq := atomic.AddInt64(&server.seq, 1)
					ev := &protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_POST, Message: m, EventAt: timestamppb.Now()}
					if err := stream.Send(ev); err != nil {
						return err
					}
					fmt.Println("SubscribeTopic: sent history ev seq", ev.SequenceNumber)
					break
				}
			}
		}
	}

	// stream new events
	fmt.Println("SubscribeTopic: entering event loop for subscriber", id)
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case ev := <-eventChan:
			// filter by topic ids if provided
			if len(req.TopicId) > 0 {
				matched := false
				for _, tid := range req.TopicId {
					if ev.Message != nil && ev.Message.TopicId == tid {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			if err := stream.Send(ev); err != nil {
				return err
			}
		}
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
	if p > 0.6 {
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
