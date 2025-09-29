package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/service"

	"github.com/go-chi/chi/v5"
)

// ChatHandler handles HTTP requests for chat and settings.
type ChatHandler struct {
	chatService     *service.ChatService
	settingsService *service.SettingsService
}

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
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve settings", err)
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
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if err := h.settingsService.Save(r.Context(), &newSettings); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save settings", err)
		return
	}

	slog.Info("Application settings updated")
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
	userID := "default-user" // This will be replaced with auth later.
	chats, err := h.chatService.ListChats(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chats", err)
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
		respondWithError(w, http.StatusNotFound, "Chat not found", err)
		return
	}
	respondWithJSON(w, http.StatusOK, fullChat)
}

// HandleStreamMessage godoc
// @Summary      Create a message and stream the response
// @Description  Sends a new message and initiates a real-time stream of the assistant's response.
// @Tags         Chats
// @Accept       json
// @Produce      text/event-stream
// @Param        message  body  service.CreateMessageRequest  true  "Message Request"
// @Success      200      {object} model.StreamResponse "Stream of response chunks"
// @Failure      400      {object} ErrorResponse "Sent as a stream error event"
// @Router       /v1/chats/messages [post]
func (h *ChatHandler) HandleStreamMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req service.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Error decoding request body", "error", err)
		sendStreamError(w, "Invalid request body")
		return
	}

	streamChan := make(chan model.StreamResponse)
	go h.chatService.HandleNewMessage(r.Context(), &req, streamChan)

	for chunk := range streamChan {
		if r.Context().Err() != nil {
			slog.Info("Client disconnected.")
			break
		}
		jsonData, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	slog.Debug("Finished streaming response.")
}

// HandleRegenerateMessage godoc
// @Summary      Regenerate a message
// @Description  Creates a new response for a previous user prompt.
// @Tags         Chats
// @Accept       json
// @Produce      text/event-stream
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
			slog.Info("Client disconnected during regeneration.")
			break
		}
		jsonData, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	slog.Debug("Finished streaming regenerated response.")
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
// @Failure      500     {object}  ErrorResponse
// @Router       /v1/chats/{chatID}/title [put]
func (h *ChatHandler) UpdateChatTitle(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	var req UpdateTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if err := h.chatService.UpdateChatTitle(r.Context(), chatID, req.Title); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update chat title", err)
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
// @Failure      500     {object}  ErrorResponse
// @Router       /v1/chats/{chatID} [delete]
func (h *ChatHandler) HandleDeleteChat(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if err := h.chatService.DeleteChat(r.Context(), chatID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete chat", err)
		return
	}
	respondWithJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}