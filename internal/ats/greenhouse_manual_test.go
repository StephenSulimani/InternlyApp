//go:build manual

package ats

import (
	"context"
	"os"
	"testing"
)

const defaultGreenhouseBoard = "https://boards.greenhouse.io/stripe"

// TestManualGreenhouse_LiveBoard hits a real Greenhouse Job Board API.
//
// Run explicitly:
//
//	go test -tags=manual ./internal/ats -run TestManualGreenhouse_LiveBoard -v -timeout=30s
//
// Optional env:
//   - GREENHOUSE_BOARD: board URL (default: https://boards.greenhouse.io/stripe)
func TestManualGreenhouse_LiveBoard(t *testing.T) {
	boardURL := os.Getenv("GREENHOUSE_BOARD")
	if boardURL == "" {
		boardURL = defaultGreenhouseBoard
	}

	jobs, err := NewGreenhouse(nil).Scrape(context.Background(), Board{
		Name:    NameGreenhouse,
		URL:     boardURL,
		Company: "manual",
	})
	if err != nil {
		t.Fatalf("scrape failed: %v", err)
	}

	t.Logf("fetched %d board jobs from %s", len(jobs), boardURL)
	if len(jobs) == 0 {
		t.Fatal("expected at least one posting on the live board")
	}

	job := jobs[0]
	if job.ApplicationLink == "" {
		t.Fatal("expected application link")
	}
	if job.RoleTitle == nil || *job.RoleTitle == "" {
		t.Fatal("expected role title")
	}
	if job.SourceName != NameGreenhouse {
		t.Fatalf("source = %q", job.SourceName)
	}
	if job.Description != nil && *job.Description != "" {
		t.Logf("first=%q desc_len=%d link=%s", *job.RoleTitle, len(*job.Description), job.ApplicationLink)
		return
	}
	t.Logf("first=%q (no description) link=%s", *job.RoleTitle, job.ApplicationLink)
}
