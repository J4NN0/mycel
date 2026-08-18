package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/maximhq/bifrost/core/schemas"
)

type ConversationSummary struct {
	ID      string
	Preview string
}

type Store interface {
	Load(ctx context.Context, conversationID string) ([]llm.Message, error)
	Len(ctx context.Context, conversationID string) (int64, error)
	Append(ctx context.Context, conversationID string, messages ...llm.Message) error
	Replace(ctx context.Context, conversationID string, messages []llm.Message) error
	ActiveConversation(ctx context.Context) (string, error)
	NewConversation(ctx context.Context) (string, error)
	SetActiveConversation(ctx context.Context, conversationID string) error
	ListConversations(ctx context.Context, excludeID string, limit int) ([]ConversationSummary, error)
}

func (c *Client) Load(ctx context.Context, conversationID string) ([]llm.Message, error) {
	raw, err := c.rdb.LRange(ctx, historyKey(conversationID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lrange: %w", err)
	}

	messages := make([]llm.Message, 0, len(raw))
	for _, r := range raw {
		var m llm.Message
		err = json.Unmarshal([]byte(r), &m)
		if err != nil {
			return nil, fmt.Errorf("unmarshal history entry: %w", err)
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func (c *Client) Len(ctx context.Context, conversationID string) (int64, error) {
	n, err := c.rdb.LLen(ctx, historyKey(conversationID)).Result()
	if err != nil {
		return 0, fmt.Errorf("redis llen: %w", err)
	}
	return n, nil
}

func (c *Client) Append(ctx context.Context, conversationID string, messages ...llm.Message) error {
	if len(messages) == 0 {
		return nil
	}

	entries, err := encodeMessages(messages)
	if err != nil {
		return err
	}

	err = c.rdb.RPush(ctx, historyKey(conversationID), entries...).Err()
	if err != nil {
		return fmt.Errorf("redis rpush: %w", err)
	}

	return nil
}

func (c *Client) Replace(ctx context.Context, conversationID string, messages []llm.Message) error {
	entries, err := encodeMessages(messages)
	if err != nil {
		return err
	}

	key := historyKey(conversationID)
	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, key)
	if len(entries) > 0 {
		pipe.RPush(ctx, key, entries...)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis replace history: %w", err)
	}

	return nil
}

func encodeMessages(messages []llm.Message) ([]any, error) {
	entries := make([]any, len(messages))
	for i, m := range messages {
		data, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("marshal history entry: %w", err)
		}
		entries[i] = data
	}
	return entries, nil
}

func (c *Client) ActiveConversation(ctx context.Context) (string, error) {
	id, err := c.rdb.Get(ctx, activeConvKey).Result()
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, goredis.Nil) {
		return "", fmt.Errorf("redis get active conversation: %w", err)
	}

	return c.NewConversation(ctx)
}

func (c *Client) NewConversation(ctx context.Context) (string, error) {
	seq, err := c.rdb.Incr(ctx, convSeqKey).Result()
	if err != nil {
		return "", fmt.Errorf("redis incr conversation seq: %w", err)
	}

	id := strconv.FormatInt(seq, 10)
	err = c.rdb.Set(ctx, activeConvKey, id, 0).Err()
	if err != nil {
		return "", fmt.Errorf("redis set active conversation: %w", err)
	}

	return id, nil
}

func (c *Client) SetActiveConversation(ctx context.Context, conversationID string) error {
	n, err := c.Len(ctx, conversationID)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("conversation %s has no messages", conversationID)
	}

	err = c.rdb.Set(ctx, activeConvKey, conversationID, 0).Err()
	if err != nil {
		return fmt.Errorf("redis set active conversation: %w", err)
	}

	return nil
}

func (c *Client) ListConversations(ctx context.Context, excludeID string, limit int) ([]ConversationSummary, error) {
	seqStr, err := c.rdb.Get(ctx, convSeqKey).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get conversation seq: %w", err)
	}

	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse conversation seq: %w", err)
	}

	var summaries []ConversationSummary
	for id := seq; id >= 1 && len(summaries) < limit; id-- {
		conversationID := strconv.FormatInt(id, 10)
		if conversationID == excludeID {
			continue
		}

		n, err := c.Len(ctx, conversationID)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			continue
		}

		messages, err := c.Load(ctx, conversationID)
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, ConversationSummary{ID: conversationID, Preview: firstUserMessage(messages)})
	}

	return summaries, nil
}

func firstUserMessage(messages []llm.Message) string {
	for _, m := range messages {
		if m.Role == schemas.ChatMessageRoleUser {
			return m.Content
		}
	}
	if len(messages) > 0 {
		return messages[len(messages)-1].Content
	}
	return ""
}
