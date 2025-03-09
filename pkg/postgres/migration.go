package postgres

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(dbURL string) error {
	migrator, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("Migration: postgres connect error: %s", err)
	}

	err = migrator.Up()
	defer migrator.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: up error: %w", err)
	}

	return nil
}
