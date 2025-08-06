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

	// Dependencies
	repo := repository.NewRedisRepository(rdb)
	ollamaProvider := llm.NewOllamaProvider(cfg.OllamaURL)
	
	// Chat dependencies
	chatService := service.NewChatService(repo, ollamaProvider)
	chatHandler := api.NewChatHandler(chatService, cfg)

	// Model management dependencies
	modelService := service.NewModelService(ollamaProvider)
	modelHandler := api.NewModelHandler(modelService)

	// FIX IS HERE: Pass both handlers to the router, as required.
	router := api.NewRouter(chatHandler, modelHandler)

	server := &http.Server{
		Addr:    ":8000",
		Handler: router,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout: 0,
		IdleTimeout: 120 * time.Second, 
	}

	log.Println("Starting server on port 8000...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}