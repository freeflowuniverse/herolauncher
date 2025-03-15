package vfsdav

import (
	"net/http"

	"github.com/emersion/go-webdav"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// Server represents a WebDAV server using a VFS backend
type Server struct {
	adapter *Adapter
	handler *webdav.Handler
	addr    string
}

// NewServer creates a new WebDAV server using the provided VFS implementation
func NewServer(vfsImpl vfs.VFSImplementation, addr string) *Server {
	adapter := New(vfsImpl)
	handler := &webdav.Handler{
		FileSystem: adapter,
	}

	return &Server{
		adapter: adapter,
		handler: handler,
		addr:    addr,
	}
}

// ListenAndServe starts the WebDAV server
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.handler)
}

// Handler returns the underlying http.Handler for the WebDAV server
// This can be used to mount the WebDAV server on a custom HTTP server
func (s *Server) Handler() http.Handler {
	return s.handler
}
