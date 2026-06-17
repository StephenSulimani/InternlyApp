package main

import (
	"context"
	"fmt"
	"net/http"

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

	apiHost := utils.GetEnv("API_HOST", "localhost")
	apiPort := utils.GetEnv("API_PORT", "8080")

	log.Info("Running Migrations")
	pool, err := db.OpenPool(context.Background(), db.ConfigFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	log.Info("Connected to DB")

	handler := server.NewHandler(log, pool)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", apiHost, apiPort), handler))
}
