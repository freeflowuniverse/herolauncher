package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces/telnet"
)

func main() {
	// Parse command line flags
	socketPath := flag.String("socket", "/tmp/processmanager.sock", "Path to the Unix domain socket")
	secret := flag.String("secret", "", "Authentication secret for the telnet server")
	flag.Parse()

	// Validate flags
	if *secret == "" {
		log.Fatal("Error: secret is required")
	}

	// Create process manager
	processManager := processmanager.NewProcessManager(*secret)

	// Create telnet adapter
	telnetAdapter := telnet.NewTelnetAdapter(processManager)

	// Start telnet server
	fmt.Printf("Starting process manager telnet server on socket: %s\n", *socketPath)
	err := telnetAdapter.Start(*socketPath)
	if err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigChan
	fmt.Printf("Received signal %v, shutting down...\n", sig)

	// Stop telnet server
	err = telnetAdapter.Stop()
	if err != nil {
		log.Printf("Error stopping telnet server: %v", err)
	}

	fmt.Println("Process manager shutdown complete")
}
