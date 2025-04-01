package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/herohandler"
)

func main() {
	// Start the herohandler in a goroutine
	go startHeroHandler()

	// Run a simple telnet test
	time.Sleep(1 * time.Second) // Give the server time to start
	fmt.Println("\n=== Testing HeroHandler Telnet Server ===")
	fmt.Println("The herohandler is running with a process manager handler registered.")
	fmt.Println("You can connect to the telnet server at:")
	fmt.Println("- Unix socket: /tmp/hero.sock")
	fmt.Println("- TCP address: localhost:8023")
	fmt.Println("\nYou can use the following command to connect:")
	fmt.Println("  telnet localhost 8023")
	fmt.Println("  nc -U /tmp/hero.sock")

	// Start interactive shell
	fmt.Println("\n=== Interactive Shell ===")
	fmt.Println("Type 'help' to see available commands")
	fmt.Println("Type 'exit' to quit")

	startInteractiveShell()
}

func startInteractiveShell() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)

		switch input {
		case "exit", "quit":
			fmt.Println("Exiting...")
			// Send termination signal to the herohandler goroutine
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return

		case "help":
			showHelp()

		case "status":
			fmt.Println("HeroHandler is running")
			fmt.Println("Telnet server is active at:")
			fmt.Println("- Unix socket: /tmp/hero.sock")
			fmt.Println("- TCP address: localhost:8023")

		case "actions":
			showSupportedActions()

		case "test":
			runTestScript()

		default:
			if input != "" {
				fmt.Println("Unknown command. Type 'help' to see available commands.")
			}
		}
	}
}

func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help     - Show this help message")
	fmt.Println("  status   - Show herohandler status")
	fmt.Println("  actions  - Show supported actions for registered handlers")
	fmt.Println("  test     - Run a test heroscript")
	fmt.Println("  exit     - Exit the program")
}

func showSupportedActions() {
	// We need to implement this function to get supported actions
	// Since we can't directly access the factory field, we'll use the telnet interface
	script := "!!core.actions"
	
	// Try TCP first, then Unix socket if TCP fails
	result, err := Send(script, "localhost:8023", false)
	if err != nil {
		fmt.Printf("TCP connection failed, trying Unix socket: %v\n", err)
		result, err = Send(script, "/tmp/hero.sock", true)
		if err != nil {
			fmt.Printf("Error getting supported actions: %v\n", err)
			return
		}
	}
	
	fmt.Println("Supported actions by actor:")
	fmt.Println(result)
}

// Send connects to the telnet server and sends a command, returning the response
func Send(command string, address string, isUnixSocket bool) (string, error) {
	var conn net.Conn
	var err error

	// Connect to the server based on the address type
	if isUnixSocket {
		conn, err = net.Dial("unix", address)
	} else {
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return "", fmt.Errorf("error connecting to server: %v", err)
	}
	defer conn.Close()

	// Create a reader for the connection
	reader := bufio.NewReader(conn)

	// Read the welcome message
	_, err = reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading welcome message: %v", err)
	}

	// Send the command
	fmt.Fprintf(conn, "%s\n", command)

	// Read the response with a timeout
	result := ""
	ch := make(chan string)
	errCh := make(chan error)

	go func() {
		var response strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				errCh <- fmt.Errorf("error reading response: %v", err)
				return
			}
			response.WriteString(line)
			
			// If we've received a complete response, break
			if strings.Contains(line, "\n") && strings.TrimSpace(line) == "" {
				break
			}
		}
		ch <- response.String()
	}()

	select {
	case result = <-ch:
		return result, nil
	case err = <-errCh:
		return "", err
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("timeout waiting for response")
	}
}

func runTestScript() {
	// Simple test script for the process manager
	script := `
!!process.list format:json
`

	fmt.Println("Running test script:")
	fmt.Println(script)

	// Send the script to the telnet server
	// Try TCP first, then Unix socket if TCP fails
	result, err := Send(script, "localhost:8023", false)
	if err != nil {
		fmt.Printf("TCP connection failed, trying Unix socket: %v\n", err)
		result, err = Send(script, "/tmp/hero.sock", true)
		if err != nil {
			fmt.Printf("Unix socket connection failed: %v\n", err)
			
			// We can't directly access the factory field, so we'll just report the error
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	fmt.Println("Result:")
	fmt.Println(result)
}

func startHeroHandler() {
	if err := herohandler.Init(); err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}
	if err := herohandler.StartTelnet(); err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}
}
