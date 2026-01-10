package messageboardserver

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"context"
	"fmt"
	// "fmt"
	// "io"
)

const (
	MAX_DATA_IN_PACKET = 5
)

// old tail sends the data to the new tail
// the master node blocks till that happens

func (server *MessageBoardServer) TransferData(ctx context.Context, in *pbRaz.TransferDataRequest) (*pbRaz.ACK, error) {
	fmt.Println("transfering data")
	server.MessageStorage.Lockable.Mu.Lock()
	for _, message := range in.Messages {
		messageData := &MessageData{Message: message, Dirty: false}
		server.MessageStorage.Lockable.Data[messageData.Id] = messageData
	}
	server.MessageStorage.Lockable.Mu.Unlock()
	server.UserStorage.Lockable.Mu.Lock()
	for _, user := range in.Users {
		userData := &UserData{User: user, Dirty: false}
		server.UserStorage.Lockable.Data[userData.Id] = userData
	}
	server.UserStorage.Lockable.Mu.Unlock()
	server.TopicStorage.Lockable.Mu.Lock()
	for _, topic := range in.Topics {
		topicData := &TopicData{Topic: topic, Dirty: false}
		server.TopicStorage.Lockable.Data[topicData.Id] = topicData
	}
	server.TopicStorage.Lockable.Mu.Unlock()
	fmt.Println("ended data tarnsfer")
	return &pbRaz.ACK{}, nil

}
