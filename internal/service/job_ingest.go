package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/internal/db"
	"go.uber.org/zap"
)

type JobSource interface {
	Scrape(log *zap.SugaredLogger) ([]db.Job, error)
}

type IngestResult struct {
	Scraped           int
	Inserted          int
	SkippedDuplicates int
	Failed            int
}

type JobIngestService struct {
	store JobStore
}

func NewJobIngestService(store JobStore) *JobIngestService {
	return &JobIngestService{store: store}
}

func (s *JobIngestService) Ingest(ctx context.Context, log *zap.SugaredLogger, source JobSource) (IngestResult, error) {
	jobs, err := source.Scrape(log)
	if err != nil {
		return IngestResult{}, fmt.Errorf("scrape: %w", err)
	}

	result := IngestResult{Scraped: len(jobs)}
	for _, job := range jobs {
		_, err := s.store.CreateJob(ctx, db.ToCreateParams(job))
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				result.SkippedDuplicates++
				continue
			}
			log.Errorw("create job failed", "application_link", job.ApplicationLink, "error", err)
			result.Failed++
			continue
		}
		result.Inserted++
	}

	return result, nil
}
