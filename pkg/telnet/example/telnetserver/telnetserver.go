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
	// Create a new telnet server with authentication, command handlers and debug mode
	server := telnet.NewServer(
		// Authentication handler
		func(secret string) bool {
			// Print the received secret for debugging
			fmt.Printf("Received authentication attempt: '%s'\n", secret)
			return secret == "1234" // This is the password to use
		},
		// Command handler
		handleCommand,
		// Enable debug mode
		true,
	)

	// Set up connection callbacks
	server.SetOnSessionConnect(func(session *telnet.Session) {
		fmt.Println("Session connected")
	})
	server.SetOnSessionDisconnect(func(session *telnet.Session) {
		fmt.Println("Session disconnected")
	})

	// Start the server
	port := 8026 // Using a different port to avoid conflicts
	address := fmt.Sprintf(":%d", port)
	err := server.Start(address)
	if err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}
	defer server.Stop()

	fmt.Println(fmt.Sprintf("Telnet server started on port %d", port))
	fmt.Println(fmt.Sprintf("Use 'telnet localhost %d' to connect", port))
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
}

// handleCommand processes commands from sessions
func handleCommand(session *telnet.Session, command string) error {
	// Debug output to see what command is being received
	fmt.Printf("DEBUG: Command received: '%s'\n", command)
	
	// Split the command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	// Process command based on the first part
	switch parts[0] {
	case "hello":
		session.PrintlnGreen("Hello there!")
		return nil

	case "status":
		// Example of using the table formatter
		headers := []string{"Process", "Status", "PID", "Uptime"}
		rows := [][]string{
			{"web-server", "Running", "1234", "2h 15m"},
			{"database", "Stopped", "-", "-"},
			{"cache", "Running", "5678", "30m"},
		}

		table := telnet.FormatTable(headers, rows, session.IsInteractive())
		session.Write(table)
		return nil

	case "error":
		// Example of error formatting
		session.Write(telnet.FormatError(fmt.Errorf("this is a test error"), session.IsInteractive()))
		return nil

	case "success":
		// Example of success formatting
		session.Write(telnet.FormatSuccess("Operation completed successfully", session.IsInteractive()))
		return nil

	case "result":
		// Example of result formatting
		content := "This is a test result\nWith multiple lines\n"
		session.Write(telnet.FormatResult(content, "job-123", session.IsInteractive()))
		return nil

	case "!!echo":
		// Echo function for debugging
		session.PrintlnGreen("Echo mode activated. Type text and it will be echoed back.")
		session.PrintlnGreen("Send an empty line to exit echo mode.")
		
		// Read lines until an empty line is received
		for {
			echoLine, err := session.ReadLine(false)
			if err != nil {
				return err
			}
			
			// Exit echo mode on empty line
			if echoLine == "" {
				session.PrintlnGreen("Echo mode deactivated.")
				break
			}
			
			// Echo the line back
			session.Println("ECHO: " + echoLine)
		}
		return nil

	default:
		// Unknown command
		session.PrintlnYellow(fmt.Sprintf("Unknown command: %s", command))
		session.Println("Try 'hello', 'status', 'error', 'success', 'result', or '!!echo'")
		return nil
	}
}
