package main

import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	protobufRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"PS_projekt/internal"
	messageboardserver "PS_projekt/messageBoardServer"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func Server(url string) {
	lis, err := net.Listen("tcp", url)
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	msgBoardServer := messageboardserver.NewMessageBoardServer()
	protobufRaz.RegisterMessageBoardServer(grpcServer, msgBoardServer.MessageBoardServer)
	if hostname, err := os.Hostname(); err != nil {
		panic(err)
	} else {
		fmt.Println("grpc listening at ", hostname, url)
	}

	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}

}

func AddNode(nodeUrl, masterUrl string, id int64) {
	ch := make(chan struct{})
	grpcServer := grpc.NewServer()
	internalNodeServer := internal.NewChainNode(nil, id, nodeUrl)
	protobufInt.RegisterChainNodeServer(grpcServer, internalNodeServer)
	ls, err := net.Listen("tcp", nodeUrl)
	if err != nil {
		panic(err)
	}
	go grpcServer.Serve(ls)
	time.Sleep(100 * time.Millisecond)
	if err := notifyMaster(id, masterUrl, nodeUrl); err != nil {
		panic(err)
	}

	fmt.Println("chain node started listening")
	<-ch
}

func StartMasterServer(url string, id int64) {

	grpcServer := grpc.NewServer()
	masterNode := internal.NewMasterNode(id, url)
	protobufInt.RegisterMasterNodeServer(grpcServer, masterNode)
	ls, err := net.Listen("tcp", url)
	if err != nil {
		panic(err)
	}
	fmt.Println("master server listening on ", url)
	go func(masterNode *internal.MasterNode) {
		for {
			fmt.Println("head", masterNode.Head)
			fmt.Println("tail ", masterNode.Tail)

			fmt.Println()
			if err := masterNode.CheckHealth(); status.Code(err) == codes.Unavailable {
				fmt.Println("should remove node")
			} else {
				fmt.Println(err)
			}
			time.Sleep(time.Second)

		}
	}(masterNode)
	if err := grpcServer.Serve(ls); err != nil {
		panic(err)
	}
}

func notifyMaster(nodeId int64, masterAddr, nodeAddr string) error {
	conn, err := grpc.NewClient(masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	fmt.Println("tried to notify server")
	grpcClient := protobufInt.NewMasterNodeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := grpcClient.SignalAlive(ctx, &protobufInt.SignalAliveRequest{NodeId: nodeId, NodeUrl: nodeAddr}); err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}
