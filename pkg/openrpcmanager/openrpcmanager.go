// Package openrpcmanager provides functionality for managing and handling OpenRPC method calls.
package openrpcmanager

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// RPCHandler is a function that handles an OpenRPC method call
type RPCHandler func(params json.RawMessage) (interface{}, error)

// CallLog represents a log entry for an RPC method call
type CallLog struct {
	Timestamp   time.Time     `json:"timestamp"`   // When the call was made
	Method      string        `json:"method"`      // Method that was called
	Params      interface{}   `json:"params"`      // Parameters passed to the method (may be redacted for security)
	Duration    time.Duration `json:"duration"`    // How long the call took to execute
	Status      string        `json:"status"`      // Success or error
	ErrorMsg    string        `json:"error,omitempty"` // Error message if status is error
	Authenticated bool        `json:"authenticated"` // Whether the call was authenticated
}

// OpenRPCManager manages OpenRPC method handlers and processes requests
type OpenRPCManager struct {
	handlers map[string]RPCHandler
	schema   OpenRPCSchema
	mutex    sync.RWMutex
	secret   string
	
	// Call logging
	callLogs      []CallLog
	callLogsMutex sync.RWMutex
	maxCallLogs   int // Maximum number of call logs to keep
}

// NewOpenRPCManager creates a new OpenRPC manager with the given schema and handlers
func NewOpenRPCManager(schema OpenRPCSchema, handlers map[string]RPCHandler, secret string) (*OpenRPCManager, error) {
	manager := &OpenRPCManager{
		handlers:    make(map[string]RPCHandler),
		schema:      schema,
		secret:      secret,
		callLogs:    make([]CallLog, 0, 100),
		maxCallLogs: 1000, // Default to keeping the last 1000 calls
	}

	// Validate that all methods in the schema have corresponding handlers
	for _, method := range schema.Methods {
		handler, exists := handlers[method.Name]
		if !exists {
			return nil, fmt.Errorf("missing handler for method '%s' defined in schema", method.Name)
		}
		manager.handlers[method.Name] = handler
	}

	// Check for handlers that don't have a corresponding method in the schema
	for name := range handlers {
		found := false
		for _, method := range schema.Methods {
			if method.Name == name {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("handler '%s' has no corresponding method in schema", name)
		}
	}

	// Add the discovery method
	manager.handlers["rpc.discover"] = manager.handleDiscovery
	
	// Add the introspection method
	manager.handlers["rpc.introspect"] = manager.handleIntrospection

	return manager, nil
}

// handleDiscovery implements the OpenRPC service discovery method
func (m *OpenRPCManager) handleDiscovery(params json.RawMessage) (interface{}, error) {
	return m.schema, nil
}

// handleIntrospection implements the OpenRPC service introspection method
// It returns information about recent RPC calls for monitoring and debugging
func (m *OpenRPCManager) handleIntrospection(params json.RawMessage) (interface{}, error) {
	m.callLogsMutex.RLock()
	defer m.callLogsMutex.RUnlock()
	
	// Parse parameters to see if we need to filter or limit results
	var requestParams struct {
		Limit int    `json:"limit,omitempty"`
		Method string `json:"method,omitempty"`
		Status string `json:"status,omitempty"`
	}
	
	// Default limit to 100 if not specified
	requestParams.Limit = 100
	
	// Try to parse parameters, but don't fail if they're invalid
	if len(params) > 0 {
		_ = json.Unmarshal(params, &requestParams)
	}
	
	// Apply limit
	if requestParams.Limit <= 0 || requestParams.Limit > m.maxCallLogs {
		requestParams.Limit = 100
	}
	
	// Create a copy of the logs to avoid holding the lock while filtering
	allLogs := make([]CallLog, len(m.callLogs))
	copy(allLogs, m.callLogs)
	
	// Filter logs based on parameters
	filteredLogs := []CallLog{}
	for i := len(allLogs) - 1; i >= 0 && len(filteredLogs) < requestParams.Limit; i-- {
		log := allLogs[i]
		
		// Apply method filter if specified
		if requestParams.Method != "" && log.Method != requestParams.Method {
			continue
		}
		
		// Apply status filter if specified
		if requestParams.Status != "" && log.Status != requestParams.Status {
			continue
		}
		
		filteredLogs = append(filteredLogs, log)
	}
	
	// Create response
	response := struct {
		Logs []CallLog `json:"logs"`
		Total int      `json:"total"`
		Filtered int   `json:"filtered"`
	}{
		Logs: filteredLogs,
		Total: len(allLogs),
		Filtered: len(filteredLogs),
	}
	
	return response, nil
}

// RegisterHandler is deprecated. Use NewOpenRPCManager with a complete set of handlers instead.
// This method is kept for backward compatibility but will return an error for methods not in the schema.
func (m *OpenRPCManager) RegisterHandler(method string, handler RPCHandler) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if handler already exists
	if _, exists := m.handlers[method]; exists {
		return fmt.Errorf("handler for method '%s' already registered", method)
	}

	// Check if method exists in schema
	found := false
	for _, schemaMethod := range m.schema.Methods {
		if schemaMethod.Name == method {
			found = true
			break
		}
	}

	if !found && method != "rpc.discover" {
		return fmt.Errorf("method '%s' not defined in schema", method)
	}

	m.handlers[method] = handler
	return nil
}

