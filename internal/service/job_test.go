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

func TestJobService_Stats(t *testing.T) {
	t.Run("returns stats from store", func(t *testing.T) {
		store := &mockJobStore{
			stats: db.GetJobsStatsRow{
				TotalJobs:      12,
				AddedThisWeek:  3,
				TotalCompanies: 8,
			},
		}
		svc := NewJobService(store)

		stats, err := svc.Stats(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if stats.TotalJobs != 12 || stats.AddedThisWeek != 3 || stats.TotalCompanies != 8 {
			t.Fatalf("stats = %+v", stats)
		}
		if store.statsCalls != 1 {
			t.Fatalf("statsCalls = %d, want 1", store.statsCalls)
		}
	})

	t.Run("store error", func(t *testing.T) {
		store := &mockJobStore{statsErr: errors.New("db down")}
		svc := NewJobService(store)

		_, err := svc.Stats(context.Background())
		if !errors.Is(err, ErrGetJobsStats) {
			t.Fatalf("err = %v, want ErrGetJobsStats", err)
		}
	})
}

func TestJobService_Search(t *testing.T) {
	t.Run("uses default limit and offset", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobService(store)

		page, err := svc.Search(context.Background(), JobListQuery{})
		if err != nil {
			t.Fatal(err)
		}
		if page.Limit != DefaultJobsLimit || page.Offset != 0 || page.Total != 0 {
			t.Fatalf("page = %+v", page)
		}
		if store.searchParams.RowLimit != DefaultJobsLimit || store.searchParams.RowOffset != 0 {
			t.Fatalf("params = %+v", store.searchParams)
		}
		if store.searchParams.SortBy != "posted" || store.searchParams.SortDir != "desc" {
			t.Fatalf("sort = %+v", store.searchParams)
		}
	})

	t.Run("passes filters and like pattern", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobService(store)

		_, err := svc.Search(context.Background(), JobListQuery{
			Q:            "intern%",
			Type:         " Internship ",
			Location:     "Remote",
			Source:       "Greenhouse",
			RecencyHours: 24,
			Limit:        10,
			Offset:       20,
		})
		if err != nil {
			t.Fatal(err)
		}

		got := store.searchParams
		if got.Q != `%intern\%%` {
			t.Fatalf("q = %q", got.Q)
		}
		if got.FilterType != "Internship" || got.FilterLocation != "%Remote%" || got.FilterSource != "Greenhouse" {
			t.Fatalf("params = %+v", got)
		}
		if got.RecencyHours != 24 || got.RowLimit != 10 || got.RowOffset != 20 {
			t.Fatalf("params = %+v", got)
		}
		if got.SortBy != "posted" || got.SortDir != "desc" {
			t.Fatalf("sort = %s %s", got.SortBy, got.SortDir)
		}
	})

	t.Run("passes sort", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobService(store)

		_, err := svc.Search(context.Background(), JobListQuery{SortBy: "company", SortDir: "asc"})
		if err != nil {
			t.Fatal(err)
		}
		if store.searchParams.SortBy != "company" || store.searchParams.SortDir != "asc" {
			t.Fatalf("params = %+v", store.searchParams)
		}
	})

	t.Run("caps limit at max", func(t *testing.T) {
		store := &mockJobStore{}
		svc := NewJobService(store)

		page, err := svc.Search(context.Background(), JobListQuery{Limit: 500})
		if err != nil {
			t.Fatal(err)
		}
		if page.Limit != MaxJobsLimit || store.searchParams.RowLimit != MaxJobsLimit {
			t.Fatalf("limit = %d params = %+v", page.Limit, store.searchParams)
		}
	})

	t.Run("returns jobs and total", func(t *testing.T) {
		total := int64(9)
		store := &mockJobStore{
			jobs: []db.Job{
				{ApplicationLink: "https://jobs.example.com/1"},
				{ApplicationLink: "https://jobs.example.com/2"},
				{ApplicationLink: "https://jobs.example.com/3"},
			},
			searchTotal: &total,
		}
		svc := NewJobService(store)

		page, err := svc.Search(context.Background(), JobListQuery{Limit: 2, Offset: 1})
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != 9 || len(page.Jobs) != 2 {
			t.Fatalf("page = %+v", page)
		}
		if page.Jobs[0].ApplicationLink != "https://jobs.example.com/2" {
			t.Fatalf("jobs = %+v", page.Jobs)
		}
	})

	t.Run("count error", func(t *testing.T) {
		store := &mockJobStore{countErr: errors.New("db down")}
		svc := NewJobService(store)

		_, err := svc.Search(context.Background(), JobListQuery{})
		if !errors.Is(err, ErrGetJobs) {
			t.Fatalf("err = %v, want ErrGetJobs", err)
		}
	})

	t.Run("search error", func(t *testing.T) {
		store := &mockJobStore{searchErr: errors.New("db down")}
		svc := NewJobService(store)

		_, err := svc.Search(context.Background(), JobListQuery{})
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

func TestNormalizeJobsOffset(t *testing.T) {
	t.Run("zero is valid", func(t *testing.T) {
		offset, err := NormalizeJobsOffset(0)
		if err != nil || offset != 0 {
			t.Fatalf("offset = %d err = %v", offset, err)
		}
	})

	t.Run("rejects negative", func(t *testing.T) {
		_, err := NormalizeJobsOffset(-1)
		if !errors.Is(err, ErrInvalidJobsOffset) {
			t.Fatalf("err = %v, want ErrInvalidJobsOffset", err)
		}
	})
}

func TestParseRecency(t *testing.T) {
	tests := []struct {
		in      string
		hours   int32
		wantErr error
	}{
		{in: "", hours: 0},
		{in: "any", hours: 0},
		{in: "24h", hours: 24},
		{in: "3d", hours: 72},
		{in: "7d", hours: 168},
		{in: " month ", wantErr: ErrInvalidJobsRecency},
	}

	for _, tt := range tests {
		hours, err := ParseRecency(tt.in)
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseRecency(%q) err = %v, want %v", tt.in, err, tt.wantErr)
			}
			continue
		}
		if err != nil || hours != tt.hours {
			t.Fatalf("ParseRecency(%q) = %d, %v want %d", tt.in, hours, err, tt.hours)
		}
	}
}

func TestParseSort(t *testing.T) {
	field, dir, err := ParseSort("", "")
	if err != nil || field != "posted" || dir != "desc" {
		t.Fatalf("default sort = %s %s %v", field, dir, err)
	}

	field, dir, err = ParseSort("company", "")
	if err != nil || field != "company" || dir != "asc" {
		t.Fatalf("company sort = %s %s %v", field, dir, err)
	}

	_, _, err = ParseSort("salary", "asc")
	if !errors.Is(err, ErrInvalidJobsSort) {
		t.Fatalf("err = %v, want ErrInvalidJobsSort", err)
	}
}

func TestJobService_Locations(t *testing.T) {
	store := &mockJobStore{locations: []string{"Remote", "New York, NY"}}
	svc := NewJobService(store)

	got, err := svc.Locations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "Remote" {
		t.Fatalf("locations = %+v", got)
	}
}
