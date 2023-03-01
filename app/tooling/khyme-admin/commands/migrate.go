package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jnkroeker/khyme/business/data/schema"
	"github.com/jnkroeker/khyme/business/sys/database"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

func Migrate(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrations complete")

	return nil
}
