package master_node

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	protobufInt "PS_projekt/github.com/david/PS_projekt/api/grpc/protobufInternal"
	"context"
	"fmt"
	"math/rand"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	nodeNotFound = fmt.Errorf("node not found")
)

type Node struct {
	msgBoardClient pbRaz.MessageBoardClient
	conn           *grpc.ClientConn
	Url            string
	Id             int64
	Next           *Node
	Prev           *Node
}

func NewNode(msgBoardClient pbRaz.MessageBoardClient, conn *grpc.ClientConn, id int64, addr string) *Node {
	return &Node{msgBoardClient, conn, addr, id, nil, nil}
}

func (masterNode *MasterNode) GetRandomNode(ctx context.Context, empty *emptypb.Empty) (*protobufInt.GenerateRandomNodeResponse, error) {
	masterNode.Mu.Lock()
	defer masterNode.Mu.Unlock()
	chainLen := masterNode.ChainLen.Load()
	randNode := rand.Int63n(chainLen)
	var counter int64 = 0
	for node := masterNode.Head; node != nil; node = node.Next {
		if counter == randNode {
			nodeData := &protobufInt.NodeData{Id: node.Id, Address: node.Url}
			generateRandomNodeResponse := &protobufInt.GenerateRandomNodeResponse{Node: nodeData}
			return generateRandomNodeResponse, nil
		}
		counter++
	}
	return nil, nodeNotFound
}
