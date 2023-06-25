// Package userdb contains user related CRUD functionality
package userdb

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jnkroeker/khyme/business/core/user"
	"github.com/jnkroeker/khyme/business/data/order"
	"github.com/jnkroeker/khyme/business/sys/database"
	"go.uber.org/zap"
)

// Store manages the set of APIs for user database access.
type Store struct {
	log    *zap.SugaredLogger
	db     sqlx.ExtContext
	inTran bool
}

// NewStore constructs the API for data access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s *Store) WithinTran(ctx context.Context, fn func(s user.Storer) error) error {
	if s.inTran {
		return fn(s)
	}

	f := func(tx *sqlx.Tx) error {
		s := &Store{
			log:    s.log,
			db:     tx,
			inTran: true,
		}
		return fn(s)
	}

	return database.WithinTran(ctx, s.log, s.db.(*sqlx.DB), f)
}

// Create inserts a new user into the database.
func (s *Store) Create(ctx context.Context, usr user.User) error {
	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, enabled, department, date_created, date_updated)
	VALUES
		(:user_id, :name, :email, :password_hash, :roles, :enabled, :department, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, database.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", err)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s *Store) Update(ctx context.Context, usr user.User) error {
	return nil
}

// Delete removes a user from the database.
func (s *Store) Delete(ctx context.Context, usr user.User) error {
	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(ctx context.Context, filter user.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]user.User, error) {
	return []user.User{}, nil
}

// Count returns the total number of users in the database.
func (s *Store) Count(ctx context.Context, filter user.QueryFilter) (int, error) {
	return 1, nil
}

// QueryByID gets the specified user from the database.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	return user.User{}, nil
}

// QueryByIDs gets the specified users from the database.
func (s *Store) QueryByIDs(ctx context.Context, userIDs []uuid.UUID) ([]user.User, error) {
	return []user.User{}, nil
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	return user.User{}, nil
}
