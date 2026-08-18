package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/internal/auth"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

type jobsListResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Jobs []struct {
			ID          string `json:"id"`
			Company     string `json:"company"`
			Description string `json:"description"`
			Saved       bool   `json:"saved"`
		} `json:"jobs"`
		Total  int64 `json:"total"`
		Limit  int   `json:"limit"`
		Offset int   `json:"offset"`
	} `json:"data"`
}

func TestListJobsRoute(t *testing.T) {
	company := "Acme"
	role := "Engineer Intern"
	var id pgtype.UUID
	if err := id.Scan("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Fatal(err)
	}

	store := &mockJobStore{
		jobs: []db.Job{
			{
				ID:              id,
				Company:         &company,
				RoleTitle:       &role,
				ApplicationLink: "https://jobs.example.com/1",
				SourceName:      "simplify",
			},
		},
	}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())

	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs?limit=10", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var res jobsListResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !res.Success || len(res.Data.Jobs) != 1 || res.Data.Jobs[0].Company != "Acme" {
		t.Fatalf("response = %+v", res)
	}
	if res.Data.Total != 1 || res.Data.Limit != 10 || res.Data.Offset != 0 {
		t.Fatalf("page = %+v", res.Data)
	}
	if store.limit != 10 {
		t.Fatalf("limit = %d, want 10", store.limit)
	}
}

func TestListJobsRoute_requiresAuth(t *testing.T) {
	handler, _ := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	handler.ServeHTTP(rec, req)
	assertAPIError(t, rec, http.StatusUnauthorized, "Authentication required")
}

func TestListJobsRoute_errors(t *testing.T) {
	t.Run("invalid limit", func(t *testing.T) {
		handler, token := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
		rec := httptest.NewRecorder()
		req := authedRequest(http.MethodGet, "/jobs?limit=abc", token)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusBadRequest, "Invalid limit")
	})

	t.Run("invalid offset", func(t *testing.T) {
		handler, token := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
		rec := httptest.NewRecorder()
		req := authedRequest(http.MethodGet, "/jobs?offset=-1", token)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusBadRequest, "Invalid offset")
	})

	t.Run("invalid recency", func(t *testing.T) {
		handler, token := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
		rec := httptest.NewRecorder()
		req := authedRequest(http.MethodGet, "/jobs?recency=month", token)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusBadRequest, "Invalid recency")
	})

	t.Run("database error", func(t *testing.T) {
		handler, token := testAuthedJobsHandler(t, &mockJobStore{getJobsErr: errors.New("db down")}, activeTestUser())
		rec := httptest.NewRecorder()
		req := authedRequest(http.MethodGet, "/jobs", token)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusInternalServerError, "Error querying the database")
	})

	t.Run("empty result", func(t *testing.T) {
		handler, token := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
		rec := httptest.NewRecorder()
		req := authedRequest(http.MethodGet, "/jobs", token)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}

		var res jobsListResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatal(err)
		}
		if !res.Success || len(res.Data.Jobs) != 0 || res.Data.Total != 0 {
			t.Fatalf("response = %+v, want empty data", res)
		}
	})
}

func TestListJobsRoute_filtersAndOffset(t *testing.T) {
	store := &mockJobStore{
		jobs: []db.Job{
			{ApplicationLink: "https://jobs.example.com/1"},
			{ApplicationLink: "https://jobs.example.com/2"},
			{ApplicationLink: "https://jobs.example.com/3"},
		},
	}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())

	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs?q=intern&type=Internship&location=Remote&source=Greenhouse&recency=24h&limit=2&offset=1", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var res jobsListResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !res.Success || len(res.Data.Jobs) != 2 || res.Data.Total != 3 {
		t.Fatalf("response = %+v", res)
	}
	if res.Data.Limit != 2 || res.Data.Offset != 1 {
		t.Fatalf("page = %+v", res.Data)
	}

	got := store.searchParams
	if got.Q != "%intern%" || got.FilterType != "Internship" || got.FilterSource != "Greenhouse" {
		t.Fatalf("search params = %+v", got)
	}
	if !slices.Equal(got.FilterLocations, []string{"%Remote%"}) {
		t.Fatalf("locations = %v", got.FilterLocations)
	}
	if got.RecencyHours != 24 || got.RowLimit != 2 || got.RowOffset != 1 {
		t.Fatalf("search params = %+v", got)
	}
}

