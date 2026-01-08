package main

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	masterUrl string = "localhost:9000"
)

var topicId atomic.Int64

func CraqSim() {
	go Server(masterUrl)
	time.Sleep(2 * time.Second)
	id := 1
	errCh := make(chan error)
	wg := sync.WaitGroup{}
	for i := 9001; i <= 9003; i++ {
		url := fmt.Sprintf("localhost:%d")
		go AddMsgBoardServer(url, masterUrl, int64(id))
		id++
		time.Sleep(100 * time.Millisecond)
	}
	//run the clients
	wg.Add(6)
	for range 6 {
		go simulateConsumers(errCh, &wg)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			fmt.Println(err)
		}

	}

}

func fail() bool {
	if rand.Float64() < 0.7 {
		return false
	}
	return true
}

func simulateConsumers(errCh chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := grpc.NewClient(masterUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		errCh <- err
		return
	}
	defer func() {
		conn.Close()
	}()
	ticker := time.NewTicker(time.Second * 10)
	masterGrpcClient := pbRaz.NewMasterNodeClient(conn)
	_, _, err = fetchDetails(masterGrpcClient)
	if err != nil {
		errCh <- err
		return
	}

	select {
	case <-ticker.C:
		fmt.Println("client has disconnected")
		return
	default:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		if randNode, err := masterGrpcClient.GenerateRandomNode(ctx, &emptypb.Empty{}); err == nil {
			randNodeConn, err := grpc.NewClient(randNode.Node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				errCh <- err

				return
			}

			randNodeConn.Close()
			randNodeClient := pbRaz.NewMessageBoardClient(randNodeConn)
			if topics, err := randNodeClient.ListTopicsFromInner(ctx, &emptypb.Empty{}); err == nil {
				for _, topic := range topics.Topics {
					fmt.Println(topic)
				}
			} else {
				errCh <- err
				return
			}

		} else {
			errCh <- err
			return
		}
		time.Sleep(1 * time.Second)
		cancel()

	}

}

func simulateWorkers(errCh chan error, wg *sync.WaitGroup) {

	defer wg.Done()
	conn, err := grpc.NewClient(masterUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		errCh <- err
		return
	}
	defer func() {
		conn.Close()
	}()
	grpcClient := pbRaz.NewMessageBoardClient(conn)
	for range 10 {
		id := topicId.Add(1)
		topicName := fmt.Sprintf("topic%d", id-1)
		newTopicReq := pbRaz.CreateTopicRequest{Name: topicName, Version: 0}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if _, err := grpcClient.CreateTopic(ctx, &newTopicReq); err == nil {
			fmt.Println(topicName, " uploaded successfully")
		} else {
			errCh <- err
			return
		}
		_ = topicId.CompareAndSwap(10, 0)
	}

}
