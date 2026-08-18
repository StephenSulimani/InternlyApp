package main

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/stephensulimani/internlyapp/cmd/engine/modules"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

const (
	simplifyInternURL = "https://raw.githubusercontent.com/SimplifyJobs/Summer2026-Internships/refs/heads/dev/.github/scripts/listings.json"
	simplifyNGURL     = "https://raw.githubusercontent.com/SimplifyJobs/New-Grad-Positions/refs/heads/dev/.github/scripts/listings.json"
)

func main() {
	godotenv.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	log.Info("Running Migrations")
	pool, err := db.OpenPool(context.Background(), db.ConfigFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	log.Info("Connected to DB")

	log.Info("Starting Internly Scraping Engine")

	ingest := service.NewJobIngestService(db.New(pool))
	scrapers := []service.JobSource{
		&modules.Simplify{URL: simplifyInternURL, JobType: "Internship"},
		&modules.Simplify{URL: simplifyNGURL, JobType: "Full Time"},
	}

	ctx := context.Background()
	for _, scraper := range scrapers {
		result, err := ingest.Ingest(ctx, log, scraper)
		if err != nil {
			log.Fatal(err)
		}
		log.Infow("ingest complete",
			"scraped", result.Scraped,
			"inserted", result.Inserted,
			"skipped_duplicates", result.SkippedDuplicates,
			"failed", result.Failed,
			"ats_discovered", result.ATSDiscovered,
		)
	}
}
