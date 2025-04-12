package herohandler

import (
	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/core"
)

// GetFactory returns the handler factory
func (h *HeroHandler) GetFactory() *core.HandlerFactory {
	return h.factory
}

// RegisterHandler registers a handler with the factory
func (h *HeroHandler) RegisterHandler(handler core.Handler) error {
	return h.factory.RegisterHandler(handler)
}
