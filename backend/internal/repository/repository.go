package repository

import (
	"context"
	"database/sql"
	"flow-ai/backend/internal/model"
)

// Repository defines the interface for data storage operations.
// This interface makes it easy to switch database implementations.
type Repository interface {
	// Transaction control
	BeginTx(ctx context.Context) (*sql.Tx, error)

	CreateChat(ctx context.Context, chat *model.Chat) error
	GetChat(ctx context.Context, chatID string) (*model.Chat, error)
	GetChats(ctx context.Context, userID string) ([]*model.Chat, error)
	UpdateChatTitle(ctx context.Context, chatID, newTitle string) error
	DeleteChat(ctx context.Context, chatID string) error

	// Message operations
	AddMessage(ctx context.Context, message *model.Message, chatID string) error
	GetMessageByID(ctx context.Context, messageID string) (*model.Message, error) // NEW
	GetActiveMessagesByChatID(ctx context.Context, chatID string) ([]model.Message, error)
	GetLastActiveMessage(ctx context.Context, chatID string) (*model.Message, error)
	UpdateMessageContext(ctx context.Context, messageID string, ollamaContext []byte) error

	// Transactional operations
	AddMessageTx(ctx context.Context, tx *sql.Tx, message *model.Message, chatID string) error // NEW
	DeactivateBranchTx(ctx context.Context, tx *sql.Tx, messageID string) error                  // NEW
	UpdateChatTimestampTx(ctx context.Context, tx *sql.Tx, chatID string) error                   // NEW
}