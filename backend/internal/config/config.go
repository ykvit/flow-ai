package config

import (
	"encoding/json"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	RedisAddr     string `json:"redis_addr"`
	OllamaURL     string `json:"ollama_url"`
	SystemPrompt  string `json:"system_prompt"`
	MainModel     string `json:"main_model"`
	SupportModel  string `json:"support_model"`
}

// Load reads configuration from environment variables.
// This is a secure and standard way to configure Docker applications.
func Load() (*Config, error) {
	// A default config file is loaded
	file, err := os.ReadFile("./config.json")
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = json.Unmarshal(file, &cfg)

	// You can override file settings with environment variables
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.RedisAddr = redisAddr
	}
	if ollamaURL := os.Getenv("OLLAMA_URL"); ollamaURL != "" {
		cfg.OllamaURL = ollamaURL
	}
	
	return cfg, err
}