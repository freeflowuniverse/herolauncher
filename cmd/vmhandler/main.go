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
	server := handlerfactory.NewTelnetServer(factory, "secret123")

	// Get home directory for socket path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	// Create socket directory if it doesn't exist
	socketDir := filepath.Join(homeDir, ".herolauncher", "sockets")
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
	fmt.Println("\nAvailable VM commands:")
	fmt.Println("  !!vm.define name:'test_vm' cpu:4 memory:'8GB' storage:'100GB'")
	fmt.Println("  !!vm.start name:'test_vm'")
	fmt.Println("  !!vm.stop name:'test_vm'")
	fmt.Println("  !!vm.disk_add name:'test_vm' size:'50GB' type:'SSD'")
	fmt.Println("  !!vm.list")
	fmt.Println("  !!vm.status name:'test_vm'")
	fmt.Println("  !!vm.delete name:'test_vm' force:true")
	fmt.Println("\nAuthentication secret: secret123")

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
