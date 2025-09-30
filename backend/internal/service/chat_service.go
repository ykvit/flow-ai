package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/repository"

	"github.com/google/uuid"
)

// ChatService handles the core business logic for chat operations.
type ChatService struct {
	repo            repository.Repository
	llm             llm.LLMProvider
	settingsService *SettingsService
}

// CreateMessageRequest is the DTO for creating a new message.
type CreateMessageRequest struct {
	ChatID       string              `json:"chat_id"`
	Content      string              `json:"content"`
	Model        string              `json:"model"`
	SystemPrompt string              `json:"system_prompt"`
	SupportModel string              `json:"support_model"`
	Options      *llm.RequestOptions `json:"options,omitempty"`
}

// RegenerateMessageRequest is the DTO for regenerating a message.
type RegenerateMessageRequest struct {
	ChatID       string              `json:"chat_id"` // Included for context
	Model        string              `json:"model"`
	SystemPrompt string              `json:"system_prompt"`
	Options      *llm.RequestOptions `json:"options,omitempty"`
}

func NewChatService(repo repository.Repository, llm llm.LLMProvider, settingsService *SettingsService) *ChatService {
	return &ChatService{repo: repo, llm: llm, settingsService: settingsService}
}

func (s *ChatService) UpdateChatTitle(ctx context.Context, chatID, newTitle string) error {
	if newTitle == "" {
		return fmt.Errorf("title cannot be empty")
	}
	slog.Info("Manually updating title", "chat_id", chatID, "new_title", newTitle)
	return s.repo.UpdateChatTitle(ctx, chatID, newTitle)
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID string) error {
	slog.Info("Deleting chat", "chat_id", chatID)
	return s.repo.DeleteChat(ctx, chatID)
}

func (s *ChatService) ListChats(ctx context.Context, userID string) ([]*model.Chat, error) {
	return s.repo.GetChats(ctx, userID)
}

func (s *ChatService) GetFullChat(ctx context.Context, chatID string) (*model.FullChat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("could not get chat: %w", err)
	}

	messages, err := s.repo.GetActiveMessagesByChatID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("could not get messages: %w", err)
	}

	return &model.FullChat{Chat: *chat, Messages: messages}, nil
}

func (s *ChatService) resolveModels(ctx context.Context, req *CreateMessageRequest, currentSettings *Settings) (mainModel, supportModel, systemPrompt string, err error) {
	mainModel = req.Model
	if mainModel == "" {
		mainModel = currentSettings.MainModel
	} else {
		availableModels, err := s.llm.ListModels(ctx)
		if err != nil {
			slog.Warn("Could not list models to validate request model", "model", mainModel, "error", err)
		} else {
			modelNames := make([]string, len(availableModels.Models))
			for i, m := range availableModels.Models {
				modelNames[i] = m.Name
			}
			if !slices.Contains(modelNames, mainModel) {
				return "", "", "", fmt.Errorf("model '%s' specified in request is not available", mainModel)
			}
		}
	}

	if mainModel == "" {
		return "", "", "", errors.New("no main model is configured or available, please pull a model first")
	}

	supportModel = req.SupportModel
	if supportModel == "" {
		supportModel = currentSettings.SupportModel
	}

	systemPrompt = req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = currentSettings.SystemPrompt
	}
	if req.Options != nil && req.Options.System != nil {
		systemPrompt = *req.Options.System
	}

	return mainModel, supportModel, systemPrompt, nil
}

