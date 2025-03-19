package openapi

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/freeflowuniverse/herolib/lib/vfs"
)

// Server represents an OpenAPI server for VFS operations
type Server struct {
	vfsImpl vfs.VFSImplementation
	port    int
}

// NewServer creates a new OpenAPI server for the given VFS implementation
func NewServer(vfsImpl vfs.VFSImplementation, port int) *Server {
	return &Server{
		vfsImpl: vfsImpl,
		port:    port,
	}
}

// Start starts the OpenAPI server
func (s *Server) Start() error {
	// Register handlers
	http.HandleFunc("/", s.handleRequest)

	// Start server
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting VFS OpenAPI server on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleRequest handles all incoming requests and routes them to the appropriate handler
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	// Skip the leading slash for VFS operations
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Handle different HTTP methods
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r, path)
	case http.MethodPut:
		s.handlePut(w, r, path)
	case http.MethodPost:
		s.handlePost(w, r, path)
	case http.MethodDelete:
		s.handleDelete(w, r, path)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
