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
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", `{"error": "Invalid request body"}`)
		return
	}

	streamChan := make(chan llm.PullStatus)
	
	// The main function will handle closing it, preventing a double-close panic.
	go func() {
		err := h.service.Pull(r.Context(), &req, streamChan)
		if err != nil {
			log.Printf("Error during model pull service call: %v", err)
			// Attempt to send a final error message if the channel is still open.
			// This is not guaranteed to be received but is good practice.
			select {
			case streamChan <- llm.PullStatus{Error: err.Error()}:
			case <-r.Context().Done():
			}
		}
		// The goroutine *never* closes the channel.
	}()

	// The main function loop now safely closes the channel once done.
	defer func() {
		close(streamChan)
		log.Println("Finished streaming model pull and closed channel.")
	}()

	for chunk := range streamChan {
		if r.Context().Err() != nil {
			log.Println("Client disconnected during model pull.")
			break // Exit the loop, defer will close the channel.
		}
		jsonData, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}