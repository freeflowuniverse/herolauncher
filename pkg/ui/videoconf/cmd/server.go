package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/ui/videoconf"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// Try to find .env in parent directory
		parentEnvPath := filepath.Join("..", ".env")
		if _, err := os.Stat(parentEnvPath); err == nil {
			envPath = parentEnvPath
		}
	}

	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
		log.Println("Using environment variables from system")
	} else {
		log.Printf("Loaded environment variables from %s", envPath)
	}

	log.Println("Starting video conferencing UI test server...")

	// Create a new video conferencing UI server with default config
	config := videoconf.DefaultConfig()

	// Set the correct path for templates
	config.TemplatesPath = "../web/templates"
	config.StaticPath = "../web/static"

	// Override port if specified in environment
	if port := os.Getenv("PORT"); port != "" {
		log.Printf("Using port from environment: %s", port)
		// Convert port string to int
		portInt, err := strconv.Atoi(port)
		if err != nil {
			log.Printf("Warning: Invalid PORT environment variable: %s. Using default port: %d", port, config.Port)
		} else {
			config.Port = portInt
		}
	}

	vc := videoconf.New(config)

	// Setup routes
	vc.SetupRoutes()

	// Create a channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on http://localhost:%d", config.Port)
		if err := vc.Start(); err != nil {
			log.Fatalf("Error starting video conferencing UI server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")
}
