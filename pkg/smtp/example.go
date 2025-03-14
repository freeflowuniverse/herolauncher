package smtp

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mailmodel "github.com/freeflowuniverse/herolauncher/pkg/mail"
	"golang.org/x/net/context"
)

// Example shows how to use the SMTP server and process incoming emails
func Example() {
	// Create a new SMTP server with default configuration
	config := DefaultConfig()
	config.Port = 2525 // Use a non-privileged port for testing
	
	server, err := NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create SMTP server: %v", err)
	}
	
	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start SMTP server: %v", err)
		}
	}()
	
	// Wait for the server to start
	time.Sleep(1 * time.Second)
	
	// Create a processor for incoming emails
	processor := func(email *mailmodel.Email) error {
		// Print the email details
		emailJSON, err := json.MarshalIndent(email, "", "  ")
		if err != nil {
			return err
		}
		
		fmt.Printf("Received email:\n%s\n", string(emailJSON))
		return nil
	}
	
	// Process emails from the queue
	go func() {
		if err := ProcessEmails(server.GetRedisClient(), processor, 5*time.Second); err != nil {
			log.Printf("Error processing emails: %v", err)
		}
	}()
	
	// Example of how to retrieve emails from Redis directly
	go func() {
		// Wait for some emails to arrive
		time.Sleep(10 * time.Second)
		
		ctx := context.Background()
		
		// List all email IDs
		emailIDs, err := server.GetRedisClient().LRange(ctx, "mail:out", 0, -1).Result()
		if err != nil {
			log.Printf("Failed to list emails: %v", err)
			return
		}
		
		fmt.Printf("Found %d emails in the queue\n", len(emailIDs))
		
		// Process each email
		for _, mailID := range emailIDs {
			email, err := GetEmail(server.GetRedisClient(), mailID)
			if err != nil {
				log.Printf("Failed to get email %s: %v", mailID, err)
				continue
			}
			
			fmt.Printf("Email %s: From=%s, To=%v, Subject=%s\n", 
				mailID, email.From, email.To, email.Subject)
		}
	}()
	
	// Keep the server running
	select {}
}
