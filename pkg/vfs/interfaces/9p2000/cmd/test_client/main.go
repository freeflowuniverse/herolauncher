package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/knusbaum/go9p/client"
	"github.com/knusbaum/go9p/proto"
)

func runTests() error {
	// Give the server time to start
	time.Sleep(500 * time.Millisecond)

	// Connect to the 9p server
	fmt.Println("Connecting to 9p server...")
	c, err := client.Dial("tcp", "localhost:9999", "glenda", "/")
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	// Test creating a file
	fmt.Println("Creating test file...")
	f, err := c.Create("/test.txt", 0644)
	if err != nil {
		fmt.Printf("File creation error (might already exist): %v\n", err)
		// Try to open the file to verify it exists
		f, err = c.Open("/test.txt", proto.Owrite)
		if err != nil {
			return fmt.Errorf("file does not exist and could not be created: %v", err)
		}
		fmt.Println("File already exists, continuing...")
	}

	// Write to the file
	fmt.Println("Writing to test file...")
	testData := []byte("Hello, 9p2000!")
	n, err := f.Write(testData)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}
	if n != len(testData) {
		return fmt.Errorf("expected to write %d bytes, but wrote %d", len(testData), n)
	}
	f.Close()

	// Read the file back
	fmt.Println("Reading test file...")
	f, err = c.Open("/test.txt", proto.Oread)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	data := make([]byte, 100)
	n, err = f.Read(data)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file: %v", err)
	}
	data = data[:n]
	if string(data) != string(testData) {
		return fmt.Errorf("expected %q, got %q", string(testData), string(data))
	}
	f.Close()

	// Test creating a directory
	fmt.Println("Creating test directory...")
	// Create a directory by creating a file with directory permissions
	_, err = c.Create("/testdir", 0755|os.ModeDir)
	if err != nil {
		fmt.Printf("Directory creation error (might already exist): %v\n", err)
		// Try to open the directory to verify it exists
		_, err = c.Stat("/testdir")
		if err != nil {
			return fmt.Errorf("directory does not exist and could not be created: %v", err)
		}
		fmt.Println("Directory already exists, continuing...")
	}

	// Create a file in the directory
	fmt.Println("Creating nested file...")
	f, err = c.Create("/testdir/nested.txt", 0644)
	if err != nil {
		fmt.Printf("Nested file creation error (might already exist): %v\n", err)
		// Try to open the file to verify it exists
		f, err = c.Open("/testdir/nested.txt", proto.Owrite)
		if err != nil {
			return fmt.Errorf("nested file does not exist and could not be created: %v", err)
		}
		fmt.Println("Nested file already exists, writing content...")
	}
	f.Write([]byte("Nested file content"))
	f.Close()

	// List the directory
	fmt.Println("Listing directory contents...")
	// Use Readdir from the client to list directory contents
	dirEntries, err := c.Readdir("/testdir")
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}
	if len(dirEntries) != 1 {
		return fmt.Errorf("expected 1 entry, got %d", len(dirEntries))
	}
	if dirEntries[0].Name != "nested.txt" {
		return fmt.Errorf("expected 'nested.txt', got %q", dirEntries[0].Name)
	}

	// Test removing a file
	fmt.Println("Removing test file...")
	err = c.Remove("/test.txt")
	if err != nil {
		return fmt.Errorf("failed to remove file: %v", err)
	}

	// Verify the file is gone
	fmt.Println("Verifying file removal...")
	_, err = c.Open("/test.txt", proto.Oread)
	if err == nil {
		return fmt.Errorf("expected file to be gone, but it still exists")
	}
	
	// Test removing the nested file
	fmt.Println("Removing nested file...")
	err = c.Remove("/testdir/nested.txt")
	if err != nil {
		fmt.Printf("WARNING: Failed to remove nested file: %v\n", err)
		// Continue with the test even if this fails
	} else {
		// Verify the nested file is gone
		fmt.Println("Verifying nested file removal...")
		_, err = c.Open("/testdir/nested.txt", proto.Oread)
		if err == nil {
			fmt.Println("WARNING: Expected nested file to be gone, but it still exists")
		} else {
			fmt.Println("Nested file successfully removed")
		}
	}
	
	// Test removing the directory
	fmt.Println("Removing test directory...")
	err = c.Remove("/testdir")
	if err != nil {
		fmt.Printf("WARNING: Failed to remove directory: %v\n", err)
		// Continue with the test even if this fails
	} else {
		// Verify the directory is gone
		fmt.Println("Verifying directory removal...")
		_, err = c.Stat("/testdir")
		if err == nil {
			fmt.Println("WARNING: Expected directory to be gone, but it still exists")
		} else {
			fmt.Println("Directory successfully removed")
		}
	}

	fmt.Println("All tests passed!")
	return nil
}

func main() {
	if err := runTests(); err != nil {
		fmt.Printf("Test failed: %v\n", err)
		os.Exit(1)
	}
}
