package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	app_errors "flow-ai/backend/internal/errors"
	"flow-ai/backend/internal/interfaces"
	"flow-ai/backend/internal/llm"
)

// ModelHandler handles HTTP requests for managing local Ollama models.
// It serves as a bridge between the HTTP layer and the model management service.
type ModelHandler struct {
	service interfaces.ModelService
}

// NewModelHandler creates a new instance of ModelHandler.
func NewModelHandler(svc interfaces.ModelService) *ModelHandler {
	return &ModelHandler{service: svc}
}

// HandleListModels godoc
// @Summary      List local models
// @Description  Gets a list of all models available locally in Ollama.
// @Tags         Models
// @Produce      json
// @Success      200  {object}  llm.ListModelsResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /v1/models [get]
func (h *ModelHandler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.List(r.Context())
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, models)
}

// HandleShowModel godoc
// @Summary      Show model info
// @Description  Retrieves detailed information about a specific model.
// @Tags         Models
// @Accept       json
// @Produce      json
// @Param        modelRequest  body      llm.ShowModelRequest  true  "Model Name"
// @Success      200           {object}  llm.ModelInfo
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /v1/models/show [post]
func (h *ModelHandler) HandleShowModel(w http.ResponseWriter, r *http.Request) {
	var req llm.ShowModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, app_errors.ErrValidation)
		return
	}
	// Note: Validation for the model name itself happens within the Ollama provider,
	// which will return an error if the model doesn't exist.
	info, err := h.service.Show(r.Context(), &req)
	if err != nil {
		respondWithError(w, err)
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
// @Param        modelRequest  body      llm.DeleteModelRequest  true  "Model Name to Delete"
// @Success      200           {object}  StatusResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /v1/models [delete]
func (h *ModelHandler) HandleDeleteModel(w http.ResponseWriter, r *http.Request) {
	var req llm.DeleteModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, app_errors.ErrValidation)
		return
	}
	if err := h.service.Delete(r.Context(), &req); err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

// HandlePullModel godoc
// @Summary      Pull a new model
// @Description  Downloads a model from the Ollama registry. This is a streaming endpoint.
// @Tags         Models
// @Accept       json
// @Produce      application/json
// @Description  Downloads a model from the Ollama registry. This is a streaming endpoint (SSE).
// @Param        modelRequest  body      llm.PullModelRequest  true  "Model Name to Pull"
// @Success      200           {object}  llm.PullStatus "Stream of progress status"
// @Failure      400           {object}  ErrorResponse "Sent as a stream error event"
// @Router       /v1/models/pull [post]
func (h *ModelHandler) HandlePullModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req llm.PullModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Error decoding request body for model pull", "error", err)
		sendStreamError(w, "Invalid request body")
		return
	}

	streamChan := make(chan llm.PullStatus)
	// The service call is launched in a goroutine to allow the handler to immediately
	// start listening for and processing stream events.
	go func() {
		// Errors from the service are logged here, as they are also propagated
		// through the stream channel to the client.
		if err := h.service.Pull(r.Context(), &req, streamChan); err != nil {
			slog.Error("Error from model pull service", "model", req.Name, "error", err)
		}
	}()

	for chunk := range streamChan {
		if r.Context().Err() != nil {
			slog.Info("Client disconnected during model pull.", "model", req.Name)
			break
		}

		// The stream itself can contain error messages from the provider.
		// These are logged for visibility on the server-side.
		if chunk.Error != "" {
			slog.Warn("Received an error in the pull stream", "model", req.Name, "error", chunk.Error)
		}

		if err := writeStreamEvent(w, chunk); err != nil {
			slog.Warn("Could not write to model pull stream, client likely disconnected.", "error", err)
			break
		}
	}

	slog.Info("Finished streaming model pull.", "model", req.Name)
}
