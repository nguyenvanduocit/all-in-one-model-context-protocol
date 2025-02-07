package services

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"
)

// NewGChatService creates and initializes a new Google Chat service
func NewGChatService() (*chat.Service, error) {
	ctx := context.Background()

	// Initialize Google Chat API service with default credentials and required scopes
	srv, err := chat.NewService(ctx, option.WithScopes(
		chat.ChatAdminSpacesScope,
		chat.ChatSpacesScope,
		chat.ChatAdminMembershipsScope,
		chat.ChatAdminMembershipsReadonlyScope,
		chat.ChatAppMembershipsScope,
		chat.ChatAppSpacesScope,
		chat.ChatAppSpacesCreateScope,
		chat.ChatMessagesScope,
		chat.ChatMessagesCreateScope,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create chat service: %v", err)
	}

	return srv, nil
}

var DefaultGChatService = sync.OnceValue[*chat.Service](func() *chat.Service {
	srv, err := NewGChatService()
	if err != nil {
		panic(fmt.Sprintf("failed to create chat service: %v", err))
	}
	return srv
})