package smtp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mailmodel "github.com/freeflowuniverse/herolauncher/pkg/mail"
	"github.com/redis/go-redis/v9"
)

// EmailProcessor represents a function that processes an email
type EmailProcessor func(email *mailmodel.Email) error

// ProcessEmails processes emails from the mail queue
func ProcessEmails(redisClient *redis.Client, processor EmailProcessor, timeout time.Duration) error {
	ctx := context.Background()
	
	// Process emails from the queue
	for {
		// Get the next email ID from the queue with timeout
		result, err := redisClient.BLPop(ctx, timeout, "mail:out").Result()
		if err != nil {
			if err == redis.Nil {
				// No emails in the queue, continue
				continue
			}
			return fmt.Errorf("failed to get email from queue: %w", err)
		}
		
		if len(result) < 2 {
			// Invalid result format
			continue
		}
		
		// Get the email ID
		mailID := result[1]
		
		// Get the email data
		emailJSON, err := redisClient.HGet(ctx, mailID, "data").Result()
		if err != nil {
			if err == redis.Nil {
				// Email not found, continue
				continue
			}
			return fmt.Errorf("failed to get email data: %w", err)
		}
		
		// Parse the email JSON
		var email mailmodel.Email
		if err := json.Unmarshal([]byte(emailJSON), &email); err != nil {
			return fmt.Errorf("failed to parse email JSON: %w", err)
		}
		
		// Process the email
		if err := processor(&email); err != nil {
			// If processing fails, put the email back in the queue
			if err := redisClient.RPush(ctx, "mail:out", mailID).Err(); err != nil {
				return fmt.Errorf("failed to put email back in queue: %w", err)
			}
			return fmt.Errorf("failed to process email: %w", err)
		}
		
		// Delete the email from Redis
		if err := redisClient.Del(ctx, mailID).Err(); err != nil {
			return fmt.Errorf("failed to delete email: %w", err)
		}
	}
}

// GetEmail retrieves an email by ID
func GetEmail(redisClient *redis.Client, mailID string) (*mailmodel.Email, error) {
	ctx := context.Background()
	
	// Get the email data
	emailJSON, err := redisClient.HGet(ctx, mailID, "data").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("email not found")
		}
		return nil, fmt.Errorf("failed to get email data: %w", err)
	}
	
	// Parse the email JSON
	var email mailmodel.Email
	if err := json.Unmarshal([]byte(emailJSON), &email); err != nil {
		return nil, fmt.Errorf("failed to parse email JSON: %w", err)
	}
	
	return &email, nil
}

// ListEmails lists all emails in the queue
func ListEmails(redisClient *redis.Client) ([]string, error) {
	ctx := context.Background()
	
	// Get all email IDs from the queue
	return redisClient.LRange(ctx, "mail:out", 0, -1).Result()
}

// CountEmails counts the number of emails in the queue
func CountEmails(redisClient *redis.Client) (int64, error) {
	ctx := context.Background()
	
	// Count emails in the queue
	return redisClient.LLen(ctx, "mail:out").Result()
}
