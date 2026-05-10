package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *goredis.Client
}

func NewClient(ctx context.Context, addr string) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{Addr: addr})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
