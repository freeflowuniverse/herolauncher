package fakehandler

import (
	"encoding/json"
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/core"
)

// FakeResponse represents a fake response for testing
type FakeResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Code    int    `json:"code"`
}

// FakeHandler is a handler for testing different response types
type FakeHandler struct {
	core.BaseHandler
}

// NewFakeHandler creates a new fake handler
func NewFakeHandler() *FakeHandler {
	return &FakeHandler{
		BaseHandler: core.BaseHandler{
			ActorName: "fake",
		},
	}
}

// ReturnSuccess returns a success message
func (h *FakeHandler) ReturnSuccess(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	message := params.Get("message")
	if message == "" {
		message = "Success message"
	}
	return message
}

// ReturnError returns an error message
func (h *FakeHandler) ReturnError(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	message := params.Get("message")
	if message == "" {
		message = "This is a test error"
	}
	return fmt.Sprintf("Error: %s", message)
}

// ReturnJSON returns a JSON response
func (h *FakeHandler) ReturnJSON(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	message := params.Get("message")
	if message == "" {
		message = "JSON response"
	}

	status := params.Get("status")
	if status == "" {
		status = "ok"
	}
	code := params.GetIntDefault("code", 200)

	response := FakeResponse{
		Message: message,
		Status:  status,
		Code:    code,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}

	return string(jsonBytes)
}

// ReturnInvalidJSON returns an invalid JSON response
func (h *FakeHandler) ReturnInvalidJSON(script string) string {
	return "{\"message\": \"This is invalid JSON, missing closing brace"
}

// ReturnEmpty returns an empty response
func (h *FakeHandler) ReturnEmpty(script string) string {
	return ""
}

// ReturnLarge returns a large response
func (h *FakeHandler) ReturnLarge(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	size := params.GetIntDefault("size", 1000)
	if size > 10000 {
		size = 10000 // Limit to 10000 to prevent abuse
	}

	response := ""
	for i := 0; i < size; i++ {
		response += fmt.Sprintf("Line %d: This is a test line\n", i+1)
	}

	return response
}

// ReturnMalformedError returns a malformed error message
func (h *FakeHandler) ReturnMalformedError(script string) string {
	return "error: this is not in the standard format"
}

// Help returns help information
func (h *FakeHandler) Help(script string) string {
	return `Fake Handler Commands:
- fake.return_success [message:'custom message'] - Returns a success message
- fake.return_error [message:'custom error'] - Returns an error message
- fake.return_json [message:'custom message'] [status:'custom status'] [code:123] - Returns a JSON response
- fake.return_invalid_json - Returns an invalid JSON response
- fake.return_empty - Returns an empty response
- fake.return_large [size:1000] - Returns a large response with the specified number of lines
- fake.return_malformed_error - Returns a malformed error message
- fake.help - Shows this help message`
}
