package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	keepRecentMessages = 4
	summaryPrefix      = "Summary of previous conversation:\n"
)

func (a *Agent) loadHistory(ctx context.Context, sessionID, conversationID string) ([]llm.Message, error) {
	messages, err := a.history.Load(ctx, sessionID, conversationID)
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}
	if len(messages) == 0 {
		systemPrompt, err := a.promptManager.LoadSystem()
		if err != nil {
			return nil, fmt.Errorf("load system prompt: %w", err)
		}

		messages = []llm.Message{{Role: schemas.ChatMessageRoleSystem, Content: systemPrompt}}
		err = a.history.Append(ctx, sessionID, conversationID, messages...)
		if err != nil {
			return nil, fmt.Errorf("seed history: %w", err)
		}
	}

	a.log.Debugf("[%s] History loaded: %d message(s)", sessionID, len(messages))

	return messages, nil
}

func (a *Agent) storeHistory(ctx context.Context, sessionID, conversationID, text string, result llm.Response) error {
	err := a.history.Append(ctx, sessionID, conversationID,
		llm.Message{Role: schemas.ChatMessageRoleUser, Content: text},
		llm.Message{Role: schemas.ChatMessageRoleAssistant, Content: result.Content},
	)
	if err != nil {
		return fmt.Errorf("append history: %w", err)
	}

	count, err := a.history.Len(ctx, sessionID, conversationID)
	if err != nil {
		return fmt.Errorf("count history: %w", err)
	}

	if a.shouldCompact(count, result.PromptTokens) {
		messages, err := a.history.Load(ctx, sessionID, conversationID)
		if err != nil {
			return fmt.Errorf("load history: %w", err)
		}

		err = a.compactHistory(ctx, sessionID, conversationID, messages)
		if err != nil {
			a.log.Errorf("[%s] Failed to compact history: %v", sessionID, err)
		}
	}

	return nil
}

func (a *Agent) shouldCompact(count int64, promptTokens int) bool {
	// A compacted history is system prompt + summary + the retained tail. At or
	// below that size there is no older content for compaction to shrink.
	const systemAndSummary = 2
	if count <= int64(keepRecentMessages+systemAndSummary) {
		return false
	}
	// Use the real context size when the provider reports it; otherwise fall
	// back to a message-count cap so history can't grow unbounded.
	if a.maxHistoryTokens > 0 && promptTokens > 0 {
		return promptTokens > a.maxHistoryTokens
	}
	return count > int64(a.maxHistoryMessages)
}

func (a *Agent) startNewConversation(ctx context.Context, sessionID string) error {
	conversationID, err := a.history.NewConversation(ctx, sessionID)
	if err != nil {
		return err
	}

	a.log.Debugf("[%s] Started new conversation %s", sessionID, conversationID)

	return nil
}

func (a *Agent) compactHistory(ctx context.Context, sessionID, conversationID string, messages []llm.Message) error {
	a.log.Debugf("[%s] Compacting history (%d messages) ...", sessionID, len(messages))

	systemPrompt, priorSummary, body := splitForCompaction(messages)

	compactPrompt, err := a.promptManager.LoadCompact()
	if err != nil {
		return fmt.Errorf("load compact prompt: %w", err)
	}

	result, err := a.provider.Chat(ctx, []llm.Message{
		{Role: schemas.ChatMessageRoleSystem, Content: compactPrompt},
		{Role: schemas.ChatMessageRoleUser, Content: buildSummaryInput(priorSummary, body)},
	}, nil)
	if err != nil {
		return fmt.Errorf("summarize history: %w", err)
	}

	compacted := assembleCompacted(systemPrompt, result.Content, messages)

	err = a.history.Replace(ctx, sessionID, conversationID, compacted)
	if err != nil {
		return fmt.Errorf("save compacted history: %w", err)
	}

	a.log.Debugf("[%s] History compacted to %d messages", sessionID, len(compacted))

	return nil
}

// splitForCompaction separates the system prompt and any summary left by a
// previous compaction from the newer messages to be folded in.
func splitForCompaction(messages []llm.Message) (systemPrompt llm.Message, priorSummary string, body []llm.Message) {
	systemPrompt = messages[0]
	body = messages[1:]
	if len(body) > 0 && body[0].Role == schemas.ChatMessageRoleSystem && strings.HasPrefix(body[0].Content, summaryPrefix) {
		priorSummary = strings.TrimPrefix(body[0].Content, summaryPrefix)
		body = body[1:]
	}
	return systemPrompt, priorSummary, body
}

func buildSummaryInput(priorSummary string, body []llm.Message) string {
	var sb strings.Builder
	if priorSummary != "" {
		sb.WriteString("EXISTING SUMMARY (established facts, preserve all of it):\n")
		sb.WriteString(priorSummary)
		sb.WriteString("\n\nNEW MESSAGES SINCE THEN:\n")
	}
	for _, m := range body {
		sb.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}
	return sb.String()
}

func assembleCompacted(systemPrompt llm.Message, summary string, messages []llm.Message) []llm.Message {
	compacted := []llm.Message{
		systemPrompt,
		{Role: schemas.ChatMessageRoleSystem, Content: summaryPrefix + summary},
	}
	start := len(messages) - keepRecentMessages
	if start < 1 {
		start = 1 // never fold the system prompt into the retained tail
	}
	return append(compacted, messages[start:]...)
}
