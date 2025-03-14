package doctree

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisClient is a wrapper around the Redis client
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// redisClient is the Redis client used by DocTree
var redisClient *RedisClient

// init initializes the Redis client when the package is imported
func init() {
	// Create a new Redis client with default settings (localhost:6379, no auth)
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // default DB
	})

	redisClient = &RedisClient{
		client: client,
		ctx:    context.Background(),
	}
}

// Del deletes a key
func (r *RedisClient) Del(key string) {
	r.client.Del(r.ctx, key)
}

// HSet sets a field in a hash
func (r *RedisClient) HSet(key, field, value string) {
	r.client.HSet(r.ctx, key, field, value)
}

// HGet gets a field from a hash
func (r *RedisClient) HGet(key, field string) (string, bool) {
	val, err := r.client.HGet(r.ctx, key, field).Result()
	if err != nil {
		return "", false
	}
	return val, true
}
