package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/mail"
	"github.com/freeflowuniverse/herolauncher/pkg/smtp"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 2525, "SMTP server port")
	host := flag.String("host", "0.0.0.0", "SMTP server host")
	domain := flag.String("domain", "localhost", "SMTP server domain")
	redisAddr := flag.String("redis-addr", "localhost:6378", "Redis server address")
	redisPassword := flag.String("redis-password", "", "Redis server password")
	redisDB := flag.Int("redis-db", 0, "Redis database number")
	flag.Parse()

	// Create SMTP server configuration
	config := smtp.DefaultConfig()
	config.Host = *host
	config.Port = *port
	config.Domain = *domain
	config.RedisAddr = *redisAddr
	config.RedisPassword = *redisPassword
	config.RedisDB = *redisDB

	// Create SMTP server
	server, err := smtp.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create SMTP server: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting SMTP server on %s:%d", config.Host, config.Port)
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start SMTP server: %v", err)
		}
	}()

	// Create a processor for incoming emails
	processor := func(email *mail.Email) error {
		// Print the email details
		emailJSON, err := json.MarshalIndent(email, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal email: %v\n", err)
			return err
		}

		fmt.Printf("Received email:\n%s\n", string(emailJSON))
		return nil
	}

	// Process emails from the queue in a goroutine
	go func() {
		log.Println("Starting email processor")
		if err := smtp.ProcessEmails(server.GetRedisClient(), processor, 5*time.Second); err != nil {
			log.Printf("Error processing emails: %v", err)
		}
	}()

	// Wait for termination signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	log.Println("Shutting down SMTP server...")
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping SMTP server: %v", err)
	}
}
