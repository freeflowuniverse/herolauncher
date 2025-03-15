package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/telnet"
)

func main() {
	// Connect to the telnet server
	conn, err := net.Dial("tcp", "localhost:8023")
	if err != nil {
		log.Fatalf("Failed to connect to telnet server: %v", err)
	}
	defer conn.Close()

	// Create a new telnet client
	client := telnet.NewClient(conn)
	defer client.Close()

	fmt.Println("Connected to telnet server. Type '1234' to authenticate.")
	fmt.Println("Type 'exit' or press Ctrl+C to quit.")

	// Read from stdin and send to server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			break
		}

		// Send the input to the server
		client.Write(input + "\n")

		// Read the response from the server
		response, err := client.ReadLine()
		if err != nil {
			log.Printf("Error reading from server: %v", err)
			break
		}

		// Print the response
		fmt.Println("Server response:", response)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from stdin: %v", err)
	}

	fmt.Println("Disconnected from server.")
}
