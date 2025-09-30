package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"time"

	"flow-ai/backend/internal/llm"
)

// Settings holds the dynamic application settings.
type Settings struct {
	SystemPrompt string `json:"system_prompt"`
	MainModel    string `json:"main_model"`
	SupportModel string `json:"support_model"`
}

type SettingsService struct {
	db  *sql.DB
	llm llm.LLMProvider
}

func NewSettingsService(db *sql.DB, llmProvider llm.LLMProvider) *SettingsService {
	return &SettingsService{db: db, llm: llmProvider}
}

// InitAndGet performs "smart initialization" on first run.
func (s *SettingsService) InitAndGet(ctx context.Context, defaultSystemPrompt string) (*Settings, error) {
	_, err := s.getFromDB(ctx)
	if err == nil {
		slog.Info("Found existing settings in database. Initialization not needed.")
		// Self-heal settings if models were removed or not present on last run.
		return s.Get(ctx)
	}

	slog.Info("No settings found in database. Discovering models from Ollama for initial setup...")

	discoveredModel := s.findLatestModel(ctx)
	if discoveredModel == "" {
		slog.Warn("Ollama has no models during initial setup. Settings will have empty model names.")
	} else {
		slog.Info("Discovered models in Ollama", "default_model", discoveredModel)
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

// Get retrieves current settings, self-healing if necessary.
func (s *SettingsService) Get(ctx context.Context) (*Settings, error) {
	settings, err := s.getFromDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve settings from DB: %w. The application might need initialization", err)
	}

	needsUpdate := false
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

	if settings.SupportModel == "" && settings.MainModel != "" {
		slog.Info("Support model is not set. Defaulting to main model", "model", settings.MainModel)
		settings.SupportModel = settings.MainModel
		needsUpdate = true
	}

	if needsUpdate {
		slog.Info("Persisting auto-updated settings to the database...")
		if err := s.saveToDB(ctx, settings); err != nil {
			slog.Error("Failed to persist auto-updated settings", "error", err)
		}
	}

	return settings, nil
}

// Save validates and persists settings.
func (s *SettingsService) Save(ctx context.Context, settings *Settings) error {
	availableModels, err := s.llm.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("could not list models from Ollama for validation: %w", err)
	}

	modelNames := make([]string, len(availableModels.Models))
	for i, m := range availableModels.Models {
		modelNames[i] = m.Name
	}

	if settings.MainModel != "" && !slices.Contains(modelNames, settings.MainModel) {
		return fmt.Errorf("main model '%s' is not available in Ollama", settings.MainModel)
	}
	if settings.SupportModel != "" && !slices.Contains(modelNames, settings.SupportModel) {
		return fmt.Errorf("support model '%s' is not available in Ollama", settings.SupportModel)
	}

	return s.saveToDB(ctx, settings)
}

func (s *SettingsService) getFromDB(ctx context.Context) (*Settings, error) {
	query := "SELECT key, value FROM settings"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settingsMap := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settingsMap[key] = value
	}

	if len(settingsMap) == 0 {
		return nil, sql.ErrNoRows
	}

	return &Settings{
		SystemPrompt: settingsMap["system_prompt"],
		MainModel:    settingsMap["main_model"],
		SupportModel: settingsMap["support_model"],
	}, nil
}

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

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range settingsMap {
		if _, err := stmt.ExecContext(ctx, key, value); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SettingsService) findLatestModel(ctx context.Context) string {
	models, err := s.llm.ListModels(ctx)
	if err != nil {
		slog.Warn("Could not get model list from Ollama during discovery.", "error", err)
		return ""
	}

	if models == nil || len(models.Models) == 0 {
		return ""
	}

	sort.Slice(models.Models, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, models.Models[i].ModifiedAt)
		t2, _ := time.Parse(time.RFC3339, models.Models[j].ModifiedAt)
		return t1.After(t2)
	})
	return models.Models[0].Name
}
