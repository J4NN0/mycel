package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/J4NN0/mycel/internal/agent"
	"github.com/J4NN0/mycel/internal/cli"
	"github.com/J4NN0/mycel/internal/config"
	"github.com/J4NN0/mycel/internal/llm"
	"github.com/J4NN0/mycel/internal/logger"
	"github.com/J4NN0/mycel/internal/prompt"
	"github.com/J4NN0/mycel/internal/telegram"
)

const (
	appName    = "Mycel"
	promptsDir = "prompts"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log := logger.New(appName)

	appConfig, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("config reading failed: %v", err)
		return
	}

	llmProvider, err := llm.NewProvider(log, appConfig)
	if err != nil {
		log.Fatalf("provider initialization failed: %v", err)
		return
	}
	defer llmProvider.Shutdown()

	bot, err := telegram.NewBot(appConfig.TelegramBotToken, log)
	if err != nil {
		log.Fatalf("telegram bot initialization failed: %v", err)
		return
	}

	cli := cli.New(log)

	promptManager := prompt.NewManager(promptsDir)
	persona, err := promptManager.LoadPersona()
	if err != nil {
		log.Fatalf("prompt loading failed: %v", err)
		return
	}

	ag := agent.New(log, llmProvider, persona, bot, cli)
	err = ag.Run(ctx)
	if err != nil {
		log.Fatalf("agent error: %v", err)
	}
}
