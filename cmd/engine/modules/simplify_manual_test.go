//go:build manual

package modules

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

const defaultSimplifyInternURL = "https://raw.githubusercontent.com/SimplifyJobs/Summer2026-Internships/refs/heads/dev/.github/scripts/listings.json"

// TestManualSimplify_LiveFeed hits the real Simplify listings JSON feed.
//
// Run explicitly:
//
//	go test -tags=manual ./cmd/engine/modules -run TestManualSimplify_LiveFeed -v -timeout=30s
//
// Optional env:
//   - SIMPLIFY_URL: override listings endpoint
func TestManualSimplify_LiveFeed(t *testing.T) {
	url := os.Getenv("SIMPLIFY_URL")
	if url == "" {
		url = defaultSimplifyInternURL
	}

	scraper := &Simplify{URL: url, JobType: "Internship"}
	jobs := scraper.Scrape(zap.NewNop().Sugar())

	if len(jobs) == 0 {
		t.Fatal("expected at least one job from live Simplify feed")
	}

	job := jobs[0]
	if job.ApplicationLink == "" {
		t.Fatal("expected application link")
	}
	if job.Company == nil || *job.Company == "" {
		t.Fatal("expected company name")
	}
	if job.RoleTitle == nil || *job.RoleTitle == "" {
		t.Fatal("expected role title")
	}
	if len(job.Locations) == 0 {
		t.Fatal("expected at least one location")
	}

	t.Logf("fetched %d jobs; first=%q at %q", len(jobs), *job.RoleTitle, *job.Company)
}
