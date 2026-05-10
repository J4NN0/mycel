package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/J4NN0/mycel/internal/llm"
)

const keyPrefix = "history:"

type History interface {
	Load(ctx context.Context, sessionID string) ([]llm.Message, error)
	Save(ctx context.Context, sessionID string, messages []llm.Message) error
	Clear(ctx context.Context, sessionID string) error
}

type Client struct {
	rdb *goredis.Client
}

func NewClient(ctx context.Context, addr string) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{Addr: addr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Load(ctx context.Context, sessionID string) ([]llm.Message, error) {
	data, err := c.rdb.Get(ctx, keyPrefix+sessionID).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}
	var messages []llm.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("unmarshal history: %w", err)
	}
	return messages, nil
}

func (c *Client) Save(ctx context.Context, sessionID string, messages []llm.Message) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	if err := c.rdb.Set(ctx, keyPrefix+sessionID, data, 0).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (c *Client) Clear(ctx context.Context, sessionID string) error {
	if err := c.rdb.Del(ctx, keyPrefix+sessionID).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}
