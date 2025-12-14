package tests

import (
	protobufRazpravljalnica "PS_projekt/api/grpc/protobufRazpravljalnica"
	messageboardserver "PS_projekt/messageBoardServer"
	"fmt"
	"testing"
)

func TestUserCreation(t *testing.T) {
	server := messageboardserver.NewMessageBoardServer()
	names := []string{" david   ", "   ", "		", " stefan"}
	for _, name := range names {
		userReq := protobufRazpravljalnica.CreateUserRequest{Name: name}
		if user, err := server.CreateUser(t.Context(), &userReq); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(user)
		}
	}

}
