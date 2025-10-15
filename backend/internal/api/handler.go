package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	app_errors "flow-ai/backend/internal/errors"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/service"

	"github.com/go-chi/chi/v5"
)

// ChatHandler encapsulates the HTTP transport logic for chat and settings-related endpoints.
// It acts as a translator between the HTTP layer and the business logic (service) layer.
type ChatHandler struct {
	chatService     *service.ChatService
	settingsService *service.SettingsService
}

// NewChatHandler creates a new instance of ChatHandler with its required service dependencies.
func NewChatHandler(chatSvc *service.ChatService, settingsSvc *service.SettingsService) *ChatHandler {
	return &ChatHandler{
		chatService:     chatSvc,
		settingsService: settingsSvc,
	}
}

// GetSettings godoc
// @Summary      Get application settings
// @Description  Retrieves the current global settings for the application.
// @Tags         Settings
// @Produce      json
// @Success      200  {object}  service.Settings
// @Failure      500  {object}  ErrorResponse
// @Router       /v1/settings [get]
func (h *ChatHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsService.Get(r.Context())
	if err != nil {
		// Delegate error handling to the centralized `respondWithError` function,
		// which maps business-layer errors to appropriate HTTP status codes.
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, settings)
}

// UpdateSettings godoc
// @Summary      Update application settings
// @Description  Updates the global settings. Models must be available in Ollama.
// @Tags         Settings
// @Accept       json
// @Produce      json
// @Param        settings  body      service.Settings  true  "New settings to apply"
// @Success      200       {object}  StatusResponse
// @Failure      400       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /v1/settings [post]
func (h *ChatHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var newSettings service.Settings
	if err := json.NewDecoder(r.Body).Decode(&newSettings); err != nil {
		// If JSON decoding fails, it's a client-side malformed request.
		respondWithError(w, app_errors.ErrValidation)
		return
	}

	// Perform struct-level validation based on the `validate` tags.
	if err := validateRequest(&newSettings); err != nil {
		respondWithError(w, err)
		return
	}

	if err := h.settingsService.Save(r.Context(), &newSettings); err != nil {
		respondWithError(w, err)
		return
	}

	slog.Info("Application settings updated", "main_model", newSettings.MainModel, "support_model", newSettings.SupportModel)
	respondWithJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

// GetChats godoc
// @Summary      List all chats
// @Description  Retrieves a list of all chats, sorted by the most recently updated.
// @Tags         Chats
// @Produce      json
// @Success      200  {array}   model.Chat
// @Failure      500  {object}  ErrorResponse
// @Router       /v1/chats [get]
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	// In the current single-user model, we fetch all available chats.
	// When authentication is added, user identity will be extracted from the
	// request context (e.g., from a JWT middleware) and passed to the service layer.
	chats, err := h.chatService.ListChats(r.Context())
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, chats)
}

// GetChat godoc
// @Summary      Get a single chat
// @Description  Retrieves the full history for a single chat's active branch.
// @Tags         Chats
// @Produce      json
// @Param        chatID  path      string  true  "Chat ID"
// @Success      200     {object}  model.FullChat
// @Failure      404     {object}  ErrorResponse
// @Router       /v1/chats/{chatID} [get]
func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	fullChat, err := h.chatService.GetFullChat(r.Context(), chatID)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, fullChat)
}

