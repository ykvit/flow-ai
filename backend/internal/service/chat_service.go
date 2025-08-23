package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/repository"
)

// ChatService handles the core business logic for chat operations.
type ChatService struct {
	repo            repository.Repository
	llm             llm.LLMProvider
	settingsService *SettingsService // Dependency for dynamic settings
}

// NewChatService creates a new ChatService.
func NewChatService(repo repository.Repository, llm llm.LLMProvider, settingsService *SettingsService) *ChatService {
	return &ChatService{repo: repo, llm: llm, settingsService: settingsService}
}

// UpdateChatTitle handles the logic for manually updating a chat's title.
func (s *ChatService) UpdateChatTitle(ctx context.Context, chatID, newTitle string) error {
    if newTitle == "" {
        return fmt.Errorf("title cannot be empty")
    }
	log.Printf("Manually updating title for chat %s to '%s'", chatID, newTitle)
	return s.repo.UpdateChatTitle(ctx, chatID, newTitle)
}

// DeleteChat handles the logic for deleting a chat and all its related data.
func (s *ChatService) DeleteChat(ctx context.Context, chatID string) error {
	log.Printf("Deleting chat %s", chatID)
	return s.repo.DeleteChat(ctx, chatID)
}

// ListChats retrieves all chats for a given user.
func (s *ChatService) ListChats(ctx context.Context, userID string) ([]*model.Chat, error) {
	return s.repo.GetChats(ctx, userID)
}

// GetFullChat retrieves the full history of a single chat.
func (s *ChatService) GetFullChat(ctx context.Context, chatID string) (*model.FullChat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil { return nil, fmt.Errorf("could not get chat: %w", err) }
	messages, err := s.repo.GetMessages(ctx, chatID)
	if err != nil { return nil, fmt.Errorf("could not get messages: %w", err) }
	return &model.FullChat{Chat: *chat, Messages: messages}, nil
}

// HandleNewMessage is now updated with model validation.
func (s *ChatService) HandleNewMessage(
	ctx context.Context,
	req *CreateMessageRequest,
	streamChan chan<- model.StreamResponse,
) {
	defer close(streamChan)

	currentSettings, err := s.settingsService.Get(ctx)
	if err != nil {
		log.Printf("CRITICAL: Could not get settings for new message: %v", err)
		streamChan <- model.StreamResponse{Error: "Could not load application settings"}
		return
	}

	modelToUse := req.Model
	if modelToUse == "" {
		modelToUse = currentSettings.MainModel
	} else {
		availableModels, err := s.llm.ListModels(ctx)
		if err != nil {
			log.Printf("WARN: Could not list models to validate request model '%s': %v", modelToUse, err)
		} else {
			modelNames := make([]string, len(availableModels.Models))
			for i, m := range availableModels.Models {
				modelNames[i] = m.Name
			}
			if !slices.Contains(modelNames, modelToUse) {
				errorMsg := fmt.Sprintf("Model '%s' specified in request is not available", modelToUse)
				log.Printf("ERROR: %s", errorMsg)
				streamChan <- model.StreamResponse{Error: errorMsg}
				return
			}
		}
	}
	
	supportModelToUse := req.SupportModel
	if supportModelToUse == "" {
		supportModelToUse = currentSettings.SupportModel
	}
	systemPromptToUse := req.SystemPrompt
	if systemPromptToUse == "" {
		systemPromptToUse = currentSettings.SystemPrompt
	}
    if req.Options != nil && req.Options.System != nil {
		systemPromptToUse = *req.Options.System
	}
	isNewChat := req.ChatID == ""
	chatID := req.ChatID
	var chat *model.Chat
	if isNewChat {
		chatID = uuid.NewString()
		chat = &model.Chat{ID: chatID, UserID: "default-user", Title: truncate(req.Content, 50), Model: modelToUse, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := s.repo.CreateChat(ctx, chat); err != nil {
			log.Printf("Error creating chat: %v", err); streamChan <- model.StreamResponse{Error: "Could not create chat"}; return
		}
	} else {
		chat, err = s.repo.GetChat(ctx, chatID)
		if err != nil {
			log.Printf("Error getting chat %s: %v", chatID, err); streamChan <- model.StreamResponse{Error: "Could not find chat"}; return
		}
	}
	userMessage := &model.Message{ID: uuid.NewString(), Role: "user", Content: req.Content, Timestamp: time.Now()}
	if err := s.repo.AddMessage(ctx, chatID, userMessage); err != nil { log.Printf("Error adding user message to chat %s: %v", chatID, err) }
	history, err := s.repo.GetMessages(ctx, chatID)
	if err != nil { log.Printf("Error getting message history for chat %s: %v", chatID, err) }
	llmMessages := []llm.Message{{Role: "system", Content: systemPromptToUse}}
	for _, msg := range history {
		llmMessages = append(llmMessages, llm.Message{Role: msg.Role, Content: msg.Content})
	}
	ollamaContext, _ := s.repo.GetOllamaContext(ctx, chatID)
	llmReq := &llm.GenerateRequest{
		Model:    modelToUse,
		Messages: llmMessages,
		Context:  ollamaContext,
		Options:  req.Options,
	}
	var fullResponse strings.Builder
	var finalContext json.RawMessage
	var finalStats *llm.GenerationStats
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
			finalContext = chunk.Context
			finalStats = chunk.Stats
		}
	}
	var metadata json.RawMessage
	if finalStats != nil {
		metadataBytes, err := json.Marshal(finalStats)
		if err != nil {
			log.Printf("Error marshaling assistant message metadata: %v", err)
		} else {
			metadata = metadataBytes
		}
	}
	assistantMessage := &model.Message{
		ID:        uuid.NewString(),
		Role:      "assistant",
		Content:   fullResponse.String(),
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	if err := s.repo.AddMessage(ctx, chatID, assistantMessage); err != nil {
		log.Printf("CRITICAL: Failed to save assistant message to chat %s: %v", chatID, err)
		return
	}
	log.Printf("Successfully saved assistant message %s to chat %s.", assistantMessage.ID, chatID)
	if finalContext != nil {
		if err := s.repo.SetOllamaContext(ctx, chatID, finalContext); err != nil {
			log.Printf("Error setting Ollama context for chat %s: %v", chatID, err)
		}
	}
	if isNewChat {
		go s.generateTitle(context.Background(), chatID, supportModelToUse, userMessage.Content, assistantMessage.Content)
	}
}

func (s *ChatService) generateTitle(ctx context.Context, chatID, supportModel, userQuery, assistantResponse string) {
	log.Printf("Generating title for chat %s using chat-based approach...", chatID)
	messages := []llm.Message{
		{
			Role: "system",
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

type CreateMessageRequest struct {
	ChatID       string             `json:"chat_id"`
	Content      string             `json:"content"`
	Model        string             `json:"model"`
	SystemPrompt string             `json:"system_prompt"`
	SupportModel string             `json:"support_model"`
	Options      *llm.RequestOptions `json:"options,omitempty"`
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }; runes := []rune(s); if len(runes) <= n { return s }; return string(runes[:n])
}