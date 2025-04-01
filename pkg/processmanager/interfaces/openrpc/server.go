package openrpc

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
)

// Server represents the Process Manager OpenRPC server
type Server struct {
	processManager interfaces.ProcessManagerInterface
	socketPath     string
	openRPCMgr     *openrpcmanager.OpenRPCManager
	unixServer     *openrpcmanager.UnixServer
	isRunning      bool
}

// NewServer creates a new Process Manager OpenRPC server
func NewServer(processManager interfaces.ProcessManagerInterface, socketPath string) (*Server, error) {
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

	// Create a new handler - no authentication now
	handler := NewHandler(processManager)

	// Create a new OpenRPC manager - using empty string for auth (no authentication)
	openRPCMgr, err := openrpcmanager.NewOpenRPCManager(schema, handler.GetHandlers(), "")
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRPC manager: %w", err)
	}

	// Create a new Unix server
	unixServer, err := openrpcmanager.NewUnixServer(openRPCMgr, socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Unix server: %w", err)
	}

	return &Server{
		processManager: processManager,
		socketPath:     socketPath,
		openRPCMgr:     openRPCMgr,
		unixServer:     unixServer,
		isRunning:      false,
	}, nil
}

// Start starts the Process Manager OpenRPC server
func (s *Server) Start() error {
	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Start the Unix server
	if err := s.unixServer.Start(); err != nil {
		return fmt.Errorf("failed to start Unix server: %w", err)
	}

	s.isRunning = true
	log.Printf("Process Manager OpenRPC server started on socket: %s", s.socketPath)
	return nil
}

// Stop stops the Process Manager OpenRPC server
func (s *Server) Stop() error {
	if !s.isRunning {
		return fmt.Errorf("server is not running")
	}

	// Stop the Unix server
	if err := s.unixServer.Stop(); err != nil {
		return fmt.Errorf("failed to stop Unix server: %w", err)
	}

	s.isRunning = false
	log.Printf("Process Manager OpenRPC server stopped")
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
