package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("Redis Connection Test")
	fmt.Println("====================")

	// Create a configuration for Redis
	config := &stats.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to ping Redis
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Error connecting to Redis: %v\n", err)
		fmt.Println("\nPlease ensure Redis is running with:")
		fmt.Println("  - docker run --name redis -p 6379:6379 -d redis")
		fmt.Println("  - or install Redis locally and start the service")
		os.Exit(1)
	}

	fmt.Printf("Successfully connected to Redis at %s\n", config.RedisAddr)
	fmt.Printf("Response: %s\n", pong)

	// Test basic operations
	fmt.Println("\nTesting basic Redis operations...")

	// Set a key
	err = client.Set(ctx, "test:key", "Hello from HeroLauncher!", 1*time.Minute).Err()
	if err != nil {
		fmt.Printf("Error setting key: %v\n", err)
		os.Exit(1)
	}

	// Get the key
	val, err := client.Get(ctx, "test:key").Result()
	if err != nil {
		fmt.Printf("Error getting key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Retrieved value: %s\n", val)
	fmt.Println("Redis connection test successful!")
}
