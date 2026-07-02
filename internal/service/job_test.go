package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stephensulimani/internlyapp/internal/db"
)

func TestJobService_List(t *testing.T) {
	t.Run("uses default limit", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobService(store)

		_, err := svc.List(context.Background(), 0)
		if err != nil {
			t.Fatal(err)
		}
		if store.limit != DefaultJobsLimit {
			t.Fatalf("limit = %d, want %d", store.limit, DefaultJobsLimit)
		}
	})

	t.Run("passes requested limit", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobService(store)

		_, err := svc.List(context.Background(), 10)
		if err != nil {
			t.Fatal(err)
		}
		if store.limit != 10 {
			t.Fatalf("limit = %d, want 10", store.limit)
		}
	})

	t.Run("caps limit at max", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobService(store)

		_, err := svc.List(context.Background(), 500)
		if err != nil {
			t.Fatal(err)
		}
		if store.limit != MaxJobsLimit {
			t.Fatalf("limit = %d, want %d", store.limit, MaxJobsLimit)
		}
	})

	t.Run("returns jobs from store", func(t *testing.T) {
		company := "Acme"
		store := &mockJobStore{
			jobs: []db.Job{
				{Company: &company, ApplicationLink: "https://jobs.example.com/1"},
			},
		}
		svc := NewJobService(store)

		jobs, err := svc.List(context.Background(), 5)
		if err != nil {
			t.Fatal(err)
		}
		if len(jobs) != 1 || jobs[0].ApplicationLink != "https://jobs.example.com/1" {
			t.Fatalf("jobs = %+v", jobs)
		}
	})

	t.Run("returns empty slice when no jobs", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobService(store)

		jobs, err := svc.List(context.Background(), 5)
		if err != nil {
			t.Fatal(err)
		}
		if len(jobs) != 0 {
			t.Fatalf("jobs = %v, want empty slice", jobs)
		}
	})

	t.Run("store applies limit to results", func(t *testing.T) {
		store := &mockJobStore{
			jobs: []db.Job{
				{ApplicationLink: "https://jobs.example.com/1"},
				{ApplicationLink: "https://jobs.example.com/2"},
				{ApplicationLink: "https://jobs.example.com/3"},
			},
		}
		svc := NewJobService(store)

		jobs, err := svc.List(context.Background(), 2)
		if err != nil {
			t.Fatal(err)
		}
		if len(jobs) != 2 {
			t.Fatalf("len(jobs) = %d, want 2", len(jobs))
		}
		if jobs[0].ApplicationLink != "https://jobs.example.com/1" || jobs[1].ApplicationLink != "https://jobs.example.com/2" {
			t.Fatalf("jobs = %+v", jobs)
		}
	})

	t.Run("store error", func(t *testing.T) {
		store := &mockJobStore{getJobsErr: errors.New("db down")}
		svc := NewJobService(store)

		_, err := svc.List(context.Background(), 5)
		if !errors.Is(err, ErrGetJobs) {
			t.Fatalf("err = %v, want ErrGetJobs", err)
		}
	})
}

func TestNormalizeJobsLimit(t *testing.T) {
	t.Run("defaults invalid values", func(t *testing.T) {
		limit, err := NormalizeJobsLimit(0)
		if err != nil || limit != DefaultJobsLimit {
			t.Fatalf("limit = %d, err = %v", limit, err)
		}
	})

	t.Run("rejects above max", func(t *testing.T) {
		_, err := NormalizeJobsLimit(MaxJobsLimit + 1)
		if !errors.Is(err, ErrInvalidJobsLimit) {
			t.Fatalf("err = %v, want ErrInvalidJobsLimit", err)
		}
	})
}