// UnregisterHandler removes a handler for the specified method
// Note: This will make the service non-compliant with its schema
func (m *OpenRPCManager) UnregisterHandler(method string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if handler exists
	if _, exists := m.handlers[method]; !exists {
		return fmt.Errorf("handler for method '%s' not found", method)
	}

	// Don't allow unregistering the discovery method
	if method == "rpc.discover" {
		return fmt.Errorf("cannot unregister the discovery method 'rpc.discover'")
	}

	delete(m.handlers, method)
	return nil
}

// HandleRequest processes an RPC request for the specified method
func (m *OpenRPCManager) HandleRequest(method string, params json.RawMessage) (interface{}, error) {
	// Start timing the request
	startTime := time.Now()
	
	// Create a call log entry
	callLog := CallLog{
		Timestamp:     startTime,
		Method:        method,
		Authenticated: false,
	}
	
	// Parse params for logging, but don't fail if we can't
	var parsedParams interface{}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &parsedParams); err == nil {
			callLog.Params = parsedParams
		} else {
			// If we can't parse the params, just store them as a string
			callLog.Params = string(params)
		}
	}
	
	// Find the handler
	m.mutex.RLock()
	handler, exists := m.handlers[method]
	m.mutex.RUnlock()

	if !exists {
		// Log the error
		callLog.Status = "error"
		callLog.ErrorMsg = fmt.Sprintf("method '%s' not found", method)
		callLog.Duration = time.Since(startTime)
		
		// Add to call logs
		m.logCall(callLog)
		
		return nil, fmt.Errorf("method '%s' not found", method)
	}

	// Execute the handler
	result, err := handler(params)
	
	// Complete the call log
	callLog.Duration = time.Since(startTime)
	if err != nil {
		callLog.Status = "error"
		callLog.ErrorMsg = err.Error()
	} else {
		callLog.Status = "success"
	}
	
	// Add to call logs
	m.logCall(callLog)
	
	return result, err
}

// logCall adds a call log entry to the call logs, maintaining the maximum size
func (m *OpenRPCManager) logCall(log CallLog) {
	m.callLogsMutex.Lock()
	defer m.callLogsMutex.Unlock()
	
	// Add the log to the call logs
	m.callLogs = append(m.callLogs, log)
	
	// Trim the call logs if they exceed the maximum size
	if len(m.callLogs) > m.maxCallLogs {
		m.callLogs = m.callLogs[len(m.callLogs)-m.maxCallLogs:]
	}
}

