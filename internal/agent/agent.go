package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/J4NN0/mycel/internal/logger"
	"github.com/maximhq/bifrost/core/schemas"
)

// Platform represents any channel the agent can communicate through.
type Platform interface {
	Run(ctx context.Context, handler MessageHandler) error
}

// MessageHandler is the callback a Platform uses to deliver an incoming message and receive the agent's reply.
type MessageHandler func(ctx context.Context, sessionID, text string) (string, error)

type Agent struct {
	log       logger.Logger
	provider  llm.Provider
	persona   string
	platforms []Platform
	mu        sync.Mutex
	histories map[string][]llm.Message
}

func New(log logger.Logger, provider llm.Provider, persona string, platforms ...Platform) *Agent {
	return &Agent{
		log:       log,
		provider:  provider,
		persona:   persona,
		platforms: platforms,
		histories: make(map[string][]llm.Message),
	}
}

func (a *Agent) Run(ctx context.Context) error {
	errCh := make(chan error, len(a.platforms))
	var wg sync.WaitGroup

	for _, p := range a.platforms {
		wg.Add(1)
		go func(p Platform) {
			defer wg.Done()
			if err := p.Run(ctx, a.reply); err != nil {
				a.log.Warningf("Platform %T failed: %v", p, err)
				errCh <- err
			}
		}(p)
	}

	wg.Wait()
	close(errCh)

	return <-errCh
}

func (a *Agent) reply(ctx context.Context, sessionID, text string) (string, error) {
	response, isCmd := a.handleCommand(sessionID, text)
	if isCmd {
		return response, nil
	}

	a.mu.Lock()
	messages := a.loadHistory(sessionID)
	a.mu.Unlock()

	messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleUser, Content: text})

	response, err := a.provider.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("agent reply: %w", err)
	}

	a.mu.Lock()
	a.storeHistory(sessionID, text, response)
	a.mu.Unlock()

	return response, nil
}

func (a *Agent) loadHistory(sessionID string) []llm.Message {
	if len(a.histories[sessionID]) == 0 {
		a.histories[sessionID] = []llm.Message{{Role: schemas.ChatMessageRoleSystem, Content: a.persona}}
	}

	messages := make([]llm.Message, len(a.histories[sessionID]))
	copy(messages, a.histories[sessionID])

	return messages
}

func (a *Agent) storeHistory(sessionID, text, response string) {
	a.histories[sessionID] = append(a.histories[sessionID],
		llm.Message{Role: schemas.ChatMessageRoleUser, Content: text},
		llm.Message{Role: schemas.ChatMessageRoleAssistant, Content: response},
	)
}

func (a *Agent) clearHistory(sessionID string) {
	a.mu.Lock()
	delete(a.histories, sessionID)
	a.mu.Unlock()
}
