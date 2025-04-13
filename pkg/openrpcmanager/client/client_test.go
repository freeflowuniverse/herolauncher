package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	BaseClient
}

// TestMethod is a test method that returns a greeting
func (c *MockClient) TestMethod(name string) (string, error) {
	params := map[string]string{"name": name}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	result, err := c.Request("test.method", paramsJSON, "")
	if err != nil {
		return "", err
	}

	// Convert result to string
	greeting, ok := result.(string)
	if !ok {
		return "", ErrUnexpectedResponse
	}

	return greeting, nil
}

// SecureMethod is a test method that requires authentication
func (c *MockClient) SecureMethod() (map[string]interface{}, error) {
	result, err := c.Request("secure.method", json.RawMessage("{}"), c.secret)
	if err != nil {
		return nil, err
	}

	// Convert result to map
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, ErrUnexpectedResponse
	}

	return data, nil
}

func TestClient(t *testing.T) {
	// Create a temporary socket path
	tempDir, err := os.MkdirTemp("", "openrpc-client-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "openrpc.sock")
	secret := "test-secret"

	// Create test schema and handlers
	schema := openrpcmanager.OpenRPCSchema{
		OpenRPC: "1.2.6",
		Info: openrpcmanager.InfoObject{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Methods: []openrpcmanager.MethodObject{
			{
				Name: "test.method",
				Params: []openrpcmanager.ContentDescriptorObject{
					{
						Name:   "name",
						Schema: openrpcmanager.SchemaObject{"type": "string"},
					},
				},
				Result: &openrpcmanager.ContentDescriptorObject{
					Name:   "result",
					Schema: openrpcmanager.SchemaObject{"type": "string"},
				},
			},
			{
				Name:   "secure.method",
				Params: []openrpcmanager.ContentDescriptorObject{},
				Result: &openrpcmanager.ContentDescriptorObject{
					Name:   "result",
					Schema: openrpcmanager.SchemaObject{"type": "object"},
				},
			},
		},
	}

	handlers := map[string]openrpcmanager.RPCHandler{
		"test.method": func(params json.RawMessage) (interface{}, error) {
			var request struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(params, &request); err != nil {
				return nil, err
			}
			return "Hello, " + request.Name + "!", nil
		},
		"secure.method": func(params json.RawMessage) (interface{}, error) {
			return map[string]interface{}{
				"secure": true,
				"data":   "sensitive information",
			}, nil
		},
	}

	// Create and start OpenRPC manager and Unix server
	manager, err := openrpcmanager.NewOpenRPCManager(schema, handlers, secret)
	if err != nil {
		t.Fatalf("Failed to create OpenRPCManager: %v", err)
	}

	server, err := openrpcmanager.NewUnixServer(manager, socketPath)
	if err != nil {
		t.Fatalf("Failed to create UnixServer: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start UnixServer: %v", err)
	}
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	client := &MockClient{
		BaseClient: BaseClient{
			socketPath: socketPath,
			secret:     secret,
		},
	}

	// Test Discover method
	t.Run("Discover", func(t *testing.T) {
		schema, err := client.Discover()
		if err != nil {
			t.Fatalf("Discover failed: %v", err)
		}

		if schema.OpenRPC != "1.2.6" {
			t.Errorf("Expected OpenRPC version 1.2.6, got: %s", schema.OpenRPC)
		}

		if len(schema.Methods) < 2 {
			t.Errorf("Expected at least 2 methods, got: %d", len(schema.Methods))
		}

		// Check if our test methods are in the schema
		foundTestMethod := false
		foundSecureMethod := false
		for _, method := range schema.Methods {
			if method.Name == "test.method" {
				foundTestMethod = true
			}
			if method.Name == "secure.method" {
				foundSecureMethod = true
			}
		}

		if !foundTestMethod {
			t.Error("test.method not found in schema")
		}
		if !foundSecureMethod {
			t.Error("secure.method not found in schema")
		}
	})

	// Test TestMethod
	t.Run("TestMethod", func(t *testing.T) {
		greeting, err := client.TestMethod("World")
		if err != nil {
			t.Fatalf("TestMethod failed: %v", err)
		}

		expected := "Hello, World!"
		if greeting != expected {
			t.Errorf("Expected greeting %q, got: %q", expected, greeting)
		}
	})

	// Test Introspect method
	t.Run("Introspect", func(t *testing.T) {
		// Make several requests to generate logs
		_, err := client.TestMethod("World")
		if err != nil {
			t.Fatalf("TestMethod failed: %v", err)
		}

		_, err = client.SecureMethod()
		if err != nil {
			t.Fatalf("SecureMethod failed: %v", err)
		}

		// Test introspection
		response, err := client.Introspect(10, "", "")
		if err != nil {
			t.Fatalf("Introspect failed: %v", err)
		}

		// Verify we have logs
		if response.Total < 2 {
			t.Errorf("Expected at least 2 logs, got: %d", response.Total)
		}

		// Test filtering by method
		response, err = client.Introspect(10, "test.method", "")
		if err != nil {
			t.Fatalf("Introspect with method filter failed: %v", err)
		}

		// Verify filtering works
		for _, log := range response.Logs {
			if log.Method != "test.method" {
				t.Errorf("Expected only test.method logs, got: %s", log.Method)
			}
		}

		// Test filtering by status
		response, err = client.Introspect(10, "", "success")
		if err != nil {
			t.Fatalf("Introspect with status filter failed: %v", err)
		}

		// Verify status filtering works
		for _, log := range response.Logs {
			if log.Status != "success" {
				t.Errorf("Expected only success logs, got: %s", log.Status)
			}
		}
	})

	// Test SecureMethod with valid secret
	t.Run("SecureMethod", func(t *testing.T) {
		data, err := client.SecureMethod()
		if err != nil {
			t.Fatalf("SecureMethod failed: %v", err)
		}

		secure, ok := data["secure"].(bool)
		if !ok || !secure {
			t.Errorf("Expected secure to be true, got: %v", data["secure"])
		}

		sensitiveData, ok := data["data"].(string)
		if !ok || sensitiveData != "sensitive information" {
			t.Errorf("Expected data to be 'sensitive information', got: %v", data["data"])
		}
	})

	// Test SecureMethod with invalid secret
	t.Run("SecureMethod with invalid secret", func(t *testing.T) {
		invalidClient := &MockClient{
			BaseClient: BaseClient{
				socketPath: socketPath,
				secret:     "wrong-secret",
			},
		}

		_, err := invalidClient.SecureMethod()
		if err == nil {
			t.Error("Expected error for invalid secret, but got nil")
		}
	})

	// Test non-existent method
	t.Run("Non-existent method", func(t *testing.T) {
		_, err := client.Request("non.existent", json.RawMessage("{}"), "")
		if err == nil {
			t.Error("Expected error for non-existent method, but got nil")
		}
	})
}
