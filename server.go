package main

import (
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"os"
)

func Server(url string) {
	grpcServer := grpc.NewServer()

}
