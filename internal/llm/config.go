package llm

import (
	"context"
	"fmt"

	"github.com/J4NN0/mycel/internal/config"
	"github.com/maximhq/bifrost/core/schemas"
)

// providerConfig needs to implement GetConfiguredProviders, GetKeysForProvider, and GetConfigForProvider
type providerConfig struct {
	config config.Config
}

func (p *providerConfig) GetConfiguredProviders() ([]schemas.ModelProvider, error) {
	return []schemas.ModelProvider{schemas.Ollama}, nil
}

func (p *providerConfig) GetKeysForProvider(ctx context.Context, providerKey schemas.ModelProvider) ([]schemas.Key, error) {
	if providerKey == schemas.Ollama {
		return []schemas.Key{{
			Value:  schemas.EnvVar{},
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
				BaseURL:                        "http://localhost:11434",
				DefaultRequestTimeoutInSeconds: 30,
			},
			ConcurrencyAndBufferSize: schemas.DefaultConcurrencyAndBufferSize,
		}, nil
	}
	return nil, fmt.Errorf("provider %s not supported", provider)
}
