package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/paramsparser"
	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/playbook"
)

// Handler interface defines methods that all handlers must implement
type Handler interface {
	GetActorName() string
	Play(script string, handler interface{}) (string, error)
}

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	ActorName string
}

// GetActorName returns the actor name for this handler
func (h *BaseHandler) GetActorName() string {
	return h.ActorName
}

// Play processes all actions for this handler's actor
func (h *BaseHandler) Play(script string, handler interface{}) (string, error) {
	pb, err := playbook.NewFromText(script)
	if err != nil {
		return "", fmt.Errorf("failed to parse heroscript: %v", err)
	}

	// Find all actions for this actor
	actions, err := pb.FindActions(0, h.ActorName, "", playbook.ActionTypeUnknown)
	if err != nil {
		return "", fmt.Errorf("failed to find actions: %v", err)
	}

	if len(actions) == 0 {
		return "", fmt.Errorf("no actions found for actor: %s", h.ActorName)
	}

	var results []string

	// Process each action
	for _, action := range actions {
		// Convert action name to method name (e.g., "disk_add" -> "DiskAdd")
		methodName := convertToMethodName(action.Name)

		// Get the method from the handler
		method := reflect.ValueOf(handler).MethodByName(methodName)
		if !method.IsValid() {
			return "", fmt.Errorf("action not supported: %s.%s", h.ActorName, action.Name)
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
func (h *BaseHandler) ParseParams(script string) (*paramsparser.ParamsParser, error) {
	pb, err := playbook.NewFromText(script)
	if err != nil {
		return nil, fmt.Errorf("failed to parse heroscript: %v", err)
	}

	// Get the first action
	if len(pb.Actions) == 0 {
		return nil, fmt.Errorf("no actions found in script")
	}

	// Get the first action
	action := pb.Actions[0]

	// Check if the action is for this handler
	if action.Actor != h.ActorName {
		return nil, fmt.Errorf("action actor '%s' does not match handler actor '%s'", action.Actor, h.ActorName)
	}

	// The action already has a ParamsParser, so we can just return it
	return action.Params, nil
}
