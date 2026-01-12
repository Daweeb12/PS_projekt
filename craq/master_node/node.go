package master_node

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
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

func (masterNode *MasterNode) GenerateRandomNode(ctx context.Context, empty *emptypb.Empty) (*pbRaz.GenerateRandomNodeResponse, error) {
	masterNode.Mu.Lock()
	defer masterNode.Mu.Unlock()
	chainLen := masterNode.ChainLen.Load()
	randNode := rand.Int63n(chainLen)
	var counter int64 = 0
	if masterNode.Head == nil {
		return nil, nodeNotFound
	}
	for node := masterNode.Head; node != nil; node = node.Next {
		if counter == randNode {
			nodeData := &pbRaz.NodeData{Id: node.Id, Address: node.Url}
			generateRandomNodeResponse := &pbRaz.GenerateRandomNodeResponse{Node: nodeData}
			return generateRandomNodeResponse, nil
		}
		counter++
	}

	return nil, nodeNotFound
}
