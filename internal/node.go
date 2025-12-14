package internal

import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	"context"
	"errors"
	"fmt"
	"time"
)

type ChainNode struct {
	protobufInt.UnimplementedChainNodeServer
	Id   int64
	addr string
}

func NewChainNode(id int64, addr string) *ChainNode {
	return &ChainNode{protobufInt.UnimplementedChainNodeServer{}, id, addr}
}

func (node *ChainNode) HeartBeat(ctx context.Context, in *protobufInt.HearthBeatRequest) (*protobufInt.HearthBeatResponse, error) {
	return &protobufInt.HearthBeatResponse{Id: node.Id}, nil
}

type MasterNode struct {
	Id          int64
	addr        string
	grpcClients []protobufInt.ChainNodeClient
}

func (masterNode *MasterNode) SendHeatBeats() {
	for i, grpClient := range masterNode.grpcClients {
		go masterNode.sendHeartBeat(grpClient, i)
	}
}

func (masterNode *MasterNode) sendHeartBeat(grpcClient protobufInt.ChainNodeClient, nodeNum int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	hearbeatReq := protobufInt.HearthBeatRequest{Id: 0}
	if _, err := grpcClient.HearthBeat(ctx, &hearbeatReq); errors.Is(err, context.DeadlineExceeded) {
		fmt.Printf("node %d: is dead", nodeNum)
	}
}
