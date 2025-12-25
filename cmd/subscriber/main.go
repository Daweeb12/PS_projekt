package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	protobufRazpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	url := flag.String("url", "localhost:9001", "gRPC server address")
	flag.Parse()

	conn, err := grpc.Dial(*url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := protobufRazpravljalnica.NewMessageBoardClient(conn)
	ctx := context.Background()
	stream, err := client.SubscribeTopic(ctx, &protobufRazpravljalnica.SubscribeTopicRequest{TopicId: nil, FromMessageId: 0})
	if err != nil {
		panic(err)
	}
	fmt.Println("subscribed, waiting for events")

	for {
		ev, err := stream.Recv()
		if err != nil {
			fmt.Fprintln(os.Stderr, "stream closed:", err)
			return
		}

		msgId := int64(0)
		txt := ""
		if ev.Message != nil {
			msgId = ev.Message.GetId()
			txt = ev.Message.GetText()
		}

		fmt.Printf("EVENT: op=%v seq=%d msg_id=%v text=%q\n", ev.Op, ev.SequenceNumber, msgId, txt)
	}
}
