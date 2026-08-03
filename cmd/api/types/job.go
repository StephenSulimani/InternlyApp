package types

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type JobListing struct {
	ID              string   `json:"id"`
	Company         string   `json:"company"`
	RoleTitle       string   `json:"role_title"`
	Locations       []string `json:"locations"`
	JobType         string   `json:"job_type"`
	ApplicationLink string   `json:"application_link"`
	FirstSeen       string   `json:"first_seen,omitempty"`
	SourceName      string   `json:"source_name"`
}

type jobsListResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    []JobListing `json:"data"`
}

func JobListingFrom(j db.Job) JobListing {
	listing := JobListing{
		ID:              j.ID.String(),
		Company:         derefString(j.Company),
		RoleTitle:       derefString(j.RoleTitle),
		Locations:       j.Locations,
		JobType:         derefString(j.JobType),
		ApplicationLink: j.ApplicationLink,
		SourceName:      j.SourceName,
	}

	if j.FirstSeen.Valid {
		listing.FirstSeen = j.FirstSeen.Time.UTC().Format(time.RFC3339)
	}

	if listing.Locations == nil {
		listing.Locations = []string{}
	}

	return listing
}

func JobListingsFrom(jobs []db.Job) []JobListing {
	listings := make([]JobListing, 0, len(jobs))
	for _, job := range jobs {
		listings = append(listings, JobListingFrom(job))
	}
	return listings
}

func WriteJobsList(w http.ResponseWriter, jobs []JobListing) {
	body, err := json.Marshal(jobsListResponse{
		Success: true,
		Message: "Jobs retrieved",
		Data:    jobs,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

type JobStats struct {
	TotalJobs      int64  `json:"total_jobs"`
	AddedThisWeek  int64  `json:"added_this_week"`
	TotalCompanies int64  `json:"total_companies"`
	LastUpdated    string `json:"last_updated,omitempty"`
}

type jobsStatsResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    JobStats `json:"data"`
}

func JobStatsFrom(row db.GetJobsStatsRow) JobStats {
	stats := JobStats{
		TotalJobs:      row.TotalJobs,
		AddedThisWeek:  row.AddedThisWeek,
		TotalCompanies: row.TotalCompanies,
	}
	if row.LastUpdated.Valid {
		stats.LastUpdated = row.LastUpdated.Time.UTC().Format(time.RFC3339)
	}
	return stats
}

func WriteJobsStats(w http.ResponseWriter, stats JobStats) {
	body, err := json.Marshal(jobsStatsResponse{
		Success: true,
		Message: "Job stats retrieved",
		Data:    stats,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
