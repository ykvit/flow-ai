package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "flow-ai/backend/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(chatHandler *ChatHandler, modelHandler *ModelHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // Chi's default logger is fine for dev
	r.Use(middleware.Recoverer)

	// Swagger documentation route
	r.Get("/api/swagger/*", httpSwagger.WrapHandler)

	// API version 1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Group for routes with a standard request timeout
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

		// Group for long-running streaming routes that should NOT have a timeout
		r.Group(func(r chi.Router) {
			r.Post("/chats/messages", chatHandler.HandleStreamMessage)
			r.Post("/chats/{chatID}/messages/{messageID}/regenerate", chatHandler.HandleRegenerateMessage)
			r.Post("/models/pull", modelHandler.HandlePullModel)
		})
	})

	// In a production setup with Nginx, this file server is not strictly necessary,
	// but it's useful for simplified local development without the reverse proxy.
	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	r.Handle("/*", http.StripPrefix("/", fileServer))

	return r
}