package ats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/internal/db"
)

const greenhouseAPIBase = "https://boards-api.greenhouse.io"

// Greenhouse scrapes a company board via Greenhouse's public Job Board API.
//
//	GET {APIBase}/v1/boards/{token}/jobs?content=true
//
// Returns every posting on the board (not only intern/new-grad). Ingest
// enriches existing Simplify rows by apply URL and inserts early-career jobs
// that Simplify missed.
//
// APIBase defaults to https://boards-api.greenhouse.io. Override it in tests.
type Greenhouse struct {
	HTTP    *http.Client
	APIBase string
}

func NewGreenhouse(client *http.Client) *Greenhouse {
	if client == nil {
		client = NewClient()
	}
	return &Greenhouse{HTTP: client}
}

func (g *Greenhouse) Name() string { return NameGreenhouse }

type greenhouseJobsResponse struct {
	Jobs []greenhouseJob `json:"jobs"`
}

type greenhouseJob struct {
	ID          int64              `json:"id"`
	Title       string             `json:"title"`
	AbsoluteURL string             `json:"absolute_url"`
	UpdatedAt   string             `json:"updated_at"`
	Content     string             `json:"content"`
	Location    greenhouseLocation `json:"location"`
	Offices     []greenhouseOffice `json:"offices"`
}

type greenhouseLocation struct {
	Name string `json:"name"`
}

type greenhouseOffice struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func (g *Greenhouse) Scrape(ctx context.Context, board Board) ([]db.Job, error) {
	if g.HTTP == nil {
		g.HTTP = NewClient()
	}

	token, err := greenhouseToken(board.URL)
	if err != nil {
		return nil, err
	}

	req, err := newJSONRequest(ctx, http.MethodGet, g.jobsURL(token))
	if err != nil {
		return nil, fmt.Errorf("greenhouse request: %w", err)
	}

	resp, err := g.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("greenhouse fetch %s: %w", token, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, fmt.Errorf("greenhouse read %s: %w", token, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusNotFound, http.StatusGone:
		return nil, fmt.Errorf("%w: greenhouse %s: status %d", ErrBoardGone, token, resp.StatusCode)
	default:
		return nil, fmt.Errorf("greenhouse %s: unexpected status %d", token, resp.StatusCode)
	}

	var payload greenhouseJobsResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("greenhouse decode %s: %w", token, err)
	}

	jobs := make([]db.Job, 0, len(payload.Jobs))
	for _, raw := range payload.Jobs {
		job, ok := mapGreenhouseJob(board, raw)
		if !ok {
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (g *Greenhouse) jobsURL(token string) string {
	u, err := url.Parse(strings.TrimRight(g.base(), "/") + "/v1/boards/" + url.PathEscape(token) + "/jobs")
	if err != nil {
		return strings.TrimRight(g.base(), "/") + "/v1/boards/" + url.PathEscape(token) + "/jobs?content=true"
	}
	q := u.Query()
	q.Set("content", "true")
	u.RawQuery = q.Encode()
	return u.String()
}

func (g *Greenhouse) base() string {
	if g.APIBase != "" {
		return g.APIBase
	}
	return greenhouseAPIBase
}

func greenhouseToken(boardURL string) (string, error) {
	link, ok := parseLink(boardURL)
	if !ok {
		return "", fmt.Errorf("greenhouse: invalid board url %q", boardURL)
	}
	token, ok := greenhousePath(link)
	if !ok {
		return "", fmt.Errorf("greenhouse: no board token in %q", boardURL)
	}
	return token, nil
}

func mapGreenhouseJob(board Board, raw greenhouseJob) (db.Job, bool) {
	title := strings.TrimSpace(raw.Title)
	applyURL := strings.TrimSpace(raw.AbsoluteURL)
	if title == "" || applyURL == "" {
		return db.Job{}, false
	}

	company := strings.TrimSpace(board.Company)
	job := db.Job{
		SourceUrl:       board.URL,
		SourceName:      NameGreenhouse,
		FirstSeen:       greenhouseFirstSeen(raw.UpdatedAt),
		ApplicationLink: applyURL,
		RoleTitle:       strPtr(title),
		Locations:       greenhouseLocations(raw),
		IsAts:           boolPtr(true),
	}
	if jobType, ok := ClassifyEarlyCareer(title); ok {
		job.JobType = strPtr(jobType)
	}
	if company != "" {
		job.Company = strPtr(company)
	}
	if desc := HTMLToText(raw.Content); desc != "" {
		job.Description = strPtr(desc)
	}
	return job, true
}

func greenhouseLocations(raw greenhouseJob) []string {
	seen := make(map[string]struct{})
	var out []string

	add := func(value string) {
		for _, part := range strings.Split(value, ";") {
			loc := strings.TrimSpace(part)
			if loc == "" {
				continue
			}
			key := strings.ToLower(loc)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, loc)
		}
	}

	add(raw.Location.Name)
	if len(out) > 0 {
		return out
	}
	for _, office := range raw.Offices {
		if office.Location != "" {
			add(office.Location)
			continue
		}
		add(office.Name)
	}
	if out == nil {
		return []string{}
	}
	return out
}

func greenhouseFirstSeen(updatedAt string) pgtype.Timestamptz {
	if t, ok := parseGreenhouseTime(updatedAt); ok {
		return pgtype.Timestamptz{Time: t, Valid: true}
	}
	return pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
}

func parseGreenhouseTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func strPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }
