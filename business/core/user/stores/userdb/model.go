package userdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jnkroeker/khyme/business/core/user"
)

// dbUser represents the structure we need for moving data
// between the app and the database
type dbUser struct {
	ID           uuid.UUID      `db:"user_id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	PasswordHash []byte         `db:"password_hash"`
	Enabled      bool           `db:"enabled"`
	Department   sql.NullString `db:"department"`
	DateCreated  time.Time      `db:"date_created"`
	DateUpdated  time.Time      `db:"date_updated"`
	// Roles        dbarray.String `db:"roles"`
}

func toDBUser(usr user.User) dbUser {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return dbUser{
		ID:    usr.ID,
		Name:  usr.Name,
		Email: usr.Email.Address,
		// Roles:        roles,
		PasswordHash: usr.PasswordHash,
		Department: sql.NullString{
			String: usr.Department,
			Valid:  usr.Department != "",
		},
		Enabled:     usr.Enabled,
		DateCreated: usr.DateCreated.UTC(),
		DateUpdated: usr.DateUpdated.UTC(),
	}
}
