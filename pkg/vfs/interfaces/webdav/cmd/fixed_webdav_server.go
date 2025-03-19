package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/net/webdav"
)

func main() {
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

	// Create a WebDAV handler
	webdavHandler := &webdav.Handler{
		FileSystem: webdav.Dir(absRootDir),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WebDAV Error: %s %s - %v", r.Method, r.URL.Path, err)
			} else {
				log.Printf("WebDAV: %s %s", r.Method, r.URL.Path)
			}
		},
	}

	// Create a handler that adds CORS headers and logs requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers to allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, PUT, PATCH, POST, DELETE, OPTIONS, PROPFIND, MKCOL, MOVE, COPY")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Depth, User-Agent, X-File-Size, X-Requested-With, If-Modified-Since, X-File-Name, Cache-Control, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "DAV, content-length, Allow")

		// Handle OPTIONS requests for CORS preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Log the request
		log.Printf("Request: %s %s", r.Method, r.URL.Path)

		// Serve the request
		webdavHandler.ServeHTTP(w, r)
	})

	// Start the server
	log.Printf("Serving WebDAV from local directory: %s", absRootDir)
	log.Printf("Starting WebDAV server on %s", *addr)
	log.Printf("Connect to the server using: http://localhost%s", *addr)
	err = http.ListenAndServe(*addr, handler)
	if err != nil {
		log.Fatalf("Error starting WebDAV server: %v", err)
	}
}
