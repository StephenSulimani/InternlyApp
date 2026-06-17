package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stephensulimani/internlyapp/internal/db"
)

// SetupPostgres connects to Postgres for API integration tests.
// DSN resolution order:
//  1. TEST_DATABASE_URL
//  2. .env at the repository root (POSTGRES_* or TEST_DATABASE_URL)
//  3. POSTGRES_* already present in the environment
//
// Skips the test when no DSN can be resolved (e.g. CI without a database).
func SetupPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dsn := resolveTestDSN(t)
	if dsn == "" {
		t.Skip("no test database configured; set TEST_DATABASE_URL or POSTGRES_* in .env")
	}

	root := RepoRoot(t)
	migrations := fmt.Sprintf("file://%s", filepath.Join(root, "internal", "db", "migrations"))
	if err := db.RunMigrationsFrom(migrations, dsn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

func resolveTestDSN(t *testing.T) string {
	t.Helper()

	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}

	root := RepoRoot(t)
	_ = godotenv.Load(filepath.Join(root, ".env"))

	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}

	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" || dbName == "" {
		return ""
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, dbName,
	)
}

func InstallRejectUserInsertTrigger(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		CREATE OR REPLACE FUNCTION test_reject_user_insert() RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'test insert rejection';
		END;
		$$ LANGUAGE plpgsql;

		DROP TRIGGER IF EXISTS test_reject_user_insert ON users;

		CREATE TRIGGER test_reject_user_insert
			BEFORE INSERT ON users
			FOR EACH ROW
			EXECUTE FUNCTION test_reject_user_insert();
	`)
	if err != nil {
		t.Fatalf("install reject insert trigger: %v", err)
	}
}

func RemoveRejectUserInsertTrigger(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		DROP TRIGGER IF EXISTS test_reject_user_insert ON users;
		DROP FUNCTION IF EXISTS test_reject_user_insert();
	`)
	if err != nil {
		t.Fatalf("remove reject insert trigger: %v", err)
	}
}

func TruncateUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(), "TRUNCATE users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate users: %v", err)
	}
}
