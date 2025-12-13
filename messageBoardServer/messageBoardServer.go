package messageboardserver

import (
	protobufRazpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"
	"PS_projekt/storage"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	randMu           sync.Mutex
	topicErr         = fmt.Errorf("topic does not exist")
	userExistsErr    = fmt.Errorf("user does not exist")
	messageExistsErr = fmt.Errorf("message does not exist")
)

// generate uint32 uids
func GenerateRand32[T any](s *storage.LockableMap[int64, T]) int64 {
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
	UserStorage  *storage.LockableMap[int64, *protobufRazpravljalnica.User]
	TopicStorage *storage.LockableMap[int64, *protobufRazpravljalnica.Topic]
	//map of [3]{messageId , userId , topicId}
	MessageStorage *storage.LockableMap[int64, *protobufRazpravljalnica.Message]
	//map to track used message ids
}

// initialize the storage
func NewMessageBoardServer() *MessageBoardServer {
	userStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.User]()
	topicStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.Topic]()
	messageStorage := storage.NewLockableMap[int64, *protobufRazpravljalnica.Message]()
	return &MessageBoardServer{protobufRazpravljalnica.UnimplementedMessageBoardServer{}, userStorage, topicStorage, messageStorage}
}

// generates random user id and adds to map
// returns error if the user name consists of spaces
func (server *MessageBoardServer) CreateUser(ctx context.Context, in *protobufRazpravljalnica.CreateUserRequest) (*protobufRazpravljalnica.User, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("username can not consist of empty spaces")
	}
	id := GenerateRand32(server.UserStorage)
	user := &protobufRazpravljalnica.User{Id: id, Name: in.Name}
	server.UserStorage.Put(int64(id), user)
	return user, nil
}

func (server *MessageBoardServer) CreateTopic(ctx context.Context, in *protobufRazpravljalnica.CreateTopicRequest) (*protobufRazpravljalnica.Topic, error) {
	if name := strings.TrimSpace(in.Name); name == "" {
		return nil, fmt.Errorf("username can not consist of empty spaces")
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

func (server *MessageBoardServer) DeleteMessage(ctx context.Context, in *protobufRazpravljalnica.DeleteMessageRequest) (emptypb.Empty, error) {
	server.MessageStorage.Delete(in.MessageId)
	return emptypb.Empty{}, nil
}

func (server *MessageBoardServer) ListTopics(ctx context.Context, empty *emptypb.Empty) *protobufRazpravljalnica.ListTopicsResponse {
	topics := server.TopicStorage.GetAllValues()
	listTopicResponse := &protobufRazpravljalnica.ListTopicsResponse{Topics: topics}
	return listTopicResponse
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

func (server *MessageBoardServer) GetMessages(ctx context.Context, in *protobufRazpravljalnica.GetMessagesRequest) *protobufRazpravljalnica.GetMessagesResponse {
	messages := server.MessageStorage.GetAllValues()
	getMessagesResponse := protobufRazpravljalnica.GetMessagesResponse{Messages: messages}
	return &getMessagesResponse
}
