package messageboardserver

import (
// pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
// "context"
// "fmt"
// "io"
// "time"

// "google.golang.org/grpc"
)

const (
	MAX_DATA_IN_PACKET = 5
)

// // old tail sends the data to the new tail
// // the master node blocks till that happens
// func (server *MessageBoardServer) SignalNewTail(ctx context.Context, in *pbRaz.SyncTailsRequest) (*pbRaz.SyncTailsACK, error) {
// 	if server == nil {
// 		return nil, fmt.Errorf("server is nilprotobufInt")
// 	}
// 	fmt.Println("master node has signaled that a new tail has been added")
// 	conn, err := grpc.NewClient(in.NewTail.Address)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer conn.Close()
// 	fmt.Println("data transfer has been opened")
// 	grpcClient := pbRaz.NewMessageBoardClient(conn)
// 	ctx2, cancel := context.WithTimeout(context.Background(), time.Second*10)
// 	defer cancel()
// 	stream, err := grpcClient.TransferData(ctx2)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := server.transferMessages(stream); err != nil {
// 		return nil, err
// 	}

// 	if err := server.transferTopics(stream); err != nil {
// 		return nil, err
// 	}

// 	if err := server.transferUsers(stream); err != nil {
// 		return nil, err
// 	}

// 	fmt.Println("data transfer has ended")

// 	return &pbRaz.SyncTailsACK{Succ: true}, nil

// }

// func (server *MessageBoardServer) TransferMessages(stream pbRaz.MessageBoard_TransferDataClient) error {
// 	messages := server.MessageStorage.GetAllValues()
// 	data := pbRaz.UploadData{}
// 	messageBuffer := []*pbRaz.Message{}
// 	for i, message := range messages {
// 		if len(messageBuffer) == MAX_DATA_IN_PACKET || i+1 == len(messageBuffer) {
// 			data.Messages = messageBuffer
// 			if err := stream.Send(&data); err != nil {
// 				return err
// 			}
// 			data.Messages = []*pbRaz.Message{}
// 		}
// 		messageBuffer = append(messageBuffer, message)
// 	}
// 	return nil
// }

// func (server *MessageBoardServer) transferTopics(stream pbRaz.MessageBoard_TransferDataClient) error {
// 	topics := server.TopicStorage.GetAllValues()
// 	data := pbRaz.UploadData{}
// 	topicBuffer := []*pbRaz.Topic{}
// 	for i, topic := range topics {
// 		if len(topicBuffer) == MAX_DATA_IN_PACKET || i+1 == len(topicBuffer) {
// 			data.Topic = topicBuffer
// 			if err := stream.Send(&data); err != nil {
// 			}
// 		}
// 		topicBuffer = append(topicBuffer, topic)
// 	}
// 	return nil
// }

// func (server *MessageBoardServer) transferUsers(stream pbRaz.MessageBoard_TransferDataClient) error {
// 	users := server.UserStorage.GetAllValues()
// 	data := pbRaz.UploadData{}
// 	userBuffer := []*pbRaz.User{}
// 	for i, user := range users {
// 		if len(userBuffer) == MAX_DATA_IN_PACKET || i+1 == len(userBuffer) {
// 			data.Users = userBuffer
// 			if err := stream.Send(&data); err != nil {
// 			}
// 		}
// 		userBuffer = append(userBuffer, user)
// 	}
// 	return nil
// }

// func (server *MessageBoardServer) TransferData(stream pbRaz.MessageBoard_TransferDataServer) error {
// 	if server == nil {
// 		return fmt.Errorf("the server is nil")
// 	}
// 	fmt.Println("new tail started receiving data")
// 	for {
// 		if data, err := stream.Recv(); err == io.EOF {
// 			fmt.Println("data stream has finished")
// 			return stream.SendAndClose(&pbRaz.UploadACK{Succ: true})
// 		} else if err != nil {
// 			return err
// 		} else {
// 			server.processData(data)
// 		}

// 	}
// }

// func (server *MessageBoardServer) processData(data *pbRaz.UploadData) {
// 	for _, topic := range data.Topic {
// 		server.TopicStorage.SetValForKey(topic.Id, topic)
// 	}
// 	for _, user := range data.Users {
// 		server.UserStorage.SetValForKey(user.Id, user)
// 	}
// 	for _, msg := range data.Messages {
// 		server.MessageStorage.SetValForKey(msg.Id, msg)
// 	}
// }
