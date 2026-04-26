package llm

import (
	"context"
	"fmt"

	"github.com/J4NN0/mycel/internal/config"
	"github.com/J4NN0/mycel/internal/logger"
	"github.com/J4NN0/mycel/internal/ollama"
	"github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
)

type Provider interface {
	Chat(ctx context.Context, msg string) (string, error)
	Shutdown()
}

type llm struct {
	log      logger.Logger
	provider schemas.ModelProvider
	bifrost  *bifrost.Bifrost
	model    string
}

func NewProvider(log logger.Logger, config config.Config, ollamaChecker ollama.Checker) (Provider, error) {
	pc := &providerConfig{
		config: config,
	}
	if pc.config.Provider == schemas.Ollama {
		err := ollamaChecker.Ensure(config.LlmModel)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure llama provider: %w", err)
		}
	}

	bifrostClient, err := bifrost.Init(context.Background(), schemas.BifrostConfig{
		Account: pc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bifrost client: %w", err)
	}

	return &llm{
		log:      log,
		provider: config.Provider,
		bifrost:  bifrostClient,
		model:    config.LlmModel,
	}, nil
}

func (l *llm) Shutdown() {
	l.bifrost.Shutdown()
}
