package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stephensulimani/internlyapp/internal/db"
)

// SetupPostgres connects to Postgres for database-backed tests.
// DSN resolution order:
//  1. TEST_DATABASE_URL in the environment
//  2. TEST_DATABASE_URL in .env at the repository root
//
// Skips the test when TEST_DATABASE_URL is unset so plain `go test` never
// touches the development database via POSTGRES_* credentials.
func SetupPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dsn := resolveTestDSN(t)
	if dsn == "" {
		t.Skip("no test database configured; set TEST_DATABASE_URL in .env or the environment")
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

// WithTestTx runs fn inside a database transaction that is always rolled back.
// Use this for any test that inserts or updates rows so nothing persists.
func WithTestTx(t *testing.T, pool *pgxpool.Pool, fn func(ctx context.Context, tx pgx.Tx, queries *db.Queries)) {
	t.Helper()

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}

	t.Cleanup(func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			t.Errorf("rollback transaction: %v", err)
		}
	})

	fn(ctx, tx, db.New(tx))
}

// TruncateUsersTx clears users inside the current test transaction.
// Committed rows reappear after rollback, so dev data is never permanently removed.
func TruncateUsersTx(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()

	if _, err := tx.Exec(ctx, "TRUNCATE users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate users in test transaction: %v", err)
	}
}

func resolveTestDSN(t *testing.T) string {
	t.Helper()

	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}

	root := RepoRoot(t)
	_ = godotenv.Load(filepath.Join(root, ".env"))

	return os.Getenv("TEST_DATABASE_URL")
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
