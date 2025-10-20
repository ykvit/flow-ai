package interfaces

import (
	"context"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/service"
)

// This file defines the interfaces for our core services.
// Depending on these interfaces, instead of concrete implementations, allows for
// decoupling (e.g., API layer from Service layer) and easier testing via mocking.

// ChatService defines the contract for chat-related business logic.
type ChatService interface {
	UpdateChatTitle(ctx context.Context, chatID, newTitle string) error
	DeleteChat(ctx context.Context, chatID string) error
	ListChats(ctx context.Context) ([]*model.Chat, error)
	GetFullChat(ctx context.Context, chatID string) (*model.FullChat, error)
	HandleNewMessage(ctx context.Context, req *service.CreateMessageRequest, streamChan chan<- model.StreamResponse)
	RegenerateMessage(ctx context.Context, chatID string, originalAssistantMessageID string, req *service.RegenerateMessageRequest, streamChan chan<- model.StreamResponse)
}

// ModelService defines the contract for model management logic.
type ModelService interface {
	List(ctx context.Context) (*llm.ListModelsResponse, error)
	Pull(ctx context.Context, req *llm.PullModelRequest, ch chan<- llm.PullStatus) error
	Delete(ctx context.Context, req *llm.DeleteModelRequest) error
	Show(ctx context.Context, req *llm.ShowModelRequest) (*llm.ModelInfo, error)
}

// SettingsService defines the contract for managing application settings.
type SettingsService interface {
	InitAndGet(ctx context.Context, defaultSystemPrompt string) (*service.Settings, error)
	Get(ctx context.Context) (*service.Settings, error)
	Save(ctx context.Context, settings *service.Settings) error
}
