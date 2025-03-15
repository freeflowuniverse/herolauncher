package vfswebdav

import (
	"log"
	"net/http"

	"github.com/emersion/go-webdav"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// Server represents a WebDAV server that uses a VFS implementation as its backend
type Server struct {
	vfsImpl vfs.VFSImplementation
	handler *webdav.Handler
	addr    string
}

// Ensure FileSystem implements webdav.FileSystem
var _ webdav.FileSystem = (*FileSystem)(nil)

// NewServer creates a new WebDAV server with the given VFS implementation
func NewServer(vfsImpl vfs.VFSImplementation, addr string) *Server {
	fs := NewFileSystem(vfsImpl)
	handler := &webdav.Handler{
		FileSystem: fs,
	}

	return &Server{
		vfsImpl: vfsImpl,
		handler: handler,
		addr:    addr,
	}
}

// Handler returns the HTTP handler for the WebDAV server
func (s *Server) Handler() http.Handler {
	// Create a handler that adds CORS headers and logs requests
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("WebDAV: %s %s", r.Method, r.URL.Path)

		// Serve the request
		s.handler.ServeHTTP(w, r)
	})
}

// Start starts the WebDAV server
func (s *Server) Start() error {
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
		log.Printf("WebDAV: %s %s", r.Method, r.URL.Path)

		// Serve the request
		s.handler.ServeHTTP(w, r)
	})

	return http.ListenAndServe(s.addr, handler)
}

// StartTLS starts the WebDAV server with TLS
func (s *Server) StartTLS(certFile, keyFile string) error {
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
		log.Printf("WebDAV: %s %s", r.Method, r.URL.Path)

		// Serve the request
		s.handler.ServeHTTP(w, r)
	})

	return http.ListenAndServeTLS(s.addr, certFile, keyFile, handler)
}
