package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/maximhq/bifrost/core/schemas"
)

const envFileName = ".env"

type Config struct {
	General  General
	Platform Platform
	Tool     Tool
}

type General struct {
	Provider           schemas.ModelProvider `envconfig:"PROVIDER" required:"true"`
	LlmModel           string                `envconfig:"LLM_MODEL" required:"true"`
	Persona            string                `envconfig:"PERSONA" default:"neutral"`
	MaxHistoryMessages int                   `envconfig:"MAX_HISTORY_MESSAGES" default:"20"`
	MaxHistoryTokens   int                   `envconfig:"MAX_HISTORY_TOKENS" default:"6000"`
	RedisAddr          string                `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	LogLevel           string                `envconfig:"LOG_LEVEL" default:"info"`
}

func ReadConfig() (Config, error) {
	if err := loadEnvFile(); err != nil {
		return Config{}, err
	}

	cfg := Config{}
	for _, section := range []any{&cfg.General, &cfg.Platform, &cfg.Tool} {
		if err := envconfig.Process("", section); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

// loadEnvFile loads ~/.config/mycel/.env into the process environment, if present,
// so `mycel` can be run from any directory once installed. Variables already set
// in the environment (e.g. exported in the shell) are left untouched.
func loadEnvFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}

	path := filepath.Join(home, ".config", "mycel", envFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return godotenv.Load(path)
}
