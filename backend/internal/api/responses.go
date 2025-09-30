package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// ErrorResponse represents a standard error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// StatusResponse represents a standard status response.
type StatusResponse struct {
	Status string `json:"status"`
}

// UpdateTitleRequest is the structure for the manual title update request.
type UpdateTitleRequest struct {
	Title string `json:"title"`
}

func respondWithError(w http.ResponseWriter, code int, message string, err error) {
	slog.Warn(message, "code", code, "error", err)
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal JSON response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(response); err != nil {
		slog.Error("Failed to write JSON response", "error", err)
	}
}

func sendStreamError(w http.ResponseWriter, message string) {
	slog.Warn("Sending stream error", "message", message)
	errorPayload := ErrorResponse{Error: message}
	jsonData, _ := json.Marshal(errorPayload)
	// The 'event' field tells the client's EventSource to dispatch a specific event.
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", string(jsonData))
}
