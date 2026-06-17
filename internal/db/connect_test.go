package db

import "testing"

func TestConfig_DSN(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "internly",
		Password: "secret",
		DBName:   "internly",
	}

	want := "postgres://internly:secret@localhost:5432/internly?sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}
