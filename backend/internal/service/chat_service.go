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

	app_errors "flow-ai/backend/internal/errors"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/repository"

	"github.com/google/uuid"
)

// ChatService encapsulates the core business logic for chat operations.
// It orchestrates interactions between the repository, LLM provider, and other services.
type ChatService struct {
	repo            repository.Repository
	llm             llm.LLMProvider
	settingsService *SettingsService
}

// CreateMessageRequest is the DTO for creating a new message. Includes validation tags.
type CreateMessageRequest struct {
	ChatID       string              `json:"chat_id,omitempty" example:"4b3b5a34-571f-47e3-abd1-a7dbee9d92fe"`
	Content      string              `json:"content" validate:"required,min=1" example:"What is the difference between SQL and NoSQL databases?"`
	Model        string              `json:"model,omitempty" example:"qwen3:8b"`
	SystemPrompt string              `json:"system_prompt,omitempty"`
	SupportModel string              `json:"support_model,omitempty"`
	Options      *llm.RequestOptions `json:"options,omitempty"`
}

// RegenerateMessageRequest is the DTO for regenerating a message.
type RegenerateMessageRequest struct {
	ChatID       string `json:"chat_id,omitempty"` // Included for client-side context.
	Model        string `json:"model,omitempty" example:"mistral:7b"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	// Allows overriding generation parameters, e.g., for a more creative response.
	Options *llm.RequestOptions `json:"options,omitempty"`
}

// NewChatService creates a new instance of ChatService.
func NewChatService(repo repository.Repository, llm llm.LLMProvider, settingsService *SettingsService) *ChatService {
	return &ChatService{repo: repo, llm: llm, settingsService: settingsService}
}

func (s *ChatService) UpdateChatTitle(ctx context.Context, chatID, newTitle string) error {
	slog.Info("Manually updating title", "chat_id", chatID, "new_title", newTitle)
	err := s.repo.UpdateChatTitle(ctx, chatID, newTitle)
	// Translate the repository-level error to a domain-level error.
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: chat with id %s", app_errors.ErrNotFound, chatID)
	}
	return err
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID string) error {
	slog.Info("Deleting chat", "chat_id", chatID)
	err := s.repo.DeleteChat(ctx, chatID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: chat with id %s", app_errors.ErrNotFound, chatID)
	}
	return err
}

func (s *ChatService) ListChats(ctx context.Context, userID string) ([]*model.Chat, error) {
	return s.repo.GetChats(ctx, userID)
}

func (s *ChatService) GetFullChat(ctx context.Context, chatID string) (*model.FullChat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: chat with id %s", app_errors.ErrNotFound, chatID)
		}
		return nil, fmt.Errorf("could not get chat: %w", err)
	}

	messages, err := s.repo.GetActiveMessagesByChatID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("could not get messages: %w", err)
	}

	return &model.FullChat{Chat: *chat, Messages: messages}, nil
}

// resolveModels determines the final models and system prompt to use for a request,
// layering request-specific overrides on top of global settings.
func (s *ChatService) resolveModels(ctx context.Context, req *CreateMessageRequest, currentSettings *Settings) (mainModel, supportModel, systemPrompt string, err error) {
	mainModel = req.Model
	if mainModel == "" {
		mainModel = currentSettings.MainModel
	} else {
		// If a model is specified in the request, validate that it's available.
		availableModels, err := s.llm.ListModels(ctx)
		if err != nil {
			// Non-critical error; proceed without validation but log a warning.
			slog.Warn("Could not list models to validate request-specific model", "model", mainModel, "error", err)
		} else {
			modelNames := make([]string, len(availableModels.Models))
			for i, m := range availableModels.Models {
				modelNames[i] = m.Name
			}
			if !slices.Contains(modelNames, mainModel) {
				return "", "", "", fmt.Errorf("%w: model '%s' specified in request is not available", app_errors.ErrValidation, mainModel)
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
	// `req.Options.System` is an alternative way to set the system prompt, often used by LLM clients.
	if req.Options != nil && req.Options.System != nil {
		systemPrompt = *req.Options.System
	}

	return mainModel, supportModel, systemPrompt, nil
}

// HandleNewMessage is the main entry point for processing a new user message.
// It manages chat creation, history retrieval, and streaming the LLM response.
// Errors are sent via the stream channel, not returned directly.
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
		// For new chats, use a truncated version of the first message as a temporary title.
		chat := &model.Chat{ID: chatID, UserID: "default-user", Title: truncate(req.Content, 50), Model: modelToUse, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		if err := s.repo.CreateChat(ctx, chat); err != nil {
			slog.Error("Error creating chat", "error", err)
			streamChan <- model.StreamResponse{Error: "Could not create chat"}
			return
		}
	}

	lastMessage, err := s.repo.GetLastActiveMessage(ctx, chatID)
	// This is not a fatal error; it just means there's no previous context to send.
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
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
		// Log the error but don't stop; we can still try to get a response from the LLM.
		slog.Error("Error adding user message", "chat_id", chatID, "error", err)
	}

	history, err := s.repo.GetActiveMessagesByChatID(ctx, chatID)
	if err != nil {
		slog.Warn("Error getting message history for chat", "chat_id", chatID, "error", err)
	}

	// Construct the payload for the LLM provider, including the system prompt and history.
	llmMessages := []llm.Message{{Role: "system", Content: systemPromptToUse}}
	for _, msg := range history {
		llmMessages = append(llmMessages, llm.Message{Role: msg.Role, Content: msg.Content})
	}

	llmReq := &llm.GenerateRequest{
		Model:    modelToUse,
		Messages: llmMessages,
		Context:  ollamaContext, // Pass the context from the previous turn for stateful conversation.
		Options:  req.Options,
	}

	var fullResponse strings.Builder
	var finalContext json.RawMessage
	var finalStats *llm.GenerationStats
	llmStreamChan := make(chan llm.StreamResponse)
	// The actual LLM call is run in a goroutine to allow this function to process the stream.
	go func() {
		if err := s.llm.GenerateStream(ctx, llmReq, llmStreamChan); err != nil {
			slog.Error("LLM stream generation failed", "error", err)
		}
	}()

	// Consume from the LLM stream and forward to the client.
	for chunk := range llmStreamChan {
		streamChan <- model.StreamResponse{Content: chunk.Content, Done: chunk.Done, Error: chunk.Error}
		if chunk.Error != "" {
			break // Stop processing on LLM error.
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

	// Persist the complete assistant message to the database.
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

	// If it was a new chat, spawn a background task to generate a better title.
	if isNewChat {
		go s.generateTitle(context.Background(), chatID, supportModelToUse, userMessage.Content, assistantMessage.Content)
	}
}

// RegenerateMessage handles the complex logic of creating a new conversational branch.
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

	// The entire regeneration process is performed within a single database transaction
	// to ensure data consistency.
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		slog.Error("Regenerate failed to begin transaction", "error", err)
		streamChan <- model.StreamResponse{Error: "Database error"}
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Failed to rollback regeneration transaction", "error", err)
		}
	}()

	originalMsg, err := s.repo.GetMessageByID(ctx, originalAssistantMessageID)
	if err != nil || originalMsg.Role != "assistant" || originalMsg.ParentID == nil {
		streamChan <- model.StreamResponse{Error: "Original message not found or invalid"}
		return
	}

	// Mark the old conversational branch (the original message and its children) as inactive.
	if err := s.repo.DeactivateBranchTx(ctx, tx, originalAssistantMessageID); err != nil {
		slog.Error("Regenerate failed to deactivate branch", "error", err)
		streamChan <- model.StreamResponse{Error: "Database error during regeneration"}
		return
	}

	// Retrieve the now-current active history to send to the LLM.
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

	// --- Streaming logic (similar to HandleNewMessage) ---
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
			return // The transaction will be rolled back by the defer statement.
		}
		fullResponse.WriteString(chunk.Content)
		if chunk.Done {
			finalContext = chunk.Context
			finalStats = chunk.Stats
		}
	}
	slog.Debug("Finished streaming regenerated response from LLM.")
	// --- End of streaming logic ---

	var metadata json.RawMessage
	if finalStats != nil {
		metadata, _ = json.Marshal(finalStats)
	}

	// Create the new assistant message, linking it to the same parent as the original.
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
		// Context update happens outside the transaction as it's not critical for consistency.
		if err := s.repo.UpdateMessageContext(ctx, newAssistantMessage.ID, finalContext); err != nil {
			slog.Warn("Error setting Ollama context for new message", "message_id", newAssistantMessage.ID, "error", err)
		}
	}
}

// generateTitle is a fire-and-forget background task to generate a chat title using an LLM.
func (s *ChatService) generateTitle(ctx context.Context, chatID, supportModel, userQuery, assistantResponse string) {
	slog.Info("Generating title", "chat_id", chatID)

	// A specific, structured prompt to coax the model into returning clean JSON.
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

	// The response from the LLM is often noisy; attempt to extract a valid JSON object.
	jsonString := extractJSON(resp.Response)
	type TitleResponse struct {
		Title string `json:"title"`
	}
	var titleResp TitleResponse
	var newTitle string

	if jsonString != "" {
		if err := json.Unmarshal([]byte(jsonString), &titleResp); err != nil {
			slog.Warn("Found JSON-like string but failed to parse for title, cleaning raw string", "chat_id", chatID, "error", err)
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

// extractJSON is a best-effort attempt to find a JSON object within a string.
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

// cleanRawTitle removes common noise (like markdown code blocks) from LLM responses.
func cleanRawTitle(s string) string {
	s = strings.Split(s, "<think>")[0] // Some models add reasoning in <think> tags.
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

// truncate safely truncates a string to a given number of runes.
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
