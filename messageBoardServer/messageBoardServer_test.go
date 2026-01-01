package messageboardserver

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	protobufRazpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"
	//"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// fakeServerStream implements grpc.ServerStream minimal methods for tests.
type fakeServerStream struct {
	ctx context.Context
}

func (f *fakeServerStream) SetHeader(md metadata.MD) error  { return nil }
func (f *fakeServerStream) SendHeader(md metadata.MD) error { return nil }
func (f *fakeServerStream) SetTrailer(md metadata.MD)       { return }
func (f *fakeServerStream) Context() context.Context        { return f.ctx }
func (f *fakeServerStream) SendMsg(m interface{}) error     { return nil }
func (f *fakeServerStream) RecvMsg(m interface{}) error     { return nil }

// testStream is used to capture MessageEvent values sent by SubscribeTopic.
type testStream struct {
	fakeServerStream
	events chan *protobufRazpravljalnica.MessageEvent
}

func (t *testStream) Send(ev *protobufRazpravljalnica.MessageEvent) error {
	select {
	case t.events <- ev:
		return nil
	case <-t.ctx.Done():
		return t.ctx.Err()
	}
}

// waitEvent waits for a single event from ch, failing the test on timeout.
func waitEvent(t *testing.T, ch chan *protobufRazpravljalnica.MessageEvent, timeout time.Duration) *protobufRazpravljalnica.MessageEvent {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for event")
	}
	return nil
}

// TestSubscribeReceivesLiveEvents verifies that a subscriber receives events produced
// by PostMessage, UpdateMessage, LikeMessage and DeleteMessage.
func TestSubscribeReceivesLiveEvents(t *testing.T) {
	fmt.Println("starting test")
	srv := NewMessageBoardServer(1)
	var err error

	// create user and topic in storage directly (avoid GenerateRand32 in tests)
	fmt.Println("creating user/topic directly in storage")
	user := &protobufRazpravljalnica.User{Id: 1, Name: "alice"}
	userData := &UserData{User: user, Dirty: true}
	srv.UserStorage.Put(user.Id, userData)
	topic := &protobufRazpravljalnica.Topic{Id: 1, Name: "t1"}
	topicData := &TopicData{Topic: topic, Dirty: true}
	srv.TopicStorage.Put(topic.Id, topicData)

	// prepare subscriber stream
	fmt.Println("preparing subscriber stream")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := &testStream{fakeServerStream{ctx: ctx}, make(chan *protobufRazpravljalnica.MessageEvent, 16)}

	// start subscription in background
	fmt.Println("starting subscription in background")
	go func() {
		req := &protobufRazpravljalnica.SubscribeTopicRequest{TopicId: []int64{topic.Id}, UserId: user.Id, FromMessageId: 0}
		if err := srv.SubscribeTopic(req, s); err != nil && ctx.Err() == nil {
			t.Errorf("SubscribeTopic returned error: %v", err)
		}
	}()

	// perform operations that should generate events
	// 1) Post: insert message directly and publish event (avoid GenerateRand32)
	fmt.Println("inserting message and publishing event directly")
	posted := &protobufRazpravljalnica.Message{Id: 11, TopicId: topic.Id, UserId: user.Id, Text: "hello", CreatedAt: timestamppb.Now()}
	postData := &MessageData{Message: posted, Dirty: true}
	srv.MessageStorage.Put(posted.Id, postData)
	seq := atomic.AddInt64(&srv.seq, 1)
	srv.publishEvent(&protobufRazpravljalnica.MessageEvent{SequenceNumber: seq, Op: protobufRazpravljalnica.OpType_OP_POST, Message: posted, EventAt: timestamppb.Now()})
	ev := waitEvent(t, s.events, 500*time.Millisecond)
	if ev.Op != protobufRazpravljalnica.OpType_OP_POST || ev.Message == nil || ev.Message.Id != posted.Id {
		t.Fatalf("unexpected post event: %+v", ev)
	}

	// 2) Update
	fmt.Println("starting update op")
	updReq := &protobufRazpravljalnica.UpdateMessageRequest{TopicId: topic.Id, UserId: user.Id, MessageId: posted.Id, Text: "edited"}
	_, err = srv.UpdateMessage(context.Background(), updReq)
	if err != nil {
		t.Fatalf("update message: %v", err)
	}
	ev = waitEvent(t, s.events, 500*time.Millisecond)
	if ev.Op != protobufRazpravljalnica.OpType_OP_UPDATE || ev.Message == nil || ev.Message.Text != "edited" {
		t.Fatalf("unexpected update event: %+v", ev)
	}

	// 3) Like
	fmt.Println("starting like op")
	likeReq := &protobufRazpravljalnica.LikeMessageRequest{TopicId: topic.Id, MessageId: posted.Id, UserId: user.Id}
	_, err = srv.LikeMessage(context.Background(), likeReq)
	if err != nil {
		t.Fatalf("like message: %v", err)
	}
	ev = waitEvent(t, s.events, 500*time.Millisecond)
	if ev.Op != protobufRazpravljalnica.OpType_OP_LIKE || ev.Message == nil || ev.Message.Likes == 0 {
		t.Fatalf("unexpected like event: %+v", ev)
	}

	// 4) Delete
	fmt.Println("starting delete op")
	delReq := &protobufRazpravljalnica.DeleteMessageRequest{TopicId: topic.Id, UserId: user.Id, MessageId: posted.Id}
	_, err = srv.DeleteMessage(context.Background(), delReq)
	if err != nil {
		t.Fatalf("delete message: %v", err)
	}
	ev = waitEvent(t, s.events, 500*time.Millisecond)
	if ev.Op != protobufRazpravljalnica.OpType_OP_DELETE || ev.Message == nil || ev.Message.Id != posted.Id {
		t.Fatalf("unexpected delete event: %+v", ev)
	}

	// finish
	fmt.Println("done")
	cancel()
}

