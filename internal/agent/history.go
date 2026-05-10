package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/maximhq/bifrost/core/schemas"
)

const keepRecentMessages = 4

func (a *Agent) loadHistory(ctx context.Context, sessionID string) ([]llm.Message, error) {
	messages, err := a.history.Load(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}
	if len(messages) == 0 {
		persona, err := a.prompts.LoadPersona()
		if err != nil {
			return nil, fmt.Errorf("load persona: %w", err)
		}
		messages = []llm.Message{{Role: schemas.ChatMessageRoleSystem, Content: persona}}
		err = a.history.Save(ctx, sessionID, messages)
		if err != nil {
			return nil, fmt.Errorf("seed history: %w", err)
		}
	}

	a.log.Debugf("[%s] History loaded: %d message(s)", sessionID, len(messages))

	return messages, nil
}

func (a *Agent) storeHistory(ctx context.Context, sessionID, text, response string) error {
	messages, err := a.history.Load(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}

	messages = append(messages,
		llm.Message{Role: schemas.ChatMessageRoleUser, Content: text},
		llm.Message{Role: schemas.ChatMessageRoleAssistant, Content: response},
	)

	err = a.history.Save(ctx, sessionID, messages)
	if err != nil {
		return err
	}

	if len(messages) > a.maxHistoryMessages {
		err = a.compactHistory(ctx, sessionID, messages)
		if err != nil {
			a.log.Errorf("[%s] Failed to compact history: %v", sessionID, err)
		}
	}

	return nil
}

func (a *Agent) clearHistory(ctx context.Context, sessionID string) error {
	err := a.history.Clear(ctx, sessionID)
	if err != nil {
		return err
	}

	a.log.Debugf("[%s] History cleared", sessionID)

	return nil
}

func (a *Agent) compactHistory(ctx context.Context, sessionID string, messages []llm.Message) error {
	a.log.Debugf("[%s] Compacting history (%d messages) ...", sessionID, len(messages))

	var sb strings.Builder
	for _, m := range messages[1:] { // skip system prompt
		sb.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}

	summaryRequest := []llm.Message{
		{
			Role:    schemas.ChatMessageRoleSystem,
			Content: a.prompts.LoadCompact(),
		},
		{
			Role:    schemas.ChatMessageRoleUser,
			Content: sb.String(),
		},
	}

	summary, err := a.provider.Chat(ctx, summaryRequest)
	if err != nil {
		return fmt.Errorf("summarize history: %w", err)
	}

	personaSystemPrompt := messages[0]
	compacted := []llm.Message{
		personaSystemPrompt,
		{Role: schemas.ChatMessageRoleSystem, Content: "Summary of previous conversation:\n" + summary},
	}
	compacted = append(compacted, messages[len(messages)-keepRecentMessages:]...)

	err = a.history.Save(ctx, sessionID, compacted)
	if err != nil {
		return fmt.Errorf("save compacted history: %w", err)
	}

	a.log.Debugf("[%s] History compacted to %d messages", sessionID, len(compacted))

	return nil
}
