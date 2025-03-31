package openrpcmanager

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// RPCRequest represents an incoming RPC request
type RPCRequest struct {
	Method     string          `json:"method"`
	Params     json.RawMessage `json:"params"`
	ID         interface{}     `json:"id,omitempty"`
	Secret     string          `json:"secret,omitempty"`
	JSONRPC    string          `json:"jsonrpc"`
}

// RPCResponse represents an outgoing RPC response
type RPCResponse struct {
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
	JSONRPC string          `json:"jsonrpc"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// UnixServer represents a Unix socket server for the OpenRPC manager
type UnixServer struct {
	manager     *OpenRPCManager
	socketPath  string
	listener    net.Listener
	connections map[net.Conn]bool
	mutex       sync.Mutex
	wg          sync.WaitGroup
	done        chan struct{}
}

// NewUnixServer creates a new Unix socket server for the OpenRPC manager
func NewUnixServer(manager *OpenRPCManager, socketPath string) (*UnixServer, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(socketPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Remove socket if it already exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing socket: %w", err)
		}
	}

	return &UnixServer{
		manager:     manager,
		socketPath:  socketPath,
		connections: make(map[net.Conn]bool),
		done:        make(chan struct{}),
	}, nil
}

// Start starts the Unix socket server
func (s *UnixServer) Start() error {
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}
	s.listener = listener

	// Set socket permissions
	if err := os.Chmod(s.socketPath, 0660); err != nil {
		s.listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

// Stop stops the Unix socket server
func (s *UnixServer) Stop() error {
	close(s.done)

	// Close the listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all connections
	s.mutex.Lock()
	for conn := range s.connections {
		conn.Close()
	}
	s.mutex.Unlock()

	// Wait for all goroutines to finish
	s.wg.Wait()

	// Remove the socket file
	os.Remove(s.socketPath)

	return nil
}

// acceptConnections accepts incoming connections
func (s *UnixServer) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.done:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.done:
					return
				default:
					fmt.Printf("Error accepting connection: %v\n", err)
					continue
				}
			}

			s.mutex.Lock()
			s.connections[conn] = true
			s.mutex.Unlock()

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a client connection
func (s *UnixServer) handleConnection(conn net.Conn) {
	defer func() {
		s.mutex.Lock()
		delete(s.connections, conn)
		s.mutex.Unlock()
		conn.Close()
		s.wg.Done()
	}()

	buf := make([]byte, 4096)
	for {
		select {
		case <-s.done:
			return
		default:
			n, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading from connection: %v\n", err)
				}
				return
			}

			if n > 0 {
				go s.handleRequest(conn, buf[:n])
			}
		}
	}
}

// handleRequest processes an RPC request
func (s *UnixServer) handleRequest(conn net.Conn, data []byte) {
	var req RPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendErrorResponse(conn, nil, -32700, "Parse error", err)
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		s.sendErrorResponse(conn, req.ID, -32600, "Invalid Request", "Invalid JSON-RPC version")
		return
	}

	var result interface{}
	var err error

	// Check if authentication is required
	if req.Secret != "" {
		result, err = s.manager.HandleRequestWithAuthentication(req.Method, req.Params, req.Secret)
	} else {
		result, err = s.manager.HandleRequest(req.Method, req.Params)
	}

	if err != nil {
		s.sendErrorResponse(conn, req.ID, -32603, "Internal error", err.Error())
		return
	}

	// Send success response
	response := RPCResponse{
		Result:  result,
		ID:      req.ID,
		JSONRPC: "2.0",
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		s.sendErrorResponse(conn, req.ID, -32603, "Internal error", err.Error())
		return
	}

	conn.Write(responseData)
}

// sendErrorResponse sends an error response
func (s *UnixServer) sendErrorResponse(conn net.Conn, id interface{}, code int, message string, data interface{}) {
	response := RPCResponse{
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID:      id,
		JSONRPC: "2.0",
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Error marshaling error response: %v\n", err)
		return
	}

	conn.Write(responseData)
}
