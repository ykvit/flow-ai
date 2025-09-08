package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"flow-ai/backend/internal/api"
	"flow-ai/backend/internal/config"
	"flow-ai/backend/internal/database"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/repository"
	"flow-ai/backend/internal/service"
)

func main() {
	bootstrapCfg, err := config.LoadBootstrapConfig()
	if err != nil {
		log.Fatalf("Failed to load bootstrap configuration: %v", err)
	}

	waitForOllama(bootstrapCfg.OllamaURL)

	db, err := database.InitDB(bootstrapCfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Successfully connected to SQLite database.")

	repo := repository.NewSQLiteRepository(db)
	ollamaProvider := llm.NewOllamaProvider(bootstrapCfg.OllamaURL)
	
	settingsService := service.NewSettingsService(db, ollamaProvider)
	
	// The service now discovers settings or creates them dynamically.
	appSettings, err := settingsService.InitAndGet(context.Background(), bootstrapCfg.SystemPrompt)
	if err != nil {
		log.Fatalf("Failed to initialize application settings: %v", err)
	}
	log.Printf("Loaded application settings. Main model is: '%s'", appSettings.MainModel)

	chatService := service.NewChatService(repo, ollamaProvider, settingsService)
	modelService := service.NewModelService(ollamaProvider)
	
	chatHandler := api.NewChatHandler(chatService, settingsService)
	modelHandler := api.NewModelHandler(modelService)
	router := api.NewRouter(chatHandler, modelHandler)

	server := &http.Server{
		Addr:              ":8000",
		Handler:           router,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       120 * time.Second, 
	}

	log.Println("Starting server on port 8000...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func waitForOllama(ollamaURL string) {
	log.Println("Waiting for Ollama to be ready...")
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		resp, err := client.Get(ollamaURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Println("Ollama is ready.")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		log.Println("Ollama not ready yet, retrying in 3 seconds...")
		time.Sleep(3 * time.Second)
	}
}