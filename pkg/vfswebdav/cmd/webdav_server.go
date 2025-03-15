package webdavserver

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/freeflowuniverse/herolauncher/pkg/vfswebdav"
)

// Run starts the WebDAV server
func Run() {
	// Parse command-line flags
	addr := flag.String("addr", ":8080", "Address to listen on (e.g., :8080)")
	rootDir := flag.String("root", "", "Root directory to serve (required)")
	flag.Parse()

	// Validate required flags
	if *rootDir == "" {
		fmt.Println("Error: -root flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Convert to absolute path
	absRootDir, err := filepath.Abs(*rootDir)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	// Ensure the directory exists
	if _, err := os.Stat(absRootDir); os.IsNotExist(err) {
		log.Fatalf("Error creating local VFS: root path does not exist: %s", absRootDir)
	}

	// Create a local VFS implementation
	vfsImpl, err := vfslocal.New(absRootDir)
	if err != nil {
		log.Fatalf("Error creating local VFS: %v", err)
	}

	// Create a WebDAV server with the VFS implementation
	server := vfswebdav.NewServer(vfsImpl, *addr)

	// Start the server
	log.Printf("Serving WebDAV from local directory: %s", absRootDir)
	log.Printf("Starting WebDAV server on %s", *addr)
	log.Printf("Connect to the server using: http://localhost%s", *addr)
	err = server.Start()
	if err != nil {
		log.Fatalf("Error starting WebDAV server: %v", err)
	}
}
