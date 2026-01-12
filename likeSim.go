package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func LikeSim() {
	masterUrl := "localhost:9000"
	go StartMasterServer(masterUrl, 0)
	time.Sleep(2 * time.Second)
	msgBoardUrl1 := fmt.Sprintf("localhost:%d", 9001)
	go AddMsgBoardServer(msgBoardUrl1, masterUrl, int64(1))
	time.Sleep(1000 * time.Millisecond)
	workersWg := sync.WaitGroup{}
	errChw := make(chan error, 10)
	workersWg.Add(3)
	for range 3 {
		go simulateWorkers(errChw, &workersWg)
	}
	workersWg.Wait()

	conn, err := grpc.NewClient(masterUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		conn.Close()
	}()
	masterGrpcClient := pbRaz.NewMasterNodeClient(conn)
	head, _, err := fetchDetails(masterGrpcClient)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("client has started")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	headConn, err := grpc.NewClient(head.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		return
	}
	headClient := pbRaz.NewMessageBoardClient(headConn)
	topic, err := headClient.CreateTopic(ctx, &pbRaz.CreateTopicRequest{Name: "topic1"})
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	user, err := headClient.CreateUser(ctx, &pbRaz.CreateUserRequest{Name: "user"})
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*2)
	message, err := headClient.PostMessage(ctx, &pbRaz.PostMessageRequest{TopicId: topic.Id, UserId: user.Id, Text: "text"})
	if err != nil {
		cancel()
		fmt.Println(err)
		return
	}
	cancel()
	for range 10 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*2)
		if message, err := headClient.LikeMessage(ctx, &pbRaz.LikeMessageRequest{TopicId: topic.Id, MessageId: message.Id, UserId: user.Id}); err != nil {
			fmt.Println(err)
			cancel()
			return
		} else {
			cancel()
			fmt.Println(message)
		}

	}

}

func likeFunc(grpcClient pbRaz.MessageBoardClient, in *pbRaz.LikeMessageRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := grpcClient.LikeMessage(ctx, in); err != nil {
		return err
	}
	return nil
}
