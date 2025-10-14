package api

import (
	"net/http"
	"time"

	// This blank import is required by swaggo to find the API definitions.
	_ "flow-ai/backend/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter creates and configures a new chi router with all the application's routes.
func NewRouter(chatHandler *ChatHandler, modelHandler *ModelHandler) *chi.Mux {
	r := chi.NewRouter()

	// --- Global Middleware ---
	// These are applied to every request.
	r.Use(middleware.RequestID) // Injects a unique request ID into the context.
	r.Use(middleware.RealIP)    // Sets the remote address to the real IP from proxy headers.
	r.Use(middleware.Logger)    // Logs the start and end of each request with useful info.
	r.Use(middleware.Recoverer) // Recovers from panics and returns a 500 error.

	// --- Public Routes ---
	// Routes that don't require authentication or versioning.

	// Serves the auto-generated Swagger UI for API documentation.
	r.Get("/api/swagger/*", httpSwagger.WrapHandler)

	// A simple health check endpoint. Crucial for container orchestration systems
	// like Kubernetes to perform liveness and readiness probes.
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// The response body itself is not critical, but a 200 OK status is.
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// --- API Version 1 Routes ---
	// All primary API endpoints are grouped under the /api/v1 prefix.
	r.Route("/api/v1", func(r chi.Router) {

		// Group for standard JSON API routes that should have a request timeout
		// to prevent client connections from hanging indefinitely.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(60 * time.Second))

			// --- Settings ---
			r.Get("/settings", chatHandler.GetSettings)
			r.Post("/settings", chatHandler.UpdateSettings)

			// --- Chats ---
			r.Get("/chats", chatHandler.GetChats)
			r.Get("/chats/{chatID}", chatHandler.GetChat)
			r.Put("/chats/{chatID}/title", chatHandler.UpdateChatTitle)
			r.Delete("/chats/{chatID}", chatHandler.HandleDeleteChat)

			// --- Models ---
			r.Get("/models", modelHandler.HandleListModels)
			r.Post("/models/show", modelHandler.HandleShowModel)
			r.Delete("/models", modelHandler.HandleDeleteModel)
		})

		// Group for long-running, streaming endpoints. These routes must NOT have a timeout,
		// as they are designed to hold a connection open for an extended period.
		r.Group(func(r chi.Router) {
			r.Post("/chats/messages", chatHandler.HandleStreamMessage)
			r.Post("/chats/{chatID}/messages/{messageID}/regenerate", chatHandler.HandleRegenerateMessage)
			r.Post("/models/pull", modelHandler.HandlePullModel)
		})
	})

	// --- Frontend File Server ---
	// This serves the static React frontend. In a typical production deployment,
	// this would be handled by Nginx, but it's useful for simplified local development.
	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	r.Handle("/*", http.StripPrefix("/", fileServer))

	return r
}
