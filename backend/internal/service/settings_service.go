package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices" // Go 1.21+ package for checking slice contents
	"flow-ai/backend/internal/config"
	"flow-ai/backend/internal/llm"
	
	"github.com/redis/go-redis/v9"
)

const settingsKey = "settings"

// Settings holds the dynamic application settings stored in Redis.
type Settings struct {
	SystemPrompt string `json:"system_prompt"`
	MainModel    string `json:"main_model"`
	SupportModel string `json:"support_model"`
}

type SettingsService struct {
	rdb *redis.Client
	llm llm.LLMProvider
}

// NewSettingsService now accepts an LLMProvider.
func NewSettingsService(rdb *redis.Client, llmProvider llm.LLMProvider) *SettingsService {
	return &SettingsService{rdb: rdb, llm: llmProvider}
}

// InitAndGet implements the "smart initialization" logic.
func (s *SettingsService) InitAndGet(ctx context.Context, bootstrap *config.BootstrapConfig) (*Settings, error) {
	val, err := s.rdb.Get(ctx, settingsKey).Result()
	if err == nil {
		// Settings found, unmarshal and return them.
		var settings Settings
		if err := json.Unmarshal([]byte(val), &settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal existing settings: %w", err)
		}
		log.Println("Found existing settings in Redis.")
		return &settings, nil
	}

	if err != redis.Nil {
		// An actual error occurred while talking to Redis.
		return nil, fmt.Errorf("failed to get settings from redis: %w", err)
	}

	// --- Settings not found in Redis (err == redis.Nil), perform smart initialization ---
	log.Println("No settings found in Redis. Performing smart initialization...")
	
	defaultModel := bootstrap.MainModel
	models, err := s.llm.ListModels(ctx)
	if err != nil {
		log.Printf("WARN: Could not connect to Ollama to get model list during init: %v. Using bootstrap default.", err)
	} else if len(models.Models) == 0 {
		log.Printf("WARN: Ollama is running but has no models. Using bootstrap default.")
	} else {
		defaultModel = models.Models[0].Name
		log.Printf("INFO: Automatically selected '%s' as the default model from Ollama.", defaultModel)
	}

	initialSettings := &Settings{
		SystemPrompt: bootstrap.SystemPrompt,
		MainModel:    defaultModel,
		SupportModel: defaultModel,
	}

	// Use the internal save method which doesn't do validation for initialization
	if err := s.saveToRedis(ctx, initialSettings); err != nil {
		return nil, fmt.Errorf("failed to save initial settings: %w", err)
	}
	
	log.Println("Successfully initialized and saved new settings to Redis.")
	return initialSettings, nil
}

// Get retrieves the current settings from Redis.
func (s *SettingsService) Get(ctx context.Context) (*Settings, error) {
	val, err := s.rdb.Get(ctx, settingsKey).Result()
	if err != nil {
		return nil, err
	}
	var settings Settings
	if err := json.Unmarshal([]byte(val), &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// Save is the public method for saving settings, which includes validation.
func (s *SettingsService) Save(ctx context.Context, settings *Settings) error {
	availableModels, err := s.llm.ListModels(ctx)
	if err != nil {
		log.Printf("WARN: Could not list models for validation, saving settings without check: %v", err)
	} else {
		modelNames := make([]string, len(availableModels.Models))
		for i, m := range availableModels.Models {
			modelNames[i] = m.Name
		}
		
		if !slices.Contains(modelNames, settings.MainModel) {
			return fmt.Errorf("main model '%s' not found in Ollama", settings.MainModel)
		}
		if !slices.Contains(modelNames, settings.SupportModel) {
			return fmt.Errorf("support model '%s' not found in Ollama", settings.SupportModel)
		}
	}

	return s.saveToRedis(ctx, settings)
}

// saveToRedis is a private helper to just save the data.
func (s *SettingsService) saveToRedis(ctx context.Context, settings *Settings) error {
	val, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	return s.rdb.Set(ctx, settingsKey, val, 0).Err()
}