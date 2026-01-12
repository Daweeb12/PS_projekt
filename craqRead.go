package main

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func CraqRead() {
	go StartMasterServer(masterUrl, 0)
	time.Sleep(500 * time.Millisecond)
	id := int64(1)
	for i := 9001; i <= 9003; i++ {
		url := fmt.Sprintf("localhost:%d", i)
		fmt.Println(url)
		go AddMsgBoardServer(url, masterUrl, int64(id))
		id++
		time.Sleep(250 * time.Millisecond)
	}
	errCh := make(chan error, 100)
	wg := sync.WaitGroup{}
	wg.Add(N)
	for range N {
		addTopic(errCh, &wg)
	}
	wg.Wait()
	firstRun := true
	idx := []int64{}
	for i := 9001; i <= 9003; i++ {
		url := fmt.Sprintf("localhost:%d", i)
		if conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
			fmt.Println(err)
			errCh <- err
			continue
		} else {
			client := pbRaz.NewMessageBoardClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel() // cancel after request
			resp, err := client.ListTopics(ctx, &emptypb.Empty{})
			if err != nil {
				fmt.Printf("Client %d request failed: %v", i, err)
				continue
			}
			for _, topic := range resp.Topics {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if t, err := client.ReadTopic(ctx, &pbRaz.ReadTopicRequest{Id: topic.Id}); err == nil {
					fmt.Println("topic: ", topic, "dirty: ", t.Dirty)
					if firstRun {
						idx = append(idx, topic.Id)
					}
				} else {
					fmt.Println(err)
				}
				cancel()
			}
			conn.Close()
			fmt.Println()
		}

	}
	masterNodeConn, err := grpc.NewClient(masterUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer masterNodeConn.Close()
	masterClient := pbRaz.NewMasterNodeClient(masterNodeConn)
	for _, id := range idx {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		randNodeInfo, err := masterClient.GenerateRandomNode(ctx, &emptypb.Empty{})
		if err != nil {
			panic(err)
		}
		fmt.Println("Reading from node", randNodeInfo.Node.Id)
		randNodeConn, err := grpc.NewClient(randNodeInfo.Node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		defer randNodeConn.Close()
		randClient := pbRaz.NewMessageBoardClient(randNodeConn)
		if topic, err := randClient.ReadTopic(ctx, &pbRaz.ReadTopicRequest{Id: id}); err == nil {
			if topic.Dirty {
				fmt.Println("reading from tail")
			} else {
				fmt.Println("data clean")
			}

		} else {
			fmt.Println(randNodeInfo.Node.Address)
		}

	}

}

func addTopic(errCh chan error, wg *sync.WaitGroup) {

	defer func() {
		wg.Done()
		// fmt.Println("worker finished")
	}()
	conn, err := grpc.NewClient(masterUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		errCh <- err
		return
	}
	defer func() {
		conn.Close()
	}()
	masterNodeClient := pbRaz.NewMasterNodeClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	clusterStateResponse, err := masterNodeClient.GetClusterState(ctx, &emptypb.Empty{})
	if err != nil {
		errCh <- err
	}
	headClient, conn, err := getGrpcClient(clusterStateResponse.Head.Address)
	if err != nil {
		errCh <- err
	}
	defer conn.Close()

	for i := range 10 {
		id := topicId.Add(1)
		topicName := fmt.Sprintf("topic %d", id-1)
		newTopicReq := pbRaz.CreateTopicRequest{Name: topicName, Version: int64(i)}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		// defer cancel()
		if _, err := headClient.CreateTopic(ctx, &newTopicReq); err == nil {
			// fmt.Println(topicName, " uploaded successfully")
		} else {
			// fmt.Println(err,"xx")
			errCh <- err
		}
		cancel()
	}

}
