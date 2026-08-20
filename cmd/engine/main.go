package main

import (
	"context"
	"errors"
	"time"

	"github.com/joho/godotenv"
	"github.com/stephensulimani/internlyapp/cmd/engine/modules"
	"github.com/stephensulimani/internlyapp/internal/ats"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
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

	queries := db.New(pool)
	ingest := service.NewJobIngestService(queries)
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

	ingestWorkingATSBoards(ctx, log, queries, ingest)
}

func ingestWorkingATSBoards(ctx context.Context, log *zap.SugaredLogger, queries *db.Queries, ingest *service.JobIngestService) {
	boards, err := queries.ListWorkingCompanyATS(ctx)
	if err != nil {
		log.Errorw("list working ats boards failed", "error", err)
		return
	}

	providers := ats.DefaultProviders(ats.NewClient())
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	for _, row := range boards {
		provider, ok := providers[row.AtsName]
		if !ok {
			continue
		}
		if err := limiter.Wait(ctx); err != nil {
			log.Errorw("ats rate limit wait failed", "error", err)
			return
		}

		source := &ats.BoardSource{
			Ctx:      ctx,
			Provider: provider,
			Board: ats.Board{
				Name:    row.AtsName,
				URL:     row.AtsUrl,
				Company: row.CompanyName,
			},
		}
		result, err := ingest.IngestBoard(ctx, log, source)
		if errors.Is(err, ats.ErrBoardGone) {
			if markErr := queries.SetCompanyATSWorking(ctx, db.SetCompanyATSWorkingParams{
				ID:      row.ID,
				Working: false,
			}); markErr != nil {
				log.Errorw("mark ats board not working failed", "ats_url", row.AtsUrl, "error", markErr)
			} else {
				log.Warnw("marked ats board not working", "ats", row.AtsName, "ats_url", row.AtsUrl)
			}
			continue
		}
		if err != nil {
			log.Errorw("ats ingest failed", "ats", row.AtsName, "ats_url", row.AtsUrl, "error", err)
			continue
		}
		log.Infow("ats ingest complete",
			"ats", row.AtsName,
			"company", row.CompanyName,
			"ats_url", row.AtsUrl,
			"scraped", result.Scraped,
			"inserted", result.Inserted,
			"enriched", result.Enriched,
			"skipped", result.Skipped,
			"skipped_duplicates", result.SkippedDuplicates,
			"failed", result.Failed,
		)
	}
}
