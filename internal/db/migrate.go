package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string) error {
	return RunMigrationsFrom("file://internal/db/migrations", databaseURL)
}

func RunMigrationsFrom(source string, databaseURL string) error {
	// databaseURL should look like "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
	m, err := migrate.New(
		source,
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run up migrations: %w", err)
	}

	return nil
}
