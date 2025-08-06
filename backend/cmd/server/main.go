package main

import (
	"log"
	"flow-ai/backend/internal/api"
	"flow-ai/backend/internal/config"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/repository"
	"flow-ai/backend/internal/service"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	log.Println("Successfully connected to Redis.")

	// Dependencies are now created in the correct order
	repo := repository.NewRedisRepository(rdb)
	ollamaProvider := llm.NewOllamaProvider(cfg.OllamaURL)
	chatService := service.NewChatService(repo, ollamaProvider)
	
	chatHandler := api.NewChatHandler(chatService, cfg)

	router := api.NewRouter(chatHandler)

	// FIX: Server timeouts are adjusted for long-running connections (like SSE).
	server := &http.Server{
		Addr:    ":8000",
		Handler: router,
		// ReadHeaderTimeout is a good practice to prevent slow-loris attacks.
		ReadHeaderTimeout: 20 * time.Second,
		// WriteTimeout is set to 0 (infinity) because SSE streams can be idle
		// for a long time while the model is processing.
		WriteTimeout: 0,
		// IdleTimeout is the preferred way to manage keep-alive connections.
		// A long timeout is set to keep the connection open for streaming.
		IdleTimeout: 120 * time.Second, 
	}

	log.Println("Starting server on port 8000...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}