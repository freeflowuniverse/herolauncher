package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
)

func main() {
	// Parse command line flags
	socketPath := flag.String("socket", "/tmp/processmanager.sock", "Path to the Unix domain socket")
	secret := flag.String("secret", "", "Authentication secret for the process manager")
	logsDir := flag.String("logs", "/tmp/herolauncher/process_logs", "Directory for process logs")
	flag.Parse()

	// Validate flags
	if *secret == "" {
		log.Fatal("Error: secret is required")
	}

	// Create process manager
	processManager := processmanager.NewProcessManager(*secret)

	// Set logs base path
	processManager.SetLogsBasePath(*logsDir)

	// Create the socket file to indicate the process manager is running
	// In a real implementation, this would be a socket for communication
	// For now, we just create an empty file
	if err := os.MkdirAll("./tmp", 0755); err != nil {
		log.Printf("Warning: Failed to create tmp directory: %v", err)
	}
	
	// Remove existing socket file if it exists
	if err := os.Remove(*socketPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove existing socket file: %v", err)
	}
	
	// Create an empty file to indicate the process manager is running
	f, err := os.Create(*socketPath)
	if err != nil {
		log.Fatalf("Failed to create socket file: %v", err)
	}
	f.Close()

	fmt.Printf("Process manager started. Socket path: %s\n", *socketPath)
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigChan
	fmt.Printf("Received signal %v, shutting down...\n", sig)

	// Clean up
	os.Remove(*socketPath)

	fmt.Println("Process manager shutdown complete")
}
