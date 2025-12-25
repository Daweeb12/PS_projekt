package main

import (
	protobufInt "PS_projekt/api/grpc/protobufInternal"
	razpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Client(url string) {
	fmt.Println("Client using gRPC connection.")
	fmt.Printf("Client connecting to URL %s.\n", url)

	// grpc connection with server
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// initiate gRPC interface
	grpcClient := razpravljalnica.NewMessageBoardClient(conn)

	// start main loop
	mainLoop(&grpcClient)
}

func mainLoop(client *razpravljalnica.MessageBoardClient) {
	scanner := bufio.NewScanner(os.Stdin)
	var currentUserID int64

	/*
		fmt.Println("\n=== Forum Client ===")
		fmt.Println("Commands:")
		fmt.Println("  1. createuser <name>        - Create a new user")
		fmt.Println("  2. createtopic <name>       - Create a new topic")
		fmt.Println("  3. listtopics               - List all topics")
		fmt.Println("  4. getmessages <topic_id>   - Get messages from a topic")
		fmt.Println("  5. postmessage <topic_id> <text>		- Post a message")
		fmt.Println("  6. likemessage <topic_id> <msg_id>	- Like a message")
		fmt.Println("  7. updatemessage <topic_id> <msg_id> <text>	- Update a message")
		fmt.Println("  8. deletemessage <topic_id> <msg_id>		- Delete a message")
		fmt.Println("  9. setuser <user_id>       	- Set current user ID")
		fmt.Println(" 10. exit 						- Exit the application")
		fmt.Println(" 11. help 						- Get command list")
		fmt.Println()
	*/

	info()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command := strings.ToLower(parts[0])

		// Create a context with timeout for each request
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		switch command {
		case "createuser":
			if len(parts) < 2 {
				fmt.Println("Usage: createuser <name>")
				cancel()
				continue
			}
			handleCreateUser(ctx, client, parts[1])

		case "createtopic":
			if len(parts) < 2 {
				fmt.Println("Usage: createtopic <name>")
				cancel()
				continue
			}
			handleCreateTopic(ctx, client, parts[1])

		case "listtopics":
			handleListTopics(ctx, client)

		case "getmessages":
			if len(parts) < 2 {
				fmt.Println("Usage: getmessages <topic_id>")
				cancel()
				continue
			}
			topicID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				fmt.Println("Invalid topic ID")
				cancel()
				continue
			}
			handleGetMessages(ctx, client, topicID)

		case "postmessage":
			if len(parts) < 3 {
				fmt.Println("Usage: postmessage <topic_id> <text>")
				cancel()
				continue
			}
			if currentUserID == 0 {
				fmt.Println("Please set user ID first: setuser <user_id>")
				cancel()
				continue
			}
			topicID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				fmt.Println("Invalid topic ID")
				cancel()
				continue
			}
			text := strings.Join(parts[2:], " ")
			handlePostMessage(ctx, client, topicID, currentUserID, text)

		case "likemessage":
			if len(parts) < 3 {
				fmt.Println("Usage: likemessage <topic_id> <message_id>")
				cancel()
				continue
			}
			if currentUserID == 0 {
				fmt.Println("Please set user ID first: setuser <user_id>")
				cancel()
				continue
			}
			topicID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				fmt.Println("Invalid topic ID")
				cancel()
				continue
			}
			messageID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				fmt.Println("Invalid message ID")
				cancel()
				continue
			}
			handleLikeMessage(ctx, client, topicID, messageID, currentUserID)

		case "updatemessage":
			if len(parts) < 4 {
				fmt.Println("Usage: updatemessage <topic_id> <message_id> <text>")
				cancel()
				continue
			}
			if currentUserID == 0 {
				fmt.Println("Please set user ID first: setuser <user_id>")
				cancel()
				continue
			}
			topicID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				fmt.Println("Invalid topic ID")
				cancel()
				continue
			}
			messageID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				fmt.Println("Invalid message ID")
				cancel()
				continue
			}
			text := strings.Join(parts[3:], " ")
			handleUpdateMessage(ctx, client, topicID, messageID, currentUserID, text)

		case "deletemessage":
			if len(parts) < 3 {
				fmt.Println("Usage: deletemessage <topic_id> <message_id>")
				cancel()
				continue
			}
			if currentUserID == 0 {
				fmt.Println("Please set user ID first: setuser <user_id>")
				cancel()
				continue
			}
			topicID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				fmt.Println("Invalid topic ID")
				cancel()
				continue
			}
			messageID, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				fmt.Println("Invalid message ID")
				cancel()
				continue
			}
			handleDeleteMessage(ctx, client, topicID, messageID, currentUserID)

		case "setuser":
			if len(parts) < 2 {
				fmt.Println("Usage: setuser <user_id>")
				cancel()
				continue
			}
			userID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				fmt.Println("Invalid user ID")
				cancel()
				continue
			}
			currentUserID = userID
			fmt.Printf("Current user set to: %d\n", currentUserID)

		case "exit":
			fmt.Println("Exiting...")
			cancel()
			return

		case "help":
			info()
			return

		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}

		cancel()
	}
}

