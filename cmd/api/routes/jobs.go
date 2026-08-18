package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
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

	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	query, err := jobsQueryFromRequest(r)
	if err != nil {
		writeJobsQueryError(w, err)
		return
	}
	query.UserID = user.ID

	page, err := jobsService.Search(r.Context(), query)
	if err != nil {
		writeJobsError(w, log, err)
		return
	}

	types.WriteJobsPage(w, types.JobsPageFrom(page.Jobs, page.SavedIDs, page.Total, page.Limit, page.Offset))
}

func SaveJob(w http.ResponseWriter, r *http.Request) {
	writeSaveJob(w, r, true)
}

func UnsaveJob(w http.ResponseWriter, r *http.Request) {
	writeSaveJob(w, r, false)
}

func writeSaveJob(w http.ResponseWriter, r *http.Request, saved bool) {
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

	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	jobID, err := jobIDFromPath(r)
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, "Invalid job id")
		return
	}

	if saved {
		err = jobsService.Save(r.Context(), user.ID, jobID)
	} else {
		err = jobsService.Unsave(r.Context(), user.ID, jobID)
	}
	if err != nil {
		writeJobsError(w, log, err)
		return
	}

	if saved {
		types.WriteJobSaved(w, true, "Job saved")
		return
	}
	types.WriteJobSaved(w, false, "Job unsaved")
}

func jobIDFromPath(r *http.Request) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(mux.Vars(r)["id"]); err != nil || !id.Valid {
		return pgtype.UUID{}, service.ErrInvalidJobID
	}
	return id, nil
}

func jobsQueryFromRequest(r *http.Request) (service.JobListQuery, error) {
	limit, err := jobsLimitFromRequest(r)
	if err != nil {
		return service.JobListQuery{}, err
	}

	offset, err := jobsOffsetFromRequest(r)
	if err != nil {
		return service.JobListQuery{}, err
	}

	recencyHours, err := service.ParseRecency(r.URL.Query().Get("recency"))
	if err != nil {
		return service.JobListQuery{}, err
	}

	sortBy, sortDir, err := service.ParseSort(r.URL.Query().Get("sort"), r.URL.Query().Get("order"))
	if err != nil {
		return service.JobListQuery{}, err
	}

	savedOnly, err := service.ParseSavedFilter(r.URL.Query().Get("saved"))
	if err != nil {
		return service.JobListQuery{}, err
	}

	return service.JobListQuery{
		Q:            r.URL.Query().Get("q"),
		Type:         r.URL.Query().Get("type"),
		Location:     r.URL.Query().Get("location"),
		Source:       r.URL.Query().Get("source"),
		RecencyHours: recencyHours,
		SavedOnly:    savedOnly,
		SortBy:       sortBy,
		SortDir:      sortDir,
		Limit:        limit,
		Offset:       offset,
	}, nil
}

func jobsLimitFromRequest(r *http.Request) (int, error) {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return service.DefaultJobsLimit, nil
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, service.ErrInvalidJobsLimit
	}

	return service.NormalizeJobsLimit(parsed)
}

func jobsOffsetFromRequest(r *http.Request) (int, error) {
	raw := r.URL.Query().Get("offset")
	if raw == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, service.ErrInvalidJobsOffset
	}

	return service.NormalizeJobsOffset(parsed)
}

func writeJobsQueryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidJobsLimit):
		types.WriteError(w, http.StatusBadRequest, "Invalid limit")
	case errors.Is(err, service.ErrInvalidJobsOffset):
		types.WriteError(w, http.StatusBadRequest, "Invalid offset")
	case errors.Is(err, service.ErrInvalidJobsRecency):
		types.WriteError(w, http.StatusBadRequest, "Invalid recency")
	case errors.Is(err, service.ErrInvalidJobsSort):
		types.WriteError(w, http.StatusBadRequest, "Invalid sort")
	case errors.Is(err, service.ErrInvalidJobsSaved):
		types.WriteError(w, http.StatusBadRequest, "Invalid saved filter")
	default:
		types.WriteError(w, http.StatusBadRequest, "Invalid query")
	}
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

func JobLocations(w http.ResponseWriter, r *http.Request) {
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

	locations, err := jobsService.Locations(r.Context())
	if err != nil {
		writeJobsError(w, log, err)
		return
	}

	types.WriteJobLocations(w, locations)
}
