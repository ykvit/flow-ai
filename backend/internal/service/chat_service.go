package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ChatService struct {
	repo repository.Repository
	llm  llm.LLMProvider
}

func NewChatService(repo repository.Repository, llm llm.LLMProvider) *ChatService {
	return &ChatService{repo: repo, llm: llm}
}

// --- NEW Request Struct with Options ---

// CreateMessageRequest is the structure for a new message request from the client.
// It now includes an optional 'Options' field to override generation parameters.
type CreateMessageRequest struct {
	ChatID       string             `json:"chat_id"`
	Content      string             `json:"content"`
	Model        string             `json:"model"`
	SystemPrompt string             `json:"system_prompt"`
	SupportModel string             `json:"support_model"`
	Options      *llm.RequestOptions `json:"options,omitempty"` // NEW FIELD
}


// --- Chat Management Methods ---

func (s *ChatService) ListChats(ctx context.Context, userID string) ([]*model.Chat, error) {
	return s.repo.GetChats(ctx, userID)
}

func (s *ChatService) GetFullChat(ctx context.Context, chatID string) (*model.FullChat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("could not get chat: %w", err)
	}
	messages, err := s.repo.GetMessages(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("could not get messages: %w", err)
	}
	return &model.FullChat{Chat: *chat, Messages: messages}, nil
}

func (s *ChatService) UpdateChatTitle(ctx context.Context, chatID, newTitle string) error {
	if newTitle == "" {
		return fmt.Errorf("title cannot be empty")
	}
	log.Printf("Manually updating title for chat %s to '%s'", chatID, newTitle)
	return s.repo.UpdateChatTitle(ctx, chatID, newTitle)
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID string) error {
	log.Printf("Deleting chat %s", chatID)
	return s.repo.DeleteChat(ctx, chatID)
}

// --- Message Handling ---

func (s *ChatService) HandleNewMessage(
	ctx context.Context,
	req *CreateMessageRequest,
	streamChan chan<- model.StreamResponse,
) {
	defer close(streamChan)

	isNewChat := req.ChatID == ""
	chatID := req.ChatID
	var chat *model.Chat
	var err error

	if isNewChat {
		chatID = uuid.NewString()
		chat = &model.Chat{ID: chatID, UserID: "default-user", Title: truncate(req.Content, 50), Model: req.Model, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := s.repo.CreateChat(ctx, chat); err != nil {
			log.Printf("Error creating chat: %v", err)
			streamChan <- model.StreamResponse{Error: "Could not create chat"}
			return
		}
	} else {
		chat, err = s.repo.GetChat(ctx, chatID)
		if err != nil {
			log.Printf("Error getting chat %s: %v", chatID, err)
			streamChan <- model.StreamResponse{Error: "Could not find chat"}
			return
		}
		chat.Model = req.Model
	}

	userMessage := &model.Message{ID: uuid.NewString(), Role: "user", Content: req.Content, Timestamp: time.Now()}
	if err := s.repo.AddMessage(ctx, chatID, userMessage); err != nil {
		log.Printf("Error adding user message to chat %s: %v", chatID, err)
	}

	history, err := s.repo.GetMessages(ctx, chatID)
	if err != nil {
		log.Printf("Error getting message history for chat %s: %v", chatID, err)
	}

	// FIX: Prioritize system prompt from options, then from request, then default.
	systemPrompt := req.SystemPrompt
	if req.Options != nil && req.Options.System != nil {
		systemPrompt = *req.Options.System
	}

	llmMessages := []llm.Message{{Role: "system", Content: systemPrompt}}
	for _, msg := range history {
		llmMessages = append(llmMessages, llm.Message{Role: msg.Role, Content: msg.Content})
	}
	ollamaContext, _ := s.repo.GetOllamaContext(ctx, chatID)

	// FIX: Pass the options to the LLM request.
	llmReq := &llm.GenerateRequest{
		Model:    req.Model,
		Messages: llmMessages,
		Context:  ollamaContext,
		Options:  req.Options, // PASSING THE OPTIONS
	}

	var fullResponse strings.Builder
	var finalContext json.RawMessage

	llmStreamChan := make(chan llm.StreamResponse)
	go s.llm.GenerateStream(ctx, llmReq, llmStreamChan)

	for chunk := range llmStreamChan {
		modelChunk := model.StreamResponse{
			Content: chunk.Content,
			Done:    chunk.Done,
			Context: chunk.Context,
			Error:   chunk.Error,
		}
		streamChan <- modelChunk

		if modelChunk.Error != "" {
			log.Printf("Stream error from LLM: %s", modelChunk.Error)
			break
		}

		fullResponse.WriteString(modelChunk.Content)
		if modelChunk.Done {
			finalContext = modelChunk.Context
		}
	}

	assistantMessage := &model.Message{ID: uuid.NewString(), Role: "assistant", Content: fullResponse.String(), Timestamp: time.Now()}
	if err := s.repo.AddMessage(ctx, chatID, assistantMessage); err != nil {
		log.Printf("Error adding assistant message to chat %s: %v", chatID, err)
	}
	if finalContext != nil {
		if err := s.repo.SetOllamaContext(ctx, chatID, finalContext); err != nil {
			log.Printf("Error setting Ollama context for chat %s: %v", chatID, err)
		}
	}

	if isNewChat {
		go s.generateTitle(context.Background(), chat.ID, req.SupportModel, userMessage.Content, assistantMessage.Content)
	}
}

func (s *ChatService) generateTitle(ctx context.Context, chatID, supportModel, userQuery, assistantResponse string) {
	log.Printf("Generating title for chat %s using chat-based approach...", chatID)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are an expert at creating short, concise titles for conversations. Respond with only the title, and nothing else.",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("Based on the following conversation, what would be a good title?\n\n---\nUser: %s\n\nAssistant: %s\n---",
				truncate(userQuery, 150),
				truncate(assistantResponse, 200),
			),
		},
	}

	req := &llm.GenerateRequest{
		Model:    supportModel,
		Messages: messages,
	}

	resp, err := s.llm.Generate(ctx, req)
	if err != nil {
		log.Printf("Failed to generate title for chat %s: %v", chatID, err)
		return
	}

	log.Printf("Raw title response for chat %s: '%s'", chatID, resp.Response)

	newTitle := strings.TrimSpace(resp.Response)
	newTitle = strings.Trim(newTitle, `"'`)

	if newTitle != "" {
		if err := s.repo.UpdateChatTitle(ctx, chatID, newTitle); err != nil {
			log.Printf("Failed to update chat %s with new title: %v", chatID, err)
		} else {
			log.Printf("Successfully updated title for chat %s to: '%s'", chatID, newTitle)
		}
	} else {
		log.Printf("Title for chat %s was not updated because the generated title was empty after cleaning.", chatID)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}