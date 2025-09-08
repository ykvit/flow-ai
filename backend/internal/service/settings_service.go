package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"slices"
	"sort" // We need this to find the latest model
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
	settings, err := s.Get(ctx)
	if err == nil {
		log.Println("Found existing settings in database.")
		return settings, nil
	}

	log.Println("No settings found in database. Discovering models from Ollama...")
	
	var discoveredModel string
	
	models, err := s.llm.ListModels(ctx)
	if err != nil {
		log.Printf("WARN: Could not get model list from Ollama during init: %v. Using empty default.", err)
	}

	if models == nil || len(models.Models) == 0 {
		// Critical case: Ollama is running but has no models.
		log.Println("WARN: Ollama has no models! The application will run, but you must pull a model via the API before chat will work.")
		discoveredModel = "" // Set to empty string
	} else {
		// Sort models by modification date to find the most recent one.
		sort.Slice(models.Models, func(i, j int) bool {
			t1, _ := time.Parse(time.RFC3339, models.Models[i].ModifiedAt)
			t2, _ := time.Parse(time.RFC3339, models.Models[j].ModifiedAt)
			return t1.After(t2)
		})
		discoveredModel = models.Models[0].Name
		log.Printf("INFO: Discovered %d models in Ollama. Setting '%s' as default.", len(models.Models), discoveredModel)
	}

	initialSettings := &Settings{
		SystemPrompt: defaultSystemPrompt,
		MainModel:    discoveredModel,
		SupportModel: discoveredModel, 
	}

	if err := s.saveToDB(ctx, initialSettings); err != nil {
		return nil, fmt.Errorf("failed to save initial settings: %w", err)
	}
	
	log.Println("Successfully initialized and saved new settings to database.")
	return initialSettings, nil
}

// Get retrieves the current settings from the database.
func (s *SettingsService) Get(ctx context.Context) (*Settings, error) {
	query := "SELECT key, value FROM settings"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	settingsMap := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil { return nil, err }
		settingsMap[key] = value
	}

	if len(settingsMap) == 0 { return nil, fmt.Errorf("no settings found") }

	return &Settings{
		SystemPrompt: settingsMap["system_prompt"],
		MainModel:    settingsMap["main_model"],
		SupportModel: settingsMap["support_model"],
	}, nil
}

// Save is the public method for updating settings via API, which includes strict validation.
func (s *SettingsService) Save(ctx context.Context, settings *Settings) error {
	availableModels, err := s.llm.ListModels(ctx)
	if err != nil { return fmt.Errorf("could not list models from Ollama for validation: %w", err) } 
	
	modelNames := make([]string, len(availableModels.Models))
	for i, m := range availableModels.Models { modelNames[i] = m.Name }
	
	if !slices.Contains(modelNames, settings.MainModel) {
		return fmt.Errorf("main model '%s' is not available in Ollama", settings.MainModel)
	}
	if !slices.Contains(modelNames, settings.SupportModel) {
		return fmt.Errorf("support model '%s' is not available in Ollama", settings.SupportModel)
	}

	return s.saveToDB(ctx, settings)
}

// saveToDB is a private helper to persist settings without validation.
func (s *SettingsService) saveToDB(ctx context.Context, settings *Settings) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil { return err }
	defer tx.Rollback()

	settingsMap := map[string]string{
		"system_prompt": settings.SystemPrompt,
		"main_model":    settings.MainModel,
		"support_model": settings.SupportModel,
	}
	
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value")
	if err != nil { return err }
	defer stmt.Close()

	for key, value := range settingsMap {
		if _, err := stmt.ExecContext(ctx, key, value); err != nil { return err }
	}

	return tx.Commit()
}