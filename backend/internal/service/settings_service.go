package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

// NewSettingsService now accepts a sql.DB connection.
func NewSettingsService(db *sql.DB, llmProvider llm.LLMProvider) *SettingsService {
	return &SettingsService{db: db, llm: llmProvider}
}

// InitAndGet implements the "smart initialization" logic.
func (s *SettingsService) InitAndGet(ctx context.Context, defaultSystemPrompt string) (*Settings, error) {
	// Try to get settings. It's okay if they don't exist yet.
	_, err := s.getFromDB(ctx)
	if err == nil {
		log.Println("Found existing settings in database. Initialization not needed.")
		// We still call Get() to ensure settings are valid and self-healed if needed.
		return s.Get(ctx)
	}

	log.Println("No settings found in database. Discovering models from Ollama for initial setup...")

	discoveredModel := s.findLatestModel(ctx)
	if discoveredModel == "" {
		log.Println("WARN: Ollama has no models during initial setup. Settings will have empty model names.")
	} else {
		log.Printf("INFO: Discovered %d models in Ollama. Setting '%s' as default.", 1, discoveredModel)
	}

	initialSettings := &Settings{
		SystemPrompt: defaultSystemPrompt,
		MainModel:    discoveredModel,
		SupportModel: discoveredModel, // Default support model is the same as main.
	}

	if err := s.saveToDB(ctx, initialSettings); err != nil {
		return nil, fmt.Errorf("failed to save initial settings: %w", err)
	}

	log.Println("Successfully initialized and saved new settings to database.")
	return initialSettings, nil
}

// Get retrieves the current settings from the database.
// NEW: This method is now "self-healing". If it finds empty model settings,
// it tries to discover available models and update the settings automatically.
func (s *SettingsService) Get(ctx context.Context) (*Settings, error) {
	settings, err := s.getFromDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve settings from DB: %w. The application might need initialization", err)
	}

	needsUpdate := false

	// If the main model is not set, try to find one.
	// This happens if the app started before any models were pulled.
	if settings.MainModel == "" {
		log.Println("Main model is not set. Attempting to auto-discover from Ollama...")
		discoveredModel := s.findLatestModel(ctx)
		if discoveredModel != "" {
			settings.MainModel = discoveredModel
			log.Printf("Discovered and set main model to: %s", discoveredModel)
			needsUpdate = true
		} else {
			log.Println("WARN: Could not discover any models to set as main model.")
		}
	}

	// If the support model is not set, default it to the main model.
	if settings.SupportModel == "" && settings.MainModel != "" {
		log.Printf("Support model is not set. Defaulting to main model: %s", settings.MainModel)
		settings.SupportModel = settings.MainModel
		needsUpdate = true
	}

	// If we made changes, save them back to the database.
	if needsUpdate {
		log.Println("Persisting auto-updated settings to the database...")
		if err := s.saveToDB(ctx, settings); err != nil {
			// Log the error but return the in-memory settings so the app can continue.
			log.Printf("ERROR: Failed to persist auto-updated settings: %v", err)
		}
	}

	return settings, nil
}

// Save is the public method for updating settings via API, which includes strict validation.
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

// getFromDB is a private helper to fetch settings without extra logic.
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
		return nil, sql.ErrNoRows // Use a standard error for "not found"
	}

	return &Settings{
		SystemPrompt: settingsMap["system_prompt"],
		MainModel:    settingsMap["main_model"],
		SupportModel: settingsMap["support_model"],
	}, nil
}

// saveToDB is a private helper to persist settings without validation.
func (s *SettingsService) saveToDB(ctx context.Context, settings *Settings) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

// findLatestModel is a helper to get the most recently modified model from Ollama.
func (s *SettingsService) findLatestModel(ctx context.Context) string {
	models, err := s.llm.ListModels(ctx)
	if err != nil {
		log.Printf("WARN: Could not get model list from Ollama during discovery: %v.", err)
		return ""
	}

	if models == nil || len(models.Models) == 0 {
		return ""
	}

	// Sort models by modification date to find the most recent one.
	sort.Slice(models.Models, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, models.Models[i].ModifiedAt)
		t2, _ := time.Parse(time.RFC3339, models.Models[j].ModifiedAt)
		return t1.After(t2)
	})
	return models.Models[0].Name
}