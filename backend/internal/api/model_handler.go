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

// HandleListModels godoc
// @Summary      List local models
// @Description  Gets a list of all models available locally in Ollama.
// @Tags         Models
// @Produce      json
// @Success      200  {object}  llm.ListModelsResponse
// @Failure      500  {object}  map[string]string
// @Router       /models [get]
func (h *ModelHandler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.List(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve models")
		return
	}
	respondWithJSON(w, http.StatusOK, models)
}

// HandleShowModel godoc
// @Summary      Show model info
// @Description  Retrieves detailed information about a specific model, including its Modelfile and parameters.
// @Tags         Models
// @Accept       json
// @Produce      json
// @Param        modelRequest  body  llm.ShowModelRequest  true  "Model Name"
// @Success      200           {object}  llm.ModelInfo
// @Failure      400           {object}  map[string]string
// @Failure      404           {object}  map[string]string
// @Router       /models/show [post]
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

// HandleDeleteModel godoc
// @Summary      Delete a local model
// @Description  Deletes a model from the local Ollama storage.
// @Tags         Models
// @Accept       json
// @Produce      json
// @Param        modelRequest  body  llm.DeleteModelRequest  true  "Model Name to Delete"
// @Success      200           {object}  map[string]string
// @Failure      400           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Router       /models [delete]
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

// HandlePullModel godoc
// @Summary      Pull a new model
// @Description  Downloads a model from the Ollama registry. This is a streaming endpoint.
// @Tags         Models
// @Accept       json
// @Produce      text/event-stream
// @Param        modelRequest  body  llm.PullModelRequest  true  "Model Name to Pull"
// @Success      200           {object}  llm.PullStatus "Stream of progress status"
// @Failure      400           {object}  map[string]string "Sent as a stream error event"
// @Router       /models/pull [post]
func (h *ModelHandler) HandlePullModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req llm.PullModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		sendStreamError(w, "Invalid request body")
		return
	}

	streamChan := make(chan llm.PullStatus)
	
	// REFACTORED: We start the service call in a goroutine.
	// The service layer (or below) is responsible for closing the channel.
	go func() {
		err := h.service.Pull(r.Context(), &req, streamChan)
		if err != nil {
			log.Printf("Error from model pull service: %v", err)
			// If an error happens before the stream even starts,
			// the channel will be closed immediately and the loop below won't run.
			// The client will just see the connection close. This is acceptable.
		}
	}()

	// This loop now safely reads from the channel until it's closed by the producer.
	for chunk := range streamChan {
		// Check if the client has disconnected.
		if r.Context().Err() != nil {
			log.Println("Client disconnected during model pull.")
			break // Exit the loop.
		}

		if chunk.Error != "" {
			log.Printf("Received an error in the pull stream: %s", chunk.Error)
			// We can choose to forward this to the client as an error event if we want.
			sendStreamError(w, chunk.Error)
		}

		jsonData, err := json.Marshal(chunk)
		if err != nil {
			log.Printf("Error marshalling pull status: %v", err)
			continue // Skip this chunk
		}

		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	
	log.Println("Finished streaming model pull.")
}