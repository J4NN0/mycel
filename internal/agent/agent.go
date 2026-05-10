package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/J4NN0/mycel/internal/logger"
	"github.com/J4NN0/mycel/internal/prompt"
	"github.com/J4NN0/mycel/internal/redis"
	"github.com/maximhq/bifrost/core/schemas"
)

// Platform represents any channel the agent can communicate through.
type Platform interface {
	Run(ctx context.Context, handler MessageHandler) error
}

// MessageHandler is the callback a Platform uses to deliver an incoming message and receive the agent's reply.
type MessageHandler func(ctx context.Context, sessionID, text string) (string, error)

type Agent struct {
	log                logger.Logger
	provider           llm.Provider
	history            redis.History
	prompts            *prompt.Manager
	maxHistoryMessages int
	platforms          []Platform
}

func New(log logger.Logger, provider llm.Provider, history redis.History, prompts *prompt.Manager, maxHistoryMessages int, platforms ...Platform) *Agent {
	return &Agent{
		log:                log,
		provider:           provider,
		history:            history,
		prompts:            prompts,
		maxHistoryMessages: maxHistoryMessages,
		platforms:          platforms,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	errCh := make(chan error, len(a.platforms))
	var wg sync.WaitGroup

	for _, p := range a.platforms {
		wg.Add(1)
		go func(p Platform) {
			defer wg.Done()
			err := p.Run(ctx, a.reply)
			if err != nil {
				a.log.Errorf("platform %T failed: %v", p, err)
				errCh <- err
			}
		}(p)
	}

	wg.Wait()
	close(errCh)

	return <-errCh
}

func (a *Agent) reply(ctx context.Context, sessionID, text string) (string, error) {
	response, isCmd := a.handleCommand(ctx, sessionID, text)
	if isCmd {
		return response, nil
	}

	messages, err := a.loadHistory(ctx, sessionID)
	if err != nil {
		return "", err
	}
	messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleUser, Content: text})

	a.log.Debugf("[%s] Generating response ...", sessionID)
	response, err = a.provider.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("agent reply: %w", err)
	}

	err = a.storeHistory(ctx, sessionID, text, response)
	if err != nil {
		a.log.Errorf("[%s] Failed to store history: %v", sessionID, err)
	}

	return response, nil
}
