package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/internal/auth"
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
	if !res.Success || len(res.Data) != 1 || res.Data[0].Company != "Acme" {
		t.Fatalf("response = %+v", res)
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

	t.Run("limit above max", func(t *testing.T) {
		handler, token := testAuthedJobsHandler(t, &mockJobStore{}, activeTestUser())
		rec := httptest.NewRecorder()
		req := authedRequest(http.MethodGet, "/jobs?limit=500", token)
		handler.ServeHTTP(rec, req)
		assertAPIError(t, rec, http.StatusBadRequest, "Invalid limit")
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
		if !res.Success || len(res.Data) != 0 {
			t.Fatalf("response = %+v, want empty data", res)
		}
	})
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

	return router, token
}
