package chain_node

import (
	// protobufInt "PS_projekt/api/grpc/protobufInternal"
	// pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	// messageboardserver "PS_projekt/messageBoardServer"
	// "PS_projekt/storage"
	// "context"
	// "fmt"
	"math/rand"
	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
	// //"errors"
	// //"fmt"
	// //"time"
)

// type ChainNode struct {
// 	protobufInt.ChainNodeServer
// 	*messageboardserver.MessageBoardServer
// 	Id               int64
// 	addr             string
// 	NodeData         storage.LockableMap[int64, Data]
// 	LastCleanVersion storage.LockableMap[int64, int64]
// 	clientNext       protobufInt.ChainNodeClient
// 	clientPrev       protobufInt.ChainNodeClient
// 	clientTail       protobufInt.ChainNodeClient
// }

// func NewChainNode(conn *grpc.ClientConn, id int64, addr string) *ChainNode {
// 	valueVersionDict := storage.NewLockableMap[int64, Data]()
// 	versionDict := storage.NewLockableMap[int64, int64]()
// 	return &ChainNode{protobufInt.UnimplementedChainNodeServer{}, messageboardserver.NewMessageBoardServer(), id, addr, *valueVersionDict, *versionDict, nil, nil, nil}
// }

// func (node *ChainNode) HeartBeat(ctx context.Context, in *protobufInt.HearthBeatRequest) (*protobufInt.HearthBeatResponse, error) {
// 	return &protobufInt.HearthBeatResponse{Id: node.Id}, nil
// }

// func (node *ChainNode) AssignChainNode(ctx context.Context, in *protobufInt.AssignRequest) (*protobufInt.ACK, error) {
// 	nextUrl, prevUrl, tailUrl := in.NextUrl, in.PrevUrl, in.TailUrl
// 	if nextUrl != "" {
// 		conn, err := grpc.NewClient(nextUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 		if err != nil {
// 			return nil, err
// 		}
// 		grpClient := protobufInt.NewChainNodeClient(conn)
// 		node.clientNext = grpClient

// 	}

// 	if prevUrl != "" {
// 		conn, err := grpc.NewClient(prevUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 		if err != nil {
// 			return nil, err
// 		}
// 		grpClient := protobufInt.NewChainNodeClient(conn)
// 		node.clientPrev = grpClient

// 	}

// 	if tailUrl != "" {
// 		conn, err := grpc.NewClient(tailUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 		if err != nil {
// 			return nil, err
// 		}
// 		grpClient := protobufInt.NewChainNodeClient(conn)
// 		node.clientTail = grpClient
// 	}
// 	return &protobufInt.ACK{}, nil
// }

// func (node *ChainNode) Forward(ctx context.Context, in *protobufInt.ForwardRequest) (*protobufInt.ACK, error) {
// 	//send the version od the first node and wait for ack
// 	if node == nil {
// 		return nil, fmt.Errorf("node can not be nil")
// 	}
// 	//fetch the lastest version if you are forwarding from the head
// 	//fetch the current version and

// 	if node.clientPrev == nil {
// 		version := node.GetDataAndVersion(in.Id)
// 		in.Version = version
// 	}
// 	if node.clientNext != nil {
// 		data := NewData(in.Id, in.Version)
// 		node.SetData(int64(data.Id), data)
// 	} else if node.clientNext == nil {
// 		fmt.Printf("the tail has received the information")
// 		fmt.Printf("ack: %d %d \n", in.Id, in.Version)
// 		data := NewData(in.Id, in.Version)
// 		node.SetData(int64(data.Id), data)
// 		ackResonse := &protobufInt.ACK{Version: data.Version, Success: true}
// 		return ackResonse, nil
// 	}
// 	if fail() == true {
// 		fmt.Printf("node %d has failed to send information\n", node.Id)
// 		return &protobufInt.ACK{Success: false}, fmt.Errorf("failed to forward")
// 	}
// 	fmt.Printf("node %d has forwarded the information\n", node.Id)

// 	ack, err := node.clientNext.Forward(ctx, in)
// 	if err != nil {
// 		fmt.Printf("node %d has failed to receive the information", node.Id)
// 		return nil, err
// 	}
// 	if ack != nil && ack.Success {
// 		version, id := ack.Version, ack.Id
// 		node.NodeData.SetValForKey(id, Data{Version: version, Id: id})
// 		fmt.Println("set ", ack.Version, " as the last clean version")
// 		node.LastCleanVersion.SetValForKey(id, version)
// 	}
// 	fmt.Printf("node %d has received confirmation of the update\n", node.Id)
// 	data := NewData(in.Id, in.Version)
// 	node.SetData(int64(data.Id), data)

// 	return ack, err
// }

// func (node *ChainNode) SetData(id int64, data *Data) {
// 	node.NodeData.SetValForKey(id, *data)
// }

// func (node *ChainNode) Read(ctx context.Context, in *protobufInt.ReadRequest) (*protobufInt.ReadResponse, error) {
// 	if node == nil {
// 		return nil, fmt.Errorf("node does not exist")
// 	}
// 	fmt.Println("node tries reading from the data source")
// 	if node.clientTail == nil && node.clientNext != nil {
// 		return nil, fmt.Errorf("no tail found")
// 	}
// 	//check last clean version
// 	//if lqst clean version < current version
// 	//->read from the tail
// 	//else read from the current node
// 	cleanVersion, _ := node.LastCleanVersion.GetValByKey(in.Id)
// 	object, _ := node.NodeData.GetValByKey(in.Id)
// 	ok := true

// 	fmt.Println("cleanVersion ", cleanVersion)
// 	fmt.Println("object version ", object.Version)

// 	if cleanVersion == object.Version {
// 		fmt.Println("object has the newest version")
// 		return &protobufInt.ReadResponse{Ok: ok, Data: object.Id}, nil
// 	} else if resp, err := node.clientTail.Read(ctx, in); err != nil {
// 		fmt.Println("read from the tail")
// 		return resp, nil
// 	} else {
// 		return resp, nil
// 	}
// }

func fail() bool {
	p := rand.Float32()
	if p < 0.5 {
		return false
	}
	return true
}
