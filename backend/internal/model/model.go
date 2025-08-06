package model

import (
	"encoding/json"
	"time"
)

// Chat stores metadata about a conversation.
type Chat struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Model     string    `json:"model"`
}

// Message stores a single message in a chat.
// It now includes an optional Metadata field for storing generation stats.
type Message struct {
	ID        string          `json:"id"`
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

// FullChat includes the chat metadata and all its messages.
type FullChat struct {
	Chat
	Messages []Message `json:"messages"`
}

// StreamResponse is the structure for a single chunk in a streaming response.
// This is now a shared model.
type StreamResponse struct {
	Content string          `json:"content"`
	Done    bool            `json:"done"`
	Context json.RawMessage `json:"context,omitempty"`
	Error   string          `json:"error,omitempty"`
}