package db

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string) {
	RunMigrationsFrom("file://internal/db/migrations", databaseURL)
}

func RunMigrationsFrom(source string, databaseURL string) {
	// databaseURL should look like "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
	m, err := migrate.New(
		source,
		databaseURL,
	)
	if err != nil {
		log.Fatal("Could not create migrate instance: ", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("Could not run up migrations: ", err)
	}

	log.Println("Migrations applied successfully!")
}
