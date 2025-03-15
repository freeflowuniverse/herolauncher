// Example for using the telnet server with the process manager
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
)

func main() {
	// Create a new process manager with a secret
	secret := "mysecret" // In production, use a secure secret
	pm := processmanager.NewProcessManager(secret)

	// Create a telnet adapter for the process manager
	adapter := processmanager.NewTelnetAdapter(pm)

	// Start the telnet server on port 8023
	err := adapter.Start(":8023")
	if err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}
	defer adapter.Stop()

	fmt.Println("Process Manager started")
	fmt.Println("Telnet server listening on port 8023")
	fmt.Println("Connect with: telnet localhost 8023")
	fmt.Println("Use secret:", secret)
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
}
