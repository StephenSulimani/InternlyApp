package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/internal/db"
	"go.uber.org/zap"
)

type stubJobSource struct {
	jobs []db.Job
	err  error
}

func (s stubJobSource) Scrape(log *zap.SugaredLogger) ([]db.Job, error) {
	return s.jobs, s.err
}

func TestJobIngestService_Ingest(t *testing.T) {
	log := zap.NewNop().Sugar()
	job := db.Job{ApplicationLink: "https://jobs.example.com/1", Company: strPtr("Acme")}

	t.Run("inserts scraped jobs", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobIngestService(store)

		result, err := svc.Ingest(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Scraped != 1 || result.Inserted != 1 {
			t.Fatalf("result = %+v, want scraped=1 inserted=1", result)
		}
		if len(store.createCalls) != 1 {
			t.Fatalf("create calls = %d, want 1", len(store.createCalls))
		}
	})

	t.Run("skips duplicate jobs", func(t *testing.T) {
		attempts := 0
		store := &mockJobStore{
			createJob: func(ctx context.Context, arg db.CreateJobParams) (db.Job, error) {
				attempts++
				if attempts == 1 {
					return db.Job{}, nil
				}
				return db.Job{}, &pgconn.PgError{Code: "23505"}
			},
		}
		svc := NewJobIngestService(store)

		result, err := svc.Ingest(context.Background(), log, stubJobSource{jobs: []db.Job{job, job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Inserted != 1 || result.SkippedDuplicates != 1 {
			t.Fatalf("result = %+v, want inserted=1 skipped=1", result)
		}
	})

	t.Run("counts other insert failures", func(t *testing.T) {
		store := &mockJobStore{createErr: errors.New("db down")}
		svc := NewJobIngestService(store)

		result, err := svc.Ingest(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Failed != 1 || result.Inserted != 0 {
			t.Fatalf("result = %+v, want failed=1", result)
		}
	})

	t.Run("returns scrape error", func(t *testing.T) {
		svc := NewJobIngestService(&mockJobStore{})

		_, err := svc.Ingest(context.Background(), log, stubJobSource{err: errors.New("network down")})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("discovers ats board from application link", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobIngestService(store)
		company := "Stripe"
		job := db.Job{
			ApplicationLink: "https://boards.greenhouse.io/stripe/jobs/12345",
			Company:         &company,
		}

		result, err := svc.Ingest(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.ATSDiscovered != 1 {
			t.Fatalf("ATSDiscovered = %d, want 1", result.ATSDiscovered)
		}
		if len(store.atsCalls) != 1 {
			t.Fatalf("atsCalls = %d, want 1", len(store.atsCalls))
		}
		got := store.atsCalls[0]
		if got.AtsName != "greenhouse" || got.AtsUrl != "https://boards.greenhouse.io/stripe" || got.CompanyName != "Stripe" {
			t.Fatalf("ats upsert = %+v", got)
		}
	})

	t.Run("discovers ats even when job is a duplicate", func(t *testing.T) {
		store := &mockJobStore{
			createErr: &pgconn.PgError{Code: "23505"},
		}
		svc := NewJobIngestService(store)
		job := db.Job{ApplicationLink: "https://jobs.lever.co/openai/abcd"}

		result, err := svc.Ingest(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.SkippedDuplicates != 1 || result.ATSDiscovered != 1 {
			t.Fatalf("result = %+v", result)
		}
	})

	t.Run("skips non-ats links", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobIngestService(store)
		job := db.Job{ApplicationLink: "https://www.linkedin.com/jobs/view/1"}

		result, err := svc.Ingest(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.ATSDiscovered != 0 || len(store.atsCalls) != 0 {
			t.Fatalf("unexpected ats discovery: %+v calls=%d", result, len(store.atsCalls))
		}
	})
}

func strPtr(s string) *string {
	return &s
}
