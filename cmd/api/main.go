package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stephensulimani/internlyapp/cmd/api/server"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/utils"
	"go.uber.org/zap"
)

func main() {
	godotenv.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	api_host := utils.GetEnv("API_HOST", "localhost")
	api_port := utils.GetEnv("API_PORT", "8080")

	db := attachDatabase(log)

	defer db.Close()

	handler := server.NewHandler(log, db)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", api_host, api_port), handler))
}

func attachDatabase(log *zap.SugaredLogger) *pgxpool.Pool {
	pg_db := utils.GetEnv("POSTGRES_DB", "internly")
	pg_user := utils.GetEnv("POSTGRES_USER", "")
	pg_pass := utils.GetEnv("POSTGRES_PASSWORD", "")
	pg_host := utils.GetEnv("POSTGRES_HOST", "localhost")
	pg_port := utils.GetEnv("POSTGRES_PORT", "5432")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pg_user, pg_pass, pg_host, pg_port, pg_db)

	log.Info("Running Migrations")
	db.RunMigrations(dsn)
	log.Info("Connected to DB")

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}

	return pool
}