func handleCreateUser(ctx context.Context, client *razpravljalnica.MessageBoardClient, name string) {
	req := &razpravljalnica.CreateUserRequest{Name: name}
	user, err := (*client).CreateUser(ctx, req)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	fmt.Printf("User created: ID=%d, Name=%s\n", user.Id, user.Name)
}

func handleCreateTopic(ctx context.Context, client *razpravljalnica.MessageBoardClient, name string) {
	req := &razpravljalnica.CreateTopicRequest{Name: name}
	topic, err := (*client).CreateTopic(ctx, req)
	if err != nil {
		fmt.Printf("Error creating topic: %v\n", err)
		return
	}
	fmt.Printf("Topic created: ID=%d, Name=%s\n", topic.Id, topic.Name)
}

func handleListTopics(ctx context.Context, client *razpravljalnica.MessageBoardClient) {
	resp, err := (*client).ListTopics(ctx, &emptypb.Empty{})
	if err != nil {
		fmt.Printf("Error listing topics: %v\n", err)
		return
	}

	if len(resp.Topics) == 0 {
		fmt.Println("No topics available.")
		return
	}

	fmt.Println("\n=== Topics ===")
	for _, topic := range resp.Topics {
		fmt.Printf("ID: %d | Name: %s\n", topic.Id, topic.Name)
	}
	fmt.Println()
}

func handleGetMessages(ctx context.Context, client *razpravljalnica.MessageBoardClient, topicID int64) {
	req := &razpravljalnica.GetMessagesRequest{
		TopicId:       topicID,
		FromMessageId: 0,
		Limit:         100,
	}
	resp, err := (*client).GetMessages(ctx, req)
	if err != nil {
		fmt.Printf("Error getting messages: %v\n", err)
		return
	}

	if len(resp.Messages) == 0 {
		fmt.Println("No messages in this topic.")
		return
	}

	fmt.Printf("\n=== Messages in Topic %d ===\n", topicID)
	for _, msg := range resp.Messages {
		fmt.Printf("ID: %d | User: %d | Likes: %d\n", msg.Id, msg.UserId, msg.Likes)
		fmt.Printf("Text: %s\n", msg.Text)
		if msg.CreatedAt != nil {
			fmt.Printf("Created: %s\n", msg.CreatedAt.AsTime().Format(time.RFC3339))
		}
		fmt.Println("---")
	}
	fmt.Println()
}

func handlePostMessage(ctx context.Context, client *razpravljalnica.MessageBoardClient, topicID, userID int64, text string) {
	req := &razpravljalnica.PostMessageRequest{
		TopicId: topicID,
		UserId:  userID,
		Text:    text,
	}
	msg, err := (*client).PostMessage(ctx, req)
	if err != nil {
		fmt.Printf("Error posting message: %v\n", err)
		return
	}
	fmt.Printf("Message posted: ID=%d, Likes=%d\n", msg.Id, msg.Likes)
}

