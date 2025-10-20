// Package app is the heart of the application, responsible for initializing
// and wiring together all the different components (database, services, API router).
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

// App holds all the long-lived components of the application, such as the
// database connection and the HTTP server.
//
// Encapsulating these components in a struct makes the application's
// dependency graph explicit and manageable. It's a key part of the refactoring
// that made the initialization logic testable.
type App struct {
	Config *config.Config
	DB     *sql.DB
	Server *http.Server
}

// NewApp creates and wires up all application components based on the provided config.
// It initializes dependencies in a specific order: DB -> Repository -> Services -> API Handlers.
//
// WHY THIS FUNCTION EXISTS (REFACTORING CONTEXT):
// Previously, all this logic was inside the `Run` function. By extracting it
// into `NewApp`, we created a function that builds the entire application stack
// *without* starting the web server. This is crucial for testing. Our `app_test.go`
// can now call `NewApp` to verify that the entire application can be initialized
// without errors, giving us high confidence and test coverage for this critical path.
func NewApp(cfg *config.Config) (*App, error) {
	// Wait for the external Ollama service to be available before proceeding.
	// This prevents the application from starting in a broken state if its
	// core dependency is not ready.
	waitForOllama(cfg.OllamaURL)

	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	slog.Info("Successfully connected to SQLite database.")

	// --- Dependency Injection ---
	// Create concrete implementations of our interfaces.
	repo := repository.NewSQLiteRepository(db)
	ollamaProvider := llm.NewOllamaProvider(cfg.OllamaURL)

	// Services are instantiated with their dependencies.
	settingsService := service.NewSettingsService(db, ollamaProvider)

	// Initialize settings on first run, which is a critical startup step.
	// If this fails, we can't proceed, so we close the DB and return the error.
	appSettings, err := settingsService.InitAndGet(context.Background(), cfg.InitialSystemPrompt)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("Failed to close database connection during initial setup error", "error", closeErr)
		}
		return nil, err
	}
	slog.Info("Loaded application settings", "main_model", appSettings.MainModel)

	// The ChatService depends on the SettingsService, demonstrating inter-service dependency.
	chatService := service.NewChatService(repo, ollamaProvider, settingsService)
	modelService := service.NewModelService(ollamaProvider)

	// API Handlers are instantiated with the services they depend on.
	// Go automatically recognizes that concrete types like `*service.ChatService`
	// satisfy the `interfaces.ChatService` expected by `NewChatHandler`.
	chatHandler := api.NewChatHandler(chatService, settingsService)
	modelHandler := api.NewModelHandler(modelService)

	// The router ties HTTP routes to specific handler methods.
	router := api.NewRouter(chatHandler, modelHandler)

	server := &http.Server{
		Addr:              ":8000",
		Handler:           router,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      0, // Disabled for streaming endpoints like chat messages.
		IdleTimeout:       120 * time.Second,
	}

	// Return the fully constructed (but not yet running) application.
	return &App{
		Config: cfg,
		DB:     db,
		Server: server,
	}, nil
}

// Run is the main entry point for the application. It orchestrates the entire lifecycle:
// configuration loading, application setup, and starting the HTTP server.
func Run() int {
	// 1. Load configuration from .env file and environment variables.
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use default logger as slog is not yet configured.
		slog.Error("Failed to load configuration", "error", err)
		return 1
	}

	// 2. Set up structured logging based on the config.
	setupLogger(cfg.LogLevel)
	logConfigSource()

	// 3. Initialize all application components.
	app, err := NewApp(cfg)
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		return 1
	}
	// Ensure the database connection is gracefully closed on exit.
	defer func() {
		if err := app.DB.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()

	// 4. Start the server and block until it's closed.
	slog.Info("Starting server", "port", 8000)
	if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server failed", "error", err)
		return 1
	}

	return 0
}

// logConfigSource logs whether the configuration was loaded from a file or from defaults/env vars.
// This is useful for debugging configuration issues.
func logConfigSource() {
	configFileUsed := viper.ConfigFileUsed()
	if configFileUsed != "" {
		slog.Info("Successfully loaded configuration from file.", "file", configFileUsed)
	} else {
		slog.Info("Configuration file not found. Using environment variables and defaults.")
	}
}

// setupLogger configures the global structured logger (`slog`) for the application.
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

// waitForOllama is a simple blocking health check. It ensures that the application
// does not start until its critical dependency (Ollama) is responsive.
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
		// Ensure the response body is always closed to prevent leaking connections,
		// even on non-200 responses.
		if resp != nil {
			if bErr := resp.Body.Close(); bErr != nil {
				slog.Warn("Failed to close response body in ollama health check (retry path)", "error", bErr)
			}
		}
		slog.Debug("Ollama not ready yet, retrying in 3 seconds...", "url", ollamaURL, "error", err)
		time.Sleep(3 * time.Second)
	}
}
