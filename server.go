package main

import (
	protobufRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
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
