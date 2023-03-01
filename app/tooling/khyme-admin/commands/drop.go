package commands

import (
	"fmt"

	"github.com/jnkroeker/khyme/business/data/schema"
	"github.com/jnkroeker/khyme/business/sys/database"
)

func DropTables(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	if err := schema.DeleteAll(db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("drop tables complete")

	return nil
}
