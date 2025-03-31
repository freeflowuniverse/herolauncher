package openrpcmanager

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// createTestSchema creates a test OpenRPC schema
func createTestSchema() OpenRPCSchema {
	return OpenRPCSchema{
		OpenRPC: "1.2.6",
		Info: InfoObject{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Methods: []MethodObject{
			{
				Name: "echo",
				Params: []ContentDescriptorObject{
					{
						Name:   "message",
						Schema: SchemaObject{"type": "object"},
					},
				},
				Result: &ContentDescriptorObject{
					Name:   "result",
					Schema: SchemaObject{"type": "object"},
				},
			},
			{
				Name: "add",
				Params: []ContentDescriptorObject{
					{
						Name:   "numbers",
						Schema: SchemaObject{"type": "object"},
					},
				},
				Result: &ContentDescriptorObject{
					Name:   "result",
					Schema: SchemaObject{"type": "number"},
				},
			},
			{
				Name: "secure.method",
				Params: []ContentDescriptorObject{},
				Result: &ContentDescriptorObject{
					Name:   "result",
					Schema: SchemaObject{"type": "string"},
				},
			},
		},
	}
}

// createTestHandlers creates test handlers for the schema
func createTestHandlers() map[string]RPCHandler {
	return map[string]RPCHandler{
		"echo": func(params json.RawMessage) (interface{}, error) {
			var data interface{}
			if err := json.Unmarshal(params, &data); err != nil {
				return nil, err
			}
			return data, nil
		},
		"add": func(params json.RawMessage) (interface{}, error) {
			var numbers struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}
			if err := json.Unmarshal(params, &numbers); err != nil {
				return nil, err
			}
			return numbers.A + numbers.B, nil
		},
		"secure.method": func(params json.RawMessage) (interface{}, error) {
			return "secure data", nil
		},
	}
}

// TestNewOpenRPCManager tests the creation of a new OpenRPC manager
func TestNewOpenRPCManager(t *testing.T) {
	secret := "test-secret"
	schema := createTestSchema()
	handlers := createTestHandlers()

	manager, err := NewOpenRPCManager(schema, handlers, secret)
	if err != nil {
		t.Fatalf("Failed to create OpenRPCManager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager is nil")
	}

	if manager.GetSecret() != secret {
		t.Errorf("Secret mismatch. Expected: %s, Got: %s", secret, manager.GetSecret())
	}

	if manager.handlers == nil {
		t.Error("handlers map not initialized")
	}

	// Test the default manager for backward compatibility
	defaultManager := NewDefaultOpenRPCManager(secret)
	if defaultManager == nil {
		t.Fatal("Default manager is nil")
	}

	if defaultManager.GetSecret() != secret {
		t.Errorf("Secret mismatch in default manager. Expected: %s, Got: %s", secret, defaultManager.GetSecret())
	}
}

// TestRegisterHandler tests registering a handler to the OpenRPC manager
func TestRegisterHandler(t *testing.T) {
	manager := NewDefaultOpenRPCManager("test-secret")

	// Define a mock handler
	mockHandler := func(params json.RawMessage) (interface{}, error) {
		return "mock response", nil
	}

	// Add a method to the schema
	manager.schema.Methods = append(manager.schema.Methods, MethodObject{
		Name:   "test.method",
		Params: []ContentDescriptorObject{},
		Result: &ContentDescriptorObject{
			Name:   "result",
			Schema: SchemaObject{"type": "string"},
		},
	})

	// Register the handler
	err := manager.RegisterHandler("test.method", mockHandler)
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// Check if handler was registered
	if _, exists := manager.handlers["test.method"]; !exists {
		t.Error("Handler was not registered")
	}

	// Try to register the same handler again, should fail
	err = manager.RegisterHandler("test.method", mockHandler)
	if err == nil {
		t.Error("Expected error when registering duplicate handler, but got nil")
	}

	// Try to register a handler for a method not in the schema
	err = manager.RegisterHandler("not.in.schema", mockHandler)
	if err == nil {
		t.Error("Expected error when registering handler for method not in schema, but got nil")
	}
}

