package main

import (
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher"
)

// @title HeroLauncher API
// @version 1.0
// @description API for HeroLauncher - a modular service manager
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@freeflowuniverse.org
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:9001
// @BasePath /api
func main() {
	// Use the default configuration
	config := herolauncher.DefaultConfig()

	// Create a new HeroLauncher instance
	launcher := herolauncher.New(config)

	// Start the server
	log.Printf("Starting HeroLauncher server on port %s", config.Port)
	if err := launcher.Start(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
