package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/freeflowuniverse/herolauncher/pkg/vfsdav"
)

func main() {
	// Create a temporary directory for the VFS
	tempDir, err := os.MkdirTemp("", "vfsdav-example")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample file in the temp directory
	sampleFile := filepath.Join(tempDir, "sample.txt")
	err = os.WriteFile(sampleFile, []byte("Hello, WebDAV!"), 0644)
	if err != nil {
		log.Fatalf("Failed to write sample file: %v", err)
	}

	// Create a local VFS implementation
	vfsImpl, err := vfslocal.New(tempDir)
	if err != nil {
		log.Fatalf("Failed to create VFS: %v", err)
	}

	// Create and start the WebDAV server
	addr := "localhost:8080"
	server := vfsdav.NewServer(vfsImpl, addr)
	
	fmt.Printf("WebDAV server started at http://%s\n", addr)
	fmt.Printf("Serving files from: %s\n", tempDir)
	fmt.Printf("Sample file available at: http://%s/sample.txt\n", addr)
	fmt.Println("Press Ctrl+C to stop the server")
	
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
