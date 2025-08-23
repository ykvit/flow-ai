package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"flow-ai/backend/internal/api"
	"flow-ai/backend/internal/config"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/repository"
	"flow-ai/backend/internal/service"
)

func main() {
	bootstrapCfg, err := config.LoadBootstrapConfig()
	if err != nil {
		log.Fatalf("Failed to load bootstrap configuration: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: bootstrapCfg.RedisAddr})
	log.Println("Successfully connected to Redis.")

	// Create providers and repositories first
	repo := repository.NewRedisRepository(rdb)
	ollamaProvider := llm.NewOllamaProvider(bootstrapCfg.OllamaURL)
	
	// NEW: Pass the ollamaProvider to the SettingsService
	settingsService := service.NewSettingsService(rdb, ollamaProvider)
	
	// InitAndGet now performs smart initialization
	appSettings, err := settingsService.InitAndGet(context.Background(), bootstrapCfg)
	if err != nil {
		log.Fatalf("Failed to initialize application settings: %v", err)
	}
	log.Printf("Loaded application settings. Main model is: %s", appSettings.MainModel)

	// Create other services with their dependencies
	chatService := service.NewChatService(repo, ollamaProvider, settingsService)
	modelService := service.NewModelService(ollamaProvider)
	
	// Create handlers
	chatHandler := api.NewChatHandler(chatService, settingsService)
	modelHandler := api.NewModelHandler(modelService)

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