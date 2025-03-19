package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces/openapi"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to run the OpenAPI server on")
	dbPath := flag.String("db", "./vfsdb", "Path to the vfsdb database")
	flag.Parse()

	// Create the VFS implementation using vfsdb
	vfsImpl, err := vfsdb.NewVFSDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to create VFSDB: %v", err)
	}
	defer vfsImpl.destroy()

	// Print information
	fmt.Printf("Starting VFS OpenAPI server on port %d\n", *port)
	fmt.Printf("Using VFSDB at path: %s\n", *dbPath)
	fmt.Println("API Endpoints:")
	fmt.Println("- GET    /{path}          - Read file/list directory/read symlink")
	fmt.Println("- PUT    /{path}          - Create or update file")
	fmt.Println("- POST   /{path}?op=mkdir - Create directory")
	fmt.Println("- POST   /{path}?op=append - Append to file")
	fmt.Println("- POST   /{path}?op=symlink&target={target} - Create symlink")
	fmt.Println("- POST   /{path}?op=copy&src={src} - Copy file/directory")
	fmt.Println("- POST   /{path}?op=move&src={src} - Move file/directory")
	fmt.Println("- POST   /{path}?op=rename&old={old} - Rename file/directory")
	fmt.Println("- DELETE /{path}          - Delete file/directory/symlink")

	// Start the OpenAPI server
	err = openapi.StartVFSOpenAPIServer(vfsImpl, *port)
	if err != nil {
		log.Fatalf("Failed to start OpenAPI server: %v", err)
	}
}