func (s *ChatService) HandleNewMessage(
	ctx context.Context,
	req *CreateMessageRequest,
	streamChan chan<- model.StreamResponse,
) {
	defer close(streamChan)

	currentSettings, err := s.settingsService.Get(ctx)
	if err != nil {
		slog.Error("Could not get settings for new message", "error", err)
		streamChan <- model.StreamResponse{Error: "Could not load application settings"}
		return
	}

	modelToUse, supportModelToUse, systemPromptToUse, err := s.resolveModels(ctx, req, currentSettings)
	if err != nil {
		streamChan <- model.StreamResponse{Error: err.Error()}
		return
	}

	isNewChat := req.ChatID == ""
	chatID := req.ChatID

	if isNewChat {
		chatID = uuid.NewString()
		chat := &model.Chat{ID: chatID, UserID: "default-user", Title: truncate(req.Content, 50), Model: modelToUse, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		if err := s.repo.CreateChat(ctx, chat); err != nil {
			slog.Error("Error creating chat", "error", err)
			streamChan <- model.StreamResponse{Error: "Could not create chat"}
			return
		}
	}

	lastMessage, err := s.repo.GetLastActiveMessage(ctx, chatID)
	if err != nil {
		slog.Warn("Error getting last message for chat", "chat_id", chatID, "error", err)
	}

	var parentID *string
	var ollamaContext json.RawMessage
	if lastMessage != nil {
		parentID = &lastMessage.ID
		ollamaContext = lastMessage.Context
	}

	userMessage := &model.Message{ID: uuid.NewString(), ParentID: parentID, Role: "user", Content: req.Content, Timestamp: time.Now().UTC()}
	if err := s.repo.AddMessage(ctx, userMessage, chatID); err != nil {
		slog.Error("Error adding user message", "chat_id", chatID, "error", err)
	}

	history, err := s.repo.GetActiveMessagesByChatID(ctx, chatID)
	if err != nil {
		slog.Warn("Error getting message history for chat", "chat_id", chatID, "error", err)
	}

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
	go func() {
		if err := s.llm.GenerateStream(ctx, llmReq, llmStreamChan); err != nil {
			slog.Error("LLM stream generation failed", "error", err)
		}
	}()

	for chunk := range llmStreamChan {
		modelChunk := model.StreamResponse{Content: chunk.Content, Done: chunk.Done, Context: chunk.Context, Error: chunk.Error}
		streamChan <- modelChunk

		if chunk.Error != "" {
			break
		}
		fullResponse.WriteString(chunk.Content)
		if chunk.Done {
			finalContext = chunk.Context
			finalStats = chunk.Stats
		}
	}
	slog.Debug("Finished streaming response from LLM.")

	var metadata json.RawMessage
	if finalStats != nil {
		metadata, _ = json.Marshal(finalStats)
	}

	assistantMessage := &model.Message{
		ID:        uuid.NewString(),
		ParentID:  &userMessage.ID,
		Role:      "assistant",
		Content:   fullResponse.String(),
		Model:     &modelToUse,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}

	if err := s.repo.AddMessage(ctx, assistantMessage, chatID); err != nil {
		slog.Error("Failed to save assistant message", "chat_id", chatID, "error", err)
		return
	}

	if finalContext != nil {
		if err := s.repo.UpdateMessageContext(ctx, assistantMessage.ID, finalContext); err != nil {
			slog.Warn("Error setting Ollama context for message", "message_id", assistantMessage.ID, "error", err)
		}
	}

	if isNewChat {
		go s.generateTitle(context.Background(), chatID, supportModelToUse, userMessage.Content, assistantMessage.Content)
	}
}

func (s *ChatService) RegenerateMessage(
	ctx context.Context,
	chatID string,
	originalAssistantMessageID string,
	req *RegenerateMessageRequest,
	streamChan chan<- model.StreamResponse,
) {
	defer close(streamChan)

	currentSettings, err := s.settingsService.Get(ctx)
	if err != nil {
		slog.Error("Could not get settings for regeneration", "error", err)
		streamChan <- model.StreamResponse{Error: "Could not load application settings"}
		return
	}

	modelToUse := req.Model
	if modelToUse == "" {
		modelToUse = currentSettings.MainModel
	}

	systemPromptToUse := req.SystemPrompt
	if systemPromptToUse == "" {
		systemPromptToUse = currentSettings.SystemPrompt
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		slog.Error("Regenerate failed to begin transaction", "error", err)
		streamChan <- model.StreamResponse{Error: "Database error"}
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			slog.Error("Failed to rollback regeneration transaction", "error", err)
		}
	}()

	originalMsg, err := s.repo.GetMessageByID(ctx, originalAssistantMessageID)
	if err != nil || originalMsg.Role != "assistant" || originalMsg.ParentID == nil {
		streamChan <- model.StreamResponse{Error: "Original message not found or invalid"}
		return
	}

	if err := s.repo.DeactivateBranchTx(ctx, tx, originalAssistantMessageID); err != nil {
		slog.Error("Regenerate failed to deactivate branch", "error", err)
		streamChan <- model.StreamResponse{Error: "Database error during regeneration"}
		return
	}

	history, err := s.repo.GetActiveMessagesByChatIDTx(ctx, tx, chatID)
	if err != nil {
		slog.Error("Regenerate failed to get history", "chat_id", chatID, "error", err)
		streamChan <- model.StreamResponse{Error: "Could not retrieve message history"}
		return
	}

	llmMessages := []llm.Message{{Role: "system", Content: systemPromptToUse}}
	for _, msg := range history {
		llmMessages = append(llmMessages, llm.Message{Role: msg.Role, Content: msg.Content})
	}

	llmReq := &llm.GenerateRequest{
		Model:    modelToUse,
		Messages: llmMessages,
		Options:  req.Options,
	}

	slog.Debug("Ollama regeneration request payload", "payload", llmReq)

	var fullResponse strings.Builder
	var finalContext json.RawMessage
	var finalStats *llm.GenerationStats
	llmStreamChan := make(chan llm.StreamResponse)
	go func() {
		if err := s.llm.GenerateStream(ctx, llmReq, llmStreamChan); err != nil {
			slog.Error("LLM stream regeneration failed", "error", err)
		}
	}()

	for chunk := range llmStreamChan {
		streamChan <- model.StreamResponse{Content: chunk.Content, Done: chunk.Done, Error: chunk.Error}
		if chunk.Error != "" {
			return
		}
		fullResponse.WriteString(chunk.Content)
		if chunk.Done {
			finalContext = chunk.Context
			finalStats = chunk.Stats
		}
	}
	slog.Debug("Finished streaming regenerated response from LLM.")

	var metadata json.RawMessage
	if finalStats != nil {
		metadata, _ = json.Marshal(finalStats)
	}

	newAssistantMessage := &model.Message{
		ID:        uuid.NewString(),
		ParentID:  originalMsg.ParentID,
		Role:      "assistant",
		Content:   fullResponse.String(),
		Model:     &modelToUse,
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
	}

	if err := s.repo.AddMessageTx(ctx, tx, newAssistantMessage, chatID); err != nil {
		slog.Error("Failed to save regenerated message", "chat_id", chatID, "error", err)
		return
	}

	if err := s.repo.UpdateChatTimestampTx(ctx, tx, chatID); err != nil {
		slog.Error("Failed to update chat timestamp after regeneration", "chat_id", chatID, "error", err)
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit regeneration transaction", "error", err)
		return
	}

	if finalContext != nil {
		if err := s.repo.UpdateMessageContext(ctx, newAssistantMessage.ID, finalContext); err != nil {
			slog.Warn("Error setting Ollama context for new message", "message_id", newAssistantMessage.ID, "error", err)
		}
	}
}

