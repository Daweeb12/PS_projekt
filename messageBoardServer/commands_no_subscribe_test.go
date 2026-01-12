package messageboardserver

import (
	"context"
	"testing"

	protobufRazpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestCreateListRead exercises CreateUser/CreateTopic and ListTopics/Read* helpers.
func TestCreateListRead(t *testing.T) {
	srv := NewMessageBoardServer(10)

	// Create user
	ureq := &protobufRazpravljalnica.CreateUserRequest{Name: "u1"}
	u, err := srv.CreateUser(context.Background(), ureq)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Read user
	ru, err := srv.ReadUser(context.Background(), &protobufRazpravljalnica.ReadUserRequest{Id: u.Id})
	if err != nil {
		t.Fatalf("ReadUser failed: %v", err)
	}
	if ru.User == nil || ru.User.Name != "u1" {
		t.Fatalf("unexpected user read back: %+v", ru.User)
	}

	// Create topic
	treg := &protobufRazpravljalnica.CreateTopicRequest{Name: "top1"}
	tp, err := srv.CreateTopic(context.Background(), treg)
	if err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}

	// List topics
	lresp, err := srv.ListTopics(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("ListTopics failed: %v", err)
	}
	found := false
	for _, tpi := range lresp.Topics {
		if tpi.Id == tp.Id && tpi.Name == "top1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created topic not present in ListTopics")
	}

	// Read topic
	rt, err := srv.ReadTopic(context.Background(), &protobufRazpravljalnica.ReadTopicRequest{Id: tp.Id})
	if err != nil {
		t.Fatalf("ReadTopic failed: %v", err)
	}
	if rt.Topic == nil || rt.Topic.Name != "top1" {
		t.Fatalf("unexpected topic read back: %+v", rt.Topic)
	}
}

// TestPostUpdateLikeDelete exercises message lifecycle operations.
func TestPostUpdateLikeDelete(t *testing.T) {
	srv := NewMessageBoardServer(11)

	// prepare user/topic
	u, err := srv.CreateUser(context.Background(), &protobufRazpravljalnica.CreateUserRequest{Name: "poster"})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	tp, err := srv.CreateTopic(context.Background(), &protobufRazpravljalnica.CreateTopicRequest{Name: "forum"})
	if err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}

	// Post message
	pmReq := &protobufRazpravljalnica.PostMessageRequest{TopicId: tp.Id, UserId: u.Id, Text: "hello"}
	m, err := srv.PostMessage(context.Background(), pmReq)
	if err != nil {
		t.Fatalf("PostMessage failed: %v", err)
	}
	if m.Text != "hello" || m.TopicId != tp.Id || m.UserId != u.Id {
		t.Fatalf("unexpected posted message: %+v", m)
	}

	// GetMessages should include the message
	gm, err := srv.GetMessages(context.Background(), &protobufRazpravljalnica.GetMessagesRequest{TopicId: tp.Id, FromMessageId: 0, Limit: 10})
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
	found := false
	for _, msg := range gm.Messages {
		if msg.Id == m.Id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("posted message not found in GetMessages")
	}

	// Update message
	umReq := &protobufRazpravljalnica.UpdateMessageRequest{TopicId: tp.Id, UserId: u.Id, MessageId: m.Id, Text: "edited"}
	um, err := srv.UpdateMessage(context.Background(), umReq)
	if err != nil {
		t.Fatalf("UpdateMessage failed: %v", err)
	}
	if um.Text != "edited" {
		t.Fatalf("expected edited text, got %q", um.Text)
	}

	// Like message
	lkReq := &protobufRazpravljalnica.LikeMessageRequest{TopicId: tp.Id, MessageId: m.Id, UserId: u.Id}
	lk, err := srv.LikeMessage(context.Background(), lkReq)
	if err != nil {
		t.Fatalf("LikeMessage failed: %v", err)
	}
	if lk.Likes < 1 {
		t.Fatalf("expected Likes>=1 after like, got %d", lk.Likes)
	}

	// ReadMessage returns the message data
	rd, err := srv.ReadMessage(context.Background(), &protobufRazpravljalnica.ReadMessageRequest{Id: m.Id})
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}
	if rd.Msg == nil || rd.Msg.Id != m.Id {
		t.Fatalf("unexpected ReadMessage result: %+v", rd)
	}

	// Delete message
	_, err = srv.DeleteMessage(context.Background(), &protobufRazpravljalnica.DeleteMessageRequest{TopicId: tp.Id, UserId: u.Id, MessageId: m.Id})
	if err != nil {
		t.Fatalf("DeleteMessage failed: %v", err)
	}

	// After delete, GetMessages should not return the message
	gm2, err := srv.GetMessages(context.Background(), &protobufRazpravljalnica.GetMessagesRequest{TopicId: tp.Id, FromMessageId: 0, Limit: 10})
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
	for _, msg := range gm2.Messages {
		if msg.Id == m.Id {
			t.Fatalf("message still present after delete: %d", m.Id)
		}
	}

	// Ensure ReadMessage now returns error
	if _, err := srv.ReadMessage(context.Background(), &protobufRazpravljalnica.ReadMessageRequest{Id: m.Id}); err == nil {
		t.Fatalf("expected error reading deleted message, got nil")
	}

	// sanity: create another message to ensure seq and storage still functional
	pm2 := &protobufRazpravljalnica.PostMessageRequest{TopicId: tp.Id, UserId: u.Id, Text: "again"}
	m2, err := srv.PostMessage(context.Background(), pm2)
	if err != nil {
		t.Fatalf("second PostMessage failed: %v", err)
	}
	if m2 == nil || m2.Text != "again" {
		t.Fatalf("unexpected second message: %+v", m2)
	}
	_ = timestamppb.Now() // keep imports used
}
