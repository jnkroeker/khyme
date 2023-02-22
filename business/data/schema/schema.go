package schema

import (
	"context"
	_ "embed" // Calls init function
	"fmt"

	"github.com/ardanlabs/darwin"
	"github.com/jmoiron/sqlx"
)

// tell compiler at build time to read the files and place content in these variables
// binary always has the schema built in
var (
	//go:embed sql/schema.sql
	schemaDoc string

	//go:embed sql/seed.sql
	seedDoc string

	//go:embed sql/delete.sql
	deleteDoc string
)

// Migrate attempts to bring the schema
/*
 * TODO: re-incorporate database.StatusCheck when you can figure out why
 *       Readiness debug probe is failing
 */
func Migrate(ctx context.Context, db *sqlx.DB) error {
	// if err := database.StatusCheck(ctx, db); err != nil {
	// 	return fmt.Errorf("status check database: %w", err)
	// }

	driver, err := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
	if err != nil {
		return fmt.Errorf("construct Darwin driver: %w", err)
	}

	d := darwin.New(driver, darwin.ParseMigrations(schemaDoc))
	return d.Migrate()
}

// Seed runs the set of seed-data queries against db. The queries are run in a
// transaction and rolled back if any fail.
/*
 * TODO: re-incorporate database.StatusCheck when you can figure out why
 *       Readiness debug probe is failing
 */
func Seed(ctx context.Context, db *sqlx.DB) error {
	// if err := database.StatusCheck(ctx, db); err != nil {
	// 	return fmt.Errorf("status check database: %w", err)
	// }

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seedDoc); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// DeleteAll runs a set of drop-table queries against the db.
// The queries are run in a transaction and rolled back if any fail.
func DeleteAll(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(deleteDoc); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
