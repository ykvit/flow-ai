package api

import (
	"encoding/json"
	"fmt"
	"log"
	"flow-ai/backend/internal/config"
	"flow-ai/backend/internal/model"
	"flow-ai/backend/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ChatHandler struct {
	service *service.ChatService
	cfg     *config.Config
}

func NewChatHandler(svc *service.ChatService, cfg *config.Config) *ChatHandler {
	return &ChatHandler{service: svc, cfg: cfg}
}

// ... (GetSettings, UpdateSettings, GetChats, GetChat functions are IDENTICAL to the last version) ...
func (h *ChatHandler) GetSettings(w http.ResponseWriter, r *http.Request) { respondWithJSON(w, http.StatusOK, h.cfg) }
func (h *ChatHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var newCfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	h.cfg.SystemPrompt = newCfg.SystemPrompt
	h.cfg.MainModel = newCfg.MainModel
	h.cfg.SupportModel = newCfg.SupportModel
	log.Println("Settings updated.")
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID := "default-user"
	chats, err := h.service.ListChats(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chats")
		return
	}
	respondWithJSON(w, http.StatusOK, chats)
}
func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	chatID := chi.URLParam(r, "chatID")
	fullChat, err := h.service.GetFullChat(r.Context(), chatID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chat not found")
		return
	}
	respondWithJSON(w, http.StatusOK, fullChat)
}


// HandleStreamMessage now creates a channel of model.StreamResponse
func (h *ChatHandler) HandleStreamMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req service.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", `{"error": "Invalid request body"}`)
		return
	}
	if req.Model == "" { req.Model = h.cfg.MainModel }
	if req.SystemPrompt == "" { req.SystemPrompt = h.cfg.SystemPrompt }
	if req.SupportModel == "" { req.SupportModel = h.cfg.SupportModel }

	// IMPORTANT: The channel now uses the shared model type
	streamChan := make(chan model.StreamResponse)
	go h.service.HandleNewMessage(r.Context(), &req, streamChan)

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

func respondWithError(w http.ResponseWriter, code int, message string) { respondWithJSON(w, code, map[string]string{"error": message}) }
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}