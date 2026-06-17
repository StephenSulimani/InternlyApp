package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/internal/db"
	"go.uber.org/zap"
)

type Simplify struct {
	URL     string
	JobType string
}

func (s *Simplify) Scrape(log *zap.SugaredLogger) ([]db.Job, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch listings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, resp.Status)
	}

	var rawJSON []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rawJSON); err != nil {
		return nil, fmt.Errorf("decode listings: %w", err)
	}

	jobs, skipped, err := parseSimplifyListings(rawJSON, req.URL.String(), s.JobType)
	if err != nil {
		return nil, err
	}
	if skipped > 0 {
		log.Warnf("skipped %d simplify listings with invalid shape", skipped)
	}

	return jobs, nil
}

func parseSimplifyListings(raw []map[string]any, sourceURL, jobType string) ([]db.Job, int, error) {
	jobs := make([]db.Job, 0, len(raw))
	skipped := 0

	for _, entry := range raw {
		job, err := parseSimplifyListing(entry, sourceURL, jobType)
		if err != nil {
			skipped++
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, skipped, nil
}

func parseSimplifyListing(job map[string]any, sourceURL, jobType string) (db.Job, error) {
	locations, err := stringSliceField(job, "locations")
	if err != nil {
		return db.Job{}, fmt.Errorf("locations: %w", err)
	}

	companyName, err := stringField(job, "company_name")
	if err != nil {
		return db.Job{}, fmt.Errorf("company_name: %w", err)
	}

	role, err := stringField(job, "title")
	if err != nil {
		return db.Job{}, fmt.Errorf("title: %w", err)
	}

	applicationLink, err := stringField(job, "url")
	if err != nil {
		return db.Job{}, fmt.Errorf("url: %w", err)
	}

	seen, err := unixTimeField(job, "date_updated")
	if err != nil {
		return db.Job{}, fmt.Errorf("date_updated: %w", err)
	}

	return db.Job{
		FirstSeen: pgtype.Timestamptz{
			Time:  seen,
			Valid: true,
		},
		SourceUrl:       sourceURL,
		SourceName:      "Simplify",
		ApplicationLink: applicationLink,
		Company:         &companyName,
		RoleTitle:       &role,
		Locations:       locations,
		JobType:         &jobType,
	}, nil
}

func stringField(job map[string]any, key string) (string, error) {
	value, ok := job[key]
	if !ok {
		return "", fmt.Errorf("missing %q", key)
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return "", fmt.Errorf("invalid %q", key)
	}
	return text, nil
}

func stringSliceField(job map[string]any, key string) ([]string, error) {
	value, ok := job[key]
	if !ok {
		return nil, fmt.Errorf("missing %q", key)
	}

	switch locations := value.(type) {
	case []any:
		out := make([]string, 0, len(locations))
		for _, loc := range locations {
			text, ok := loc.(string)
			if !ok {
				return nil, fmt.Errorf("invalid %q entry", key)
			}
			out = append(out, text)
		}
		return out, nil
	case []string:
		return locations, nil
	default:
		return nil, fmt.Errorf("invalid %q", key)
	}
}

func unixTimeField(job map[string]any, key string) (time.Time, error) {
	value, ok := job[key]
	if !ok {
		return time.Time{}, fmt.Errorf("missing %q", key)
	}

	switch ts := value.(type) {
	case float64:
		return time.Unix(int64(ts), 0), nil
	case int64:
		return time.Unix(ts, 0), nil
	case json.Number:
		parsed, err := ts.Int64()
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid %q", key)
		}
		return time.Unix(parsed, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid %q", key)
	}
}
