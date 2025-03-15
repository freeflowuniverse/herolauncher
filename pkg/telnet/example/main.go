package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/telnet"
)

func main() {
	// Create a new telnet server with authentication and command handlers
	server := telnet.NewServer(
		// Authentication handler
		func(secret string) bool {
			return secret == "1234" // Replace with actual secret
		},
		// Command handler
		handleCommand,
	)

	// Set up connection callbacks
	server.SetOnClientConnect(func(client *telnet.Client) {
		fmt.Println("Client connected")
	})
	server.SetOnClientDisconnect(func(client *telnet.Client) {
		fmt.Println("Client disconnected")
	})

	// Start the server
	err := server.Start(":8023")
	if err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}
	defer server.Stop()

	fmt.Println("Telnet server started on port 8023")
	fmt.Println("Use 'telnet localhost 8023' to connect")
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
}

// handleCommand processes commands from clients
func handleCommand(client *telnet.Client, command string) error {
	// Split the command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	// Process command based on the first part
	switch parts[0] {
	case "hello":
		client.PrintlnGreen("Hello there!")
		return nil

	case "status":
		// Example of using the table formatter
		headers := []string{"Process", "Status", "PID", "Uptime"}
		rows := [][]string{
			{"web-server", "Running", "1234", "2h 15m"},
			{"database", "Stopped", "-", "-"},
			{"cache", "Running", "5678", "30m"},
		}

		table := telnet.FormatTable(headers, rows, client.IsInteractive())
		client.Write(table)
		return nil

	case "error":
		// Example of error formatting
		client.Write(telnet.FormatError(fmt.Errorf("this is a test error"), client.IsInteractive()))
		return nil

	case "success":
		// Example of success formatting
		client.Write(telnet.FormatSuccess("Operation completed successfully", client.IsInteractive()))
		return nil

	case "result":
		// Example of result formatting
		content := "This is a test result\nWith multiple lines\n"
		client.Write(telnet.FormatResult(content, "job-123", client.IsInteractive()))
		return nil

	default:
		// Unknown command
		client.PrintlnYellow(fmt.Sprintf("Unknown command: %s", command))
		client.Println("Try 'hello', 'status', 'error', 'success', or 'result'")
		return nil
	}
}