func TestListJobsRoute_multipleLocations(t *testing.T) {
	store := &mockJobStore{}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())

	query := url.Values{}
	query.Set("location", "NYC, Manhattan, New York")
	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs?"+query.Encode(), token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	want := []string{"%NYC%", "%Manhattan%", "%New York%"}
	if !slices.Equal(store.searchParams.FilterLocations, want) {
		t.Fatalf("locations = %v, want %v", store.searchParams.FilterLocations, want)
	}
}

func TestListJobsRoute_quotedLocationWithComma(t *testing.T) {
	store := &mockJobStore{}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())

	query := url.Values{}
	query.Set("location", `"New York, NY", Remote`)
	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs?"+query.Encode(), token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	want := []string{"%New York, NY%", "%Remote%"}
	if !slices.Equal(store.searchParams.FilterLocations, want) {
		t.Fatalf("locations = %v, want %v", store.searchParams.FilterLocations, want)
	}
}

func TestListJobsRoute_sort(t *testing.T) {
	store := &mockJobStore{}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())
	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs?sort=company&order=asc", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if store.searchParams.SortBy != "company" || store.searchParams.SortDir != "asc" {
		t.Fatalf("search params = %+v", store.searchParams)
	}
}

func TestListJobsRoute_invalidSort(t *testing.T) {
	handler, token := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs?sort=salary", token)
	handler.ServeHTTP(rec, req)
	assertAPIError(t, rec, http.StatusBadRequest, "Invalid sort")
}

func TestJobLocationsRoute(t *testing.T) {
	store := &mockJobStore{locations: []string{"Remote", "New York, NY"}}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())

	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs/locations", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Success bool     `json:"success"`
		Data    []string `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !res.Success || len(res.Data) != 2 || res.Data[0] != "Remote" {
		t.Fatalf("response = %+v", res)
	}
}

func TestJobStatsRoute(t *testing.T) {
	var lastUpdated pgtype.Timestamptz
	if err := lastUpdated.Scan(time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	store := &mockJobStore{
		stats: db.GetJobsStatsRow{
			TotalJobs:      100,
			AddedThisWeek:  7,
			TotalCompanies: 40,
			LastUpdated:    lastUpdated,
		},
	}
	handler, _ := testAuthedJobsHandler(t, store, activeTestUser())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/stats", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Success bool `json:"success"`
		Data    struct {
			TotalJobs      int64  `json:"total_jobs"`
			AddedThisWeek  int64  `json:"added_this_week"`
			TotalCompanies int64  `json:"total_companies"`
			LastUpdated    string `json:"last_updated"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.Data.TotalJobs != 100 || res.Data.AddedThisWeek != 7 || res.Data.TotalCompanies != 40 {
		t.Fatalf("response = %+v", res)
	}
	if res.Data.LastUpdated == "" {
		t.Fatal("expected last_updated")
	}
}

func TestJobStatsRoute_databaseError(t *testing.T) {
	handler, _ := testAuthedJobsHandler(t, &mockJobStore{statsErr: errors.New("db down")}, activeTestUser())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/stats", nil)
	handler.ServeHTTP(rec, req)
	assertAPIError(t, rec, http.StatusInternalServerError, "Error querying the database")
}

func TestBoardPreviewRoute_isPublic(t *testing.T) {
	company := "Acme"
	store := &mockJobStore{
		jobs: []db.Job{
			{Company: &company, ApplicationLink: "https://jobs.example.com/1"},
		},
	}
	handler, _ := testAuthedJobsHandler(t, store, activeTestUser())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/board/preview", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRequireAuth_inactiveUser(t *testing.T) {
	user := activeTestUser()
	user.IsActive = false
	handler, token := testAuthedJobsHandler(t, &mockJobStore{}, user)

	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs", token)
	handler.ServeHTTP(rec, req)
	assertAPIError(t, rec, http.StatusForbidden, "Account pending activation")
}

