package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client is a wrapper around redis.Client that provides enhanced functionality
type Client struct {
	*redis.Client
}

// NewClient creates a new enhanced Redis client
func NewClient(options *redis.Options) *Client {
	return &Client{
		Client: redis.NewClient(options),
	}
}

// NewClientWithAddr creates a new enhanced Redis client with the given address
func NewClientWithAddr(addr string, db int) *Client {
	return NewClient(&redis.Options{
		Addr:         addr,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
}

// Keys is an enhanced implementation that uses SCAN for better pattern matching
// This fixes issues with pattern matching in older Redis versions
func (c *Client) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(ctx)
	
	// Use SCAN to emulate KEYS for better pattern matching
	go func() {
		var cursor uint64
		var keys []string
		var allKeys []string
		var err error

		for {
			keys, cursor, err = c.Client.Scan(ctx, cursor, pattern, 10).Result()
			if err != nil {
				cmd.SetErr(err)
				return
			}

			allKeys = append(allKeys, keys...)
			if cursor == 0 {
				break
			}
		}

		cmd.SetVal(allKeys)
	}()

	return cmd
}
