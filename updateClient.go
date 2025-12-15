package main

import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func UpdateClient(url string) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	grpcClient := protobufInt.NewMasterNodeClient(conn)
	for {
		fmt.Println("Update called")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		ack, err := grpcClient.Update(ctx, &protobufInt.UpdateRequest{})
		if err != nil {
			fmt.Println(err)
		}
		cancel()

		if ack.Success {
			fmt.Println("information updated")
		}
		time.Sleep(time.Second * 2)
	}
}
