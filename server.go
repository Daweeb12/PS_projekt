package main

import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	protobufRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"PS_projekt/internal"
	messageboardserver "PS_projekt/messageBoardServer"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
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

func ChainNodeServer(url string, id int64) error {
	grpcServer := grpc.NewServer()
	internalNodeServer := internal.NewChainNode(id, url)
	protobufInt.RegisterChainNodeServer(grpcServer, internalNodeServer)
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	ls, err := net.Listen("tcp", url)
	if err != nil {
		return err
	}
	if err := grpcServer.Serve(ls); err != nil {
		return err
	}
	fmt.Println(hostname, ": listening on ", url)
	return nil

}
