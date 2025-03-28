package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/ui/book"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
)

func main() {
	// Parse command line arguments
	var contentPath string
	var configPath string
	var port string = "3002" // Default port

	if len(os.Args) > 1 {
		contentPath = os.Args[1]
	} else {
		contentPath = "./content"
	}

	if len(os.Args) > 2 {
		configPath = os.Args[2]
	}

	if len(os.Args) > 3 {
		port = os.Args[3]
	}

	// Ensure content path exists
	if _, err := os.Stat(contentPath); os.IsNotExist(err) {
		log.Fatalf("Content path does not exist: %s", contentPath)
	}

	// Get absolute path for content directory
	absContentPath, err := filepath.Abs(contentPath)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	// Create a local VFS instance
	localVFS, err := vfslocal.New(absContentPath)
	if err != nil {
		log.Fatalf("Error creating local VFS: %v", err)
	}

	log.Printf("Serving content from: %s", absContentPath)

	// Create a new book server
	wikiServer, err := book.NewWikiServer(contentPath, configPath, localVFS)
	if err != nil {
		log.Fatalf("Error creating book server: %v", err)
	}

	// Setup routes
	wikiServer.SetupRoutes()

	// Start the server
	log.Printf("Starting book server on port %s", port)
	if err := wikiServer.Start(port); err != nil {
		log.Fatalf("Error starting book server: %v", err)
	}
}
