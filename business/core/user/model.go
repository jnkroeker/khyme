package user

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/jnkroeker/khyme/business/core/event"
)

// User represents information about an individual user.
type User struct {
	ID           uuid.UUID
	Name         string
	Email        mail.Address
	Roles        []Role
	PasswordHash []byte
	Department   string
	Enabled      bool
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	Name            string
	Email           mail.Address
	Roles           []Role
	Department      string
	Password        string
	PasswordConfirm string
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	Name            *string
	Email           *mail.Address
	Roles           []Role
	Department      *string
	Password        *string
	PasswordConfirm *string
	Enabled         *bool
}

// UpdatedEvent constructs an event for when a user is updated.
func (uu UpdateUser) UpdatedEvent(userID uuid.UUID) event.Event {
	return event.Event{}
}