package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"time"

	app_errors "flow-ai/backend/internal/errors"
	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/repository"
)

// Settings holds the dynamic, user-configurable application settings.
type Settings struct {
	SystemPrompt string `json:"system_prompt" example:"You are a helpful assistant that always answers in Markdown format."`
	// The primary model for new chats. Must be an available local model.
	MainModel string `json:"main_model" validate:"required" example:"qwen3:8b"`
	// A model for background tasks like title generation. Can be the same as the main model.
	SupportModel string `json:"support_model" example:"gemma3:4b"`
}

// SettingsService provides methods for managing application settings.
// It includes logic for smart initialization and self-healing.
type SettingsService struct {
	db  *sql.DB
	llm llm.LLMProvider
}

// NewSettingsService creates a new instance of SettingsService.
func NewSettingsService(db *sql.DB, llmProvider llm.LLMProvider) *SettingsService {
	return &SettingsService{db: db, llm: llmProvider}
}

// InitAndGet performs a "smart initialization" on the first application run.
// If settings are not found in the database, it discovers available Ollama models
// and creates a default configuration.
func (s *SettingsService) InitAndGet(ctx context.Context, defaultSystemPrompt string) (*Settings, error) {
	_, err := s.getFromDB(ctx)
	// If settings already exist, no initialization is needed.
	if err == nil {
		slog.Info("Found existing settings in database. Initialization not needed.")
		return s.Get(ctx)
	}

	slog.Info("No settings found in database. Discovering models from Ollama for initial setup...")

	discoveredModel := s.findLatestModel(ctx)
	if discoveredModel == "" {
		slog.Warn("Ollama has no models during initial setup. Settings will have empty model names.")
	} else {
		slog.Info("Discovered models in Ollama, selecting the most recent as default.", "default_model", discoveredModel)
	}

	initialSettings := &Settings{
		SystemPrompt: defaultSystemPrompt,
		MainModel:    discoveredModel,
		SupportModel: discoveredModel,
	}

	if err := s.saveToDB(ctx, initialSettings); err != nil {
		return nil, fmt.Errorf("failed to save initial settings: %w", err)
	}

	slog.Info("Successfully initialized and saved new settings to database.")
	return initialSettings, nil
}

// Get retrieves current settings. It includes "self-healing" logic to automatically
// select a model if the configured one is missing or not set.
func (s *SettingsService) Get(ctx context.Context) (*Settings, error) {
	settings, err := s.getFromDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve settings from DB: %w. The application might need initialization", err)
	}

	needsUpdate := false
	// Self-heal: If the main model is not set, try to find one.
	if settings.MainModel == "" {
		slog.Info("Main model is not set. Attempting to auto-discover from Ollama...")
		discoveredModel := s.findLatestModel(ctx)
		if discoveredModel != "" {
			settings.MainModel = discoveredModel
			slog.Info("Discovered and set main model", "model", discoveredModel)
			needsUpdate = true
		} else {
			slog.Warn("Could not discover any models to set as main model.")
		}
	}

	// Self-heal: If the support model is not set, default to the main model.
	if settings.SupportModel == "" && settings.MainModel != "" {
		slog.Info("Support model is not set. Defaulting to main model", "model", settings.MainModel)
		settings.SupportModel = settings.MainModel
		needsUpdate = true
	}

	if needsUpdate {
		slog.Info("Persisting auto-updated settings to the database...")
		// This save is best-effort; a failure here is logged but not returned as a critical error.
		if err := s.saveToDB(ctx, settings); err != nil {
			slog.Error("Failed to persist auto-updated settings", "error", err)
		}
	}

	return settings, nil
}

// Save validates the provided settings against available Ollama models and persists them.
func (s *SettingsService) Save(ctx context.Context, settings *Settings) error {
	availableModels, err := s.llm.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("could not list models from Ollama for validation: %w", err)
	}

	modelNames := make([]string, len(availableModels.Models))
	for i, m := range availableModels.Models {
		modelNames[i] = m.Name
	}

	// Ensure the selected models actually exist locally.
	if settings.MainModel != "" && !slices.Contains(modelNames, settings.MainModel) {
		return fmt.Errorf("%w: main model '%s' is not available in Ollama", app_errors.ErrValidation, settings.MainModel)
	}
	if settings.SupportModel != "" && !slices.Contains(modelNames, settings.SupportModel) {
		return fmt.Errorf("%w: support model '%s' is not available in Ollama", app_errors.ErrValidation, settings.SupportModel)
	}

	return s.saveToDB(ctx, settings)
}

// getFromDB is a private helper for retrieving settings from the key-value table.
func (s *SettingsService) getFromDB(ctx context.Context) (*Settings, error) {
	query := "SELECT key, value FROM settings"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Failed to close rows in getFromDB", "error", err)
		}
	}()

	settingsMap := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settingsMap[key] = value
	}

	// If the map is empty, it means the settings table has no rows.
	if len(settingsMap) == 0 {
		return nil, repository.ErrNotFound
	}

	return &Settings{
		SystemPrompt: settingsMap["system_prompt"],
		MainModel:    settingsMap["main_model"],
		SupportModel: settingsMap["support_model"],
	}, nil
}

// saveToDB is a private helper for persisting settings using an UPSERT operation.
func (s *SettingsService) saveToDB(ctx context.Context, settings *Settings) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			slog.Error("Failed to rollback save settings transaction", "error", err)
		}
	}()

	settingsMap := map[string]string{
		"system_prompt": settings.SystemPrompt,
		"main_model":    settings.MainModel,
		"support_model": settings.SupportModel,
	}

	// ADD THIS BLOCK TO MAKE THE ORDER DETERMINISTIC
	keys := make([]string, 0, len(settingsMap))
	for k := range settingsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value")
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			slog.Error("Failed to close statement in saveToDB", "error", err)
		}
	}()

	// ITERATE OVER SORTED KEYS
	for _, key := range keys {
		if _, err := stmt.ExecContext(ctx, key, settingsMap[key]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// findLatestModel discovers available Ollama models and returns the name of the
// most recently modified one.
func (s *SettingsService) findLatestModel(ctx context.Context) string {
	models, err := s.llm.ListModels(ctx)
	if err != nil {
		slog.Warn("Could not get model list from Ollama during discovery.", "error", err)
		return ""
	}

	if models == nil || len(models.Models) == 0 {
		return ""
	}

	// Sort models by modification date, descending, to find the newest one.
	sort.Slice(models.Models, func(i, j int) bool {
		// Parsing errors are ignored for simplicity; a zero-time will sort incorrectly but won't crash.
		t1, _ := time.Parse(time.RFC3339, models.Models[i].ModifiedAt)
		t2, _ := time.Parse(time.RFC3339, models.Models[j].ModifiedAt)
		return t1.After(t2)
	})
	return models.Models[0].Name
}
