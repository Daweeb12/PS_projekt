package messageboardserver

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	// "PS_projekt/storage"
	"context"
	// "fmt"
	// "math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"errors"
	//"fmt"
	//"time"
)

func (node *MessageBoardServer) HeartBeat(ctx context.Context, in *pbRaz.HearthBeatRequest) (*pbRaz.HearthBeatResponse, error) {
	return &pbRaz.HearthBeatResponse{Id: node.Id}, nil
}

func (node *MessageBoardServer) AssignChainNode(ctx context.Context, in *pbRaz.AssignRequest) (*pbRaz.ACK, error) {
	nextUrl, prevUrl, tailUrl := in.NextUrl, in.PrevUrl, in.TailUrl
	if nextUrl != "" {
		conn, err := grpc.NewClient(nextUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		grpClient := pbRaz.NewMessageBoardClient(conn)
		node.ClientNext = grpClient

	}

	if prevUrl != "" {
		conn, err := grpc.NewClient(prevUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		grpClient := pbRaz.NewMessageBoardClient(conn)
		node.ClientPrev = grpClient

	}

	if tailUrl != "" {
		conn, err := grpc.NewClient(tailUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		grpClient := pbRaz.NewMessageBoardClient(conn)
		node.ClientTail = grpClient
	}
	return &pbRaz.ACK{}, nil
}
