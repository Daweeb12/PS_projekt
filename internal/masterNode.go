package internal

import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	//"errors"
	//"fmt"
	//"time"
)

type MasterNode struct {
	protobufInt.UnimplementedMasterNodeServer
	Id          int64
	Url         string
	Chain       *Node
	Tail        *Node
	Head        *Node
	clientConns []protobufInt.ChainNodeClient
}

type Node struct {
	protobufInt.ChainNodeServer
	clientConn protobufInt.ChainNodeClient
	Url        string
	Id         int64
	Next       *Node
	Prev       *Node
}

func NewMasterNode(id int64, addr string) *MasterNode {
	return &MasterNode{protobufInt.UnimplementedMasterNodeServer{}, id, addr, nil, nil, nil, []protobufInt.ChainNodeClient{}}
}

func (masterNode *MasterNode) SignalAlive(ctx context.Context, in *protobufInt.SignalAliveRequest) (*protobufInt.SignalAliveResponse, error) {
	id, nodeUrl := in.NodeId, in.NodeUrl
	signalAliveResponse := &protobufInt.SignalAliveResponse{MasterUrl: masterNode.Url}
	masterNode.openNewClient(nodeUrl, id)
	return signalAliveResponse, nil
}

func (masterNode *MasterNode) openNewClient(nodeUrl string, id int64) error {
	conn, err := grpc.NewClient(nodeUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	grpcClient := protobufInt.NewChainNodeClient(conn)
	masterNode.clientConns = append(masterNode.clientConns, grpcClient)
	node := &Node{clientConn: grpcClient, Next: nil, Prev: nil, Url: nodeUrl, Id: id}
	fmt.Println(node.clientConn)
	masterNode.AddNode(node)

	return nil
}

func (masterNode *MasterNode) CheckHealth() error {
	i := 0
	for node := masterNode.Head; node != nil; node = node.Next {
		grpcClient := node.clientConn
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := grpcClient.HeartBeat(ctx, &protobufInt.HearthBeatRequest{Id: 0})
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("node ", i, " has stopped responding")
		} else if status.Code(err) == codes.Unavailable {
			masterNode.RemoveNode(node)
		} else if err != nil {
			return err
		}
		i++
	}
	return nil
}

func (masterNode *MasterNode) RemoveNode(target *Node) {

	for node := masterNode.Head; node != nil; node = node.Next {
		if target == node && node.Prev == nil {
			if node.Next == nil {
				masterNode.Head = nil
				masterNode.Tail = nil
				masterNode.Chain = nil
				return
			}
			node.Next.Prev = nil
			masterNode.Head = node.Next
			masterNode.Chain = node.Next
			return
		} else if target == node && node.Next == nil {
			node.Prev.Next = nil
			masterNode.Tail = node.Prev
		} else if target == node {
			node.Next.Prev = node.Prev.Next
			node.Prev.Next = node.Next
			return
		}
	}

}

func (masterNode *MasterNode) AddNode(node *Node) error {
	if masterNode.Tail == nil {
		fmt.Println("added first node")
		masterNode.Head = node
		masterNode.Tail = node
		masterNode.Tail = node
		fmt.Println(node)
		return nil
	}
	fmt.Println("added to tail")
	fmt.Println(node)
	oldTail := masterNode.Tail
	oldTail.Next = node
	node.Prev = oldTail
	masterNode.Tail = node

	assignmentPrevTailReq := protobufInt.AssignRequest{Id: node.Id, NextUrl: node.Url, PrevUrl: ""}
	assignmentNewTailReq := protobufInt.AssignRequest{Id: masterNode.Tail.Id, NextUrl: "", PrevUrl: oldTail.Url}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := oldTail.clientConn.AssignChainNode(ctx, &assignmentPrevTailReq); err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := node.clientConn.AssignChainNode(ctx, &assignmentNewTailReq); err != nil {
		return err
	}
	return nil
}

func (masterNode *MasterNode) Update(ctx context.Context, in *protobufInt.UpdateRequest) (*protobufInt.ACK, error) {

	if masterNode.Head == nil {
		return nil, fmt.Errorf("no known chain nodes")
	}

	forwarReq := &protobufInt.ForwardRequest{}

	ack, err := masterNode.Head.clientConn.Forward(ctx, forwarReq)

	return ack, err
}
