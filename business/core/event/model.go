package event

import (
	"context"
	"fmt"
)

// HandleFunc represents a function that can handle an event
type HandleFunc func(context.Context, Event) error

// Event represents an event between core domains
type Event struct {
	Source    string
	Type      string
	RawParams []byte
}

// String implements the Stringer interface
func (e Event) String() string {
	return fmt.Sprintf(
		"Event{Source:%#v, Type:%#v, RawParams:%#v}",
		e.Source, e.Type, string(e.RawParams),
	)
}
