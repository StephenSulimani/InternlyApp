package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/internal/ats"
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
	ATSDiscovered     int
}

type JobIngestService struct {
	store JobIngestStore
}

func NewJobIngestService(store JobIngestStore) *JobIngestService {
	return &JobIngestService{store: store}
}

func (s *JobIngestService) Ingest(ctx context.Context, log *zap.SugaredLogger, source JobSource) (IngestResult, error) {
	jobs, err := source.Scrape(log)
	if err != nil {
		return IngestResult{}, fmt.Errorf("scrape: %w", err)
	}

	seenBoards := make(map[string]struct{})
	result := IngestResult{Scraped: len(jobs)}
	for _, job := range jobs {
		board, isATS := ats.Discover(job.ApplicationLink)
		job.IsAts = boolPtr(isATS)

		_, err := s.store.CreateJob(ctx, db.ToCreateParams(job))
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				result.SkippedDuplicates++
			} else {
				log.Errorw("create job failed", "application_link", job.ApplicationLink, "error", err)
				result.Failed++
				continue
			}
		} else {
			result.Inserted++
		}

		if isATS && s.recordATS(ctx, log, job, board, seenBoards) {
			result.ATSDiscovered++
		}
	}

	return result, nil
}

func (s *JobIngestService) recordATS(ctx context.Context, log *zap.SugaredLogger, job db.Job, board ats.Board, seen map[string]struct{}) bool {
	if _, exists := seen[board.URL]; exists {
		return false
	}

	company := ""
	if job.Company != nil {
		company = *job.Company
	}
	if company == "" {
		company = board.URL
	}

	_, err := s.store.UpsertCompanyATS(ctx, db.UpsertCompanyATSParams{
		CompanyName: company,
		AtsName:     board.Name,
		AtsUrl:      board.URL,
	})
	if err != nil {
		log.Errorw("upsert company ats failed", "ats_url", board.URL, "error", err)
		return false
	}

	seen[board.URL] = struct{}{}
	return true
}

func boolPtr(b bool) *bool {
	return &b
}
