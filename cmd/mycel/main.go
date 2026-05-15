package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/J4NN0/mycel/internal/agent"
	"github.com/J4NN0/mycel/internal/cli"
	"github.com/J4NN0/mycel/internal/config"
	"github.com/J4NN0/mycel/internal/llm"
	"github.com/J4NN0/mycel/internal/logger"
	"github.com/J4NN0/mycel/internal/prompt"
	"github.com/J4NN0/mycel/internal/redis"
	"github.com/J4NN0/mycel/internal/telegram"
)

const (
	appName    = "Mycel"
	promptsDir = "prompts"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appConfig, err := config.ReadConfig()
	if err != nil {
		fmt.Printf("Config reading failed: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(appName, appConfig.LogLevel)

	llmProvider, err := llm.NewProvider(log, appConfig)
	if err != nil {
		log.Fatalf("llm provider initialization failed: %v", err)
		return
	}
	defer llmProvider.Shutdown()

	redisClient, err := redis.NewClient(ctx, appConfig.RedisAddr)
	if err != nil {
		log.Fatalf("redis initialization failed: %v", err)
		return
	}
	defer redisClient.Close()

	bot, err := telegram.NewBot(appConfig.TelegramBotToken, log)
	if err != nil {
		log.Fatalf("telegram bot initialization failed: %v", err)
		return
	}

	cliChat := cli.New(log)

	promptManager := prompt.NewManager(promptsDir, appConfig.Persona)

	ag := agent.New(log, llmProvider, redisClient, promptManager, appConfig.Objective, appConfig.MaxHistoryMessages, bot, cliChat)
	err = ag.Run(ctx)
	if err != nil {
		log.Fatalf("agent error: %v", err)
	}
}
