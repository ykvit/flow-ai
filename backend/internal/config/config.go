package config

import (
	"encoding/json"
	"os"
)

type BootstrapConfig struct {
	RedisAddr     string `json:"redis_addr"`
	OllamaURL     string `json:"ollama_url"`
	SystemPrompt  string `json:"system_prompt"`
	MainModel     string `json:"main_model"`
	SupportModel  string `json:"support_model"`
}

// LoadBootstrapConfig now prioritizes environment variables over the config file.
func LoadBootstrapConfig() (*BootstrapConfig, error) {
	cfg := &BootstrapConfig{}

	// --- 1. Load from Environment Variables (Primary method for Docker) ---
	cfg.RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.OllamaURL = getEnv("OLLAMA_URL", "http://localhost:11434")
	cfg.SystemPrompt = getEnv("INITIAL_SYSTEM_PROMPT", "You are a helpful assistant.")
	cfg.MainModel = getEnv("INITIAL_MAIN_MODEL", "qwen:0.5b")
	cfg.SupportModel = getEnv("INITIAL_SUPPORT_MODEL", "qwen:0.5b")

	// --- 2. Attempt to load from file (Fallback for local development) ---
	// This allows local development without setting env vars, but Docker will use env vars.
	file, err := os.ReadFile("./config.json")
	if err == nil {
		// If file exists, unmarshal it. Values from the file will override
		// any defaults set above, but NOT values explicitly set by other env vars.
		var fileCfg BootstrapConfig
		if json.Unmarshal(file, &fileCfg) == nil {
			if cfg.RedisAddr == "localhost:6379" { cfg.RedisAddr = fileCfg.RedisAddr }
			if cfg.OllamaURL == "http://localhost:11434" { cfg.OllamaURL = fileCfg.OllamaURL }
			if cfg.SystemPrompt == "You are a helpful assistant." { cfg.SystemPrompt = fileCfg.SystemPrompt }
			if cfg.MainModel == "qwen:0.5b" { cfg.MainModel = fileCfg.MainModel }
			if cfg.SupportModel == "qwen:0.5b" { cfg.SupportModel = fileCfg.SupportModel }
		}
	}
	
	// Re-check env vars to ensure they have the highest priority
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" { cfg.RedisAddr = redisAddr }
	if ollamaURL := os.Getenv("OLLAMA_URL"); ollamaURL != "" { cfg.OllamaURL = ollamaURL }

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}