// HandleStreamMessage godoc
// @Summary      Create a message and stream the response
// @Description  Sends a new message and initiates a real-time stream of the assistant's response.
// @Tags         Chats
// @Accept       json
// @Produce      application/json
// @Description  Sends a new message and initiates a real-time stream of the assistant's response (SSE).
// @Param        message  body  service.CreateMessageRequest  true  "Message Request"
// @Success      200      {object} model.StreamResponse "Stream of response chunks"
// @Failure      400      {object} ErrorResponse "Sent as a stream error event"
// @Router       /v1/chats/messages [post]
func (h *ChatHandler) HandleStreamMessage(w http.ResponseWriter, r *http.Request) {
	// Set headers for Server-Sent Events (SSE).
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req service.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Error decoding stream request body", "error", err)
		sendStreamError(w, "Invalid request body")
		return
	}

	// For streaming endpoints, validation errors are also sent over the event stream
	// to ensure a consistent communication channel with the client.
	if err := validateRequest(&req); err != nil {
		slog.Warn("Stream request validation failed", "error", err)
		sendStreamError(w, err.Error())
		return
	}

	streamChan := make(chan model.StreamResponse)
	// Launch the business logic in a separate goroutine to not block the handler.
	go h.chatService.HandleNewMessage(r.Context(), &req, streamChan)

	// Loop and send stream chunks to the client as they arrive.
	for chunk := range streamChan {
		// The client context is checked to detect if the client has disconnected.
		if r.Context().Err() != nil {
			slog.Info("Client disconnected, stopping stream.")
			break
		}
		if err := writeStreamEvent(w, chunk); err != nil {
			// This error typically means the client closed the connection.
			slog.Warn("Could not write to stream, client likely disconnected.", "error", err)
			break
		}
	}

	slog.Debug("Finished streaming response.")
}

// HandleRegenerateMessage godoc
// @Summary      Regenerate a message
// @Description  Creates a new response for a previous user prompt.
// @Tags         Chats
// @Accept       json
// @Produce      application/json
// @Description  Creates a new response for a previous user prompt (SSE).
// @Param        chatID    path      string                              true  "Chat ID"
// @Param        messageID path      string                              true  "The ID of the assistant message to regenerate"
// @Param        regenRequest body   service.RegenerateMessageRequest    true  "Regeneration options"
// @Success      200       {object}  model.StreamResponse "Stream of new response chunks"
// @Failure      400       {object}  ErrorResponse "Sent as a stream error event"
// @Failure      404       {object}  ErrorResponse "Sent as a stream error event"
// @Router       /v1/chats/{chatID}/messages/{messageID}/regenerate [post]
func (h *ChatHandler) HandleRegenerateMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	chatID := chi.URLParam(r, "chatID")
	messageID := chi.URLParam(r, "messageID")

	var req service.RegenerateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendStreamError(w, "Invalid request payload")
		return
	}

	streamChan := make(chan model.StreamResponse)
	go h.chatService.RegenerateMessage(r.Context(), chatID, messageID, &req, streamChan)

	for chunk := range streamChan {
		if r.Context().Err() != nil {
			slog.Info("Client disconnected during regeneration.", "chatID", chatID)
			break
		}
		if err := writeStreamEvent(w, chunk); err != nil {
			slog.Warn("Could not write to regeneration stream, client likely disconnected.", "error", err, "chatID", chatID)
			break
		}
	}

	slog.Debug("Finished streaming regenerated response.", "chatID", chatID)
}

// UpdateChatTitle godoc
// @Summary      Update a chat's title
// @Description  Manually renames a chat.
// @Tags         Chats
// @Accept       json
// @Produce      json
// @Param        chatID  path      string              true  "Chat ID"
// @Param        title   body      UpdateTitleRequest  true  "New title"
// @Success      200     {object}  StatusResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /v1/chats/{chatID}/title [put]
func (h *ChatHandler) UpdateChatTitle(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	var req UpdateTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, app_errors.ErrValidation)
		return
	}

	if err := validateRequest(&req); err != nil {
		respondWithError(w, err)
		return
	}

	if err := h.chatService.UpdateChatTitle(r.Context(), chatID, req.Title); err != nil {
		respondWithError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

// HandleDeleteChat godoc
// @Summary      Delete a chat
// @Description  Permanently deletes a chat and all its associated messages.
// @Tags         Chats
// @Produce      json
// @Param        chatID  path      string  true  "Chat ID"
// @Success      200     {object}  StatusResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /v1/chats/{chatID} [delete]
func (h *ChatHandler) HandleDeleteChat(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if err := h.chatService.DeleteChat(r.Context(), chatID); err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}
