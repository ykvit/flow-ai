package app

import (
	"context"
	"errors"
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

// Run initializes and starts the application. It returns an exit code.
func Run() int {
	setupLogger()

	bootstrapCfg, err := config.LoadBootstrapConfig()
	if err != nil {
		slog.Error("Failed to load bootstrap configuration", "error", err)
		return 1
	}

	waitForOllama(bootstrapCfg.OllamaURL)

	db, err := database.InitDB(bootstrapCfg.DatabasePath)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return 1
	}
	defer db.Close()
	slog.Info("Successfully connected to SQLite database.")

	repo := repository.NewSQLiteRepository(db)
	ollamaProvider := llm.NewOllamaProvider(bootstrapCfg.OllamaURL)

	settingsService := service.NewSettingsService(db, ollamaProvider)

	appSettings, err := settingsService.InitAndGet(context.Background(), bootstrapCfg.SystemPrompt)
	if err != nil {
		slog.Error("Failed to initialize application settings", "error", err)
		return 1
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
		WriteTimeout:      0, // Disabled for streaming endpoints
		IdleTimeout:       120 * time.Second,
	}

	slog.Info("Starting server", "port", 8000)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server failed", "error", err)
		return 1
	}

	return 0
}

func setupLogger() {
	var level slog.Level
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		level = slog.LevelDebug
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