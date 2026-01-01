package master_node

import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"sync"

	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	ChainLen    atomic.Int64
	Mu          sync.Mutex
}

var (
	reconfigCh chan struct{}
)

func NewMasterNode(id int64, addr string) *MasterNode {
	return &MasterNode{protobufInt.UnimplementedMasterNodeServer{}, id, addr, nil, nil, nil, []protobufInt.ChainNodeClient{}, atomic.Int64{}, sync.Mutex{}}
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
	grpcClient := pbRaz.NewMessageBoardClient(conn)
	node := NewNode(grpcClient, conn, id, nodeUrl)
	fmt.Println(node.msgBoardClient)
	masterNode.AddNode(node)

	return nil
}

func (masterNode *MasterNode) CheckHealth() error {
	i := 0
	for node := masterNode.Head; node != nil; node = node.Next {
		grpcClient := node.msgBoardClient
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := grpcClient.HeartBeat(ctx, &pbRaz.HearthBeatRequest{})
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
	masterNode.ChainLen.Add(-1)
	masterNode.Mu.Lock()
	defer masterNode.Mu.Unlock()
	if target != nil && target.conn != nil {
		defer target.conn.Close()
	}

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
	masterNode.ChainLen.Add(1)
	masterNode.Mu.Lock()
	defer masterNode.Mu.Unlock()
	if masterNode.Tail == nil {
		fmt.Println("added first node")
		masterNode.Head = node
		masterNode.Tail = node
		masterNode.Tail = node
		fmt.Println(node)
		return nil
	}
	//errCh := make(chan error)
	//	go masterNode.reconfigTail(errCh)
	//	if err := <-errCh; err != nil {
	//		return err
	//	}
	fmt.Println("added to tail")
	fmt.Println(node)
	oldTail := masterNode.Tail
	oldTail.Next = node
	node.Prev = oldTail
	masterNode.Tail = node

	assignmentPrevTailReq := pbRaz.AssignRequest{Id: node.Id, NextUrl: node.Url, PrevUrl: "", TailUrl: node.Url}
	assignmentNewTailReq := pbRaz.AssignRequest{Id: masterNode.Tail.Id, NextUrl: "", PrevUrl: oldTail.Url, TailUrl: ""}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := oldTail.msgBoardClient.AssignChainNode(ctx, &assignmentPrevTailReq); err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := node.msgBoardClient.AssignChainNode(ctx, &assignmentNewTailReq); err != nil {
		return err
	}
	return nil
}

func (masterNode *MasterNode) reconfigTail(errorCh chan error) {
	defer close(reconfigCh)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ack, err := masterNode.Tail.msgBoardClient.SignalNewTail(ctx, &pbRaz.SyncTailsRequest{})
	if err != nil {
		errorCh <- err
		return
	}
	if !ack.Succ {
		errorCh <- fmt.Errorf("did not receive succ ack ")
	} else {
		errorCh <- nil
	}

}

func (masterNode *MasterNode) GetClusterState(ctx context.Context, empty *emptypb.Empty) (*protobufInt.GetClusterStateResponse, error) {
	if masterNode == nil {
		return nil, fmt.Errorf("master node is nil")
	}
	clusterStateResp := &protobufInt.GetClusterStateResponse{}
	fmt.Println("masterNode trying to send the head and the tail")
	fmt.Println("head", masterNode.Head)
	fmt.Println("tail", masterNode.Tail)
	if masterNode.Head != nil {
		headInfo := &protobufInt.NodeData{Id: masterNode.Head.Id, Address: masterNode.Head.Url}
		clusterStateResp.Head = headInfo
	}

	if masterNode.Tail != nil {
		tailInfo := &protobufInt.NodeData{Id: masterNode.Tail.Id, Address: masterNode.Tail.Url}
		clusterStateResp.Tail = tailInfo
	}
	return clusterStateResp, nil
}
