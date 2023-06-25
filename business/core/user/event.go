package user

import "github.com/google/uuid"

// EventSource represents the source of the given event.
const EventSource = "user"

// Set of user related events.
const (
	EventUpdated = "UserUpdated"
)

// ======================================================

// EventParamsUpdated is the event parameters for the updated event.
type EventParamsUpdated struct {
	UserID uuid.UUID
	UpdateUser
}
