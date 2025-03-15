package vfswebdav

import (
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
	return s.handler
}

// Start starts the WebDAV server
func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.handler)
}

// StartTLS starts the WebDAV server with TLS
func (s *Server) StartTLS(certFile, keyFile string) error {
	return http.ListenAndServeTLS(s.addr, certFile, keyFile, s.handler)
}