// TestHandleRequest tests handling an RPC request
func TestHandleRequest(t *testing.T) {
	schema := createTestSchema()
	handlers := createTestHandlers()
	manager, err := NewOpenRPCManager(schema, handlers, "test-secret")
	if err != nil {
		t.Fatalf("Failed to create OpenRPCManager: %v", err)
	}

	// Test the echo handler
	testParams := json.RawMessage(`{"message":"hello world"}`)
	result, err := manager.HandleRequest("echo", testParams)
	if err != nil {
		t.Fatalf("Failed to handle request: %v", err)
	}

	// Convert result to map for comparison
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got: %T", result)
	}

	if resultMap["message"] != "hello world" {
		t.Errorf("Expected 'hello world', got: %v", resultMap["message"])
	}

	// Test the add handler
	addParams := json.RawMessage(`{"a":5,"b":7}`)
	addResult, err := manager.HandleRequest("add", addParams)
	if err != nil {
		t.Fatalf("Failed to handle add request: %v", err)
	}

	// Check the result type and value
	resultValue, ok := addResult.(float64)
	if !ok {
		t.Fatalf("Expected result type float64, got: %T", addResult)
	}
	if resultValue != float64(12) {
		t.Errorf("Expected 12, got: %v", resultValue)
	}

	// Test with non-existent method
	_, err = manager.HandleRequest("nonexistent", testParams)
	if err == nil {
		t.Error("Expected error for non-existent method, but got nil")
	}

	// Test the discovery method
	discoveryResult, err := manager.HandleRequest("rpc.discover", json.RawMessage(`{}`)) 
	if err != nil {
		t.Fatalf("Failed to handle discovery request: %v", err)
	}

	// Verify the discovery result is the schema
	discoverySchema, ok := discoveryResult.(OpenRPCSchema)
	if !ok {
		t.Fatalf("Expected OpenRPCSchema result, got: %T", discoveryResult)
	}

	if discoverySchema.OpenRPC != schema.OpenRPC {
		t.Errorf("Expected OpenRPC version %s, got: %s", schema.OpenRPC, discoverySchema.OpenRPC)
	}

	if len(discoverySchema.Methods) != len(schema.Methods) {
		t.Errorf("Expected %d methods, got: %d", len(schema.Methods), len(discoverySchema.Methods))
	}
}

// TestHandleRequestWithAuthentication tests handling a request with authentication
func TestHandleRequestWithAuthentication(t *testing.T) {
	secret := "test-secret"
	schema := createTestSchema()
	handlers := createTestHandlers()
	manager, err := NewOpenRPCManager(schema, handlers, secret)
	if err != nil {
		t.Fatalf("Failed to create OpenRPCManager: %v", err)
	}

	// Test with correct secret
	result, err := manager.HandleRequestWithAuthentication("secure.method", json.RawMessage(`{}`), secret)
	if err != nil {
		t.Fatalf("Failed to handle authenticated request: %v", err)
	}

	if result != "secure data" {
		t.Errorf("Expected 'secure data', got: %v", result)
	}

	// Test with incorrect secret
	_, err = manager.HandleRequestWithAuthentication("secure.method", json.RawMessage(`{}`), "wrong-secret")
	if err == nil {
		t.Error("Expected authentication error, but got nil")
	}
}

// TestUnregisterHandler tests removing a handler from the OpenRPC manager
func TestUnregisterHandler(t *testing.T) {
	manager := NewDefaultOpenRPCManager("test-secret")

	// Define a mock handler
	mockHandler := func(params json.RawMessage) (interface{}, error) {
		return "mock response", nil
	}

	// Add a method to the schema
	manager.schema.Methods = append(manager.schema.Methods, MethodObject{
		Name:   "test.method",
		Params: []ContentDescriptorObject{},
		Result: &ContentDescriptorObject{
			Name:   "result",
			Schema: SchemaObject{"type": "string"},
		},
	})

	// Register the handler
	manager.RegisterHandler("test.method", mockHandler)

	// Unregister the handler
	err := manager.UnregisterHandler("test.method")
	if err != nil {
		t.Fatalf("Failed to unregister handler: %v", err)
	}

	// Check if handler was unregistered
	if _, exists := manager.handlers["test.method"]; exists {
		t.Error("Handler was not unregistered")
	}

	// Try to unregister non-existent handler
	err = manager.UnregisterHandler("nonexistent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent handler, but got nil")
	}

	// Try to unregister the discovery method
	err = manager.UnregisterHandler("rpc.discover")
	if err == nil {
		t.Error("Expected error when unregistering discovery method, but got nil")
	}
}

