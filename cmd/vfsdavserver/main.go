package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/freeflowuniverse/herolauncher/pkg/vfsdav"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "WebDAV server port")
	host := flag.String("host", "localhost", "WebDAV server host")
	rootDir := flag.String("dir", "", "Root directory for WebDAV server (defaults to temp directory if not specified)")
	verbose := flag.Bool("verbose", true, "Enable verbose logging")
	flag.Parse()

	// Set up the root directory
	var rootPath string
	if *rootDir == "" {
		// Create a temporary directory if no root directory is specified
		tempDir, err := os.MkdirTemp("", "vfsdav-server")
		if err != nil {
			log.Fatalf("Failed to create temporary directory: %v", err)
		}
		rootPath = tempDir
		log.Printf("Created temporary directory for WebDAV server: %s", rootPath)
		
		// Clean up the temporary directory when the server exits
		defer os.RemoveAll(rootPath)
	} else {
		// Use the specified directory
		var err error
		rootPath, err = filepath.Abs(*rootDir)
		if err != nil {
			log.Fatalf("Failed to get absolute path: %v", err)
		}
		
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(rootPath, 0755); err != nil {
			log.Fatalf("Failed to create root directory: %v", err)
		}
		
		log.Printf("Using directory for WebDAV server: %s", rootPath)
	}
	
	// Create a local VFS backend
	vfsImpl, err := vfslocal.New(rootPath)
	if err != nil {
		log.Fatalf("Failed to create VFS: %v", err)
	}
	
	// Create the server address
	addr := fmt.Sprintf("%s:%d", *host, *port)
	
	// Create and configure the WebDAV server
	server := vfsdav.NewServer(vfsImpl, addr)
	
	// Set up a logger for WebDAV requests if verbose mode is enabled
	if *verbose {
		// Wrap the handler to add logging
		originalHandler := server.Handler()
		loggingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("WebDAV: %s %s", r.Method, r.URL.Path)
			originalHandler.ServeHTTP(w, r)
		})
		
		// Replace the server's handler with our logging handler
		http.Handle("/", loggingHandler)
	}
	
	// Start the server in a goroutine
	go func() {
		log.Printf("Starting WebDAV server at http://%s", addr)
		log.Printf("Serving files from: %s", rootPath)
		
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Create some example files in the root directory for testing
	if *rootDir == "" {
		if err := createExampleFiles(rootPath); err != nil {
			log.Printf("Warning: Failed to create example files: %v", err)
		}
	}
	
	// Wait for termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Println("Server is running. Press Ctrl+C to stop.")
	<-stop
	
	log.Println("Server stopped")
}

// createExampleFiles creates some example files in the root directory for testing
func createExampleFiles(rootPath string) error {
	// Create a text file
	textFilePath := filepath.Join(rootPath, "hello.txt")
	if err := os.WriteFile(textFilePath, []byte("Hello, WebDAV!"), 0644); err != nil {
		return fmt.Errorf("failed to create text file: %w", err)
	}
	log.Printf("Created example text file: %s", textFilePath)

	// Create a binary file (a small PNG image)
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	pngFilePath := filepath.Join(rootPath, "small-image.png")
	if err := os.WriteFile(pngFilePath, pngData, 0644); err != nil {
		return fmt.Errorf("failed to create PNG file: %w", err)
	}
	log.Printf("Created example PNG file: %s", pngFilePath)

	// Create a directory
	dirPath := filepath.Join(rootPath, "test-dir")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	log.Printf("Created example directory: %s", dirPath)

	// Create a file in the directory
	dirFilePath := filepath.Join(dirPath, "file-in-dir.txt")
	if err := os.WriteFile(dirFilePath, []byte("This is a file in a directory"), 0644); err != nil {
		return fmt.Errorf("failed to create file in directory: %w", err)
	}
	log.Printf("Created example file in directory: %s", dirFilePath)

	// Create files with different sizes
	fileSizes := map[string]int{
		"file-10kb.dat":  10 * 1024,
		"file-100kb.dat": 100 * 1024,
		"file-500kb.dat": 500 * 1024,
		"file-1mb.dat":   1 * 1024 * 1024,
	}

	for fileName, size := range fileSizes {
		filePath := filepath.Join(rootPath, fileName)
		data := make([]byte, size)

		// Fill with a pattern for verification
		for i := 0; i < size; i++ {
			data[i] = byte((i * 37) % 256)
		}

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", fileName, err)
		}
		log.Printf("Created example file %s (%d bytes)", filePath, size)
	}

	return nil
}
