package user

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID               *uuid.UUID    `validate:"omitempty"`
	Name             *string       `validate:"omitempty,min=3"`
	Email            *mail.Address `validate:"omitempty"`
	StartCreatedDate *time.Time    `validate:"omitempty"`
	EndCreatedDate   *time.Time    `validate:"omitempty"`
}
