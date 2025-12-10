package accountutils

import "time"

type User struct {
	Id   int64
	Name string
}

func NewUser(id int64, name string) *User {
	return &User{Id: id, Name: name}
}

type Topic struct {
	Id   int64
	Name string
}

type Message struct {
	Id        int64
	TopicId   int64
	UserId    int64
	Text      string
	CreatedAt time.Time
	Likes     int32
}

type Like struct {
	MessageId int64
	TopicId   int64
	UserId    int64
}
