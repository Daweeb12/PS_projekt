package tests

import (
	protobufRazpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"
	messageboardserver "PS_projekt/messageBoardServer"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTopic(t *testing.T) {
	fmt.Println("====topic tests====")
	server := messageboardserver.NewMessageBoardServer()
	topics := []string{"movies", "cinema", "books", "comics"}
	for _, topic := range topics {
		topicReq := protobufRazpravljalnica.CreateTopicRequest{Name: topic}
		topic, err := server.CreateTopic(t.Context(), &topicReq)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(topic)

	}

	fmt.Println("====topic tests====")
}

func TestEmptyTopic(t *testing.T) {
	fmt.Println("===empty topic test#2===")
	server := messageboardserver.NewMessageBoardServer()
	topicReq := protobufRazpravljalnica.CreateTopicRequest{Name: ""}
	if topic, err := server.CreateTopic(t.Context(), &topicReq); err != nil {
		t.Error(err)
	} else {
		fmt.Println(topic)
	}

	fmt.Println("===empty topic test#2===")
}

func TestDuplicateTopic(t *testing.T) {
	fmt.Println("==== duplicate topic ====")
	server := messageboardserver.NewMessageBoardServer()
	topicReq := protobufRazpravljalnica.CreateTopicRequest{Name: "football"}
	if _, err := server.CreateTopic(t.Context(), &topicReq); err != nil {
		t.Error(err)
	}

	_, err := server.CreateTopic(t.Context(), &topicReq)
	assert.False(t, err == nil)
	fmt.Println("==== duplicate topic ====")

}
