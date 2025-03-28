package main

import (
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/webui"
)

func main() {
	// Create server with default configuration
	config := webui.DefaultConfig()
	server := webui.NewServer(config)
	
	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start Freezone Manager UI server: %v", err)
	}
}
