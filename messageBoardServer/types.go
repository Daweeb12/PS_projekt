package messageboardserver

import (
	pbRaz "PS_projekt/api/grpc/protobufRazpravljalnica"
)

type Data interface {
	IsDirty() bool
}
type MessageData struct {
	*pbRaz.Message
	Dirty bool
}

type UserData struct {
	*pbRaz.User
	Dirty bool
}

type TopicData struct {
	*pbRaz.Topic
	Dirty bool
}

func (msgData *MessageData) IsDirty() bool {
	return msgData.Dirty
}

func (topicData *TopicData) IsDirty() bool {
	return topicData.Dirty
}

func (userData *UserData) IsDirty() bool {
	return userData.Dirty
}
