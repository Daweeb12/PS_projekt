package master_node

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"

	"google.golang.org/grpc"
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

// func (masterNode *MasterNode) GetClusterState(ctx context.Context, empty *emptypb.Empty) (*protobufInt.GetClusterStateResponse, error) {
// 	if masterNode == nil {
// 		return nil, fmt.Errorf("the masterNode is nil")
// 	}
// 	clusterStateResponse := protobufInt.GetClusterStateResponse{}
// 	if masterNode.Head != nil {
// 		headInfo := protobufInt.NodeData{Address: masterNode.Head.Url, Id: masterNode.Head.Id}
// 		clusterStateResponse.Head = &headInfo
// 	}
// 	if masterNode.Tail != nil {
// 		tailInfo := protobufInt.NodeData{Address: masterNode.Tail.Url, Id: masterNode.Tail.Id}
// 		clusterStateResponse.Tail = &tailInfo
// 	}

// 	return &clusterStateResponse, nil
// }
