package model

import (
	"encoding/json"
	"time"
)

// Chat stores metadata about a conversation.
type Chat struct {
	ID        string    `json:"id" example:"4b3b5a34-571f-47e3-abd1-a7dbee9d92fe"`
	Title     string    `json:"title" example:"History of the Roman Empire"`
	CreatedAt time.Time `json:"created_at" example:"2025-09-08T14:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-09-08T14:05:00Z"`
	Model     string    `json:"model" example:"qwen:0.5b"`
}

// Message stores a single message in a chat.
type Message struct {
	ID        string          `json:"id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	ParentID  *string         `json:"parent_id,omitempty" example:"f0e9d8c7-b6a5-4321-fedc-ba9876543210"`
	Role      string          `json:"role" example:"assistant"`
	Content   string          `json:"content" example:"The Roman Empire fell in 476 AD."`
	Model     *string         `json:"model,omitempty" example:"qwen:0.5b"`
	Timestamp time.Time       `json:"timestamp" example:"2025-09-08T14:05:00Z"`
	Metadata  json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	Context   json.RawMessage `json:"-"`
}

// FullChat includes the chat metadata and all its messages.
type FullChat struct {
	Chat
	Messages []Message `json:"messages"`
}

// StreamResponse is the structure for a single chunk in a streaming response.
type StreamResponse struct {
	Content string          `json:"content" example:"Hello"`
	Done    bool            `json:"done" example:"false"`
	Context json.RawMessage `json:"context,omitempty" swaggertype:"object"`
	Error   string          `json:"error,omitempty"`
}
