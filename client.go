package main

import (
	razpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Client(url string) {
	fmt.Println("Client using gRPC connection.")
	fmt.Printf("Client connecting to URL %s.\n")

	// grpc connection with server
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// initiate environment
	contextCrud, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// initiate gRPC interface
	grpcClient := razpravljalnica.NewMessageBoardClient(conn)

	// posting example
	testing(&contextCrud, &grpcClient)
	/*
		topicCreate := razpravljalnica.CreateTopicRequest{Name: "ExampleTopic"}

		if _, err := grpcClient.CreateTopic(contextCrud, &topicCreate); err != nil {
			panic(err)
		}
		fmt.Println("Done")
	*/
}

func testing(ctx *context.Context, client *razpravljalnica.MessageBoardClient) {
	context := *ctx
	grpcClient := *client
	// posting example
	topicCreate := razpravljalnica.CreateTopicRequest{Name: "ExampleTopic"}

	if _, err := grpcClient.CreateTopic(context, &topicCreate); err != nil {
		panic(err)
	}
	fmt.Println("Done")
}
