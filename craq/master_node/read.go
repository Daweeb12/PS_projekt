package master_node

import (
// protobufInt "PS_projekt/api/grpc/protobufInternal"
// "context"
// "fmt"
// "math/rand"

// "google.golang.org/protobuf/types/known/emptypb"
)

func temp() {}

// read from a random node in the chain
// if object is clean -> read
// else read from tail
// func (masterNode *MasterNode) generateRandomNode() (string, int64, error) {
// 	if masterNode.ChainLen.Load() == 0 {
// 		return "", 0, nil
// 	}
// 	randNode := rand.Int63n(int64(masterNode.ChainLen.Load()))
// 	var i int64 = 0
// 	for node := masterNode.Head; node != nil; node = node.Next {
// 		if i == randNode {
// 			return node.Url, node.Id, nil
// 		}
// 		i++
// 	}
// 	return "", 0, fmt.Errorf("no node found")
// }

// // generate random node for reads
// // due to using CRAQ as the main algoroithm read could be made from any node
// func (masterNode *MasterNode) GenerateRandomNodeForRead(ctx context.Context, empty *emptypb.Empty) (*protobufInt.GenerateRandomNodeResponse, error) {
// 	if nodeUrl, nodeId, err := masterNode.generateRandomNode(); err != nil {
// 		return nil, err
// 	} else {
// 		nodeInfo := protobufInt.NodeData{Id: nodeId, Address: nodeUrl}
// 		return &protobufInt.GenerateRandomNodeResponse{Node: &nodeInfo}, nil
// 	}
// }
