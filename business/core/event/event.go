// Package event provides business access to events in the system.
package event

import "go.uber.org/zap"

// Core manages the set of APIs for event access.
type Core struct {
	log      *zap.SugaredLogger
	handlers map[string]map[string][]HandleFunc
}

// NewCore constructs a Core for event API access.
func NewCore(log *zap.SugaredLogger) *Core {
	return &Core{
		log:      log,
		handlers: map[string]map[string][]HandleFunc{},
	}
}
