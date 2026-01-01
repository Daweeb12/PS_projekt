package main

/*
import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"context"
	"fmt"

	//"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func UpdateClient(url string) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	grpcClient := protobufInt.NewMasterNodeClient(conn)

	for {
		headInfo, tailInfo, err := fetchDetails(grpcClient)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("HEAD: ", headInfo)
		fmt.Println("TAIL: ", tailInfo)
		fmt.Println()
		headConn, err := grpc.NewClient(headInfo.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer headConn.Close()
		headClient := pbRaz.NewMessageBoardClient(headConn)
		if user, err := sendCreateUserReq(headClient); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("CREATE USER ", user)
		}

		time.Sleep(time.Second * 2)
	}
}

func fetchDetails(grpcClient protobufInt.MasterNodeClient) (*protobufInt.NodeData, *protobufInt.NodeData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	clusterInfo, err := grpcClient.GetClusterState(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, nil, err
	}
	headInfo, tailInfo := clusterInfo.Head, clusterInfo.Tail
	return headInfo, tailInfo, nil
}

func sendCreateUserReq(grpcClient pbRaz.MessageBoardClient) (*pbRaz.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	createUserReq := &pbRaz.CreateTopicRequest{Name: "david"}
	user, err := grpcClient.CreateUser(ctx, (*pbRaz.CreateUserRequest)(createUserReq))
	return user, err
}
*/
