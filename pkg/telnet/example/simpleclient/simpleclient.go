package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	// Parse command line arguments
	host := flag.String("host", "localhost", "Host to connect to")
	port := flag.Int("port", 8025, "Port to connect to")
	password := flag.String("password", "1234", "Password for authentication")
	flag.Parse()

	// Connect to the server
	address := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("Connecting to %s...\n", address)
	
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	fmt.Println("Connected!")
	
	// Set up reader and writer
	reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(os.Stdin)
	
	// Start a goroutine to read from the server
	go func() {
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Connection closed: %v\n", err)
				os.Exit(0)
			}
			fmt.Print(message)
		}
	}()
	
	// Wait a moment for the welcome message
	time.Sleep(500 * time.Millisecond)
	
	// Send the password automatically
	fmt.Printf("Sending password: %s\n", *password)
	conn.Write([]byte(*password + "\n"))
	
	// Wait a moment for authentication
	time.Sleep(500 * time.Millisecond)
	
	// Main loop to read from stdin and send to server
	fmt.Println("\nEnter commands (type '!!exit' to quit):")
	for scanner.Scan() {
		line := scanner.Text()
		
		// Check for exit command
		if strings.TrimSpace(line) == "!!exit" {
			fmt.Println("Exiting...")
			break
		}
		
		// Send the line to the server
		conn.Write([]byte(line + "\n"))
		
		// Small delay to allow server response
		time.Sleep(100 * time.Millisecond)
	}
}
