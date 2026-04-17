package main

import (
	"context"

	"github.com/J4NN0/mycel/internal/config"
	"github.com/J4NN0/mycel/internal/llm"
	"github.com/J4NN0/mycel/internal/logger"
)

const appName = "Mycel"

func main() {
	ctx := context.Background()
	log := logger.New(appName)

	appConfig, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("config reading failed: %v", err)
		return
	}

	provider, err := llm.NewProvider(log, appConfig)
	if err != nil {
		log.Fatalf("provider initilization failed: %v", err)
		return
	}
	defer provider.Shutdown()

	response, err := provider.Chat(ctx, "Hello!")
	if err != nil {
		log.Fatalf("chat message failed: %v", err)
		return
	}
	log.Printf(response)
}
