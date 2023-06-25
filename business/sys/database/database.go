// Package database provides support for database access
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/jnkroeker/khyme/foundation/web"
	_ "github.com/lib/pq" // Calls this database driver's init function
	"go.uber.org/zap"
)

const (
	uniqueViolation = "23505"
	undefinedTable  = "42P01"
)

// Set of error variables
var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrDBDuplicatedEntry     = errors.New("duplicated entry")
	ErrForbidden             = errors.New("attempted action is not allowed")
)

// Config is the required properties to use the database
type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// Open knows how to open a database connection based on the Config
func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

// StatusCheck returns nil if it can successfully talk to the database.
// It returns a non-nil response if it can't talk to the database.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {

	// First, check if we can ping the database
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// Make sure we didn't timeout or get canceled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a light, simple query to determine connectivity
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

// WithinTran runs passed function and does commit/rollback at the end.
func WithinTran(ctx context.Context, log *zap.SugaredLogger, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	log.Info(ctx, "begin tran")
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tran: %w", err)
	}

	// We can defer the rollback since the code checks if the transaction
	// has already been committed.
	defer func() {
		if err := tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				return
			}
			log.Error(ctx, "unable to rollback tran", "msg", err)
		}
		log.Info(ctx, "rollback tran")
	}()

	if err := fn(tx); err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok && pqerr.Code == uniqueViolation {
			return ErrDBDuplicatedEntry
		}
		return fmt.Errorf("exec tran: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tran: %w", err)
	}
	log.Info(ctx, "commit tran")

	return nil
}

// NamedExecContext is a helper function to execute a CRUD operation
func NamedExecContext(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data interface{}) error {
	q := queryString(query, data)
	log.Infow("database.NamedExecContext", "traceid", web.GetTraceId(ctx), "query", q)

	if _, err := sqlx.NamedExecContext(ctx, db, query, data); err != nil {
		return err
	}

	return nil
}

// NamedQuerySlice is a helper for queries that return a collection of data to be unmarshalled into a slice
func NamedQuerySlice(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	log.Infow("database.NamedQuerySlice", "traceid", web.GetTraceId(ctx), "query", q)

	// Pass the address of a slice thru the dest param
	// Use the reflection package to determine slice type (Task or User)
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.New("must provide a pointer to a slice")
	}

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		return err
	}

	// Construct an instance of the type specified by the slice
	// from each row and add it to the slice.
	slice := val.Elem()
	for rows.Next() {
		v := reflect.New(slice.Type().Elem())
		if err := rows.StructScan(v.Interface()); err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, v.Elem()))
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters
func queryString(query string, args ...interface{}) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("%q", v)
		case byte:
			value = fmt.Sprintf("%q", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}
