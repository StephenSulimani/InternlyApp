package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephensulimani/internlyapp/internal/utils"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func ConfigFromEnv() Config {
	return Config{
		Host:     utils.GetEnv("POSTGRES_HOST", "localhost"),
		Port:     utils.GetEnv("POSTGRES_PORT", "5432"),
		User:     utils.GetEnv("POSTGRES_USER", ""),
		Password: utils.GetEnv("POSTGRES_PASSWORD", ""),
		DBName:   utils.GetEnv("POSTGRES_DB", "internly"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DBName,
	)
}

func OpenPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := cfg.DSN()
	if err := RunMigrations(dsn); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}

	return pool, nil
}
