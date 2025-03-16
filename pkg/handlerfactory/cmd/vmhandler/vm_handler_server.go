package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory"
)

// The tutorial functions are defined in tutorial.go

func main() {
	// Check if tutorial mode is requested
	addTutorialCommand()

	fmt.Println("Starting VM Handler Example")

	// Create a new handler factory
	factory := handlerfactory.NewHandlerFactory()

	// Create and register the VM handler
	vmHandler := NewVMHandler()
	err := factory.RegisterHandler(vmHandler)
	if err != nil {
		log.Fatalf("Failed to register VM handler: %v", err)
	}

	// Create a telnet server with the handler factory
	server := handlerfactory.NewTelnetServer(factory, "1234")

	// Create socket directory if it doesn't exist
	socketDir := "/tmp"
	err = os.MkdirAll(socketDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create socket directory: %v", err)
	}

	// Start the telnet server on a Unix socket
	socketPath := filepath.Join(socketDir, "vmhandler.sock")
	err = server.Start(socketPath)
	if err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}
	fmt.Printf("Telnet server started on socket: %s\n", socketPath)
	fmt.Printf("Connect with: nc -U %s\n", socketPath)

	// Also start on TCP port for easier access
	err = server.StartTCP("localhost:8024")
	if err != nil {
		log.Fatalf("Failed to start TCP telnet server: %v", err)
	}
	fmt.Println("Telnet server started on TCP: localhost:8024")
	fmt.Println("Connect with: telnet localhost 8024")

	// Print available commands
	fmt.Println("\nVM Handler started. Type '!!vm.help' to see available commands.")
	fmt.Println("Authentication secret: 1234")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop the server
	fmt.Println("Stopping server...")
	err = server.Stop()
	if err != nil {
		log.Fatalf("Failed to stop telnet server: %v", err)
	}
	fmt.Println("Telnet server stopped")
}
