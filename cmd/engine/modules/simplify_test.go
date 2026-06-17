package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestSimplify_Scrape(t *testing.T) {
	fixture := []map[string]any{
		{
			"locations":    []any{"New York, NY", "Remote"},
			"company_name": "Acme Corp",
			"title":        "Software Engineer Intern",
			"date_updated": float64(1_700_000_000),
			"url":          "https://jobs.example.com/123",
		},
	}

	t.Run("parses listings from api", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("method = %s, want GET", r.Method)
			}
			_ = json.NewEncoder(w).Encode(fixture)
		}))
		t.Cleanup(srv.Close)

		scraper := &Simplify{URL: srv.URL, JobType: "Internship"}
		jobs, err := scraper.Scrape(zap.NewNop().Sugar())
		if err != nil {
			t.Fatal(err)
		}

		if len(jobs) != 1 {
			t.Fatalf("jobs = %d, want 1", len(jobs))
		}

		job := jobs[0]
		if job.SourceName != "Simplify" {
			t.Fatalf("source = %q", job.SourceName)
		}
		if job.SourceUrl != srv.URL {
			t.Fatalf("source url = %q", job.SourceUrl)
		}
		if job.ApplicationLink != "https://jobs.example.com/123" {
			t.Fatalf("application link = %q", job.ApplicationLink)
		}
		if deref(job.Company) != "Acme Corp" {
			t.Fatalf("company = %q", deref(job.Company))
		}
		if deref(job.RoleTitle) != "Software Engineer Intern" {
			t.Fatalf("role = %q", deref(job.RoleTitle))
		}
		if deref(job.JobType) != "Internship" {
			t.Fatalf("job type = %q", deref(job.JobType))
		}
		if len(job.Locations) != 2 || job.Locations[0] != "New York, NY" {
			t.Fatalf("locations = %v", job.Locations)
		}
		wantSeen := time.Unix(1_700_000_000, 0)
		if !job.FirstSeen.Valid || !job.FirstSeen.Time.Equal(wantSeen) {
			t.Fatalf("first seen = %+v, want %v", job.FirstSeen, wantSeen)
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "upstream error", http.StatusBadGateway)
		}))
		t.Cleanup(srv.Close)

		scraper := &Simplify{URL: srv.URL, JobType: "Internship"}
		_, err := scraper.Scrape(zap.NewNop().Sugar())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("returns error on invalid json", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `{`)
		}))
		t.Cleanup(srv.Close)

		scraper := &Simplify{URL: srv.URL, JobType: "Internship"}
		_, err := scraper.Scrape(zap.NewNop().Sugar())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("returns error on unreachable host", func(t *testing.T) {
		scraper := &Simplify{URL: "http://127.0.0.1:1", JobType: "Internship"}
		_, err := scraper.Scrape(zap.NewNop().Sugar())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestParseSimplifyListings(t *testing.T) {
	t.Run("parses valid listings", func(t *testing.T) {
		raw := []map[string]any{
			{
				"locations":    []any{"Boston, MA"},
				"company_name": "Globex",
				"title":        "Backend Intern",
				"date_updated": float64(1_600_000_000),
				"url":          "https://jobs.example.com/456",
			},
		}

		jobs, skipped, err := parseSimplifyListings(raw, "https://source.example/listings.json", "Internship")
		if err != nil {
			t.Fatal(err)
		}
		if skipped != 0 {
			t.Fatalf("skipped = %d, want 0", skipped)
		}
		if len(jobs) != 1 {
			t.Fatalf("jobs = %d, want 1", len(jobs))
		}
	})

	t.Run("skips malformed listings", func(t *testing.T) {
		raw := []map[string]any{
			{
				"locations":    "not-a-slice",
				"company_name": "Bad Co",
				"title":        "Role",
				"date_updated": float64(1_600_000_000),
				"url":          "https://jobs.example.com/bad",
			},
			{
				"locations":    []any{"Seattle, WA"},
				"company_name": "Good Co",
				"title":        "SWE Intern",
				"date_updated": float64(1_600_000_001),
				"url":          "https://jobs.example.com/good",
			},
		}

		jobs, skipped, err := parseSimplifyListings(raw, "https://source.example/listings.json", "Internship")
		if err != nil {
			t.Fatal(err)
		}
		if skipped != 1 {
			t.Fatalf("skipped = %d, want 1", skipped)
		}
		if len(jobs) != 1 {
			t.Fatalf("jobs = %d, want 1", len(jobs))
		}
		if deref(jobs[0].Company) != "Good Co" {
			t.Fatalf("company = %q", deref(jobs[0].Company))
		}
	})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func TestSimplify_implementsScraper(t *testing.T) {
	var _ Scraper = (*Simplify)(nil)
}

func TestSimplify_Scrape_skippedListingsStillSucceed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"company_name": "missing fields"},
			{
				"locations":    []any{"Remote"},
				"company_name": "Good Co",
				"title":        "Intern",
				"date_updated": float64(1_600_000_000),
				"url":          "https://jobs.example.com/good",
			},
		})
	}))
	t.Cleanup(srv.Close)

	jobs, err := (&Simplify{URL: srv.URL, JobType: "Internship"}).Scrape(zap.NewNop().Sugar())
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("jobs = %d, want 1", len(jobs))
	}
}
