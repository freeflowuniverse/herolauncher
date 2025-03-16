package handlerfactory

import (
	"fmt"
	"reflect"
	"strings"

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

// HandlerFactory manages a collection of handlers
type HandlerFactory struct {
	handlers map[string]Handler
}

// NewHandlerFactory creates a new handler factory
func NewHandlerFactory() *HandlerFactory {
	return &HandlerFactory{
		handlers: make(map[string]Handler),
	}
}

// RegisterHandler registers a handler with the factory
func (f *HandlerFactory) RegisterHandler(handler Handler) error {
	actorName := handler.GetActorName()
	if actorName == "" {
		return fmt.Errorf("handler has no actor name")
	}

	if _, exists := f.handlers[actorName]; exists {
		return fmt.Errorf("handler for actor '%s' already registered", actorName)
	}

	f.handlers[actorName] = handler
	return nil
}

// GetHandler returns a handler for the specified actor
func (f *HandlerFactory) GetHandler(actorName string) (Handler, error) {
	handler, exists := f.handlers[actorName]
	if !exists {
		return nil, fmt.Errorf("no handler registered for actor: %s", actorName)
	}
	return handler, nil
}

// ProcessHeroscript processes a heroscript command
func (f *HandlerFactory) ProcessHeroscript(script string) (string, error) {
	pb, err := playbook.NewFromText(script)
	if err != nil {
		return "", fmt.Errorf("failed to parse heroscript: %v", err)
	}

	if len(pb.Actions) == 0 {
		return "", fmt.Errorf("no actions found in script")
	}

	// Group actions by actor
	actorActions := make(map[string][]*playbook.Action)
	for _, action := range pb.Actions {
		actorActions[action.Actor] = append(actorActions[action.Actor], action)
	}

	var results []string

	// Process actions for each actor
	for actorName, actions := range actorActions {
		handler, err := f.GetHandler(actorName)
		if err != nil {
			return "", err
		}

		// Create a playbook with just this actor's actions
		actorPB := playbook.New()
		for _, action := range actions {
			actorAction := actorPB.NewAction(action.CID, action.Name, action.Actor, action.Priority, action.ActionType)
			actorAction.Params = action.Params
		}

		// Process the actions
		result, err := handler.Play(actorPB.HeroScript(true), handler)
		if err != nil {
			return "", err
		}

		results = append(results, result)
	}

	return strings.Join(results, "\n"), nil
}

// GetSupportedActions returns a map of supported actions for each registered actor
func (f *HandlerFactory) GetSupportedActions() map[string][]string {
	result := make(map[string][]string)

	for actorName, handler := range f.handlers {
		handlerType := reflect.TypeOf(handler)
		
		// Get all methods of the handler
		var methods []string
		for i := 0; i < handlerType.NumMethod(); i++ {
			method := handlerType.Method(i)
			
			// Skip methods from BaseHandler and other non-action methods
			if method.Name == "GetActorName" || method.Name == "Play" || method.Name == "ParseParams" {
				continue
			}
			
			// Convert method name to action name (e.g., "DiskAdd" -> "disk_add")
			actionName := convertToActionName(method.Name)
			methods = append(methods, actionName)
		}
		
		result[actorName] = methods
	}

	return result
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
