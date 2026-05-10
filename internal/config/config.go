package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/maximhq/bifrost/core/schemas"
)

type Config struct {
	Provider           schemas.ModelProvider `envconfig:"PROVIDER" required:"true"`
	LlmModel           string                `envconfig:"LLM_MODEL" required:"true"`
	MaxHistoryMessages int                   `envconfig:"MAX_HISTORY_MESSAGES" default:"20"`
	Persona            string                `envconfig:"PERSONA" default:"influencer"`
	TelegramBotToken   string                `envconfig:"TELEGRAM_BOT_TOKEN" required:"true"`
	RedisAddr          string                `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	LogLevel           string                `envconfig:"LOG_LEVEL" default:"debug"`
}

func ReadConfig() (Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
