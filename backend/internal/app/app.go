package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"flow-ai/backend/internal/api"
	"flow-ai/backend/internal/config"
	"flow-ai/backend/internal/database"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/repository"
	"flow-ai/backend/internal/service"
)

type App struct {
	Config *config.Config
	DB     *sql.DB
	Server *http.Server
}

func NewApp(cfg *config.Config) (*App, error) {
	waitForOllama(cfg.OllamaURL)

	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	slog.Info("Successfully connected to SQLite database.")

	repo := repository.NewSQLiteRepository(db)
	ollamaProvider := llm.NewOllamaProvider(cfg.OllamaURL)
	settingsService := service.NewSettingsService(db, ollamaProvider)

	appSettings, err := settingsService.InitAndGet(context.Background(), cfg.InitialSystemPrompt)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("Failed to close database connection during initial setup error", "error", closeErr)
		}
		return nil, err
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

	return &App{
		Config: cfg,
		DB:     db,
		Server: server,
	}, nil
}

func Run() int {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return 1
	}

	setupLogger(cfg.LogLevel)
	logConfigSource()

	app, err := NewApp(cfg)
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		return 1
	}
	defer func() {
		if err := app.DB.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()

	slog.Info("Starting server", "port", 8000)
	if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server failed", "error", err)
		return 1
	}

	return 0
}

func logConfigSource() {
	configFileUsed := viper.ConfigFileUsed()
	if configFileUsed != "" {
		slog.Info("Successfully loaded configuration from file.", "file", configFileUsed)
	} else {
		slog.Info("Configuration file not found. Using environment variables and defaults.")
	}
}

func setupLogger(logLevel string) {
	var level slog.Level
	switch strings.ToUpper(logLevel) {
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
			if bErr := resp.Body.Close(); bErr != nil {
				slog.Warn("Failed to close response body in ollama health check", "error", bErr)
			}
			slog.Info("Ollama is ready.")
			return
		}
		if resp != nil {
			if bErr := resp.Body.Close(); bErr != nil {
				slog.Warn("Failed to close response body in ollama health check (retry path)", "error", bErr)
			}
		}
		slog.Debug("Ollama not ready yet, retrying in 3 seconds...", "url", ollamaURL, "error", err)
		time.Sleep(3 * time.Second)
	}
}
