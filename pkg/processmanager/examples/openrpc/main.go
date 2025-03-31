package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces/openrpc"
)

func main() {
	log.Println("Starting Process Manager OpenRPC Example")

	// Use /tmp directory for the socket as it's more reliable for Unix sockets
	// Define the socket path
	socketPath := "/tmp/process-manager.sock"
	// Remove the socket file if it already exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			log.Fatalf("Failed to remove existing socket file: %v", err)
		}
	}
	log.Printf("Using socket path: %s", socketPath)

	// Create a mock process manager
	mockPM := NewMockProcessManager()

	// Create and start the server
	server, err := openrpc.NewServer(mockPM, socketPath)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server
	err = server.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Println("Server started successfully")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait a bit for the server to start
	time.Sleep(1 * time.Second)

	// Run the client example in a goroutine
	errChan := make(chan error, 1)
	go func() {
		err := RunClientExample(socketPath, mockPM.GetSecret())
		errChan <- err
	}()

	// Wait for the client to finish or for a signal
	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("Client example failed: %v", err)
		}
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	}

	// Stop the server
	log.Println("Stopping server...")
	err = server.Stop()
	if err != nil {
		log.Printf("Failed to stop server: %v", err)
	}

	log.Println("Example completed")
}