// HandleRequestWithAuthentication processes an authenticated RPC request
func (m *OpenRPCManager) HandleRequestWithAuthentication(method string, params json.RawMessage, secret string) (interface{}, error) {
	// Start timing the request
	startTime := time.Now()
	
	// Create a call log entry
	callLog := CallLog{
		Timestamp:     startTime,
		Method:        method,
		Authenticated: true,
	}
	
	// Parse params for logging, but don't fail if we can't
	var parsedParams interface{}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &parsedParams); err == nil {
			callLog.Params = parsedParams
		} else {
			// If we can't parse the params, just store them as a string
			callLog.Params = string(params)
		}
	}
	
	// Verify the secret
	if secret != m.secret {
		// Log the authentication failure
		callLog.Status = "error"
		callLog.ErrorMsg = "authentication failed"
		callLog.Duration = time.Since(startTime)
		m.logCall(callLog)
		
		return nil, fmt.Errorf("authentication failed")
	}

	// Execute the handler
	m.mutex.RLock()
	handler, exists := m.handlers[method]
	m.mutex.RUnlock()

	if !exists {
		// Log the error
		callLog.Status = "error"
		callLog.ErrorMsg = fmt.Sprintf("method '%s' not found", method)
		callLog.Duration = time.Since(startTime)
		m.logCall(callLog)
		
		return nil, fmt.Errorf("method '%s' not found", method)
	}

	// Execute the handler
	result, err := handler(params)
	
	// Complete the call log
	callLog.Duration = time.Since(startTime)
	if err != nil {
		callLog.Status = "error"
		callLog.ErrorMsg = err.Error()
	} else {
		callLog.Status = "success"
	}
	
	// Add to call logs
	m.logCall(callLog)
	
	return result, err
}

// ListMethods returns a list of all registered method names
func (m *OpenRPCManager) ListMethods() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	methods := make([]string, 0, len(m.handlers))
	for method := range m.handlers {
		methods = append(methods, method)
	}

	return methods
}

// GetSecret returns the authentication secret
func (m *OpenRPCManager) GetSecret() string {
	return m.secret
}

// GetSchema returns the OpenRPC schema
func (m *OpenRPCManager) GetSchema() OpenRPCSchema {
	return m.schema
}

// NewDefaultOpenRPCManager creates a new OpenRPC manager with a default schema
// This is provided for backward compatibility and testing
func NewDefaultOpenRPCManager(secret string) *OpenRPCManager {
	// Create a minimal default schema
	defaultSchema := OpenRPCSchema{
		OpenRPC: "1.2.6",
		Info: InfoObject{
			Title:   "Default OpenRPC Service",
			Version: "1.0.0",
		},
		Methods: []MethodObject{
			{
				Name:        "rpc.discover",
				Description: "Returns the OpenRPC schema for this service",
				Params:      []ContentDescriptorObject{},
				Result: &ContentDescriptorObject{
					Name:        "schema",
					Description: "The OpenRPC schema",
					Schema:      SchemaObject{"type": "object"},
				},
			},
			{
				Name:        "rpc.introspect",
				Description: "Returns information about recent RPC calls for monitoring and debugging",
				Params: []ContentDescriptorObject{
					{
						Name:        "limit",
						Description: "Maximum number of call logs to return",
						Required:    false,
						Schema:      SchemaObject{"type": "integer", "default": 100},
					},
					{
						Name:        "method",
						Description: "Filter logs by method name",
						Required:    false,
						Schema:      SchemaObject{"type": "string"},
					},
					{
						Name:        "status",
						Description: "Filter logs by status (success or error)",
						Required:    false,
						Schema:      SchemaObject{"type": "string", "enum": []interface{}{"success", "error"}},
					},
				},
				Result: &ContentDescriptorObject{
					Name:        "introspection",
					Description: "Introspection data including call logs",
					Schema:      SchemaObject{"type": "object"},
				},
			},
		},
	}

	// Create the manager directly without validation since we're starting with empty methods
	return &OpenRPCManager{
		handlers: make(map[string]RPCHandler),
		schema:   defaultSchema,
		secret:   secret,
	}
}
