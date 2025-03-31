package openrpc

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces"
)

// Server represents the VFS OpenRPC server
type Server struct {
	vfsManager    interfaces.VFSManager
	socketPath    string
	secret        string
	openRPCMgr    *openrpcmanager.OpenRPCManager
	unixServer    *openrpcmanager.UnixServer
	isRunning     bool
}

// NewServer creates a new VFS OpenRPC server
func NewServer(vfsManager interfaces.VFSManager, socketPath, secret string) (*Server, error) {
	// Ensure the directory for the socket exists
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Load the OpenRPC schema
	schema, err := LoadSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenRPC schema: %w", err)
	}

	// Create a new handler
	handler := NewHandler(vfsManager, secret)

	// Create a new OpenRPC manager
	openRPCMgr, err := openrpcmanager.NewOpenRPCManager(schema, handler.GetHandlers(), secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRPC manager: %w", err)
	}

	// Create a new Unix server
	unixServer, err := openrpcmanager.NewUnixServer(openRPCMgr, socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Unix server: %w", err)
	}

	return &Server{
		vfsManager: vfsManager,
		socketPath: socketPath,
		secret:     secret,
		openRPCMgr: openRPCMgr,
		unixServer: unixServer,
		isRunning:  false,
	}, nil
}

// Start starts the VFS OpenRPC server
func (s *Server) Start() error {
	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Start the Unix server
	if err := s.unixServer.Start(); err != nil {
		return fmt.Errorf("failed to start Unix server: %w", err)
	}

	s.isRunning = true
	log.Printf("VFS OpenRPC server started on socket: %s", s.socketPath)
	return nil
}

// Stop stops the VFS OpenRPC server
func (s *Server) Stop() error {
	if !s.isRunning {
		return fmt.Errorf("server is not running")
	}

	// Stop the Unix server
	if err := s.unixServer.Stop(); err != nil {
		return fmt.Errorf("failed to stop Unix server: %w", err)
	}

	s.isRunning = false
	log.Printf("VFS OpenRPC server stopped")
	return nil
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	return s.isRunning
}

// SocketPath returns the socket path
func (s *Server) SocketPath() string {
	return s.socketPath
}
