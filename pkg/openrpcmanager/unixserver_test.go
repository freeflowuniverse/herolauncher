package openrpcmanager

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUnixServer(t *testing.T) {
	// Create a temporary socket path
	tempDir, err := os.MkdirTemp("", "openrpc-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "openrpc.sock")

	// Create OpenRPC manager
	schema := createTestSchema()
	handlers := createTestHandlers()
	manager, err := NewOpenRPCManager(schema, handlers, "test-secret")
	if err != nil {
		t.Fatalf("Failed to create OpenRPCManager: %v", err)
	}

	// Create and start Unix server
	server, err := NewUnixServer(manager, socketPath)
	if err != nil {
		t.Fatalf("Failed to create UnixServer: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start UnixServer: %v", err)
	}
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test connection
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to Unix socket: %v", err)
	}
	defer conn.Close()

	// Test echo method
	t.Run("Echo method", func(t *testing.T) {
		request := RPCRequest{
			Method:  "echo",
			Params:  json.RawMessage(`{"message":"hello world"}`),
			ID:      1,
			JSONRPC: "2.0",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		_, err = conn.Write(requestData)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// Read response
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var response RPCResponse
		if err := json.Unmarshal(buf[:n], &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check response
		if response.Error != nil {
			t.Fatalf("Received error response: %v", response.Error)
		}

		// Note: JSON unmarshaling may convert numbers to float64, so we need to check the value not exact type
		if fmt.Sprintf("%v", response.ID) != fmt.Sprintf("%v", request.ID) {
			t.Errorf("Response ID mismatch. Expected: %v, Got: %v", request.ID, response.ID)
		}

		// Check result
		resultMap, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got: %T", response.Result)
		}

		if resultMap["message"] != "hello world" {
			t.Errorf("Expected 'hello world', got: %v", resultMap["message"])
		}
	})

	// Test add method
	t.Run("Add method", func(t *testing.T) {
		request := RPCRequest{
			Method:  "add",
			Params:  json.RawMessage(`{"a":5,"b":7}`),
			ID:      2,
			JSONRPC: "2.0",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		_, err = conn.Write(requestData)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// Read response
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var response RPCResponse
		if err := json.Unmarshal(buf[:n], &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check response
		if response.Error != nil {
			t.Fatalf("Received error response: %v", response.Error)
		}

		// Note: JSON unmarshaling may convert numbers to float64, so we need to check the value not exact type
		if fmt.Sprintf("%v", response.ID) != fmt.Sprintf("%v", request.ID) {
			t.Errorf("Response ID mismatch. Expected: %v, Got: %v", request.ID, response.ID)
		}

		// Check result
		resultValue, ok := response.Result.(float64)
		if !ok {
			t.Fatalf("Expected float64 result, got: %T", response.Result)
		}

		if resultValue != float64(12) {
			t.Errorf("Expected 12, got: %v", resultValue)
		}
	})

	// Test authenticated method
	t.Run("Authenticated method", func(t *testing.T) {
		request := RPCRequest{
			Method:  "secure.method",
			Params:  json.RawMessage(`{}`),
			ID:      3,
			Secret:  "test-secret",
			JSONRPC: "2.0",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		_, err = conn.Write(requestData)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// Read response
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var response RPCResponse
		if err := json.Unmarshal(buf[:n], &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check response
		if response.Error != nil {
			t.Fatalf("Received error response: %v", response.Error)
		}

		// Note: JSON unmarshaling may convert numbers to float64, so we need to check the value not exact type
		if fmt.Sprintf("%v", response.ID) != fmt.Sprintf("%v", request.ID) {
			t.Errorf("Response ID mismatch. Expected: %v, Got: %v", request.ID, response.ID)
		}

		// Check result
		resultValue, ok := response.Result.(string)
		if !ok {
			t.Fatalf("Expected string result, got: %T", response.Result)
		}

		if resultValue != "secure data" {
			t.Errorf("Expected 'secure data', got: %v", resultValue)
		}
	})

	// Test authentication failure
	t.Run("Authentication failure", func(t *testing.T) {
		request := RPCRequest{
			Method:  "secure.method",
			Params:  json.RawMessage(`{}`),
			ID:      4,
			Secret:  "wrong-secret",
			JSONRPC: "2.0",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		_, err = conn.Write(requestData)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// Read response
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var response RPCResponse
		if err := json.Unmarshal(buf[:n], &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check response
		if response.Error == nil {
			t.Fatal("Expected error response, but got nil")
		}

		// Note: JSON unmarshaling may convert numbers to float64, so we need to check the value not exact type
		if fmt.Sprintf("%v", response.ID) != fmt.Sprintf("%v", request.ID) {
			t.Errorf("Response ID mismatch. Expected: %v, Got: %v", request.ID, response.ID)
		}

		if response.Error.Code != -32603 {
			t.Errorf("Expected error code -32603, got: %v", response.Error.Code)
		}
	})

	// Test non-existent method
	t.Run("Non-existent method", func(t *testing.T) {
		request := RPCRequest{
			Method:  "nonexistent",
			Params:  json.RawMessage(`{}`),
			ID:      5,
			JSONRPC: "2.0",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		_, err = conn.Write(requestData)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// Read response
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var response RPCResponse
		if err := json.Unmarshal(buf[:n], &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check response
		if response.Error == nil {
			t.Fatal("Expected error response, but got nil")
		}

		// Note: JSON unmarshaling may convert numbers to float64, so we need to check the value not exact type
		if fmt.Sprintf("%v", response.ID) != fmt.Sprintf("%v", request.ID) {
			t.Errorf("Response ID mismatch. Expected: %v, Got: %v", request.ID, response.ID)
		}

		if response.Error.Code != -32603 {
			t.Errorf("Expected error code -32603, got: %v", response.Error.Code)
		}
	})

	// Test discovery method
	t.Run("Discovery method", func(t *testing.T) {
		request := RPCRequest{
			Method:  "rpc.discover",
			Params:  json.RawMessage(`{}`),
			ID:      6,
			JSONRPC: "2.0",
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		_, err = conn.Write(requestData)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// Read response
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var response RPCResponse
		if err := json.Unmarshal(buf[:n], &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Check response
		if response.Error != nil {
			t.Fatalf("Received error response: %v", response.Error)
		}

		// Note: JSON unmarshaling may convert numbers to float64, so we need to check the value not exact type
		if fmt.Sprintf("%v", response.ID) != fmt.Sprintf("%v", request.ID) {
			t.Errorf("Response ID mismatch. Expected: %v, Got: %v", request.ID, response.ID)
		}

		// Check that we got a valid schema
		resultMap, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got: %T", response.Result)
		}

		if resultMap["openrpc"] != "1.2.6" {
			t.Errorf("Expected OpenRPC version 1.2.6, got: %v", resultMap["openrpc"])
		}

		methods, ok := resultMap["methods"].([]interface{})
		if !ok {
			t.Fatalf("Expected methods array, got: %T", resultMap["methods"])
		}

		if len(methods) < 3 {
			t.Errorf("Expected at least 3 methods, got: %d", len(methods))
		}
	})
}
