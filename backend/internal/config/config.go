package config

import (
	"encoding/json"
	"os"
)

// BootstrapConfig now only contains settings needed before the database is available.
type BootstrapConfig struct {
	DatabasePath string `json:"database_path"`
	OllamaURL    string `json:"ollama_url"`
	SystemPrompt string `json:"system_prompt"`
}

func LoadBootstrapConfig() (*BootstrapConfig, error) {
	cfg := &BootstrapConfig{}

	cfg.DatabasePath = getEnv("DATABASE_PATH", "/app/data/flow.db")
	cfg.OllamaURL = getEnv("OLLAMA_URL", "http://localhost:11434")
	cfg.SystemPrompt = getEnv("INITIAL_SYSTEM_PROMPT", "You are a helpful assistant.")

	file, err := os.ReadFile("./config.json")
	if err == nil {
		var fileCfg BootstrapConfig
		if json.Unmarshal(file, &fileCfg) == nil {
			if cfg.DatabasePath == "/app/data/flow.db" && fileCfg.DatabasePath != "" {
				cfg.DatabasePath = fileCfg.DatabasePath
			}
			if cfg.OllamaURL == "http://localhost:11434" {
				cfg.OllamaURL = fileCfg.OllamaURL
			}
			if cfg.SystemPrompt == "You are a helpful assistant." {
				cfg.SystemPrompt = fileCfg.SystemPrompt
			}
		}
	}
	
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		cfg.DatabasePath = dbPath
	}
	if ollamaURL := os.Getenv("OLLAMA_URL"); ollamaURL != "" {
		cfg.OllamaURL = ollamaURL
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}