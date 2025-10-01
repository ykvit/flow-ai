package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort             int    `mapstructure:"APP_PORT"`
	DatabasePath        string `mapstructure:"DATABASE_PATH"`
	OllamaURL           string `mapstructure:"OLLAMA_URL"`
	InitialSystemPrompt string `mapstructure:"INITIAL_SYSTEM_PROMPT"`
	LogLevel            string `mapstructure:"LOG_LEVEL"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("APP_PORT", 3000)
	viper.SetDefault("DATABASE_PATH", "/data/flow.db")
	viper.SetDefault("OLLAMA_URL", "http://ollama:11434")
	viper.SetDefault("INITIAL_SYSTEM_PROMPT", "You are a helpful assistant.")
	viper.SetDefault("LOG_LEVEL", "INFO")

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./backend")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {

			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
