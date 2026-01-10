package main

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
	"context"
	"fmt"
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
	go StartMasterServer(masterUrl, 0)
	time.Sleep(500 * time.Millisecond)
	id := 1
	errCh := make(chan error)
	wg := sync.WaitGroup{}
	for i := 9001; i <= 9005; i++ {
		url := fmt.Sprintf("localhost:%d", i)
		fmt.Println(url)
		go AddMsgBoardServer(url, masterUrl, int64(id))
		id++
		time.Sleep(250 * time.Millisecond)
	}
	t1 := time.Now()
	workersWg := sync.WaitGroup{}
	errChw := make(chan error, 100)
	workersWg.Add(1)
	for range 1 {
		simulateWorkers(errChw, &workersWg)
	}
	workersWg.Wait()
	fmt.Println("++++++++++++++++++++++++++++++workers finished+++++++++++++++++++++++++++++++")
	//run the clients
	fmt.Println("++++++++++++++++++++++++++++++consumers started+++++++++++++++++++++++++++++++")
	wg.Add(1000)
	for range 1000 {
		simulateConsumers(errCh, &wg)
	}
	wg.Wait()
	fmt.Println("++++++++++++++++++++++++++consumers finished++++++++++++++++++++++++++++++++")
	close(errCh)
	for err := range errCh {
		if err != nil {
			fmt.Println(err)
		}

	}
	t2 := time.Now()
	fmt.Println("elapsed time: ", t2.Sub(t1))
}

func simulateConsumers(errCh chan error, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		// fmt.Println("consumer finished")
	}()
	conn, err := grpc.NewClient(masterUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		errCh <- err
		fmt.Println(err)
		return
	}
	defer func() {
		conn.Close()
	}()
	masterGrpcClient := pbRaz.NewMasterNodeClient(conn)
	//ctxMaster , cancel := context.WithTimeout(context.Background() , 10*time.Second)
	// clusterInfo,  err :=masterGrpcClient.GetClusterState(ctxMaster, &emptypb.Empty{})
	// if err != nil {
	// 	errCh <- err
	// 	cancel()
	// }
	// cancel()
	fmt.Println("client has started")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	randNodeInfo, err := masterGrpcClient.GenerateRandomNode(ctx, &emptypb.Empty{})
	if err != nil {
		return
	}
	randNodeConn, err := grpc.NewClient(randNodeInfo.Node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}
	randNodeClient := pbRaz.NewMessageBoardClient(randNodeConn)
	defer randNodeConn.Close()
	// tailNodeConn, err := grpc.NewClient(clusterInfo.Tail.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// tailClient := pbRaz.NewMessageBoardClient(tailNodeConn)
	// defer tailNodeConn.Close()
	for range 1 {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		if topics, err := randNodeClient.ListTopicsFromInner(ctx, &emptypb.Empty{}); err == nil {
			// fmt.Println("Node: ",randNodeInfo.Node)
			// fmt.Println("===TOPICS===")
			for _, _ = range topics.Topics {
				// fmt.Println(topic)
			}
			// fmt.Println("===TOPICS===")
		} else {

			// fmt.Println("expected to come here")
		}

		cancel()

	}

}

func simulateWorkers(errCh chan error, wg *sync.WaitGroup) {

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

	for range 1 {
		id := topicId.Add(1)
		topicName := fmt.Sprintf("topic%d", id-1)
		newTopicReq := pbRaz.CreateTopicRequest{Name: topicName, Version: 0}
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

func getGrpcClient(url string) (pbRaz.MessageBoardClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	msgBoardClient := pbRaz.NewMessageBoardClient(conn)
	return msgBoardClient, conn, nil

}
