package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsnested"
	"github.com/freeflowuniverse/herolauncher/pkg/vfswebdav"
)

func main() {
	// Parse command-line flags
	addr := flag.String("addr", ":8080", "Address to listen on (e.g., :8080)")
	rootDir := flag.String("root", "", "Root directory to serve (required)")
	nestedMode := flag.Bool("nested", false, "Use nested VFS mode with multiple directories")
	nestedDirs := flag.String("dirs", "", "Comma-separated list of directories to mount in nested mode (format: mountpoint=path,mountpoint2=path2)")
	flag.Parse()

	// Validate required flags
	if *rootDir == "" && !*nestedMode {
		fmt.Println("Error: -root flag is required unless using -nested mode")
		flag.Usage()
		os.Exit(1)
	}

	var vfsImpl vfs.VFSImplementation

	if *nestedMode {
		// Create a nested VFS
		nestedVFS := vfsnested.New()
		vfsImpl = nestedVFS

		// If nested directories are specified, mount them
		if *nestedDirs != "" {
			// Parse the nested directories
			dirs := parseNestedDirs(*nestedDirs)
			for mountpoint, path := range dirs {
				// Create a local VFS for each directory
				localVFS, err := vfslocal.New(path)
				if err != nil {
					log.Fatalf("Error creating local VFS for %s: %v", path, err)
				}

				// Add the local VFS to the nested VFS
				err = nestedVFS.AddVFS(mountpoint, localVFS)
				if err != nil {
					log.Fatalf("Error adding VFS at mountpoint %s: %v", mountpoint, err)
				}

				log.Printf("Mounted %s at %s", path, mountpoint)
			}
		} else {
			fmt.Println("Warning: No directories specified for nested mode. Using empty VFS.")
		}
	} else {
		// Create a local VFS
		localVFS, err := vfslocal.New(*rootDir)
		if err != nil {
			log.Fatalf("Error creating local VFS: %v", err)
		}
		vfsImpl = localVFS
		log.Printf("Serving WebDAV from local directory: %s", *rootDir)
	}

	// Create a WebDAV server with the VFS implementation
	server := vfswebdav.NewServer(vfsImpl, *addr)

	// Start the server
	log.Printf("Starting WebDAV server on %s", *addr)
	log.Printf("Connect to the server using: http://localhost%s", *addr)
	log.Fatal(server.Start())
}

// parseNestedDirs parses a comma-separated list of mountpoint=path pairs
func parseNestedDirs(dirs string) map[string]string {
	result := make(map[string]string)
	if dirs == "" {
		return result
	}

	// Split by comma
	pairs := strings.Split(dirs, ",")
	for _, pair := range pairs {
		// Split by equals sign
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			log.Printf("Warning: Invalid directory pair format: %s (expected mountpoint=path)", pair)
			continue
		}

		mountpoint := parts[0]
		path := parts[1]

		// Ensure mountpoint starts with a slash
		if !strings.HasPrefix(mountpoint, "/") {
			mountpoint = "/" + mountpoint
		}

		// Validate the path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Printf("Warning: Directory does not exist: %s", path)
			continue
		}

		result[mountpoint] = path
	}

	return result
}
