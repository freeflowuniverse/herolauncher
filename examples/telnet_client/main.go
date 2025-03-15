// Example telnet client for testing keyboard handling and color support
package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Connect to the telnet server
	conn, err := net.Dial("tcp", "localhost:8023")
	if err != nil {
		log.Fatalf("Failed to connect to telnet server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to telnet server")
	fmt.Println("Enter the secret: mysecret")
	fmt.Println("Try using keyboard shortcuts: Ctrl+C, arrow keys, etc.")
	fmt.Println("Type '!!help' or 'h' for help")

	// Create a channel to signal when the connection is closed
	done := make(chan struct{})

	// Goroutine to read from server and print to stdout
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Connection closed")
				return
			}
			
			// Process the buffer to handle telnet protocol bytes
			processedBuf := processTelnetBytes(buf[:n])
			if len(processedBuf) > 0 {
				os.Stdout.Write(processedBuf)
			}
		}
	}()

	// Read from stdin and write to server
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				_, err = conn.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}
	}()

	// Wait for connection to close
	<-done
}

// processTelnetBytes filters out telnet protocol negotiation bytes
func processTelnetBytes(buf []byte) []byte {
	var result []byte
	for i := 0; i < len(buf); i++ {
		// Check for IAC (Interpret As Command) byte
		if buf[i] == 255 { // IAC byte
			// Skip this byte and the next two bytes (command and option)
			if i+2 < len(buf) {
				// Check if it's a 3-byte command (IAC + CMD + OPTION)
				if buf[i+1] >= 251 && buf[i+1] <= 254 {
					i += 2 // Skip command and option bytes
					continue
				}
				
				// Check if it's a subnegotiation (IAC + SB + ... + IAC + SE)
				if buf[i+1] == 250 { // SB - Subnegotiation Begin
					// Find the end of subnegotiation (IAC + SE)
					j := i + 2
					for j < len(buf)-1 {
						if buf[j] == 255 && buf[j+1] == 240 { // IAC + SE
							break
						}
						j++
					}
					
					if j < len(buf)-1 {
						i = j + 1 // Skip to after IAC + SE
						continue
					}
				}
			}
			
			// If we couldn't properly parse the command, just skip the IAC byte
			continue
		}
		
		// Add non-IAC bytes to the result
		result = append(result, buf[i])
	}
	return result
}
