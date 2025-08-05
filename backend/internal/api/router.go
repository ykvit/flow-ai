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
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api", func(r chi.Router) {
		// Settings endpoints
		r.Get("/settings", handler.GetSettings)
		r.Post("/settings", handler.UpdateSettings)

		// Chat endpoints
		r.Get("/chats", handler.GetChats)
		r.Get("/chats/{chatID}", handler.GetChat)

		// The main endpoint for sending messages and getting a stream response.
		// It can handle both new chats (no chatID) and existing ones.
		r.Post("/chats/messages", handler.HandleStreamMessage)
	})

	// This is a simple way to serve a frontend built with Vite/React/Vue.
	// For production, NGINX is a better choice.
	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	r.Handle("/*", http.StripPrefix("/", fileServer))

	return r
}