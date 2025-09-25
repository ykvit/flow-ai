package api

import (
	"encoding/json"
	"fmt"
	"log"
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
// @Description  Retrieves the current global settings for the application, such as the default model.
// @Tags         Settings
// @Produce      json
// @Success      200  {object}  service.Settings
// @Failure      500  {object}  map[string]string
// @Router       /settings [get]
func (h *ChatHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsService.Get(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve settings")
		return
	}
	respondWithJSON(w, http.StatusOK, settings)
}

// UpdateSettings godoc
// @Summary      Update application settings
// @Description  Updates the global settings. The models specified must be available in Ollama.
// @Tags         Settings
// @Accept       json
// @Produce      json
// @Param        settings  body      service.Settings  true  "New settings to apply"
// @Success      200       {object}  map[string]string
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /settings [post]
func (h *ChatHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var newSettings service.Settings
	if err := json.NewDecoder(r.Body).Decode(&newSettings); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.settingsService.Save(r.Context(), &newSettings); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save settings")
		return
	}

	log.Println("Settings updated.")
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetChats godoc
// @Summary      List all chats
// @Description  Retrieves a list of all chats for the user, sorted by the most recently updated.
// @Tags         Chats
// @Produce      json
// @Success      200  {array}   model.Chat
// @Failure      500  {object}  map[string]string
// @Router       /chats [get]
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID := "default-user" // This will be replaced with auth later.
	chats, err := h.chatService.ListChats(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chats")
		return
	}
	respondWithJSON(w, http.StatusOK, chats)
}

// GetChat godoc
// @Summary      Get a single chat
// @Description  Retrieves the full history and metadata for a single chat's active branch.
// @Tags         Chats
// @Produce      json
// @Param        chatID  path      string  true  "Chat ID"
// @Success      200     {object}  model.FullChat
// @Failure      404     {object}  map[string]string
// @Router       /chats/{chatID} [get]
func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	fullChat, err := h.chatService.GetFullChat(r.Context(), chatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chat not found")
		return
	}
	respondWithJSON(w, http.StatusOK, fullChat)
}

// HandleStreamMessage godoc
// @Summary      Create a message (and stream response)
// @Description  Sends a new message and initiates a real-time stream of the assistant's response. If chat_id is omitted, a new chat is created. The response is a server-sent event stream.
// @Tags         Chats
// @Accept       json
// @Produce      text/event-stream
// @Param        message  body  service.CreateMessageRequest  true  "Message Request"
// @Success      200      {object} model.StreamResponse "Stream of response chunks"
// @Failure      400      {object} map[string]string "Sent as a stream error event"
// @Router       /chats/messages [post]
func (h *ChatHandler) HandleStreamMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req service.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		// Use a helper to send a structured SSE error
		sendStreamError(w, "Invalid request body")
		return
	}

	streamChan := make(chan model.StreamResponse)
	go h.chatService.HandleNewMessage(r.Context(), &req, streamChan)

	for chunk := range streamChan {
		if r.Context().Err() != nil {
			log.Println("Client disconnected.")
			break
		}
		jsonData, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	log.Println("Finished streaming response.")
}

// --- Title and Chat Deletion Handlers ---

// UpdateTitleRequest is the structure for the manual title update request.
type UpdateTitleRequest struct {
	Title string `json:"title"`
}

// UpdateChatTitle godoc
// @Summary      Update a chat's title
// @Description  Manually renames a chat.
// @Tags         Chats
// @Accept       json
// @Produce      json
// @Param        chatID  path      string              true  "Chat ID"
// @Param        title   body      UpdateTitleRequest  true  "New title"
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /chats/{chatID}/title [put]
func (h *ChatHandler) UpdateChatTitle(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	var req UpdateTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.chatService.UpdateChatTitle(r.Context(), chatID, req.Title); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update chat title")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleDeleteChat godoc
// @Summary      Delete a chat
// @Description  Permanently deletes a chat and all its associated messages.
// @Tags         Chats
// @Produce      json
// @Param        chatID  path      string  true  "Chat ID"
// @Success      200     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /chats/{chatID} [delete]
func (h *ChatHandler) HandleDeleteChat(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	if err := h.chatService.DeleteChat(r.Context(), chatID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete chat")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Helper Functions ---

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func sendStreamError(w http.ResponseWriter, message string) {
	errorPayload := map[string]string{"error": message}
	jsonData, _ := json.Marshal(errorPayload)
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", string(jsonData))
}