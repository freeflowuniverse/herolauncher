package fakehandler

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/core"
)

func TestFakeHandler(t *testing.T) {
	handler := NewFakeHandler()

	// Test ReturnSuccess
	t.Run("ReturnSuccess", func(t *testing.T) {
		// Default message
		result := handler.ReturnSuccess("!!fake.return_success")
		if result != "Success message" {
			t.Errorf("Expected 'Success message', got '%s'", result)
		}

		// Custom message
		result = handler.ReturnSuccess("!!fake.return_success message:'Custom success'")
		if result != "Custom success" {
			t.Errorf("Expected 'Custom success', got '%s'", result)
		}
	})

	// Test ReturnError
	t.Run("ReturnError", func(t *testing.T) {
		// Default message
		result := handler.ReturnError("!!fake.return_error")
		if result != "Error: This is a test error" {
			t.Errorf("Expected 'Error: This is a test error', got '%s'", result)
		}

		// Custom message
		result = handler.ReturnError("!!fake.return_error message:'Custom error'")
		if result != "Error: Custom error" {
			t.Errorf("Expected 'Error: Custom error', got '%s'", result)
		}
	})

	// Test ReturnJSON
	t.Run("ReturnJSON", func(t *testing.T) {
		// Default values
		result := handler.ReturnJSON("!!fake.return_json")
		var response FakeResponse
		err := json.Unmarshal([]byte(result), &response)
		if err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}
		if response.Message != "JSON response" || response.Status != "ok" || response.Code != 200 {
			t.Errorf("Unexpected JSON response: %+v", response)
		}

		// Custom values
		result = handler.ReturnJSON("!!fake.return_json message:'Custom JSON' status:'success' code:201")
		err = json.Unmarshal([]byte(result), &response)
		if err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}
		if response.Message != "Custom JSON" || response.Status != "success" || response.Code != 201 {
			t.Errorf("Unexpected JSON response: %+v", response)
		}
	})

	// Test ReturnInvalidJSON
	t.Run("ReturnInvalidJSON", func(t *testing.T) {
		result := handler.ReturnInvalidJSON("!!fake.return_invalid_json")
		var response FakeResponse
		err := json.Unmarshal([]byte(result), &response)
		if err == nil {
			t.Error("Expected JSON parsing error, but got no error")
		}
	})

	// Test ReturnEmpty
	t.Run("ReturnEmpty", func(t *testing.T) {
		result := handler.ReturnEmpty("!!fake.return_empty")
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	// Test ReturnLarge
	t.Run("ReturnLarge", func(t *testing.T) {
		// Default size
		result := handler.ReturnLarge("!!fake.return_large")
		lines := strings.Count(result, "\n") + 1
		if lines != 1000 {
			t.Errorf("Expected 1000 lines, got %d", lines)
		}

		// Custom size
		result = handler.ReturnLarge("!!fake.return_large size:50")
		lines = strings.Count(result, "\n") + 1
		if lines != 50 {
			t.Errorf("Expected 50 lines, got %d", lines)
		}

		// Size limit
		result = handler.ReturnLarge("!!fake.return_large size:20000")
		lines = strings.Count(result, "\n") + 1
		if lines > 10000 {
			t.Errorf("Expected max 10000 lines, got %d", lines)
		}
	})

	// Test ReturnMalformedError
	t.Run("ReturnMalformedError", func(t *testing.T) {
		result := handler.ReturnMalformedError("!!fake.return_malformed_error")
		if result != "error: this is not in the standard format" {
			t.Errorf("Expected 'error: this is not in the standard format', got '%s'", result)
		}
	})

	// Test Help
	t.Run("Help", func(t *testing.T) {
		result := handler.Help("!!fake.help")
		if !strings.Contains(result, "Fake Handler Commands:") {
			t.Errorf("Expected help message to contain 'Fake Handler Commands:', got '%s'", result)
		}
	})

	// Test parameter parsing error
	t.Run("ParameterParsingError", func(t *testing.T) {
		// Create a handler with a broken parameter parser for testing
		brokenHandler := &FakeHandler{
			BaseHandler: core.BaseHandler{
				ActorName: "fake",
			},
		}
		
		// Override the ParseParams method to always return an error
		brokenHandler.BaseHandler.ParseParamsFunc = func(script string) (core.Params, error) {
			return nil, core.ErrInvalidParameters
		}
		
		// Test that the error is properly handled
		result := brokenHandler.ReturnSuccess("!!fake.return_success")
		if !strings.Contains(result, "Error parsing parameters") {
			t.Errorf("Expected error message, got '%s'", result)
		}
	})
}
