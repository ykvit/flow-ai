package repository

import (
	"context"
	"flow-ai/backend/internal/model"
)

// Repository defines the interface for data storage operations.
// This interface makes it easy to switch database implementations.
type Repository interface {
	CreateChat(ctx context.Context, chat *model.Chat) error
	GetChat(ctx context.Context, chatID string) (*model.Chat, error)
	GetChats(ctx context.Context, userID string) ([]*model.Chat, error)
	UpdateChatTitle(ctx context.Context, chatID, newTitle string) error
	DeleteChat(ctx context.Context, chatID string) error

	AddMessage(ctx context.Context, message *model.Message, chatID string) error
	GetActiveMessagesByChatID(ctx context.Context, chatID string) ([]model.Message, error)
	
	GetLastActiveMessage(ctx context.Context, chatID string) (*model.Message, error)
	UpdateMessageContext(ctx context.Context, messageID string, ollamaContext []byte) error
}