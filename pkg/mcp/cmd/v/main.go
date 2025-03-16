package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mcpv "github.com/freeflowuniverse/herolauncher/pkg/mcp/v"
)

func main() {
	// Log startup information
	log.Println("Starting MCP V Language Specs server...")

	// Create a new MCP server with stdin and stdout
	server, err := mcpv.NewMCPServer(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Create a channel to handle shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// Log server information
	log.Printf("MCP V Language Specs Server initialized and ready")

	// Start the server in a goroutine
	go func() {
		log.Println("MCP V Language Specs Server is now serving requests...")
		if err := server.Serve(); err != nil {
			log.Printf("Server error: %v", err)
			// Signal the main goroutine to shut down
			done <- syscall.SIGTERM
		}
	}()

	// Wait for termination signal
	sig := <-done
	log.Printf("Received signal: %v", sig)
	log.Println("Shutting down MCP V Language Specs server...")

	// Perform any cleanup here if needed

	log.Println("MCP V Language Specs Server shutdown complete")
}
