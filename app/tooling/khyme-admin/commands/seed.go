package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/jnkroeker/khyme/business/data/schema"
	"github.com/jnkroeker/khyme/business/sys/database"
)

func Seed(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	fmt.Println("seed data complete")

	return nil
}
