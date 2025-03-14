package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Parse command line flags
	redisAddr := flag.String("redis-addr", "localhost:6378", "Redis server address")
	username := flag.String("user", "jan", "Username to check")
	mailbox := flag.String("mailbox", "inbox", "Mailbox to check")
	debug := flag.Bool("debug", true, "Enable debug output")
	dbNum := flag.Int("db", 0, "Redis database number")
	flag.Parse()

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:         *redisAddr,
		DB:           *dbNum,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	defer redisClient.Close()

	ctx := context.Background()

	// Check connection
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Connected to Redis: %s", pong)

	// Try to get a specific key that we know exists
	specificKey := "mail:in:jan:inbox:17419716651"
	val, err := redisClient.Get(ctx, specificKey).Result()
	if err == redis.Nil {
		log.Printf("Key '%s' does not exist", specificKey)
	} else if err != nil {
		log.Printf("Error getting key '%s': %v", specificKey, err)
	} else {
		log.Printf("Found key '%s' with value length: %d", specificKey, len(val))
	}

	// List all keys using SCAN instead of KEYS
	if *debug {
		log.Println("Listing keys in Redis using SCAN:")
		var cursor uint64
		var allKeys []string
		var err error
		var keys []string

		for {
			keys, cursor, err = redisClient.Scan(ctx, cursor, "*", 10).Result()
			if err != nil {
				log.Printf("Error scanning keys: %v", err)
				break
			}

			allKeys = append(allKeys, keys...)
			if cursor == 0 {
				break
			}
		}

		log.Printf("Found %d total keys using SCAN", len(allKeys))
		for i, k := range allKeys {
			if i < 20 { // Limit output to first 20 keys
				log.Printf("Key[%d]: %s", i, k)
			}
		}
		if len(allKeys) > 20 {
			log.Printf("... and %d more keys", len(allKeys)-20)
		}
	}

	// Test different pattern formats using SCAN and KEYS
	patterns := []string{
		fmt.Sprintf("mail:in:%s:%s*", *username, strings.ToLower(*mailbox)),
		fmt.Sprintf("mail:in:%s:%s:*", *username, strings.ToLower(*mailbox)),
		fmt.Sprintf("mail:in:%s:%s/*", *username, strings.ToLower(*mailbox)),
		fmt.Sprintf("mail:in:%s:%s*", *username, *mailbox),
	}

	for _, pattern := range patterns {
		// Test with SCAN
		log.Printf("Trying pattern with SCAN: %s", pattern)
		var cursor uint64
		var keys []string
		var allKeys []string

		for {
			keys, cursor, err = redisClient.Scan(ctx, cursor, pattern, 10).Result()
			if err != nil {
				log.Printf("Error scanning with pattern %s: %v", pattern, err)
				break
			}

			allKeys = append(allKeys, keys...)
			if cursor == 0 {
				break
			}
		}
		
		log.Printf("Found %d keys with pattern %s using SCAN", len(allKeys), pattern)
		for i, key := range allKeys {
			log.Printf("  Key[%d]: %s", i, key)
		}

		// Test with our scanKeys helper function (emulated KEYS)
		log.Printf("Trying pattern with emulated KEYS: %s", pattern)
		keysResult, err := scanKeys(ctx, redisClient, pattern)
		if err != nil {
			log.Printf("Error with emulated KEYS command for pattern %s: %v", pattern, err)
		} else {
			log.Printf("Found %d keys with pattern %s using emulated KEYS", len(keysResult), pattern)
			for i, key := range keysResult {
				log.Printf("  Key[%d]: %s", i, key)
			}
		}
	}

	// Find all keys for the specified user using SCAN
	userPattern := fmt.Sprintf("mail:in:%s:*", *username)
	log.Printf("Checking all keys for user with pattern: %s using SCAN", userPattern)
	var cursor uint64
	var keys []string
	var userKeys []string

	for {
		keys, cursor, err = redisClient.Scan(ctx, cursor, userPattern, 10).Result()
		if err != nil {
			log.Printf("Error scanning user keys: %v", err)
			break
		}

		userKeys = append(userKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	log.Printf("Found %d total keys for user %s using SCAN", len(userKeys), *username)
	
	// Extract unique mailbox names
	mailboxMap := make(map[string]bool)
	for _, key := range userKeys {
		parts := strings.Split(key, ":")
		if len(parts) >= 4 {
			mailboxName := parts[3]
			
			// Handle mailbox/uid format
			if strings.Contains(mailboxName, "/") {
				mailboxParts := strings.Split(mailboxName, "/")
				mailboxName = mailboxParts[0]
			}
			
			mailboxMap[mailboxName] = true
		}
	}
	
	log.Printf("Found %d unique mailboxes for user %s:", len(mailboxMap), *username)
	for mailbox := range mailboxMap {
		log.Printf("  Mailbox: %s", mailbox)
	}
}
