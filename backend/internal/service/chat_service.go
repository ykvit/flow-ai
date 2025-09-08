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
	settingsService *SettingsService 
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
	
	// Get only the active messages for the main chat view.
	messages, err := s.repo.GetActiveMessagesByChatID(ctx, chatID)
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
		// Validate that the requested model exists.
		availableModels, err := s.llm.ListModels(ctx)
		if err != nil {
			log.Printf("WARN: Could not list models to validate request model '%s': %v", modelToUse, err)
		} else {
			modelNames := make([]string, len(availableModels.Models))
			for i, m := range availableModels.Models { modelNames[i] = m.Name }
			if !slices.Contains(modelNames, modelToUse) {
				errorMsg := fmt.Sprintf("Model '%s' specified in request is not available", modelToUse)
				streamChan <- model.StreamResponse{Error: errorMsg}
				return
			}
		}
	}
	
	supportModelToUse := req.SupportModel
	if supportModelToUse == "" { supportModelToUse = currentSettings.SupportModel }
	systemPromptToUse := req.SystemPrompt
	if systemPromptToUse == "" { systemPromptToUse = currentSettings.SystemPrompt }
    if req.Options != nil && req.Options.System != nil { systemPromptToUse = *req.Options.System }

	isNewChat := req.ChatID == ""
	chatID := req.ChatID
	
	// Create chat if it's a new conversation.
	if isNewChat {
		chatID = uuid.NewString()
		chat := &model.Chat{ID: chatID, UserID: "default-user", Title: truncate(req.Content, 50), Model: modelToUse, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		if err := s.repo.CreateChat(ctx, chat); err != nil {
			log.Printf("Error creating chat: %v", err)
			streamChan <- model.StreamResponse{Error: "Could not create chat"}
			return
		}
	}

	// Get the context and parent ID from the last message in the active thread.
	lastMessage, err := s.repo.GetLastActiveMessage(ctx, chatID)
	if err != nil { log.Printf("Error getting last message for chat %s: %v", chatID, err) }
	
	var parentID *string
	var ollamaContext json.RawMessage
	if lastMessage != nil {
		parentID = &lastMessage.ID
		ollamaContext = lastMessage.Context
	}

	// Save the user's message.
	userMessage := &model.Message{ID: uuid.NewString(), ParentID: parentID, Role: "user", Content: req.Content, Timestamp: time.Now().UTC()}
	if err := s.repo.AddMessage(ctx, userMessage, chatID); err != nil { 
		log.Printf("Error adding user message to chat %s: %v", chatID, err) 
	}

	// Prepare messages for the LLM, including history.
	history, err := s.repo.GetActiveMessagesByChatID(ctx, chatID)
	if err != nil { log.Printf("Error getting message history for chat %s: %v", chatID, err) }
	llmMessages := []llm.Message{{Role: "system", Content: systemPromptToUse}}
	for _, msg := range history {
		llmMessages = append(llmMessages, llm.Message{Role: msg.Role, Content: msg.Content})
	}
	
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
		// FIX: Manually create the model.StreamResponse instead of direct conversion.
		modelChunk := model.StreamResponse{
			Content: chunk.Content,
			Done:    chunk.Done,
			Context: chunk.Context,
			Error:   chunk.Error,
		}
		streamChan <- modelChunk
		
		if chunk.Error != "" { break }
		fullResponse.WriteString(chunk.Content)
		if chunk.Done {
			finalContext = chunk.Context
			finalStats = chunk.Stats
		}
	}

	// Save the assistant's message.
	var metadata json.RawMessage
	if finalStats != nil {
		metadata, _ = json.Marshal(finalStats)
	}
	assistantMessage := &model.Message{
		ID:        uuid.NewString(),
		ParentID:  &userMessage.ID, // Parent is the user message.
		Role:      "assistant",
		Content:   fullResponse.String(),
		Model:     &modelToUse,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}
	if err := s.repo.AddMessage(ctx, assistantMessage, chatID); err != nil {
		log.Printf("CRITICAL: Failed to save assistant message to chat %s: %v", chatID, err)
		return
	}
	
	// Update the assistant message with the final context from the LLM.
	if finalContext != nil {
		if err := s.repo.UpdateMessageContext(ctx, assistantMessage.ID, finalContext); err != nil {
			log.Printf("Error setting Ollama context for message %s: %v", assistantMessage.ID, err)
		}
	}

	if isNewChat {
		go s.generateTitle(context.Background(), chatID, supportModelToUse, userMessage.Content, assistantMessage.Content)
	}
}

func (s *ChatService) generateTitle(ctx context.Context, chatID, supportModel, userQuery, assistantResponse string) {
	log.Printf("Generating title for chat %s...", chatID)
	messages := []llm.Message{
		{Role: "system", Content: "You create short, concise titles for conversations. Respond with only the title."},
		{Role: "user", Content: fmt.Sprintf("Conversation:\nUser: %s\nAssistant: %s\n\nTitle:", truncate(userQuery, 150), truncate(assistantResponse, 200))},
	}
	req := &llm.GenerateRequest{ Model: supportModel, Messages: messages }
	resp, err := s.llm.Generate(ctx, req)
	if err != nil {
		log.Printf("Failed to generate title for chat %s: %v", chatID, err)
		return
	}
    newTitle := strings.Trim(strings.TrimSpace(resp.Response), `"'`)
	if newTitle != "" {
		if err := s.repo.UpdateChatTitle(ctx, chatID, newTitle); err != nil {
			log.Printf("Failed to update chat %s with new title: %v", chatID, err)
		} else {
			log.Printf("Successfully updated title for chat %s to: '%s'", chatID, newTitle)
		}
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