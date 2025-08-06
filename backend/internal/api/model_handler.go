package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/service"
)

// ModelHandler handles HTTP requests for model management.
type ModelHandler struct {
	service *service.ModelService
}

func NewModelHandler(svc *service.ModelService) *ModelHandler {
	return &ModelHandler{service: svc}
}

func (h *ModelHandler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.List(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve models")
		return
	}
	respondWithJSON(w, http.StatusOK, models)
}

func (h *ModelHandler) HandleShowModel(w http.ResponseWriter, r *http.Request) {
	var req llm.ShowModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	info, err := h.service.Show(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not get model info")
		return
	}
	respondWithJSON(w, http.StatusOK, info)
}

func (h *ModelHandler) HandleDeleteModel(w http.ResponseWriter, r *http.Request) {
	var req llm.DeleteModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := h.service.Delete(r.Context(), &req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete model")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandlePullModel streams the download progress of a model.
func (h *ModelHandler) HandlePullModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req llm.PullModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", `{"error": "Invalid request body"}`)
		return
	}

	streamChan := make(chan llm.PullStatus)
	go func() {
		defer close(streamChan)
		if err := h.service.Pull(r.Context(), &req, streamChan); err != nil {
			log.Printf("Error pulling model: %v", err)
			// Send the final error via the channel if it's still open
			select {
			case streamChan <- llm.PullStatus{Error: err.Error()}:
			default:
			}
		}
	}()

	for chunk := range streamChan {
		if r.Context().Err() != nil {
			log.Println("Client disconnected during model pull.")
			break
		}
		jsonData, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	log.Println("Finished streaming model pull.")
}