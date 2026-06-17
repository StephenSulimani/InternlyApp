//go:build manual

package linkedin

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stephensulimani/internlyapp/internal/testutil"
	"go.uber.org/zap"
)

// TestManualLinkedIn_Scrape launches Chromium via Playwright against live LinkedIn job search.
//
// Run explicitly:
//
//	go test -tags=manual ./cmd/engine/modules/linkedin -run TestManualLinkedIn_Scrape -v -timeout=10m
//
// Chromium path resolution:
//   - CHROMIUM_PATH env, or
//   - <repo>/Chromium.app/Contents/MacOS/Chromium
//
// Optional env:
//   - LINKEDIN_KEYWORD (default: "Software Engineer")
//   - LINKEDIN_LOCATION (default: "New York")
//   - LINKEDIN_MAX_JOBS (default: "2")
//   - LINKEDIN_HEADLESS (default: "true")
func TestManualLinkedIn_Scrape(t *testing.T) {
	chromiumPath := resolveChromiumPath(t)

	keyword := envOrDefault("LINKEDIN_KEYWORD", "Software Engineer")
	location := envOrDefault("LINKEDIN_LOCATION", "New York")
	maxJobs := envIntOrDefault("LINKEDIN_MAX_JOBS", 2)
	headless := envBoolOrDefault("LINKEDIN_HEADLESS", true)

	scraper := LinkedIn{
		ChromiumPath: chromiumPath,
		MaxJobs:      maxJobs,
		Headless:     headless,
		SearchParams: LinkedInSearchParams{
			Keyword:  keyword,
			Location: location,
		},
	}

	jobs, err := scraper.Scrape(zap.NewNop().Sugar())
	if err != nil {
		t.Fatalf("scrape failed: %v", err)
	}
	if len(jobs) == 0 {
		t.Fatal("expected at least one job from live LinkedIn scrape")
	}

	job := jobs[0]
	if job.ApplicationLink == "" {
		t.Fatal("expected application link")
	}
	if job.RoleTitle == nil || *job.RoleTitle == "" {
		t.Fatal("expected role title")
	}

	t.Logf("scraped %d jobs; first=%q (%q)", len(jobs), deref(job.RoleTitle), deref(job.Company))
}

func resolveChromiumPath(t *testing.T) string {
	t.Helper()

	if path := os.Getenv("CHROMIUM_PATH"); path != "" {
		return path
	}

	root := testutil.RepoRoot(t)
	defaultPath := filepath.Join(root, "Chromium.app", "Contents", "MacOS", "Chromium")
	if _, err := os.Stat(defaultPath); err != nil {
		t.Skip("CHROMIUM_PATH is not set and Chromium.app was not found at repo root")
	}

	return defaultPath
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