// TestIntrospection tests the introspection functionality
func TestIntrospection(t *testing.T) {
	// Create a test manager
	schema := createTestSchema()
	handlers := createTestHandlers()
	manager, err := NewOpenRPCManager(schema, handlers, "test-secret")
	assert.NoError(t, err)
	
	// The introspection handler is already registered in NewOpenRPCManager
	
	// Make some test calls to generate logs
	_, err = manager.HandleRequest("echo", json.RawMessage(`{"message":"hello"}`))
	assert.NoError(t, err)
	
	_, err = manager.HandleRequestWithAuthentication("echo", json.RawMessage(`{"message":"authenticated"}`), "test-secret")
	assert.NoError(t, err)
	
	_, err = manager.HandleRequestWithAuthentication("echo", json.RawMessage(`{"message":"auth-fail"}`), "wrong-secret")
	assert.Error(t, err)
	
	// Wait a moment to ensure timestamps are different
	time.Sleep(10 * time.Millisecond)
	
	// Call the introspection handler
	result, err := manager.handleIntrospection(json.RawMessage(`{"limit":10}`))
	assert.NoError(t, err)
	
	// Verify the result
	response, ok := result.(struct {
		Logs     []CallLog `json:"logs"`
		Total    int       `json:"total"`
		Filtered int       `json:"filtered"`
	})
	assert.True(t, ok)
	
	// Should have 3 logs (2 successful calls, 1 auth failure)
	assert.Equal(t, 3, response.Total)
	assert.Equal(t, 3, response.Filtered)
	assert.Len(t, response.Logs, 3)
	
	// Test filtering by method
	result, err = manager.handleIntrospection(json.RawMessage(`{"method":"echo"}`))
	assert.NoError(t, err)
	response, ok = result.(struct {
		Logs     []CallLog `json:"logs"`
		Total    int       `json:"total"`
		Filtered int       `json:"filtered"`
	})
	assert.True(t, ok)
	assert.Equal(t, 3, response.Total) // Total is still 3
	assert.Equal(t, 3, response.Filtered) // All 3 match the method filter
	
	// Test filtering by status
	result, err = manager.handleIntrospection(json.RawMessage(`{"status":"error"}`))
	assert.NoError(t, err)
	response, ok = result.(struct {
		Logs     []CallLog `json:"logs"`
		Total    int       `json:"total"`
		Filtered int       `json:"filtered"`
	})
	assert.True(t, ok)
	assert.Equal(t, 3, response.Total) // Total is still 3
	assert.Equal(t, 1, response.Filtered) // Only 1 error
	assert.Len(t, response.Logs, 1)
	assert.Equal(t, "error", response.Logs[0].Status)
}

// TestListMethods tests listing all registered methods
func TestListMethods(t *testing.T) {
	schema := createTestSchema()
	handlers := createTestHandlers()
	manager, err := NewOpenRPCManager(schema, handlers, "test-secret")
	if err != nil {
		t.Fatalf("Failed to create OpenRPCManager: %v", err)
	}

	// List all methods
	registeredMethods := manager.ListMethods()

	// Check if all methods plus discovery and introspection methods are listed
	expectedCount := len(schema.Methods) + 2 // +2 for rpc.discover and rpc.introspect
	if len(registeredMethods) != expectedCount {
		t.Errorf("Expected %d methods, got %d", expectedCount, len(registeredMethods))
	}

	// Check if all schema methods are in the list
	for _, methodObj := range schema.Methods {
		found := false
		for _, registeredMethod := range registeredMethods {
			if registeredMethod == methodObj.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Method %s not found in list", methodObj.Name)
		}
	}

	// Check if discovery method is in the list
	found := false
	for _, registeredMethod := range registeredMethods {
		if registeredMethod == "rpc.discover" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Discovery method 'rpc.discover' not found in list")
	}
}

// TestSchemaValidation tests that the schema validation works correctly
func TestSchemaValidation(t *testing.T) {
	secret := "test-secret"
	schema := createTestSchema()

	// Test with missing handler
	incompleteHandlers := map[string]RPCHandler{
		"echo": func(params json.RawMessage) (interface{}, error) {
			return nil, nil
		},
		// Missing "add" handler
		"secure.method": func(params json.RawMessage) (interface{}, error) {
			return nil, nil
		},
	}

	_, err := NewOpenRPCManager(schema, incompleteHandlers, secret)
	if err == nil {
		t.Error("Expected error when missing handler for schema method, but got nil")
	}

	// Test with extra handler not in schema
	extraHandlers := createTestHandlers()
	extraHandlers["not.in.schema"] = func(params json.RawMessage) (interface{}, error) {
		return nil, nil
	}

	_, err = NewOpenRPCManager(schema, extraHandlers, secret)
	if err == nil {
		t.Error("Expected error when handler has no corresponding method in schema, but got nil")
	}
}
