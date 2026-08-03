package types

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/internal/db"
)

func TestJobListingFrom(t *testing.T) {
	company := "Acme"
	role := "Engineer Intern"
	jobType := "Internship"
	var id pgtype.UUID
	if err := id.Scan("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Fatal(err)
	}

	firstSeen := pgtype.Timestamptz{}
	if err := firstSeen.Scan(time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	listing := JobListingFrom(db.Job{
		ID:              id,
		Company:         &company,
		RoleTitle:       &role,
		Locations:       []string{"New York, NY"},
		JobType:         &jobType,
		ApplicationLink: "https://jobs.example.com/1",
		FirstSeen:       firstSeen,
		SourceName:      "simplify",
	})

	if listing.ID != id.String() || listing.Company != "Acme" || listing.RoleTitle != "Engineer Intern" {
		t.Fatalf("listing = %+v", listing)
	}
	if listing.FirstSeen == "" || len(listing.Locations) != 1 {
		t.Fatalf("listing = %+v", listing)
	}
}

func TestWriteJobsList(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJobsList(rec, []JobListing{{ID: "1", Company: "Acme", Locations: []string{}}})

	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestJobStatsFrom(t *testing.T) {
	var lastUpdated pgtype.Timestamptz
	if err := lastUpdated.Scan(time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	stats := JobStatsFrom(db.GetJobsStatsRow{
		TotalJobs:      10,
		AddedThisWeek:  2,
		TotalCompanies: 5,
		LastUpdated:    lastUpdated,
	})
	if stats.TotalJobs != 10 || stats.AddedThisWeek != 2 || stats.TotalCompanies != 5 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.LastUpdated != "2026-08-03T12:00:00Z" {
		t.Fatalf("last_updated = %q", stats.LastUpdated)
	}
}

func TestWriteJobsStats(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJobsStats(rec, JobStats{TotalJobs: 1, AddedThisWeek: 1, TotalCompanies: 1})
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
}
