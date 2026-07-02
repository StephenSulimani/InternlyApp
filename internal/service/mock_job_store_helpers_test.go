package service

import (
	"context"
	"testing"

	"github.com/stephensulimani/internlyapp/internal/db"
)

func TestSliceJobs(t *testing.T) {
	jobs := []db.Job{
		{ApplicationLink: "1"},
		{ApplicationLink: "2"},
		{ApplicationLink: "3"},
	}

	t.Run("empty input", func(t *testing.T) {
		got := sliceJobs(nil, 5)
		if len(got) != 0 {
			t.Fatalf("len = %d, want 0", len(got))
		}
	})

	t.Run("limit below length", func(t *testing.T) {
		got := sliceJobs(jobs, 2)
		if len(got) != 2 || got[1].ApplicationLink != "2" {
			t.Fatalf("got = %+v", got)
		}
	})

	t.Run("limit above length", func(t *testing.T) {
		got := sliceJobs(jobs, 10)
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3", len(got))
		}
	})
}

func TestMockJobStore_GetJobsLimit(t *testing.T) {
	store := &mockJobStore{
		jobs: []db.Job{{ApplicationLink: "https://jobs.example.com/1"}},
	}

	jobs, err := store.GetJobsLimit(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("jobs = %+v", jobs)
	}
	if store.getJobsCalls != 1 || store.limit != 10 {
		t.Fatalf("calls = %d limit = %d", store.getJobsCalls, store.limit)
	}
}
