package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"flow-ai/backend/internal/api"
	"flow-ai/backend/internal/config"
	"flow-ai/backend/internal/database"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/repository"
	"flow-ai/backend/internal/service"
)

// @title           Flow-AI Backend API
// @version         0.1
// @description     This is the API server for the Flow-AI application, providing endpoints for chat, model management, and settings.
// @contact.name    API Support
// @contact.url     https://github.com/ykvit/flow-ai
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @BasePath        /api

func main() {
	setupLogger()

	bootstrapCfg, err := config.LoadBootstrapConfig()
	if err != nil {
		slog.Error("Failed to load bootstrap configuration", "error", err)
		os.Exit(1)
	}

	waitForOllama(bootstrapCfg.OllamaURL)

	db, err := database.InitDB(bootstrapCfg.DatabasePath)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Successfully connected to SQLite database.")

	repo := repository.NewSQLiteRepository(db)
	ollamaProvider := llm.NewOllamaProvider(bootstrapCfg.OllamaURL)

	settingsService := service.NewSettingsService(db, ollamaProvider)

	appSettings, err := settingsService.InitAndGet(context.Background(), bootstrapCfg.SystemPrompt)
	if err != nil {
		slog.Error("Failed to initialize application settings", "error", err)
		os.Exit(1)
	}
	slog.Info("Loaded application settings", "main_model", appSettings.MainModel)

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

	slog.Info("Starting server", "port", 8000)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

// setupLogger configures the global logger for structured JSON logging.
// The log level can be controlled via the LOG_LEVEL environment variable.
func setupLogger() {
	var level slog.Level
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
}

func waitForOllama(ollamaURL string) {
	slog.Info("Waiting for Ollama to be ready...")
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		resp, err := client.Get(ollamaURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			slog.Info("Ollama is ready.")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		slog.Debug("Ollama not ready yet, retrying in 3 seconds...")
		time.Sleep(3 * time.Second)
	}
}