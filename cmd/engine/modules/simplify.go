package modules

import (
	"encoding/json"
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

func (s *Simplify) Scrape(log *zap.SugaredLogger) []db.Job {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		log.Error(err)
		return []db.Job{}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return []db.Job{}
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
		return []db.Job{}
	}

	var jobs []db.Job

	var raw_json []map[string]any

	err = json.NewDecoder(resp.Body).Decode(&raw_json)
	if err != nil {
		log.Error(err)
		return []db.Job{}
	}

	for _, job := range raw_json {
		locations := job["locations"].([]any)
		location_str := make([]string, len(locations))
		for i, loc := range locations {
			location_str[i] = loc.(string)
		}
		company_name := job["company_name"].(string)
		role := job["title"].(string)
		seen := time.Unix(int64(job["date_updated"].(float64)), 0)
		pgTimestamptz := pgtype.Timestamptz{
			Time:  seen,
			Valid: true,
		}
		jobs = append(jobs, db.Job{
			FirstSeen:       pgTimestamptz,
			SourceUrl:       req.URL.String(),
			SourceName:      "Simplify",
			ApplicationLink: job["url"].(string),
			Company:         &company_name,
			RoleTitle:       &role,
			Locations:       location_str,
			JobType:         &s.JobType,
		})
	}

	return jobs
}
