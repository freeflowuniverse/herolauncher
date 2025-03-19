package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/freeflowuniverse/herolauncher/pkg/vfswebdav/vfsadapter"
)

func main() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "webdav-test")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a VFS implementation
	vfsImpl, err := vfslocal.New(tempDir)
	if err != nil {
		log.Fatalf("Failed to create VFS: %v", err)
	}

	// Create a WebDAV server
	server := vfsadapter.NewWebDAVServer(vfsImpl, ":8084")

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting WebDAV test server on :8084")
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// Test file content
	testContent := []byte("This is a test file for WebDAV read/write operations")
	
	// Test file path
	testFilePath := "/test-file.txt"
	
	// Full URL to the test file
	fileURL := fmt.Sprintf("http://localhost:8084%s", testFilePath)

	// Test 1: Upload a file
	log.Println("Test 1: Uploading file...")
	req, err := http.NewRequest("PUT", fileURL, bytes.NewReader(testContent))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Failed to upload file, status: %s", resp.Status)
	}
	
	log.Printf("File uploaded successfully, status: %s", resp.Status)

	// Test 2: Read the file back
	log.Println("Test 2: Reading file...")
	resp, err = http.Get(fileURL)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to read file, status: %s", resp.Status)
	}
	
	readContent, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	
	// Verify the content
	if !bytes.Equal(readContent, testContent) {
		log.Fatalf("Content mismatch: expected %q, got %q", testContent, readContent)
	}
	
	log.Printf("File read successfully, content matches")

	// Test 3: Verify the file exists on disk
	log.Println("Test 3: Verifying file on disk...")
	diskPath := filepath.Join(tempDir, testFilePath)
	fileInfo, err := os.Stat(diskPath)
	if err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}
	
	if fileInfo.IsDir() {
		log.Fatalf("Expected a file, got a directory")
	}
	
	diskContent, err := os.ReadFile(diskPath)
	if err != nil {
		log.Fatalf("Failed to read file from disk: %v", err)
	}
	
	if !bytes.Equal(diskContent, testContent) {
		log.Fatalf("Disk content mismatch: expected %q, got %q", testContent, diskContent)
	}
	
	log.Printf("File exists on disk, content matches")
	
	log.Println("All tests passed!")
}
