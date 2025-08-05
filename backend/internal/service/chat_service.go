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

func (s *ChatService) ListChats(ctx context.Context, userID string) ([]*model.Chat, error) {
	return s.repo.GetChats(ctx, userID)
}

func (s *ChatService) GetFullChat(ctx context.Context, chatID string) (*model.FullChat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil { return nil, fmt.Errorf("could not get chat: %w", err) }
	messages, err := s.repo.GetMessages(ctx, chatID)
	if err != nil { return nil, fmt.Errorf("could not get messages: %w", err) }
	return &model.FullChat{Chat: *chat, Messages: messages}, nil
}

// HandleNewMessage now performs the translation from llm.StreamResponse to model.StreamResponse
func (s *ChatService) HandleNewMessage(
	ctx context.Context,
	req *CreateMessageRequest,
	// This channel (going to the API layer) still uses the public model type
	streamChan chan<- model.StreamResponse,
) {
	defer close(streamChan)

	// --- Step 1 & 2: Get/Create Chat and Save User Message (No Changes) ---
	chatID := req.ChatID
	var chat *model.Chat
	var err error
	if chatID == "" {
		chatID = uuid.NewString()
		chat = &model.Chat{ID: chatID, UserID: "default-user", Title: truncate(req.Content, 50), Model: req.Model, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := s.repo.CreateChat(ctx, chat); err != nil {
			log.Printf("Error creating chat: %v", err); streamChan <- model.StreamResponse{Error: "Could not create chat"}; return
		}
	} else {
		chat, err = s.repo.GetChat(ctx, chatID)
		if err != nil {
			log.Printf("Error getting chat %s: %v", chatID, err); streamChan <- model.StreamResponse{Error: "Could not find chat"}; return
		}
		chat.Model = req.Model
	}
	userMessage := &model.Message{ID: uuid.NewString(), Role: "user", Content: req.Content, Timestamp: time.Now()}
	if err := s.repo.AddMessage(ctx, chatID, userMessage); err != nil { log.Printf("Error adding user message to chat %s: %v", chatID, err) }

	// --- Step 3: Prepare LLM Request (No Changes) ---
	history, err := s.repo.GetMessages(ctx, chatID)
	if err != nil { log.Printf("Error getting message history for chat %s: %v", chatID, err) }
	llmMessages := []llm.Message{{Role: "system", Content: req.SystemPrompt}}
	for _, msg := range history { llmMessages = append(llmMessages, llm.Message{Role: msg.Role, Content: msg.Content}) }
	ollamaContext, _ := s.repo.GetOllamaContext(ctx, chatID)
	llmReq := &llm.GenerateRequest{Model: req.Model, Messages: llmMessages, Context: ollamaContext}

	// --- Step 4: THE FIX - Translation Layer ---
	var fullResponse strings.Builder
	var finalContext json.RawMessage
	
	// This channel receives the PRIVATE llm.StreamResponse
	llmStreamChan := make(chan llm.StreamResponse)
	go s.llm.GenerateStream(ctx, llmReq, llmStreamChan)

	for chunk := range llmStreamChan {
		// Translate the private llm.StreamResponse to the public model.StreamResponse
		modelChunk := model.StreamResponse{
			Content: chunk.Content,
			Done:    chunk.Done,
			Context: chunk.Context,
			Error:   chunk.Error,
		}

		// Send the PUBLIC model type to the API layer's channel
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

	// --- Step 5 & 6: Save Assistant Message & Generate Title (No Changes) ---
	assistantMessage := &model.Message{ID: uuid.NewString(), Role: "assistant", Content: fullResponse.String(), Timestamp: time.Now()}
	if err := s.repo.AddMessage(ctx, chatID, assistantMessage); err != nil { log.Printf("Error adding assistant message to chat %s: %v", chatID, err) }
	if finalContext != nil {
		if err := s.repo.SetOllamaContext(ctx, chatID, finalContext); err != nil { log.Printf("Error setting Ollama context for chat %s: %v", chatID, err) }
	}
	chat.UpdatedAt = time.Now()
	if err := s.repo.UpdateChat(ctx, chat); err != nil { log.Printf("Error updating chat timestamp for chat %s: %v", chatID, err) }
	if len(history) <= 1 { go s.generateTitle(context.Background(), chat, req.SupportModel, userMessage.Content, assistantMessage.Content) }
}

func (s *ChatService) generateTitle(ctx context.Context, chat *model.Chat, supportModel, userQuery, assistantResponse string) {
	log.Printf("Generating title for chat %s...", chat.ID)
	prompt := fmt.Sprintf(`Based on the following conversation, create a very short, concise title (4-6 words). Do not use quotes.
		User: "%s"
		Assistant: "%s"
		Title:`, truncate(userQuery, 100), truncate(assistantResponse, 150))
	req := &llm.GenerateRequest{Model: supportModel, Prompt: prompt}
	resp, err := s.llm.Generate(ctx, req)
	if err != nil { log.Printf("Failed to generate title for chat %s: %v", chat.ID, err); return }
	newTitle := strings.Trim(resp.Response, ` "`)
	if newTitle != "" {
		chat.Title = newTitle
		if err := s.repo.UpdateChat(ctx, chat); err != nil {
			log.Printf("Failed to update chat %s with new title: %v", chat.ID, err)
		} else {
			log.Printf("Successfully updated title for chat %s to: '%s'", chat.ID, newTitle)
		}
	}
}

type CreateMessageRequest struct {
	ChatID string `json:"chat_id"`; Content string `json:"content"`; Model string `json:"model"`; SystemPrompt string `json:"system_prompt"`; SupportModel string `json:"support_model"`
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }; runes := []rune(s); if len(runes) <= n { return s }; return string(runes[:n])
}