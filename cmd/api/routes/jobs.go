package routes

import (
	"net/http"
	"strconv"

	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/types"
	"github.com/stephensulimani/internlyapp/internal/service"
)

func ListJobs(w http.ResponseWriter, r *http.Request) {
	log, ok := middleware.LoggerFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	jobsService, ok := jobServiceFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	limit, err := jobsLimitFromRequest(r)
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, "Invalid limit")
		return
	}

	jobs, err := jobsService.List(r.Context(), limit)
	if err != nil {
		writeJobsError(w, log, err)
		return
	}

	types.WriteJobsList(w, types.JobListingsFrom(jobs))
}

func jobsLimitFromRequest(r *http.Request) (int, error) {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return service.DefaultJobsLimit, nil
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}

	return service.NormalizeJobsLimit(parsed)
}

func JobStats(w http.ResponseWriter, r *http.Request) {
	log, ok := middleware.LoggerFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	jobsService, ok := jobServiceFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	stats, err := jobsService.Stats(r.Context())
	if err != nil {
		writeJobsError(w, log, err)
		return
	}

	types.WriteJobsStats(w, types.JobStatsFrom(stats))
}

func BoardPreview(w http.ResponseWriter, r *http.Request) {
	log, ok := middleware.LoggerFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	jobsService, ok := jobServiceFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	jobs, err := jobsService.List(r.Context(), 5)
	if err != nil {
		writeJobsError(w, log, err)
		return
	}

	types.WriteJobsList(w, types.JobListingsFrom(jobs))
}
