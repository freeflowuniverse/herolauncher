package handlers

import (
	"fmt"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory"
	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/playbook"
)

// HandlerFactory manages a collection of handlers for processing HeroScript commands
type HandlerFactory struct {
	handlers map[string]handlerfactory.Handler
}

// NewHandlerFactory creates a new handler factory
func NewHandlerFactory() *HandlerFactory {
	return &HandlerFactory{
		handlers: make(map[string]handlerfactory.Handler),
	}
}

// RegisterHandler registers a handler with the factory
func (f *HandlerFactory) RegisterHandler(handler handlerfactory.Handler) error {
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
func (f *HandlerFactory) GetHandler(actorName string) (handlerfactory.Handler, error) {
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
		// Get supported actions for this handler
		actions, err := getSupportedActions(handler)
		if err == nil && len(actions) > 0 {
			result[actorName] = actions
		}
	}

	return result
}

// getSupportedActions returns a list of supported actions for a handler
func getSupportedActions(handler handlerfactory.Handler) ([]string, error) {
	// This is a simplified implementation
	// In a real implementation, you would use reflection to get all methods
	// that match the pattern for action handlers

	// For now, we'll return an empty list
	return []string{}, nil
}
