package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Reconfig() {
	go StartMasterServer(masterUrl, 0)
	time.Sleep(2 * time.Second)
	msgBoardUrl1 := fmt.Sprintf("localhost:%d", 9001)
	go AddMsgBoardServer(msgBoardUrl1, masterUrl, int64(1))
	time.Sleep(1000 * time.Millisecond)
	workersWg := sync.WaitGroup{}
	errChw := make(chan error, 10)
	workersWg.Add(3)
	for range 3 {
		go simulateWorkers(errChw, &workersWg)
	}
	workersWg.Wait()

	conn, err := grpc.NewClient(masterUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		conn.Close()
	}()
	masterGrpcClient := pbRaz.NewMasterNodeClient(conn)
	_, tail, err := fetchDetails(masterGrpcClient)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("client has started")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	tailConn, err := grpc.NewClient(tail.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		return
	}
	tailClient := pbRaz.NewMessageBoardClient(tailConn)
	defer tailConn.Close()
	if topics, err := tailClient.ListTopics(ctx, &emptypb.Empty{}); err == nil {
		fmt.Println("TAIL: ")
		fmt.Println("===TOPICS===")
		for _, topic := range topics.Topics {
			fmt.Println(topic)
		}
		fmt.Println("===TOPICS===")
	} else {
		fmt.Println("err: ", err)

	}
	fmt.Printf("\n\n")

	msgBoardUrl2 := fmt.Sprintf("localhost:%d", 9002)
	go AddMsgBoardServer(msgBoardUrl2, masterUrl, int64(2))
	time.Sleep(1000 * time.Millisecond)

	defer func() {
		conn.Close()
	}()
	_, tail, err = fetchDetails(masterGrpcClient)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("client has started")
	ctx2, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	newTailConn, err := grpc.NewClient(tail.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		return
	}
	newTailClient := pbRaz.NewMessageBoardClient(newTailConn)
	defer newTailConn.Close()
	if topics, err := newTailClient.ListTopics(ctx2, &emptypb.Empty{}); err == nil {
		fmt.Println("TAIL: ")
		fmt.Println("===TOPICS===")
		for _, topic := range topics.Topics {
			fmt.Println(topic)
		}
		fmt.Println("===TOPICS===")
	} else {
		fmt.Println("err: ", err)

	}
	fmt.Printf("\n\n")

}