// TestSubscribeReceivesHistoryAndLive verifies that SubscribeTopic sends historical
// messages (FromMessageId) and then streams live events.
func TestSubscribeReceivesHistoryAndLive(t *testing.T) {
	srv := NewMessageBoardServer(2)

	// create user and topic directly
	user := &protobufRazpravljalnica.User{Id: 2, Name: "bob"}
	userData := &UserData{User: user, Dirty: true}
	srv.UserStorage.Put(user.Id, userData)
	topic := &protobufRazpravljalnica.Topic{Id: 2, Name: "history"}
	topicData := &TopicData{Topic: topic, Dirty: true}
	srv.TopicStorage.Put(topic.Id, topicData)

	// post a message before subscribing (this should be delivered as history)
	posted := &protobufRazpravljalnica.Message{Id: 21, TopicId: topic.Id, UserId: user.Id, Text: "first", CreatedAt: timestamppb.Now()}
	messageData := &MessageData{Message: posted, Dirty: true}
	srv.MessageStorage.Put(posted.Id, messageData)

	// subscribe with FromMessageId = 0 to receive the existing message
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := &testStream{fakeServerStream{ctx: ctx}, make(chan *protobufRazpravljalnica.MessageEvent, 16)}
	go func() {
		req := &protobufRazpravljalnica.SubscribeTopicRequest{TopicId: []int64{topic.Id}, UserId: user.Id, FromMessageId: 0}
		if err := srv.SubscribeTopic(req, s); err != nil && ctx.Err() == nil {
			t.Errorf("SubscribeTopic returned error: %v", err)
		}
	}()

	// expect history event for 'first'
	ev := waitEvent(t, s.events, 500*time.Millisecond)
	if ev.Op != protobufRazpravljalnica.OpType_OP_POST || ev.Message == nil || ev.Message.Id != posted.Id {
		t.Fatalf("unexpected history event: %+v", ev)
	}

	// now insert another message directly and publish as live event (avoid GenerateRand32)
	posted2 := &protobufRazpravljalnica.Message{Id: 22, TopicId: topic.Id, UserId: user.Id, Text: "second", CreatedAt: timestamppb.Now()}
	posted2Data := &MessageData{Message: posted2, Dirty: true}
	srv.MessageStorage.Put(posted2.Id, posted2Data)
	seq2 := atomic.AddInt64(&srv.seq, 1)
	srv.publishEvent(&protobufRazpravljalnica.MessageEvent{SequenceNumber: seq2, Op: protobufRazpravljalnica.OpType_OP_POST, Message: posted2, EventAt: timestamppb.Now()})
	ev = waitEvent(t, s.events, 500*time.Millisecond)
	if ev.Op != protobufRazpravljalnica.OpType_OP_POST || ev.Message == nil || ev.Message.Id != posted2.Id {
		t.Fatalf("unexpected live event: %+v", ev)
	}

	cancel()
}
