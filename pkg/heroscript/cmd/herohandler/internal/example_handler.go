package internal

import (
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory"
	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/handlers"
)

// ExampleHandler handles example actions
type ExampleHandler struct {
	handlers.BaseHandler
	data map[string]string
}

// NewExampleHandler creates a new example handler
func NewExampleHandler() *ExampleHandler {
	return &ExampleHandler{
		BaseHandler: handlers.BaseHandler{
			BaseHandler: handlerfactory.BaseHandler{
				ActorName: "example",
			},
		},
		data: make(map[string]string),
	}
}

// Set handles the example.set action
func (h *ExampleHandler) Set(script string) string {
	params, err := h.BaseHandler.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	key := params.Get("key")
	if key == "" {
		return "Error: key is required"
	}

	value := params.Get("value")
	if value == "" {
		return "Error: value is required"
	}

	h.data[key] = value
	return fmt.Sprintf("Set %s = %s", key, value)
}

// Get handles the example.get action
func (h *ExampleHandler) Get(script string) string {
	params, err := h.BaseHandler.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	key := params.Get("key")
	if key == "" {
		return "Error: key is required"
	}

	value, exists := h.data[key]
	if !exists {
		return fmt.Sprintf("Key '%s' not found", key)
	}

	return fmt.Sprintf("%s = %s", key, value)
}

// List handles the example.list action
func (h *ExampleHandler) List(script string) string {
	if len(h.data) == 0 {
		return "No data stored"
	}

	var result string
	for key, value := range h.data {
		result += fmt.Sprintf("%s = %s\n", key, value)
	}

	return result
}

// Delete handles the example.delete action
func (h *ExampleHandler) Delete(script string) string {
	params, err := h.BaseHandler.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	key := params.Get("key")
	if key == "" {
		return "Error: key is required"
	}

	_, exists := h.data[key]
	if !exists {
		return fmt.Sprintf("Key '%s' not found", key)
	}

	delete(h.data, key)
	return fmt.Sprintf("Deleted key '%s'", key)
}
