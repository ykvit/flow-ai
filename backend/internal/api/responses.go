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
	Title string `json:"title" example:"My Custom Chat Title"`
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

// sendStreamError sends a structured error message over an SSE stream.
// It marshals the error, writes it with an "error" event type, and flushes.
func sendStreamError(w http.ResponseWriter, message string) {
	slog.Warn("Sending stream error to client", "message", message)
	errorPayload := ErrorResponse{Error: message}

	jsonData, err := json.Marshal(errorPayload)
	if err != nil {
		// This is a server-side problem; the struct should always be marshallable.
		slog.Error("Failed to marshal stream error payload", "error", err)
		return
	}

	// The 'event' field tells the client's EventSource to dispatch a specific event.
	if _, err := fmt.Fprintf(w, "event: error\ndata: %s\n\n", string(jsonData)); err != nil {
		// This is not a server error, but an expected I/O error if the client hangs up.
		slog.Warn("Failed to write stream error, client might have disconnected", "error", err)
		return
	}

	// Flush the writer to ensure the message is sent immediately.
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// writeStreamEvent is a helper to marshal data and write it to an SSE stream.
// It handles JSON marshaling, writing to the response writer, and flushing.
// It returns an error if writing or flushing fails, which usually indicates
// that the client has disconnected.
func writeStreamEvent(w http.ResponseWriter, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("Failed to marshal stream data to JSON", "error", err)
		// We don't return the error here because the issue is with the data, not the stream.
		// A log is sufficient.
		return nil
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		// This error is important, as it signals a problem with the connection.
		return fmt.Errorf("failed to write data to stream: %w", err)
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}