func (s *ChatService) generateTitle(ctx context.Context, chatID, supportModel, userQuery, assistantResponse string) {
	slog.Info("Generating title", "chat_id", chatID)

	prompt := fmt.Sprintf(
		`Analyze the following conversation and generate a short, concise title (5 words max).
		Respond with ONLY a JSON object in the format {"title": "your generated title"}. Do not add any other text or explanations.
		
		CONVERSATION:
		User: %s
		Assistant: %s`,
		truncate(userQuery, 150),
		truncate(assistantResponse, 200),
	)

	messages := []llm.Message{{Role: "user", Content: prompt}}
	req := &llm.GenerateRequest{Model: supportModel, Messages: messages}
	resp, err := s.llm.Generate(ctx, req)
	if err != nil {
		slog.Warn("Failed to generate title", "chat_id", chatID, "error", err)
		return
	}

	slog.Debug("Raw title response from LLM", "chat_id", chatID, "response", resp.Response)

	jsonString := extractJSON(resp.Response)
	type TitleResponse struct {
		Title string `json:"title"`
	}
	var titleResp TitleResponse
	var newTitle string

	if jsonString != "" {
		if err := json.Unmarshal([]byte(jsonString), &titleResp); err != nil {
			slog.Warn("Found JSON-like string but failed to parse for title", "chat_id", chatID, "error", err)
			newTitle = cleanRawTitle(resp.Response)
		} else {
			newTitle = titleResp.Title
		}
	} else {
		slog.Warn("No JSON found in title response, cleaning raw string", "chat_id", chatID)
		newTitle = cleanRawTitle(resp.Response)
	}

	if trimmedTitle := strings.TrimSpace(newTitle); trimmedTitle != "" {
		if err := s.repo.UpdateChatTitle(ctx, chatID, trimmedTitle); err != nil {
			slog.Warn("Failed to update chat with new title", "chat_id", chatID, "error", err)
		} else {
			slog.Info("Successfully updated title", "chat_id", chatID, "title", trimmedTitle)
		}
	}
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(s, "}")
	if end == -1 || end < start {
		return ""
	}
	return s[start : end+1]
}

func cleanRawTitle(s string) string {
	s = strings.Split(s, "<think>")[0]
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
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
