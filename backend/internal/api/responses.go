package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	app_errors "flow-ai/backend/internal/errors"
)

// This file contains shared DTOs (Data Transfer Objects) for API responses
// and helper functions for sending consistent HTTP responses.

// ErrorResponse defines the standard JSON structure for error messages.
type ErrorResponse struct {
	Error string `json:"error"`
}

// StatusResponse defines a generic success response, typically for operations
// like POST, PUT, DELETE that don't need to return a full resource.
type StatusResponse struct {
	Status string `json:"status"`
}

// UpdateTitleRequest is the DTO for the manual chat title update endpoint.
// It includes validation tags to enforce business rules at the API boundary.
type UpdateTitleRequest struct {
	Title string `json:"title" validate:"required,min=1,max=100" example:"My Custom Chat Title"`
}

// respondWithError is the centralized error handling function for the API layer.
// It maps custom business-layer errors to appropriate HTTP status codes and formats
// a standard JSON error response.
func respondWithError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, app_errors.ErrNotFound):
		statusCode = http.StatusNotFound
		message = "The requested resource was not found."
	case errors.Is(err, app_errors.ErrValidation):
		statusCode = http.StatusBadRequest
		// For validation errors, the error message from the service layer
		// is already descriptive and user-friendly.
		message = err.Error()
	case errors.Is(err, app_errors.ErrConflict):
		statusCode = http.StatusConflict
		message = "A conflict occurred with the current state of the resource."
	case errors.Is(err, app_errors.ErrPermission):
		statusCode = http.StatusForbidden
		message = "You do not have permission to perform this action."
	default:
		// Any unhandled error is considered an internal server error.
		// This prevents leaking implementation details to the client.
		statusCode = http.StatusInternalServerError
		message = "An unexpected internal server error occurred."
	}

	// The original, more detailed error is logged for debugging purposes,
	// while a generic message is sent to the client.
	slog.Warn("Responding with error", "status_code", statusCode, "client_message", message, "internal_error", err)

	respondWithJSON(w, statusCode, ErrorResponse{Error: message})
}

// respondWithJSON is a low-level helper for marshaling a payload to JSON
// and writing it to the http.ResponseWriter with a given status code.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		// This indicates a server-side programming error (e.g., trying to marshal a channel).
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

// sendStreamError sends a structured error message over a Server-Sent Events (SSE) stream.
// This ensures that clients consuming streams can handle errors gracefully.
func sendStreamError(w http.ResponseWriter, message string) {
	slog.Warn("Sending stream error to client", "message", message)
	errorPayload := ErrorResponse{Error: message}

	jsonData, err := json.Marshal(errorPayload)
	if err != nil {
		slog.Error("Failed to marshal stream error payload", "error", err)
		return
	}

	// The `event: error` line allows clients to add a specific event listener
	// for errors, e.g., `eventSource.addEventListener('error', ...)`.
	if _, err := fmt.Fprintf(w, "event: error\ndata: %s\n\n", string(jsonData)); err != nil {
		// This is often an expected I/O error if the client closes the connection.
		slog.Warn("Failed to write stream error, client might have disconnected", "error", err)
		return
	}

	// Flush the writer to ensure the message is sent to the client immediately.
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// writeStreamEvent is a generic helper to marshal data and write it to an SSE stream.
// It returns an error on write failure, which is a signal that the client has disconnected.
func writeStreamEvent(w http.ResponseWriter, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("Failed to marshal stream data to JSON", "error", err)
		// Don't return an error to the caller, as the stream is still open.
		// The issue is with the data, not the connection.
		return nil
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		// A write failure here is a strong indicator of a closed connection.
		return fmt.Errorf("failed to write data to stream: %w", err)
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}