func TestListJobsRoute_savedFilter(t *testing.T) {
	var jobID pgtype.UUID
	if err := jobID.Scan("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Fatal(err)
	}
	store := &mockJobStore{
		jobs:       []db.Job{{ID: jobID, ApplicationLink: "https://jobs.example.com/1"}},
		savedAmong: []pgtype.UUID{jobID},
	}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())

	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs?saved=true", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !store.searchParams.FilterSaved {
		t.Fatalf("filter_saved = %v", store.searchParams.FilterSaved)
	}

	var res jobsListResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if len(res.Data.Jobs) != 1 || !res.Data.Jobs[0].Saved {
		t.Fatalf("response = %+v", res)
	}
}

func TestListJobsRoute_invalidSaved(t *testing.T) {
	handler, token := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/jobs?saved=maybe", token)
	handler.ServeHTTP(rec, req)
	assertAPIError(t, rec, http.StatusBadRequest, "Invalid saved filter")
}

func TestSaveJobRoute(t *testing.T) {
	var jobID pgtype.UUID
	if err := jobID.Scan("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Fatal(err)
	}
	store := &mockJobStore{jobs: []db.Job{{ID: jobID}}}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())

	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodPut, "/jobs/"+jobID.String()+"/save", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(store.saveCalls) != 1 || store.saveCalls[0].JobID != jobID {
		t.Fatalf("save calls = %+v", store.saveCalls)
	}
}

func TestSaveJobRoute_notFound(t *testing.T) {
	handler, token := testAuthedJobsHandler(t, &mockJobStore{getJobErr: pgx.ErrNoRows}, activeTestUser())
	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodPut, "/jobs/550e8400-e29b-41d4-a716-446655440000/save", token)
	handler.ServeHTTP(rec, req)
	assertAPIError(t, rec, http.StatusNotFound, "Job not found")
}

func TestSaveJobRoute_invalidID(t *testing.T) {
	handler, token := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodPut, "/jobs/not-a-uuid/save", token)
	handler.ServeHTTP(rec, req)
	assertAPIError(t, rec, http.StatusBadRequest, "Invalid job id")
}

func TestUnsaveJobRoute(t *testing.T) {
	var jobID pgtype.UUID
	if err := jobID.Scan("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Fatal(err)
	}
	store := &mockJobStore{}
	handler, token := testAuthedJobsHandler(t, store, activeTestUser())

	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodDelete, "/jobs/"+jobID.String()+"/save", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(store.unsaveCalls) != 1 || store.unsaveCalls[0].JobID != jobID {
		t.Fatalf("unsave calls = %+v", store.unsaveCalls)
	}
}

func activeTestUser() db.User {
	var id pgtype.UUID
	_ = id.Scan("550e8400-e29b-41d4-a716-446655440000")
	return db.User{
		ID:        id,
		Email:     "ada@example.com",
		FirstName: "Ada",
		LastName:  "Lovelace",
		IsActive:  true,
	}
}

func authedRequest(method, path, token string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func testAuthedJobsHandler(t *testing.T, store service.JobReader, user db.User) (http.Handler, string) {
	t.Helper()

	log := zap.NewNop().Sugar()
	tokens := auth.NewTokenIssuer("test-secret", time.Hour)
	token, err := tokens.Issue(user)
	if err != nil {
		t.Fatal(err)
	}

	users := service.NewUserService(&mockUserStore{
		getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
			u := user
			u.Email = email
			return u, nil
		},
	}, nil)

	jobs := service.NewJobService(store)
	router := mux.NewRouter()
	router.Use(middleware.LoggerContext(log))
	router.Use(UserServiceMiddleware(users))
	router.Use(JobServiceMiddleware(jobs))
	router.Use(TokenIssuerMiddleware(tokens))

	router.HandleFunc("/board/preview", BoardPreview).Methods("GET")
	router.HandleFunc("/jobs/stats", JobStats).Methods("GET")

	authed := router.NewRoute().Subrouter()
	authed.Use(RequireAuth)
	authed.HandleFunc("/jobs", ListJobs).Methods("GET")
	authed.HandleFunc("/jobs/locations", JobLocations).Methods("GET")
	authed.HandleFunc("/jobs/{id}/save", SaveJob).Methods("PUT")
	authed.HandleFunc("/jobs/{id}/save", UnsaveJob).Methods("DELETE")

	return router, token
}
