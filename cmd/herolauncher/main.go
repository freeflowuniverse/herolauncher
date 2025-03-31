package main

import (
	"fmt"
	"log"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher"
)

func main() {
	// Initialize HeroLauncher with default configuration
	config := herolauncher.DefaultConfig()

	// Override with environment variables if provided
	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}

	// Create HeroLauncher instance
	launcher := herolauncher.New(config)

	// Start the server
	fmt.Printf("Starting HeroLauncher on port %s...\n", config.Port)
	if err := launcher.Start(); err != nil {
		log.Fatalf("Failed to start HeroLauncher: %v", err)
	}
}
