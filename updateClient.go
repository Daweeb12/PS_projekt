package main

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"context"
	"fmt"

	//"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func UpdateClientTest(url string) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	grpcClient := pbRaz.NewMasterNodeClient(conn)
	topicId, userId := func() (int64, int64) {
		headInfo, tailInfo, err := fetchDetails(grpcClient)
		if err != nil {
			fmt.Println(err)
			return -1, -1
		}
		fmt.Println("HEAD: ", headInfo)
		fmt.Println("TAIL: ", tailInfo)
		fmt.Println()
		headConn, err := grpc.NewClient(headInfo.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println(err)
			return -1, -1
		}
		defer headConn.Close()
		headClient := pbRaz.NewMessageBoardClient(headConn)
		user, err := sendCreateUserReq(headClient)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("CREATE USER ", user)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if topic, err := headClient.CreateTopic(ctx, &pbRaz.CreateTopicRequest{Name: "topic"}); err != nil {
			fmt.Println(err)
			return -1, -1
		} else {
			fmt.Println("CREATE TOPIC: ", topic.Id)
			return topic.Id, user.Id
		}
	}()
	time.Sleep(10 * time.Second)
	errCh := make(chan error)
	for {
		func(chan error, int64, int64) {
			headInfo, tailInfo, err := fetchDetails(grpcClient)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("HEAD: ", headInfo)
			fmt.Println("TAIL: ", tailInfo)
			fmt.Println()
			headConn, err := grpc.NewClient(headInfo.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				fmt.Println(err)
				return
			}
			defer headConn.Close()
			headClient := pbRaz.NewMessageBoardClient(headConn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if _, err := headClient.PostMessage(ctx, &pbRaz.PostMessageRequest{TopicId: topicId, UserId: userId, Text: "this is a test"}); err != nil {
				fmt.Println(err)
				errCh <- err
			}
		}(errCh, topicId, userId)
		time.Sleep(time.Second * 2)
	}
}

func fetchDetailsTest(grpcClient pbRaz.MasterNodeClient) (*pbRaz.NodeData, *pbRaz.NodeData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	clusterInfo, err := grpcClient.GetClusterState(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, nil, err
	}
	headInfo, tailInfo := clusterInfo.Head, clusterInfo.Tail
	return headInfo, tailInfo, nil
}

func sendCreateUserReqTest(grpcClient pbRaz.MessageBoardClient) (*pbRaz.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	createUserReq := &pbRaz.CreateTopicRequest{Name: "david"}
	user, err := grpcClient.CreateUser(ctx, (*pbRaz.CreateUserRequest)(createUserReq))
	return user, err
}
