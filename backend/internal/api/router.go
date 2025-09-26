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
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))

		r.Route("/api", func(r chi.Router) {
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
	})

	// Group for long-running streaming routes that should NOT have a timeout
	r.Group(func(r chi.Router) {
		r.Post("/api/chats/messages", chatHandler.HandleStreamMessage)
		r.Post("/api/models/pull", modelHandler.HandlePullModel)
		// NEW: Route for regenerating a message
		r.Post("/api/chats/{chatID}/messages/{messageID}/regenerate", chatHandler.HandleRegenerateMessage)
	})

	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	r.Handle("/*", http.StripPrefix("/", fileServer))

	return r
}