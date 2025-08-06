package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler *ChatHandler) *chi.Mux {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Group for routes that should have a timeout.
	// These are requests that should complete quickly.
	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))

		r.Route("/api", func(r chi.Router) {
			// Settings endpoints
			r.Get("/settings", handler.GetSettings)
			r.Post("/settings", handler.UpdateSettings)

			// Chat list and chat history endpoints
			r.Get("/chats", handler.GetChats)
			r.Get("/chats/{chatID}", handler.GetChat)

			// Endpoint to manually update a chat's title
			r.Put("/chats/{chatID}/title", handler.UpdateChatTitle)

			// TODO: Add endpoint for regenerating title here later
			// r.Post("/chats/{chatID}/regenerate-title", handler.RegenerateChatTitle)
		})
	})

	// Group for long-running routes that should NOT have a timeout.
	// This is critical for model streaming.
	r.Group(func(r chi.Router) {
		// The main endpoint for sending messages and getting a stream response.
		// It does not have a timeout to allow for slow models.
		r.Post("/api/chats/messages", handler.HandleStreamMessage)
	})


	// This is a simple way to serve a frontend built with Vite/React/Vue.
	// For production, a dedicated web server like NGINX is a better choice.
	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	r.Handle("/*", http.StripPrefix("/", fileServer))

	return r
}