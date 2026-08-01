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
	promptManager      *prompt.Manager
	tools              []tool.Tool
	maxHistoryMessages int
	maxHistoryTokens   int
	platforms          []Platform
	mu                 sync.Mutex
}

func New(log logger.Logger, provider llm.Provider, history redis.List, promptManager *prompt.Manager, maxHistoryMessages, maxHistoryTokens int, agentTools []tool.Tool, platforms ...Platform) *Agent {
	return &Agent{
		log:                log,
		provider:           provider,
		history:            history,
		promptManager:      promptManager,
		tools:              agentTools,
		maxHistoryMessages: maxHistoryMessages,
		maxHistoryTokens:   maxHistoryTokens,
		platforms:          platforms,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(a.platforms))
	var wg sync.WaitGroup

	for _, p := range a.platforms {
		wg.Add(1)
		go func(p Platform) {
			defer wg.Done()
			defer cancel()
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
	a.mu.Lock()
	defer a.mu.Unlock()

	response, isCmd := a.handleCommand(ctx, sessionID, text)
	if isCmd {
		return response, nil
	}

	messages, err := a.loadHistory(ctx, sessionID)
	if err != nil {
		return "", err
	}
	messages = a.withToolPolicy(messages)
	messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleUser, Content: text})

	a.log.Debugf("[%s] Generating response ...", sessionID)
	result, err := a.provider.Chat(ctx, messages, a.tools...)
	if err != nil {
		return "", fmt.Errorf("agent reply: %w", err)
	}

	err = a.storeHistory(ctx, sessionID, text, result)
	if err != nil {
		a.log.Errorf("[%s] Failed to store history: %v", sessionID, err)
	}

	return result.Content, nil
}

func (a *Agent) withToolPolicy(messages []llm.Message) []llm.Message {
	if len(a.tools) == 0 {
		return messages
	}

	policy, err := a.promptManager.LoadTools()
	if err != nil {
		a.log.Errorf("Failed to load tool prompt: %v", err)
		return messages
	}

	policyMsg := llm.Message{Role: schemas.ChatMessageRoleSystem, Content: policy}
	if len(messages) == 0 {
		return []llm.Message{policyMsg}
	}

	withPolicy := make([]llm.Message, 0, len(messages)+1)
	withPolicy = append(withPolicy, messages[0], policyMsg)

	return append(withPolicy, messages[1:]...)
}
