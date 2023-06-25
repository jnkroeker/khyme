// Package user provides an example of a core business API.
// Right now these calls are just wrapping the data layer.
// At some point you will want auditing or something not
// specific to data/store layer.
package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"github.com/jnkroeker/khyme/business/core/event"
	"github.com/jnkroeker/khyme/business/data/order"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// ===============================================================

// Storer interface declares the behavior this package needs to persist
// and retrieve data
type Storer interface {
	WithinTran(ctx context.Context, fn func(s Storer) error) error
	Create(ctx context.Context, usr User) error
	Update(ctx context.Context, usr User) error
	Delete(ctx context.Context, usr User) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]User, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (User, error)
	QueryByIDs(ctx context.Context, userIDs []uuid.UUID) ([]User, error)
	QueryByEmail(ctx context.Context, email mail.Address) (User, error)
}

// ===============================================================

// Core manages the set of APIs for user access.
type Core struct {
	storer  Storer
	evnCore *event.Core
}

// NewCore constructs a Core for user API access.
func NewCore(evnCore *event.Core, storer Storer) *Core {
	return &Core{
		storer:  storer,
		evnCore: evnCore,
	}
}

// Create inserts a new user into the database.
func (c *Core) Create(ctx context.Context, nu NewUser) (User, error) {
	return User{}, nil
}

// Update replaces a user document in the database.
func (c *Core) Update(ctx context.Context, usr User, uu UpdateUser) (User, error) {
	return User{}, nil
}

// Delete removes a user from the database
func (c *Core) Delete(ctx context.Context, usr User) error {
	return nil
}

// Query returns a list of existing users from the database.
func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]User, error) {
	return nil, nil
}

// Count returns the total number of users in the store.
func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return 0, nil
}

// QueryByID gets the specified user from the database.
func (c *Core) QueryByID(ctx context.Context, userID uuid.UUID) (User, error) {
	return User{}, nil
}

// QueryByIDs gets the specified users from the database.
func (c *Core) QueryByIDs(ctx context.Context, userIDs []uuid.UUID) ([]User, error) {
	return nil, nil
}

// QueryByEmail gets the specified user from the database by email.
func (c *Core) QueryByEmail(ctx context.Context, email mail.Address) (User, error) {
	return User{}, nil
}

// ===============================================================

func (c *Core) Authenticate(ctx context.Context, email mail.Address, password string) (User, error) {
	usr, err := c.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return User{}, fmt.Errorf("comparehashandpassword: %w", ErrAuthenticationFailure)
	}

	return usr, nil
}
