package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/maximhq/bifrost/core/schemas"
)

type Config struct {
	Provider schemas.ModelProvider `envconfig:"PROVIDER" required:"true"`
	LlmModel string                `envconfig:"LLM_MODEL" required:"true"`
}

func ReadConfig() (Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
