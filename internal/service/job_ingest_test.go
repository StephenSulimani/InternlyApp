package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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
		if len(store.createCalls) != 1 || !isTrue(store.createCalls[0].IsAts) {
			t.Fatalf("createCalls = %+v, want IsAts true", store.createCalls)
		}
		if len(store.atsCalls) != 1 {
			t.Fatalf("atsCalls = %d, want 1", len(store.atsCalls))
		}
		got := store.atsCalls[0]
		if got.AtsName != "greenhouse" || got.AtsUrl != "https://boards.greenhouse.io/stripe" || got.CompanyName != "Stripe" {
			t.Fatalf("ats upsert = %+v", got)
		}
	})

	t.Run("sets is_ats false for non-ats links", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobIngestService(store)
		job := db.Job{ApplicationLink: "https://www.linkedin.com/jobs/view/1"}

		_, err := svc.Ingest(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if len(store.createCalls) != 1 || isTrue(store.createCalls[0].IsAts) {
			t.Fatalf("createCalls = %+v, want IsAts false", store.createCalls)
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

func TestJobIngestService_IngestBoard(t *testing.T) {
	log := zap.NewNop().Sugar()

	t.Run("enriches existing simplify job by greenhouse id", func(t *testing.T) {
		var existingID pgtype.UUID
		if err := existingID.Scan("22222222-2222-2222-2222-222222222222"); err != nil {
			t.Fatal(err)
		}
		store := &mockJobStore{
			jobs: []db.Job{{
				ID:              existingID,
				ApplicationLink: "https://job-boards.greenhouse.io/acme/jobs/1",
				SourceName:      "Simplify",
			}},
		}
		svc := NewJobIngestService(store)
		job := db.Job{
			ApplicationLink: "https://boards.greenhouse.io/acme/jobs/1",
			Company:         strPtr("Acme"),
			RoleTitle:       strPtr("Software Engineer Intern"),
			Description:     strPtr("Ship intern projects."),
		}

		result, err := svc.IngestBoard(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Enriched != 1 || result.Inserted != 0 {
			t.Fatalf("result = %+v, want enriched=1 inserted=0", result)
		}
		if len(store.createCalls) != 0 {
			t.Fatalf("create calls = %d, want 0", len(store.createCalls))
		}
		if len(store.updateDescCalls) != 1 || store.updateDescCalls[0].Description == nil || *store.updateDescCalls[0].Description != "Ship intern projects." {
			t.Fatalf("update calls = %+v", store.updateDescCalls)
		}
	})

	t.Run("inserts early-career jobs simplify missed", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobIngestService(store)
		job := db.Job{
			ApplicationLink: "https://boards.greenhouse.io/acme/jobs/9",
			Company:         strPtr("Acme"),
			RoleTitle:       strPtr("Data Intern"),
			JobType:         strPtr("Internship"),
			Description:     strPtr("Analyze data."),
		}

		result, err := svc.IngestBoard(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Inserted != 1 || result.Enriched != 0 {
			t.Fatalf("result = %+v, want inserted=1", result)
		}
		if len(store.createCalls) != 1 {
			t.Fatalf("create calls = %d, want 1", len(store.createCalls))
		}
		if store.createCalls[0].Description == nil || *store.createCalls[0].Description != "Analyze data." {
			t.Fatalf("create params = %+v", store.createCalls[0])
		}
	})

	t.Run("skips senior roles that are not already in the db", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobIngestService(store)
		job := db.Job{
			ApplicationLink: "https://boards.greenhouse.io/acme/jobs/2",
			Company:         strPtr("Acme"),
			RoleTitle:       strPtr("Staff Software Engineer"),
			Description:     strPtr("Lead the platform."),
		}

		result, err := svc.IngestBoard(context.Background(), log, stubJobSource{jobs: []db.Job{job}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Skipped != 1 || result.Inserted != 0 || result.Enriched != 0 {
			t.Fatalf("result = %+v, want skipped=1", result)
		}
		if len(store.createCalls) != 0 {
			t.Fatalf("create calls = %d, want 0", len(store.createCalls))
		}
	})

	t.Run("returns scrape error", func(t *testing.T) {
		svc := NewJobIngestService(&mockJobStore{})
		_, err := svc.IngestBoard(context.Background(), log, stubJobSource{err: errors.New("network down")})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func isTrue(v *bool) bool {
	return v != nil && *v
}
