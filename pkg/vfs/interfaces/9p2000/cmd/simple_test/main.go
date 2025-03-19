package main

import (
	"fmt"
	"os"

	"github.com/knusbaum/go9p/client"
	"github.com/knusbaum/go9p/proto"
)

func main() {
	// Connect to the 9p server
	fmt.Println("Connecting to 9p server...")
	c, err := client.Dial("tcp", "localhost:9999", "glenda", "/")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		os.Exit(1)
		return
	}
	fmt.Println("Connected to server successfully!")

	// Test listing the root directory
	fmt.Println("Listing root directory...")
	stats, err := c.Readdir("/")
	if err != nil {
		fmt.Printf("Failed to read root directory: %v\n", err)
		os.Exit(1)
		return
	}
	fmt.Printf("Root directory contains %d entries\n", len(stats))
	for _, stat := range stats {
		fmt.Printf("- %s (mode: %o)\n", stat.Name, stat.Mode)
	}

	// Test creating a simple file
	fmt.Println("\nCreating a simple file...")
	f, err := c.Create("/simple.txt", 0644)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		os.Exit(1)
		return
	}
	fmt.Println("File created successfully!")

	// Write to the file
	content := []byte("Hello, 9p2000!")
	fmt.Println("Writing to file...")
	n, err := f.Write(content)
	if err != nil {
		fmt.Printf("Failed to write to file: %v\n", err)
		os.Exit(1)
		return
	}
	fmt.Printf("Wrote %d bytes to file\n", n)
	f.Close()

	// Read the file back
	fmt.Println("\nReading file back...")
	f, err = c.Open("/simple.txt", proto.Oread)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		os.Exit(1)
		return
	}
	
	data := make([]byte, 100)
	n, err = f.Read(data)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
		return
	}
	data = data[:n]
	fmt.Printf("Read %d bytes: %s\n", n, string(data))
	f.Close()

	fmt.Println("\nSimple test completed successfully!")
}
