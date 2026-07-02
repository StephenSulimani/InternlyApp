package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

type jobsListResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    []struct {
		ID      string `json:"id"`
		Company string `json:"company"`
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
	handler := testJobsHandler(store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs?limit=10", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var res jobsListResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !res.Success || len(res.Data) != 1 || res.Data[0].Company != "Acme" {
		t.Fatalf("response = %+v", res)
	}
	if store.limit != 10 {
		t.Fatalf("limit = %d, want 10", store.limit)
	}
}

func TestListJobsRoute_errors(t *testing.T) {
	t.Run("invalid limit", func(t *testing.T) {
		handler := testJobsHandler(&mockJobStore{})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs?limit=abc", nil)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusBadRequest, "Invalid limit")
	})

	t.Run("limit above max", func(t *testing.T) {
		handler := testJobsHandler(&mockJobStore{})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs?limit=500", nil)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusBadRequest, "Invalid limit")
	})

	t.Run("database error", func(t *testing.T) {
		handler := testJobsHandler(&mockJobStore{getJobsErr: errors.New("db down")})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusInternalServerError, "Error querying the database")
	})

	t.Run("empty result", func(t *testing.T) {
		handler := testJobsHandler(&mockJobStore{})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}

		var res jobsListResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatal(err)
		}
		if !res.Success || len(res.Data) != 0 {
			t.Fatalf("response = %+v, want empty data", res)
		}
	})
}

func testJobsHandler(store service.JobReader) http.Handler {
	log := zap.NewNop().Sugar()
	jobs := service.NewJobService(store)
	router := mux.NewRouter()
	router.Use(middleware.LoggerContext(log))
	router.Use(JobServiceMiddleware(jobs))
	router.HandleFunc("/jobs", ListJobs).Methods("GET")
	return router
}
