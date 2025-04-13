package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager"
)

// Common errors
var (
	ErrConnectionFailed     = errors.New("failed to connect to OpenRPC server")
	ErrRequestFailed        = errors.New("failed to send request to OpenRPC server")
	ErrResponseFailed       = errors.New("failed to read response from OpenRPC server")
	ErrUnmarshalFailed      = errors.New("failed to unmarshal response")
	ErrUnexpectedResponse   = errors.New("unexpected response format")
	ErrRPCError             = errors.New("RPC error")
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// RPCRequest represents an outgoing RPC request
type RPCRequest struct {
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      int             `json:"id"`
	Secret  string          `json:"secret,omitempty"`
	JSONRPC string          `json:"jsonrpc"`
}

// RPCResponse represents an incoming RPC response
type RPCResponse struct {
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
	JSONRPC string      `json:"jsonrpc"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error returns a string representation of the RPC error
func (e *RPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("RPC error %d: %s - %v", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// IntrospectionResponse represents the response from the rpc.introspect method
type IntrospectionResponse struct {
	Logs     []openrpcmanager.CallLog `json:"logs"`
	Total    int                      `json:"total"`
	Filtered int                      `json:"filtered"`
}

// Client is the interface that all OpenRPC clients must implement
type Client interface {
	// Discover returns the OpenRPC schema
	Discover() (openrpcmanager.OpenRPCSchema, error)

	// Introspect returns information about recent RPC calls
	Introspect(limit int, method string, status string) (IntrospectionResponse, error)

	// Request sends a request to the OpenRPC server and returns the result
	Request(method string, params json.RawMessage, secret string) (interface{}, error)

	// Close closes the client connection
	Close() error
}

// BaseClient provides a base implementation of the Client interface
type BaseClient struct {
	socketPath string
	secret     string
	nextID     int
}

// NewClient creates a new OpenRPC client
func NewClient(socketPath, secret string) *BaseClient {
	return &BaseClient{
		socketPath: socketPath,
		secret:     secret,
		nextID:     1,
	}
}

// Discover returns the OpenRPC schema
func (c *BaseClient) Discover() (openrpcmanager.OpenRPCSchema, error) {
	result, err := c.Request("rpc.discover", json.RawMessage("{}"), "")
	if err != nil {
		return openrpcmanager.OpenRPCSchema{}, err
	}

	// Convert result to schema
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return openrpcmanager.OpenRPCSchema{}, fmt.Errorf("%w: %v", ErrUnmarshalFailed, err)
	}

	var schema openrpcmanager.OpenRPCSchema
	if err := json.Unmarshal(resultJSON, &schema); err != nil {
		return openrpcmanager.OpenRPCSchema{}, fmt.Errorf("%w: %v", ErrUnmarshalFailed, err)
	}

	return schema, nil
}

// Introspect returns information about recent RPC calls
func (c *BaseClient) Introspect(limit int, method string, status string) (IntrospectionResponse, error) {
	// Create the params object
	params := struct {
		Limit  int    `json:"limit,omitempty"`
		Method string `json:"method,omitempty"`
		Status string `json:"status,omitempty"`
	}{
		Limit:  limit,
		Method: method,
		Status: status,
	}

	// Marshal the params
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return IntrospectionResponse{}, fmt.Errorf("failed to marshal introspection params: %v", err)
	}

	// Make the request
	result, err := c.Request("rpc.introspect", paramsJSON, c.secret)
	if err != nil {
		return IntrospectionResponse{}, err
	}

	// Convert result to introspection response
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return IntrospectionResponse{}, fmt.Errorf("%w: %v", ErrUnmarshalFailed, err)
	}

	var response IntrospectionResponse
	if err := json.Unmarshal(resultJSON, &response); err != nil {
		return IntrospectionResponse{}, fmt.Errorf("%w: %v", ErrUnmarshalFailed, err)
	}

	return response, nil
}

// Request sends a request to the OpenRPC server and returns the result
func (c *BaseClient) Request(method string, params json.RawMessage, secret string) (interface{}, error) {
	// Connect to the Unix socket
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer conn.Close()

	// Create the request
	request := RPCRequest{
		Method:  method,
		Params:  params,
		ID:      c.nextID,
		Secret:  secret,
		JSONRPC: "2.0",
	}
	c.nextID++

	// Marshal the request
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Send the request
	_, err = conn.Write(requestData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	// Read the response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrResponseFailed, err)
	}

	// Parse the response
	var response RPCResponse
	if err := json.Unmarshal(buf[:n], &response); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnmarshalFailed, err)
	}

	// Check for errors
	if response.Error != nil {
		return nil, fmt.Errorf("%w: %v", ErrRPCError, response.Error)
	}

	return response.Result, nil
}

// Close closes the client connection
func (c *BaseClient) Close() error {
	// Nothing to do for the base client since we create a new connection for each request
	return nil
}
