package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/mail"
	"github.com/redis/go-redis/v9"
)

var (
	// Sample data for generating random emails
	senders = []string{
		"john.doe@example.com",
		"jane.smith@example.com",
		"bob.johnson@example.com",
		"alice.williams@example.com",
		"david.brown@example.com",
	}

	recipients = []string{
		"recipient1@example.com",
		"recipient2@example.com",
		"recipient3@example.com",
		"recipient4@example.com",
		"recipient5@example.com",
		"team@example.com",
	}

	subjects = []string{
		"Meeting Agenda for Next Week",
		"Project Update: Phase 2 Complete",
		"Quarterly Report: Q1 2025",
		"Invitation to Company Event",
		"Important Announcement",
		"Follow-up on Previous Discussion",
		"New Product Launch",
		"Customer Feedback Summary",
		"Upcoming Deadlines",
		"Team Building Activity",
	}

	messages = []string{
		"Hello,\n\nI hope this email finds you well. I wanted to discuss the upcoming project deadlines.\n\nBest regards,\n%s",
		"Hi team,\n\nPlease find attached the latest report on our quarterly performance.\n\nRegards,\n%s",
		"Dear colleague,\n\nI'm writing to invite you to our annual company retreat next month.\n\nLooking forward to your response,\n%s",
		"Greetings,\n\nThis is a reminder about the meeting scheduled for tomorrow at 10 AM.\n\nThank you,\n%s",
		"Hello,\n\nI wanted to follow up on our conversation from last week regarding the new project.\n\nBest,\n%s",
	}

	// Account names
	accounts = []string{"pol", "jan"}

	// Base folder structure with potential subfolders
	folderStructure = map[string][]string{
		"inbox":    {"important", "work", "personal"},
		"sent":     {"work", "personal", "archive"},
		"drafts":   {},
		"archive":  {"2023", "2024", "2025"},
		"work":     {"projects", "meetings", "reports"},
		"personal": {"family", "friends", "finance"},
		"projects": {"projectA", "projectB", "projectC"},
		"reports":  {"quarterly", "annual", "monthly"},
	}
)

// generateRandomEmail creates a random email
func generateRandomEmail(r *rand.Rand) *mail.Email {
	// Choose random sender and format name
	sender := senders[r.Intn(len(senders))]
	senderName := strings.Split(sender, "@")[0]
	senderName = strings.ReplaceAll(senderName, ".", " ")
	senderName = strings.Title(senderName)

	// Choose 1-3 random recipients
	numRecipients := r.Intn(3) + 1
	to := make([]string, 0, numRecipients)
	for i := 0; i < numRecipients; i++ {
		recipient := recipients[r.Intn(len(recipients))]
		// Avoid duplicates
		if !contains(to, recipient) {
			to = append(to, recipient)
		}
	}

	// Choose random subject
	subject := subjects[r.Intn(len(subjects))]

	// Choose random message and format with sender name
	message := fmt.Sprintf(messages[r.Intn(len(messages))], senderName)

	// Create email
	email := &mail.Email{
		From:        sender,
		To:          to,
		Subject:     subject,
		Message:     message,
		Attachments: []mail.Attachment{},
	}

	// Randomly add an attachment (20% chance)
	if r.Intn(5) == 0 {
		email.Attachments = append(email.Attachments, mail.Attachment{
			Filename:    "document.pdf",
			ContentType: "application/pdf",
			Data:        "SGVsbG8sIHRoaXMgaXMgYSBkdW1teSBQREYgZmlsZSBjb250ZW50Lg==", // Base64 encoded dummy content
		})
	}

	return email
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateRandomFolder creates a random folder path with up to 3 levels
func generateRandomFolder(r *rand.Rand) string {
	// Start with a base folder
	baseKeys := make([]string, 0, len(folderStructure))
	for k := range folderStructure {
		baseKeys = append(baseKeys, k)
	}
	
	folder := baseKeys[r.Intn(len(baseKeys))]
	depth := r.Intn(3) + 1 // 1-3 levels deep
	
	currentFolder := folder
	path := []string{currentFolder}
	
	// Add subfolders based on the depth
	for i := 1; i < depth; i++ {
		if subfolders, ok := folderStructure[currentFolder]; ok && len(subfolders) > 0 {
			subfolder := subfolders[r.Intn(len(subfolders))]
			path = append(path, subfolder)
			currentFolder = subfolder
		} else {
			break
		}
	}
	
	return strings.Join(path, "/")
}

func main() {
	// Parse command line flags
	redisAddr := flag.String("redis-addr", "localhost:6378", "Redis server address")
	redisPassword := flag.String("redis-password", "", "Redis server password")
	redisDB := flag.Int("redis-db", 0, "Redis database number")
	numEmails := flag.Int("num-emails", 100, "Number of emails to generate")
	flag.Parse()

	// Initialize random number generator with a source
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: *redisPassword,
		DB:       *redisDB,
	})

	// Test Redis connection
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", *redisAddr, err)
	}
	log.Printf("Successfully connected to Redis at %s", *redisAddr)

	// Generate and store random emails
	log.Printf("Generating %d random emails...", *numEmails)
	
	for i := 0; i < *numEmails; i++ {
		// Generate random email
		email := generateRandomEmail(r)
		
		// Choose random account and folder
		account := accounts[r.Intn(len(accounts))]
		folder := generateRandomFolder(r)
		
		// Generate UID based on epoch time + incrementing number
		epoch := time.Now().Unix()
		baseUID := fmt.Sprintf("%d", epoch)
		uid := baseUID
		
		// Check if the UID already exists and increment until we find a unique one
		increment := 0
		for {
			// Check if the key exists in Redis using EXISTS command
			mailKey := fmt.Sprintf("mail:in:%s:%s:%s", account, folder, uid)
			exists, err := redisClient.Exists(ctx, mailKey).Result()
			if err != nil {
				log.Printf("Failed to check if key exists: %v", err)
				break
			}
			
			if exists == 0 {
				// Key doesn't exist
				break
			}
			
			// If we get here, the key exists, so increment and try again
			increment++
			uid = fmt.Sprintf("%s%d", baseUID, increment)
		}
		
		// Convert email to JSON
		emailJSON, err := json.Marshal(email)
		if err != nil {
			log.Printf("Failed to marshal email: %v", err)
			continue
		}
		
		// Store email in Redis
		mailKey := fmt.Sprintf("mail:in:%s:%s:%s", account, folder, uid)
		if err := redisClient.Set(ctx, mailKey, string(emailJSON), 0).Err(); err != nil {
			log.Printf("Failed to store email in Redis: %v", err)
			continue
		}
		
		log.Printf("Stored email %d/%d with key: %s", i+1, *numEmails, mailKey)
	}
	
	log.Printf("Successfully generated and stored %d random emails in Redis", *numEmails)
}
