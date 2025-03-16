package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/mcpopenapi"
)

func main() {
	// Log startup information
	log.Println("Starting mcpopenapi MCP server...")

	// Create a new MCP server with stdin and stdout
	server, err := mcpopenapi.NewMCPServer(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Create a channel to handle shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// Log server information
	log.Printf("MCP OpenAPI Server initialized and ready")
	
	// Start the server in a goroutine
	go func() {
		log.Println("MCP OpenAPI Server is now serving requests...")
		if err := server.Serve(); err != nil {
			log.Printf("Server error: %v", err)
			// Signal the main goroutine to shut down
			done <- syscall.SIGTERM
		}
	}()

	// Wait for termination signal
	sig := <-done
	log.Printf("Received signal: %v", sig)
	log.Println("Shutting down mcpopenapi MCP server...")
	
	// Perform any cleanup here if needed
	
	log.Println("MCP OpenAPI Server shutdown complete")
}
