package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/J4NN0/mycel/internal/logger"
	"github.com/J4NN0/mycel/internal/prompt"
	"github.com/J4NN0/mycel/internal/redis"
	"github.com/J4NN0/mycel/internal/tool"
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
	history            redis.List
	prompts            *prompt.Manager
	tools              []tool.Tool
	objective          string
	maxHistoryMessages int
	platforms          []Platform
}

func New(log logger.Logger, provider llm.Provider, history redis.List, prompts *prompt.Manager, objective string, maxHistoryMessages int, agentTools []tool.Tool, platforms ...Platform) *Agent {
	return &Agent{
		log:                log,
		provider:           provider,
		history:            history,
		prompts:            prompts,
		tools:              agentTools,
		objective:          objective,
		maxHistoryMessages: maxHistoryMessages,
		platforms:          platforms,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	errCh := make(chan error, len(a.platforms))
	var wg sync.WaitGroup

	if a.objective != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.runObjective(ctx)
		}()
	}

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
	response, err = a.provider.Chat(ctx, messages, a.tools...)
	if err != nil {
		return "", fmt.Errorf("agent reply: %w", err)
	}

	err = a.storeHistory(ctx, sessionID, text, response)
	if err != nil {
		a.log.Errorf("[%s] Failed to store history: %v", sessionID, err)
	}

	return response, nil
}
