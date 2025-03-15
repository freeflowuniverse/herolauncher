package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs9p"
	"github.com/knusbaum/go9p"
)

func main() {
	// Parse command line arguments
	var (
		dbPath = flag.String("db", "vfs9p.db", "Path to the database file")
		addr   = flag.String("addr", "0.0.0.0:5640", "Address to listen on")
	)
	flag.Parse()

	// Create absolute path for the database
	absPath, err := filepath.Abs(*dbPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create VFS
	log.Printf("Creating VFS with database at %s", absPath)
	vfsImpl, err := vfsdb.NewFromPath(absPath)
	if err != nil {
		log.Fatalf("Failed to create VFS: %v", err)
	}

	// Create VFS9P adapter
	adapter := vfs9p.NewVFS9P(vfsImpl)

	// Serve 9p
	log.Printf("Serving 9p on %s", *addr)
	err = go9p.Serve(*addr, adapter)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
