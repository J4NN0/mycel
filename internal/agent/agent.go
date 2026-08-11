package agent

import (
	"context"
	"errors"
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

type Input struct {
	Text   string
	Images []llm.Image
}

// MessageHandler is the callback a Platform uses to deliver an incoming message and receive the
// agent's reply. A non-nil onDelta also receives the reply text in fragments, as the model
// generates it (commands never stream, only normal chat replies do).
type MessageHandler func(ctx context.Context, sessionID string, input Input, onDelta llm.StreamFunc) (*Reply, error)

// Reply is what a Platform gets back for a turn.
// Text is a normal reply to display as-is.
// Conversations is set instead when the platform should let the user pick one — e.g. /resume with
// no argument — via whatever interactive UI it has; picking one is just another MessageHandler
// call with the chosen ID appended to the same command.
type Reply struct {
	Text          string
	Conversations []Conversation
}

type Agent struct {
	log                logger.Logger
	provider           llm.Provider
	history            redis.Store
	promptManager      *prompt.Manager
	tools              []tool.Tool
	maxHistoryMessages int
	maxHistoryTokens   int
	platforms          []Platform
	mu                 sync.Mutex
}

func New(log logger.Logger, provider llm.Provider, history redis.Store, promptManager *prompt.Manager, maxHistoryMessages, maxHistoryTokens int, agentTools []tool.Tool, platforms ...Platform) *Agent {
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

func (a *Agent) reply(ctx context.Context, sessionID string, input Input, onDelta llm.StreamFunc) (*Reply, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cmdReply, err := a.handleCommand(ctx, sessionID, input.Text)
	if err != nil {
		return nil, err
	}
	if cmdReply != nil {
		return cmdReply, nil
	}

	conversationID, err := a.history.ActiveConversation(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("resolve active conversation: %w", err)
	}
	messages, err := a.loadHistory(ctx, sessionID, conversationID)
	if err != nil {
		return nil, err
	}
	messages = a.withToolPolicy(messages)
	messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleUser, Content: input.Text, Images: input.Images})

	a.log.Debugf("[%s] Generating response ...", sessionID)
	result, err := a.provider.Chat(ctx, messages, onDelta, a.tools...)
	if err != nil {
		if errors.Is(err, llm.ErrVisionUnsupported) {
			return &Reply{Text: fmt.Sprintf("I can't look at images: %s has no vision support.", a.provider.Model())}, nil
		}
		return nil, fmt.Errorf("agent reply: %w", err)
	}

	a.log.Debugf("[%s] Storing history ...", sessionID)
	err = a.storeHistory(ctx, sessionID, conversationID, input, result)
	if err != nil {
		a.log.Errorf("[%s] Failed to store history: %v", sessionID, err)
	}

	return &Reply{Text: result.Content}, nil
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
