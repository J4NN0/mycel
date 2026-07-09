package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/J4NN0/mycel/internal/llm"
)

// historyKey is the single Redis key under which the agent's entire
// conversation history is stored, shared across every platform and user.
// It's kept as a Redis list (one message per element) rather than a single
// JSON blob so that appending a message doesn't require reading and
// rewriting the whole history, and so it isn't bound by Redis's per-value
// size cap on strings.
const historyKey = "history:shared"

type List interface {
	Load(ctx context.Context) ([]llm.Message, error)
	Len(ctx context.Context) (int64, error)
	Append(ctx context.Context, messages ...llm.Message) error
	Replace(ctx context.Context, messages []llm.Message) error
	Clear(ctx context.Context) error
}

func (c *Client) Load(ctx context.Context) ([]llm.Message, error) {
	raw, err := c.rdb.LRange(ctx, historyKey, 0, -1).Result()
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

func (c *Client) Len(ctx context.Context) (int64, error) {
	n, err := c.rdb.LLen(ctx, historyKey).Result()
	if err != nil {
		return 0, fmt.Errorf("redis llen: %w", err)
	}
	return n, nil
}

func (c *Client) Append(ctx context.Context, messages ...llm.Message) error {
	if len(messages) == 0 {
		return nil
	}

	entries, err := encodeMessages(messages)
	if err != nil {
		return err
	}

	err = c.rdb.RPush(ctx, historyKey, entries...).Err()
	if err != nil {
		return fmt.Errorf("redis rpush: %w", err)
	}

	return nil
}

func (c *Client) Replace(ctx context.Context, messages []llm.Message) error {
	entries, err := encodeMessages(messages)
	if err != nil {
		return err
	}

	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, historyKey)
	if len(entries) > 0 {
		pipe.RPush(ctx, historyKey, entries...)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis replace history: %w", err)
	}

	return nil
}

func (c *Client) Clear(ctx context.Context) error {
	err := c.rdb.Del(ctx, historyKey).Err()
	if err != nil {
		return fmt.Errorf("redis del: %w", err)
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
