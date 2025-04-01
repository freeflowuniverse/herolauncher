package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/core"
	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/fakehandler"
	"github.com/freeflowuniverse/herolauncher/pkg/telnet"
)

func main() {
	// Create a temporary socket path for testing
	socketPath := "/tmp/herolauncher/fakehandler.sock"

	// Create directory if it doesn't exist
	if err := os.MkdirAll("/tmp/herolauncher", 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Remove socket file if it already exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			log.Fatalf("Failed to remove existing socket: %v", err)
		}
	}

	// Create a new fake handler
	handler := fakehandler.NewFakeHandler()

	// Create a new telnet server
	server := telnet.NewServer(socketPath)
	server.SetHandler(handler)

	// Start the server in a goroutine
	go func() {
		fmt.Printf("Starting fake handler server on %s\n", socketPath)
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(500 * time.Millisecond)

	// Create a test client to demonstrate the handler
	fmt.Println("Creating test client...")
	client, err := core.NewClient(socketPath)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Connect to the server
	fmt.Println("Connecting to server...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Test various commands
	testCommands(client)

	// Set up signal handling for graceful shutdown
	fmt.Println("\nServer is running. Press Ctrl+C to stop.")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Stop the server
	fmt.Println("Stopping server...")
	server.Stop()
	fmt.Println("Server stopped")
}

func testCommands(client *core.Client) {
	// Test help command
	fmt.Println("\n--- Help Information ---")
	response, err := client.SendCommand("!!fake.help")
	printResponse("Help", response, err)

	// Test success command
	fmt.Println("\n--- Success Message ---")
	response, err = client.SendCommand("!!fake.return_success message:'Custom success message'")
	printResponse("Success", response, err)

	// Test JSON command
	fmt.Println("\n--- JSON Response ---")
	response, err = client.SendCommand("!!fake.return_json message:'JSON message' status:'success' code:200")
	printResponse("JSON", response, err)

	// Test error command
	fmt.Println("\n--- Error Message ---")
	response, err = client.SendCommand("!!fake.return_error message:'Custom error message'")
	printResponse("Error", response, err)

	// Test empty response
	fmt.Println("\n--- Empty Response ---")
	response, err = client.SendCommand("!!fake.return_empty")
	printResponse("Empty", response, err)

	// Test large response
	fmt.Println("\n--- Large Response ---")
	response, err = client.SendCommand("!!fake.return_large size:10")
	if err != nil {
		fmt.Printf("Large Response Error: %v\n", err)
	} else {
		lines := 0
		for i, c := range response {
			if c == '\n' {
				lines++
			}
			if i > 100 {
				break
			}
		}
		fmt.Printf("Large Response: First 100 chars of response with %d lines\n", lines)
		if len(response) > 100 {
			fmt.Println(response[:100] + "...")
		} else {
			fmt.Println(response)
		}
	}

	// Test invalid JSON
	fmt.Println("\n--- Invalid JSON ---")
	response, err = client.SendCommand("!!fake.return_invalid_json")
	printResponse("Invalid JSON", response, err)

	// Test malformed error
	fmt.Println("\n--- Malformed Error ---")
	response, err = client.SendCommand("!!fake.return_malformed_error")
	printResponse("Malformed Error", response, err)
}

func printResponse(name string, response string, err error) {
	if err != nil {
		fmt.Printf("%s Error: %v\n", name, err)
	} else {
		if len(response) > 100 {
			fmt.Printf("%s Response: %s...\n", name, response[:100])
		} else {
			fmt.Printf("%s Response: %s\n", name, response)
		}
	}
}
