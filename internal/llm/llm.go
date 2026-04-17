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

type LLM struct {
	log      logger.Logger
	provider schemas.ModelProvider
	bifrost  *bifrost.Bifrost
}

func NewProvider(log logger.Logger, config config.Config, ollama *ollama.Ollama) (Provider, error) {
	pc := &providerConfig{
		config: config,
	}
	if pc.config.Provider == schemas.Ollama {
		if ollama == nil {
			return nil, fmt.Errorf("ollama config cannot be nil when provider is set to ollama")
		}

		err := ollama.Ensure("llama3.1:latest")
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

	return &LLM{
		log:      log,
		provider: config.Provider,
		bifrost:  bifrostClient,
	}, nil
}

func (l *LLM) Shutdown() {
	l.bifrost.Shutdown()
}
