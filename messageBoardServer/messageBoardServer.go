package messageboardserver

import (
	protobufRazpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"
	"PS_projekt/storage"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	randMu           sync.Mutex
	topicErr         = fmt.Errorf("topic does not exist")
	userExistsErr    = fmt.Errorf("user does not exist")
	messageExistsErr = fmt.Errorf("message does not exist")
	emptyTopicErr    = fmt.Errorf("topic can not be empty")
	emptyUsernameErr = fmt.Errorf("username can not be empty")
)

// generate uint32 uids
func GenerateRand32[T comparable](s *storage.LockableMap[int64, T]) int64 {
	randMu.Lock()
	id := rand.Uint32()
	defer randMu.Unlock()
	_, ok := s.GetValByKey(int64(id))
	for ok {
		_, ok = s.GetValByKey(int64(id))
	}
	return int64(id)
}

type MessageBoardServer struct {
	protobufRazpravljalnica.MessageBoardServer
	Id           int64
	UserStorage  *storage.LockableMap[int64, *protobufRazpravljalnica.User]
	TopicStorage *storage.LockableMap[int64, *protobufRazpravljalnica.Topic]
	//map of [3]{messageId , userId , topicId}
	MessageStorage *storage.LockableMap[int64, *protobufRazpravljalnica.Message]
	//map to track used message ids
	ConnNext   *grpc.ClientConn
	ConnPrev   *grpc.ClientConn
	ConnTail   *grpc.ClientConn
	ClientCurr protobufRazpravljalnica.MessageBoardClient
	ClientNext protobufRazpravljalnica.MessageBoardClient
	ClientPrev protobufRazpravljalnica.MessageBoardClient
	ClientTail protobufRazpravljalnica.MessageBoardClient
}

// initialize the storage
func NewMessageBoardServer(id int64) *MessageBoardServer {
	userStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.User]()
	topicStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.Topic]()
	messageStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.Message]()
	return &MessageBoardServer{protobufRazpravljalnica.UnimplementedMessageBoardServer{}, id, userStorage, topicStorage, messageStorage, nil, nil, nil, nil, nil, nil, nil}
}

// generates random user id and adds to map
// returns error if the user name consists of spaces
func (server *MessageBoardServer) CreateUser(ctx context.Context, in *protobufRazpravljalnica.CreateUserRequest) (*protobufRazpravljalnica.User, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, emptyTopicErr
	}
	if server == nil {
		return nil, fmt.Errorf("server is nil")
	}
	id := GenerateRand32(server.UserStorage)
	user := &protobufRazpravljalnica.User{Id: id, Name: in.Name}
	server.UserStorage.Put(int64(id), user)
	if server.ClientNext == nil {
		fmt.Println("the data has arrived at the tail")
		return user, nil
	}
	if fail() {
		fmt.Println("sending data to the next node has failed")
		return nil, fmt.Errorf("sending data has failed")
	}
	fmt.Println("the data has been forwarded")
	fmt.Println("next server is ", server.ClientNext)
	if user, err := server.ClientNext.CreateUser(ctx, in); status.Code(err) == codes.Unavailable {
		if server.ConnNext != nil {
			defer server.ConnNext.Close()
		}
		server.ClientNext = nil
		return nil, fmt.Errorf("the next client has stoped responding")
	} else if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		return user, nil
	}
}

func (server *MessageBoardServer) CreateTopic(ctx context.Context, in *protobufRazpravljalnica.CreateTopicRequest) (*protobufRazpravljalnica.Topic, error) {
	if name := strings.TrimSpace(in.Name); name == "" {
		return nil, emptyTopicErr
	} else {
		id := GenerateRand32(server.TopicStorage)
		topic := &protobufRazpravljalnica.Topic{Name: name, Id: id}
		server.TopicStorage.Put(id, topic)
		return topic, nil
	}
}

func (server *MessageBoardServer) PostMessage(ctx context.Context, in *protobufRazpravljalnica.PostMessageRequest) (*protobufRazpravljalnica.Message, error) {
	topicId, userId := in.TopicId, in.UserId
	if _, ok := server.UserStorage.GetValByKey(userId); !ok {
		return nil, fmt.Errorf("userId does not exist")
	}
	if _, ok := server.TopicStorage.GetValByKey(topicId); !ok {
		return nil, fmt.Errorf("topicId does not exist")
	}
	randMu.Lock()
	messageId := GenerateRand32(server.MessageStorage)
	randMu.Unlock()
	message := &protobufRazpravljalnica.Message{Id: messageId, TopicId: topicId, UserId: userId, Text: in.Text}
	server.MessageStorage.Put(messageId, message)
	return &protobufRazpravljalnica.Message{Id: messageId, TopicId: topicId, UserId: userId, Text: in.Text}, nil
}

func (server *MessageBoardServer) UpdateMessage(ctx context.Context, in *protobufRazpravljalnica.UpdateMessageRequest) (*protobufRazpravljalnica.Message, error) {
	if _, ok := server.UserStorage.GetValByKey(in.UserId); !ok {
		return nil, fmt.Errorf("user does not exist")
	}
	if _, ok := server.TopicStorage.GetValByKey(in.TopicId); !ok {
		return nil, topicErr
	}
	if msg, ok := server.MessageStorage.GetValByKey(in.MessageId); !ok {
		return nil, fmt.Errorf("message does not exsist")
	} else {
		msg.Text = in.Text
		server.MessageStorage.Put(in.MessageId, msg)
		return nil, nil
	}
}

func (server *MessageBoardServer) DeleteMessage(ctx context.Context, in *protobufRazpravljalnica.DeleteMessageRequest) (*emptypb.Empty, error) {
	server.MessageStorage.Delete(in.MessageId)
	return &emptypb.Empty{}, nil
}

func (server *MessageBoardServer) ListTopics(ctx context.Context, empty *emptypb.Empty) (*protobufRazpravljalnica.ListTopicsResponse, error) {
	topics := server.TopicStorage.GetAllValues()
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

	return message, nil
}

func (server *MessageBoardServer) GetMessages(ctx context.Context, in *protobufRazpravljalnica.GetMessagesRequest) (*protobufRazpravljalnica.GetMessagesResponse, error) {
	messages := server.MessageStorage.GetAllValues()
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
