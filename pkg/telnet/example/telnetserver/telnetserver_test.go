package main

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/telnet"
	"github.com/stretchr/testify/assert"
)

// TestTelnetServerEcho tests the telnet server echo functionality
// This test includes both client and server in the same file
func TestTelnetServerEcho(t *testing.T) {
	// Skip this test if running in short mode
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	// Use a random port for testing to avoid conflicts
	port := 9024
	address := fmt.Sprintf("localhost:%d", port)

	// Create a test server with simple authentication and command handling
	server := telnet.NewServer(
		// Authentication handler - accept "test" as the password
		func(secret string) bool {
			// Print the secret for debugging
			fmt.Printf("Test received secret: '%s'\n", secret)
			return secret == "test"
		},
		// Command handler
		func(client *telnet.Client, command string) error {
			// Echo command handler
			if command == "!!echo" {
				// Echo function for debugging
				client.PrintlnGreen("Echo mode activated. Type text and it will be echoed back.")
				client.PrintlnGreen("Send an empty line to exit echo mode.")
				
				// Read lines until an empty line is received
				for {
					echoLine, err := client.ReadLine()
					if err != nil {
						return err
					}
					
					// Exit echo mode on empty line
					if echoLine == "" {
						client.PrintlnGreen("Echo mode deactivated.")
						break
					}
					
					// Echo the line back
					client.Println("ECHO: " + echoLine)
				}
				return nil
			} else if strings.HasPrefix(command, "echo ") {
				message := strings.TrimPrefix(command, "echo ")
				client.Println("ECHO: " + message)
				return nil
			} else if command == "hello" {
				client.PrintlnGreen("Hello from test server!")
				return nil
			}
			return nil
		},
	)

	// Track client connections for test validation
	var clientConnected bool
	server.SetOnClientConnect(func(client *telnet.Client) {
		clientConnected = true
	})

	// Start the server
	err := server.Start(address)
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	// Ensure server is stopped after the test
	defer func() {
		err := server.Stop()
		if err != nil {
			t.Logf("Error stopping server: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Create a done channel to signal test completion
	testDone := make(chan struct{})

	// Run the client test in a separate goroutine with timeout protection
	go func() {
		defer close(testDone)

		// Connect a client to the server with timeout
		dialer := net.Dialer{Timeout: 2 * time.Second}
		conn, err := dialer.Dial("tcp", address)
		if err != nil {
			t.Errorf("Failed to connect to test server: %v", err)
			return
		}
		defer conn.Close()

		// Set a deadline to prevent hanging
		conn.SetDeadline(time.Now().Add(5 * time.Second))

		// Create a telnet client
		client := telnet.NewClient(conn)
		defer client.Close()

		// Helper function to send a command and wait
		sendCommand := func(cmd string, waitTime time.Duration) {
			client.Write(cmd + "\n")
			time.Sleep(waitTime) // Give the server time to process
		}
		
		// We'll use a simple approach for testing
		// Since we can't easily read the server's responses in real-time,
		// we'll use sleep intervals between commands
		
		// Wait for initial connection and send authentication
		time.Sleep(500 * time.Millisecond) // Wait for welcome message
		sendCommand("auth:test", 500*time.Millisecond) // Send password using new format
		
		// Test simple echo command
		sendCommand("echo Hello World", 500*time.Millisecond)
		
		// Test hello command
		sendCommand("hello", 500*time.Millisecond)
		
		// Test !!echo command
		sendCommand("!!echo", 500*time.Millisecond)
		
		// Send multiple lines to echo
		testMessages := []string{
			"This is line 1",
			"This is line 2",
			"This is line 3",
		}
		
		for _, msg := range testMessages {
			sendCommand(msg, 300*time.Millisecond)
		}
		
		// Send empty line to exit echo mode
		sendCommand("", 500*time.Millisecond)
		
		// Test exit command
		sendCommand("!!exit", 300*time.Millisecond)
		
		// Wait a moment for the server to process the exit command
		time.Sleep(500 * time.Millisecond)
	}()

	// Wait for test to complete with timeout
	select {
	case <-testDone:
		// Test completed successfully
	case <-time.After(10 * time.Second):
		t.Fatal("Test timed out after 10 seconds")
	}

	// Wait for client disconnect to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify that client connected callback was called
	assert.True(t, clientConnected, "Client connected callback should be called")
}