func handleLikeMessage(ctx context.Context, client *razpravljalnica.MessageBoardClient, topicID, messageID, userID int64) {
	req := &razpravljalnica.LikeMessageRequest{
		TopicId:   topicID,
		MessageId: messageID,
		UserId:    userID,
	}
	msg, err := (*client).LikeMessage(ctx, req)
	if err != nil {
		fmt.Printf("Error liking message: %v\n", err)
		return
	}
	fmt.Printf("Message liked! Total likes: %d\n", msg.Likes)
}

func handleUpdateMessage(ctx context.Context, client *razpravljalnica.MessageBoardClient, topicID, messageID, userID int64, text string) {
	req := &razpravljalnica.UpdateMessageRequest{
		TopicId:   topicID,
		UserId:    userID,
		MessageId: messageID,
		Text:      text,
	}
	msg, err := (*client).UpdateMessage(ctx, req)
	if err != nil {
		fmt.Printf("Error updating message: %v\n", err)
		return
	}
	fmt.Printf("Message updated: ID=%d, New text: %s\n", msg.Id, msg.Text)
}

func handleDeleteMessage(ctx context.Context, client *razpravljalnica.MessageBoardClient, topicID, messageID, userID int64) {
	req := &razpravljalnica.DeleteMessageRequest{
		TopicId:   topicID,
		UserId:    userID,
		MessageId: messageID,
	}
	_, err := (*client).DeleteMessage(ctx, req)
	if err != nil {
		fmt.Printf("Error deleting message: %v\n", err)
		return
	}
	fmt.Println("Message deleted successfully!")
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

func info() {
	fmt.Println("\n=== Forum Client ===")
	fmt.Println("Commands:")
	fmt.Println("  1. createuser <name>        - Create a new user")
	fmt.Println("  2. createtopic <name>       - Create a new topic")
	fmt.Println("  3. listtopics               - List all topics")
	fmt.Println("  4. getmessages <topic_id>   - Get messages from a topic")
	fmt.Println("  5. postmessage <topic_id> <text>		- Post a message")
	fmt.Println("  6. likemessage <topic_id> <msg_id>	- Like a message")
	fmt.Println("  7. updatemessage <topic_id> <msg_id> <text>	- Update a message")
	fmt.Println("  8. deletemessage <topic_id> <msg_id>		- Delete a message")
	fmt.Println("  9. setuser <user_id>       	- Set current user ID")
	fmt.Println(" 10. exit 						- Exit the application")
	fmt.Println(" 11. help 						- Get command list")
	fmt.Println()
}

// connect to master node and fetch head/tail. forward small periodic request to the head.
func UpdateClient(url string) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	grpcClient := protobufInt.NewMasterNodeClient(conn)

	for {
		headInfo, tailInfo, err := fetchDetails(grpcClient)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}
		fmt.Println("HEAD: ", headInfo)
		fmt.Println("TAIL: ", tailInfo)
		fmt.Println()
		headConn, err := grpc.NewClient(headInfo.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}

		// close connection afrerwards
		headClient := razpravljalnica.NewMessageBoardClient(headConn)
		if user, err := sendCreateUserReq(headClient); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("CREATE USER ", user)
		}

		headConn.Close()
		time.Sleep(time.Second * 2)
	}
}

func fetchDetails(grpcClient protobufInt.MasterNodeClient) (*protobufInt.NodeData, *protobufInt.NodeData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	clusterInfo, err := grpcClient.GetClusterState(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, nil, err
	}
	headInfo, tailInfo := clusterInfo.Head, clusterInfo.Tail
	return headInfo, tailInfo, nil
}

func sendCreateUserReq(grpcClient razpravljalnica.MessageBoardClient) (*razpravljalnica.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	createUserReq := &razpravljalnica.CreateUserRequest{Name: "david"}
	user, err := grpcClient.CreateUser(ctx, createUserReq)
	return user, err
}
