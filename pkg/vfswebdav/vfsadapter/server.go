package vfsadapter

import (
	"log"
	"net/http"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"golang.org/x/net/webdav"
)

// WebDAVServer represents a WebDAV server that uses a VFS implementation
type WebDAVServer struct {
	vfsImpl vfs.VFSImplementation
	addr    string
	handler http.Handler
}

// NewWebDAVServer creates a new WebDAV server using the given VFS implementation
func NewWebDAVServer(vfsImpl vfs.VFSImplementation, addr string) *WebDAVServer {
	// Create a VFS adapter
	adapter := NewVFSAdapter(vfsImpl)

	// Create a WebDAV handler
	webdavHandler := &webdav.Handler{
		FileSystem: adapter,
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, PUT, PATCH, POST, DELETE, OPTIONS, PROPFIND, MKCOL, MOVE, COPY, LOCK, UNLOCK")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Depth, User-Agent, X-File-Size, X-Requested-With, If-Modified-Since, X-File-Name, Cache-Control, Authorization, Destination, Overwrite")
		w.Header().Set("Access-Control-Expose-Headers", "DAV, content-length, Allow, ETag")
		
		// Add headers specifically for macOS Finder compatibility
		w.Header().Set("DAV", "1, 2")
		w.Header().Set("MS-Author-Via", "DAV")
		w.Header().Set("X-Content-Type-Options", "nosniff")

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

	return &WebDAVServer{
		vfsImpl: vfsImpl,
		addr:    addr,
		handler: handler,
	}
}

// Start starts the WebDAV server
func (s *WebDAVServer) Start() error {
	log.Printf("Starting WebDAV server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.handler)
}

// Handler returns the HTTP handler for the WebDAV server
func (s *WebDAVServer) Handler() http.Handler {
	return s.handler
}
