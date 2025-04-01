// @title HeroLauncher API
// @version 1.0
// @description HeroLauncher API provides endpoints for managing services, processes, and system resources
// @termsOfService http://swagger.io/terms/

// @contact.name HeroLauncher Support
// @contact.url https://github.com/freeflowuniverse/herolauncher
// @contact.email support@herolauncher.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:9021
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher"
	_ "github.com/freeflowuniverse/herolauncher/pkg/herolauncher/docs" // Import generated swagger docs
)

func main() {
	// Parse command-line flags
	portFlag := flag.String("port", "", "Port to run the server on")
	flag.Parse()

	// Initialize HeroLauncher with default configuration
	config := herolauncher.DefaultConfig()

	// Override with command-line flags if provided
	if *portFlag != "" {
		config.Port = *portFlag
	}

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
