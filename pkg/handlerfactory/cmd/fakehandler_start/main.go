package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/herohandler"
)

func main() {
	// Parse command line flags
	socketPath := flag.String("socket", "/tmp/hero.sock", "Unix socket path")
	tcpAddress := flag.String("tcp", "localhost:8023", "TCP address")
	useUnixSocket := flag.Bool("unix", true, "Use Unix socket")
	useTCP := flag.Bool("tcp-enable", false, "Use TCP")
	flag.Parse()

	// Initialize the hero handler
	err := herohandler.Init()
	if err != nil {
		fmt.Printf("Failed to initialize hero handler: %v\n", err)
		os.Exit(1)
	}
	
	// Get the default instance
	handler := herohandler.DefaultInstance

	// The fake handler is already registered in the Init() function
	fmt.Println("Using pre-registered fake handler")

	// Start the server
	if *useUnixSocket || *useTCP {
		fmt.Printf("Starting telnet server\n")
		var socketPathStr, tcpAddressStr string
		if *useUnixSocket {
			socketPathStr = *socketPath
			fmt.Printf("Unix socket: %s\n", socketPathStr)
		}
		if *useTCP {
			tcpAddressStr = *tcpAddress
			fmt.Printf("TCP address: %s\n", tcpAddressStr)
		}
		
		err = handler.StartTelnet(socketPathStr, tcpAddressStr)
		if err != nil {
			fmt.Printf("Failed to start telnet server: %v\n", err)
			os.Exit(1)
		}
	}

	// Print available commands
	factory := handler.GetFactory()
	actions := factory.GetSupportedActions()
	fmt.Println("\nAvailable commands:")
	for actor, commands := range actions {
		fmt.Printf("Actor: %s\n", actor)
		for _, command := range commands {
			fmt.Printf("  !!%s.%s\n", actor, command)
		}
	}

	fmt.Println("\nServer is running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	handler.StopTelnet()
	fmt.Println("Server stopped")
}
