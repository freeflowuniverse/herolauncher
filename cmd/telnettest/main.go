package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	// Parse command line flags
	socketPath := flag.String("socket", "/tmp/processmanager.sock", "Path to the Unix domain socket")
	secret := flag.String("secret", "test123", "Authentication secret for the telnet server")
	flag.Parse()

	// Connect to the socket
	fmt.Printf("Connecting to socket: %s\n", *socketPath)
	conn, err := net.Dial("unix", *socketPath)
	if err != nil {
		fmt.Printf("Error connecting to socket: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Create a reader for the connection
	reader := bufio.NewReader(conn)

	// Start a goroutine to read from the connection
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading from connection: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(line)
		}
	}()

	// Wait for the welcome message
	fmt.Println("Waiting for welcome message...")
	time.Sleep(1 * time.Second)

	// Send authentication
	fmt.Printf("Authenticating with secret: %s\n", *secret)
	fmt.Fprintf(conn, "auth:%s\n", *secret)

	// Wait for authentication to complete
	time.Sleep(1 * time.Second)

	// Send some test commands
	testCommands := []string{
		"  !!process.list  ", // Note the extra spaces to test trimming
		"!!help",
		"unknown command",
		"!!process.status",
	}

	for _, cmd := range testCommands {
		fmt.Printf("Sending command: '%s'\n", cmd)
		fmt.Fprintf(conn, "%s\n", cmd)
		time.Sleep(1 * time.Second)
	}

	// Wait for user input to exit
	fmt.Println("\nPress Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
