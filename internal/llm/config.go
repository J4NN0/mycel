package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/J4NN0/mycel/internal/config"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	// requestTimeout bounds a whole request.
	requestTimeout = 5 * time.Minute

	// streamIdleTimeout bounds the silence between chunks rather than the whole request.
	streamIdleTimeout = 3 * time.Minute
)

// providerConfig needs to implement GetConfiguredProviders, GetKeysForProvider, and GetConfigForProvider
type providerConfig struct {
	config config.General
}

func (p *providerConfig) GetConfiguredProviders() ([]schemas.ModelProvider, error) {
	return []schemas.ModelProvider{schemas.Ollama}, nil
}

func (p *providerConfig) GetKeysForProvider(ctx context.Context, providerKey schemas.ModelProvider) ([]schemas.Key, error) {
	if providerKey == schemas.Ollama {
		return []schemas.Key{{
			Value: schemas.EnvVar{Val: "ollama", FromEnv: false},
			OllamaKeyConfig: &schemas.OllamaKeyConfig{
				URL: schemas.EnvVar{
					Val:     ollamaBaseURL,
					FromEnv: false,
				},
			},
			Models: schemas.WhiteList{"*"},
			Weight: 1.0,
		}}, nil
	}
	return nil, fmt.Errorf("provider %s not supported", providerKey)
}

func (p *providerConfig) GetConfigForProvider(provider schemas.ModelProvider) (*schemas.ProviderConfig, error) {
	if provider == schemas.Ollama {
		return &schemas.ProviderConfig{
			NetworkConfig: schemas.NetworkConfig{
				BaseURL:                        ollamaBaseURL,
				DefaultRequestTimeoutInSeconds: int(requestTimeout.Seconds()),
				StreamIdleTimeoutInSeconds:     int(streamIdleTimeout.Seconds()),
			},
			ConcurrencyAndBufferSize: schemas.DefaultConcurrencyAndBufferSize,
		}, nil
	}
	return nil, fmt.Errorf("provider %s not supported", provider)
}
