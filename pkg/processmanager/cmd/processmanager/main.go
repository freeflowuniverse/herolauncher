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
	secret := flag.String("secret", "", "Authentication secret for the telnet server")
	flag.Parse()

	// Validate flags
	if *secret == "" {
		log.Fatal("Error: secret is required")
	}

	// Create process manager
	pm := processmanager.NewProcessManager(*secret)

	// Create telnet server
	ts := processmanager.NewTelnetServer(pm)

	// Start telnet server
	fmt.Printf("Starting process manager telnet server on socket: %s\n", *socketPath)
	err := ts.Start(*socketPath)
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
	err = ts.Stop()
	if err != nil {
		log.Printf("Error stopping telnet server: %v", err)
	}

	fmt.Println("Process manager shutdown complete")
}
