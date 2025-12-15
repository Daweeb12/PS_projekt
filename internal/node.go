package internal

import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"errors"
	//"fmt"
	//"time"
)

type ChainNode struct {
	protobufInt.ChainNodeServer
	Id         int64
	addr       string
	clientNext protobufInt.ChainNodeClient
	clientPrev protobufInt.ChainNodeClient
}

func NewChainNode(conn *grpc.ClientConn, id int64, addr string) *ChainNode {
	return &ChainNode{protobufInt.UnimplementedChainNodeServer{}, id, addr, nil, nil}
}

func (node *ChainNode) HeartBeat(ctx context.Context, in *protobufInt.HearthBeatRequest) (*protobufInt.HearthBeatResponse, error) {
	return &protobufInt.HearthBeatResponse{Id: node.Id}, nil
}

func (node *ChainNode) AssignChainNode(ctx context.Context, in *protobufInt.AssignRequest) (*protobufInt.ACK, error) {
	nextUrl, prevUrl := in.NextUrl, in.PrevUrl
	if nextUrl != "" {
		conn, err := grpc.NewClient(nextUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		grpClient := protobufInt.NewChainNodeClient(conn)
		node.clientNext = grpClient

	}

	if prevUrl != "" {
		conn, err := grpc.NewClient(prevUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		grpClient := protobufInt.NewChainNodeClient(conn)
		node.clientPrev = grpClient

	}
	return &protobufInt.ACK{}, nil
}

func (node *ChainNode) Forward(ctx context.Context, in *protobufInt.ForwardRequest) (*protobufInt.ACK, error) {
	if node == nil {
		return nil, fmt.Errorf("node can not be nil")
	}
	if node.clientNext == nil {
		fmt.Printf("the tail has received the information")
		return &protobufInt.ACK{Success: true}, nil
	}
	fmt.Printf("node %d has forwarded the information\n", node.Id)
	ack, err := node.clientNext.Forward(ctx, in)
	if err != nil {
		fmt.Printf("node %d has failed to receive the information", node.Id)
		return nil, err
	}
	fmt.Printf("node %d has received confirmation of the update\n", node.Id)
	return ack, err
}
