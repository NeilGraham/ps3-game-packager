package consoles

import (
	"fmt"

	"github.com/NeilGraham/rom-organizer/internal/common"
	"github.com/NeilGraham/rom-organizer/internal/detect"
)

// Registry manages console-specific handlers
type Registry struct {
	handlers map[detect.ConsoleType]common.ConsoleHandler
}

// NewRegistry creates a new console registry
func NewRegistry() *Registry {
	registry := &Registry{
		handlers: make(map[detect.ConsoleType]common.ConsoleHandler),
	}

	// Register available console handlers
	registry.RegisterHandler(detect.PS3, NewPS3Handler())

	return registry
}

// RegisterHandler registers a console handler for a specific console type
func (r *Registry) RegisterHandler(consoleType detect.ConsoleType, handler common.ConsoleHandler) {
	r.handlers[consoleType] = handler
}

// GetHandler returns the handler for a specific console type
func (r *Registry) GetHandler(consoleType detect.ConsoleType) (common.ConsoleHandler, error) {
	handler, exists := r.handlers[consoleType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for console type: %s", consoleType.String())
	}
	return handler, nil
}

// GetSupportedConsoles returns a list of all supported console types
func (r *Registry) GetSupportedConsoles() []detect.ConsoleType {
	consoles := make([]detect.ConsoleType, 0, len(r.handlers))
	for consoleType := range r.handlers {
		consoles = append(consoles, consoleType)
	}
	return consoles
}

// IsSupported checks if a console type is supported
func (r *Registry) IsSupported(consoleType detect.ConsoleType) bool {
	_, exists := r.handlers[consoleType]
	return exists
}
