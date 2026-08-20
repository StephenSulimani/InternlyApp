package ats

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestGreenhouse_Scrape(t *testing.T) {
	payload := greenhouseJobsResponse{
		Jobs: []greenhouseJob{
			{
				ID:          1,
				Title:       "Software Engineer Intern",
				AbsoluteURL: "https://boards.greenhouse.io/acme/jobs/1",
				UpdatedAt:   "2024-06-01T12:00:00.000Z",
				Content:     "<p>Build intern things.</p>",
				Location:    greenhouseLocation{Name: "New York, NY; Remote"},
			},
			{
				ID:          2,
				Title:       "Staff Software Engineer",
				AbsoluteURL: "https://boards.greenhouse.io/acme/jobs/2",
				UpdatedAt:   "2024-06-02T12:00:00.000Z",
				Location:    greenhouseLocation{Name: "San Francisco, CA"},
			},
			{
				ID:          3,
				Title:       "New Grad Software Engineer",
				AbsoluteURL: "https://boards.greenhouse.io/acme/jobs/3",
				UpdatedAt:   "2024-06-03T00:00:00Z",
				Offices: []greenhouseOffice{
					{Name: "NYC", Location: "New York, NY"},
					{Name: "SF", Location: "San Francisco, CA"},
				},
			},
			{
				ID:    4,
				Title: "Intern missing apply url",
			},
		},
	}

	t.Run("maps every posting with a title and apply url", func(t *testing.T) {
		var gotPath, gotQuery, gotUA, gotAccept string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotQuery = r.URL.RawQuery
			gotUA = r.Header.Get("User-Agent")
			gotAccept = r.Header.Get("Accept")
			if r.Method != http.MethodGet {
				t.Errorf("method = %s, want GET", r.Method)
			}
			_ = json.NewEncoder(w).Encode(payload)
		}))
		t.Cleanup(srv.Close)

		jobs, err := testGreenhouse(srv.URL).Scrape(context.Background(), Board{
			Name:    NameGreenhouse,
			URL:     "https://boards.greenhouse.io/acme",
			Company: "Acme",
		})
		if err != nil {
			t.Fatal(err)
		}
		if gotPath != "/v1/boards/acme/jobs" {
			t.Fatalf("path = %q, want /v1/boards/acme/jobs", gotPath)
		}
		if gotQuery != "content=true" {
			t.Fatalf("query = %q, want content=true", gotQuery)
		}
		if gotUA != defaultUserAgent {
			t.Fatalf("user-agent = %q", gotUA)
		}
		if gotAccept != "application/json" {
			t.Fatalf("accept = %q", gotAccept)
		}
		if len(jobs) != 3 {
			t.Fatalf("jobs = %d, want 3 (intern, staff, new grad)", len(jobs))
		}

		intern := jobs[0]
		if intern.SourceName != NameGreenhouse {
			t.Fatalf("source = %q", intern.SourceName)
		}
		if intern.SourceUrl != "https://boards.greenhouse.io/acme" {
			t.Fatalf("source url = %q", intern.SourceUrl)
		}
		if intern.ApplicationLink != "https://boards.greenhouse.io/acme/jobs/1" {
			t.Fatalf("application link = %q", intern.ApplicationLink)
		}
		if intern.Company == nil || *intern.Company != "Acme" {
			t.Fatalf("company = %v", intern.Company)
		}
		if intern.RoleTitle == nil || *intern.RoleTitle != "Software Engineer Intern" {
			t.Fatalf("role = %v", intern.RoleTitle)
		}
		if intern.JobType == nil || *intern.JobType != "Internship" {
			t.Fatalf("job type = %v", intern.JobType)
		}
		if intern.IsAts == nil || !*intern.IsAts {
			t.Fatal("expected is_ats true")
		}
		if intern.Description == nil || *intern.Description != "Build intern things." {
			t.Fatalf("description = %v", intern.Description)
		}
		if len(intern.Locations) != 2 || intern.Locations[0] != "New York, NY" || intern.Locations[1] != "Remote" {
			t.Fatalf("locations = %v", intern.Locations)
		}
		wantSeen := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
		if !intern.FirstSeen.Valid || !intern.FirstSeen.Time.Equal(wantSeen) {
			t.Fatalf("first seen = %+v, want %v", intern.FirstSeen, wantSeen)
		}

		staff := jobs[1]
		if staff.JobType != nil {
			t.Fatalf("staff job type = %v, want nil", staff.JobType)
		}
		if staff.ApplicationLink != "https://boards.greenhouse.io/acme/jobs/2" {
			t.Fatalf("staff application link = %q", staff.ApplicationLink)
		}

		ng := jobs[2]
		if ng.JobType == nil || *ng.JobType != "Full Time" {
			t.Fatalf("new grad job type = %v", ng.JobType)
		}
		if len(ng.Locations) != 2 || ng.Locations[0] != "New York, NY" {
			t.Fatalf("new grad locations = %v", ng.Locations)
		}
	})

	t.Run("treats 404 as board gone", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
		t.Cleanup(srv.Close)

		_, err := testGreenhouse(srv.URL).Scrape(context.Background(), Board{
			URL: "https://boards.greenhouse.io/gone",
		})
		if !errors.Is(err, ErrBoardGone) {
			t.Fatalf("err = %v, want ErrBoardGone", err)
		}
	})

	t.Run("returns error on non-404 failure", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", http.StatusBadGateway)
		}))
		t.Cleanup(srv.Close)

		_, err := testGreenhouse(srv.URL).Scrape(context.Background(), Board{
			URL: "https://boards.greenhouse.io/acme",
		})
		if err == nil || errors.Is(err, ErrBoardGone) {
			t.Fatalf("err = %v, want a retryable error", err)
		}
	})

	t.Run("returns error on invalid json", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{`)
		}))
		t.Cleanup(srv.Close)

		_, err := testGreenhouse(srv.URL).Scrape(context.Background(), Board{
			URL: "https://boards.greenhouse.io/acme",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects invalid board url", func(t *testing.T) {
		_, err := NewGreenhouse(nil).Scrape(context.Background(), Board{URL: "not-a-url"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGreenhouseToken(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://boards.greenhouse.io/stripe", "stripe"},
		{"https://job-boards.greenhouse.io/andurilindustries", "andurilindustries"},
		{"https://boards.greenhouse.io/figma/", "figma"},
	}
	for _, tc := range cases {
		got, err := greenhouseToken(tc.in)
		if err != nil {
			t.Fatalf("token(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("token(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBoardSource_Scrape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(greenhouseJobsResponse{
			Jobs: []greenhouseJob{{
				Title:       "Data Intern",
				AbsoluteURL: "https://boards.greenhouse.io/acme/jobs/9",
				Location:    greenhouseLocation{Name: "Remote"},
			}},
		})
	}))
	t.Cleanup(srv.Close)

	source := &BoardSource{
		Ctx:      context.Background(),
		Provider: testGreenhouse(srv.URL),
		Board:    Board{URL: "https://boards.greenhouse.io/acme", Company: "Acme"},
	}
	jobs, err := source.Scrape(zap.NewNop().Sugar())
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("jobs = %d, want 1", len(jobs))
	}
}

func TestDefaultProviders_includesGreenhouse(t *testing.T) {
	providers := DefaultProviders(nil)
	p, ok := providers[NameGreenhouse]
	if !ok {
		t.Fatal("expected greenhouse provider")
	}
	if p.Name() != NameGreenhouse {
		t.Fatalf("name = %q", p.Name())
	}
}

func testGreenhouse(apiBase string) *Greenhouse {
	return &Greenhouse{HTTP: NewClient(), APIBase: apiBase}
}
