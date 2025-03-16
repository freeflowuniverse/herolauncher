package handlers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory"
	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/playbook"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	handlerfactory.BaseHandler
}

// Play processes all actions for this handler's actor
func (h *BaseHandler) Play(script string, handler interface{}) (string, error) {
	pb, err := playbook.NewFromText(script)
	if err != nil {
		return "", fmt.Errorf("failed to parse heroscript: %v", err)
	}

	// Find all actions for this actor
	actions, err := pb.FindActions(0, h.GetActorName(), "", playbook.ActionTypeUnknown)
	if err != nil {
		return "", fmt.Errorf("failed to find actions: %v", err)
	}

	if len(actions) == 0 {
		return "", fmt.Errorf("no actions found for actor: %s", h.GetActorName())
	}

	var results []string

	// Process each action
	for _, action := range actions {
		// Convert action name to method name (e.g., "disk_add" -> "DiskAdd")
		methodName := convertToMethodName(action.Name)

		// Get the method from the handler
		method := reflect.ValueOf(handler).MethodByName(methodName)
		if !method.IsValid() {
			return "", fmt.Errorf("action not supported: %s.%s", h.GetActorName(), action.Name)
		}

		// Call the method with the action's heroscript
		actionScript := action.HeroScript()
		args := []reflect.Value{reflect.ValueOf(actionScript)}
		result := method.Call(args)

		// Get the result
		if len(result) > 0 {
			resultStr := result[0].String()
			results = append(results, resultStr)
		}
	}

	return strings.Join(results, "\n"), nil
}

// ParseParams parses parameters from a heroscript action
func (h *BaseHandler) ParseParams(script string) (*playbook.ParamsParser, error) {
	pb, err := playbook.NewFromText(script)
	if err != nil {
		return nil, fmt.Errorf("failed to parse heroscript: %v", err)
	}

	// Get the first action
	if len(pb.Actions) == 0 {
		return nil, fmt.Errorf("no actions found in script")
	}

	return pb.Actions[0].Params, nil
}

// Helper functions for name conversion

// convertToMethodName converts an action name to a method name
// e.g., "disk_add" -> "DiskAdd"
func convertToMethodName(actionName string) string {
	parts := strings.Split(actionName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[0:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// convertToActionName converts a method name to an action name
// e.g., "DiskAdd" -> "disk_add"
func convertToActionName(methodName string) string {
	var result strings.Builder
	for i, char := range methodName {
		if i > 0 && 'A' <= char && char <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(char)
	}
	return strings.ToLower(result.String())
}
