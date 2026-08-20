package routes

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestListSavedSearchesRoute(t *testing.T) {
	var userID, searchID pgtype.UUID
	_ = userID.Scan("22222222-2222-2222-2222-222222222222")
	_ = searchID.Scan("33333333-3333-3333-3333-333333333333")

	handler, token := testAuthedSavedSearchHandler(t, &mockSavedSearchStore{
		searches: []db.UserSavedSearch{{
			ID:      searchID,
			UserID:  userID,
			Name:    "Remote interns",
			Q:       "intern",
			SortBy:  "posted",
			SortDir: "desc",
		}},
	}, activeTestUser())

	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodGet, "/saved-searches", token)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Success bool `json:"success"`
		Data    []struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !res.Success || len(res.Data) != 1 || res.Data[0].Name != "Remote interns" {
		t.Fatalf("response = %+v", res)
	}
}

func TestCreateSavedSearchRoute(t *testing.T) {
	handler, token := testAuthedSavedSearchHandler(t, &mockSavedSearchStore{}, activeTestUser())

	body, _ := json.Marshal(map[string]any{
		"name":     "Bay Area",
		"location": "San Francisco",
		"type":     "Internship",
		"sort":     "posted",
		"order":    "desc",
	})
	rec := httptest.NewRecorder()
	req := authedRequest(http.MethodPost, "/saved-searches", token)
	req.Header.Set("Content-Type", "application/json")
	req.Body = ioNopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func testAuthedSavedSearchHandler(t *testing.T, store service.SavedSearchStore, user db.User) (http.Handler, string) {
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

	searches := service.NewSavedSearchService(store)
	router := mux.NewRouter()
	router.Use(middleware.LoggerContext(log))
	router.Use(UserServiceMiddleware(users))
	router.Use(SavedSearchServiceMiddleware(searches))
	router.Use(TokenIssuerMiddleware(tokens))

	authed := router.NewRoute().Subrouter()
	authed.Use(RequireAuth)
	authed.HandleFunc("/saved-searches", ListSavedSearches).Methods("GET")
	authed.HandleFunc("/saved-searches/{id}", DeleteSavedSearch).Methods("DELETE")

	write := authed.NewRoute().Subrouter()
	write.Use(middleware.EnsureJSONBody)
	write.HandleFunc("/saved-searches", CreateSavedSearch).Methods("POST")
	write.HandleFunc("/saved-searches/{id}", UpdateSavedSearch).Methods("PUT")

	return router, token
}

type mockSavedSearchStore struct {
	searches []db.UserSavedSearch
}

func (m *mockSavedSearchStore) ListUserSavedSearches(ctx context.Context, userID pgtype.UUID) ([]db.UserSavedSearch, error) {
	return m.searches, nil
}

func (m *mockSavedSearchStore) GetUserSavedSearch(ctx context.Context, arg db.GetUserSavedSearchParams) (db.UserSavedSearch, error) {
	return db.UserSavedSearch{}, service.ErrSavedSearchNotFound
}

func (m *mockSavedSearchStore) CreateUserSavedSearch(ctx context.Context, arg db.CreateUserSavedSearchParams) (db.UserSavedSearch, error) {
	var id pgtype.UUID
	_ = id.Scan("33333333-3333-3333-3333-333333333333")
	return db.UserSavedSearch{
		ID:       id,
		UserID:   arg.UserID,
		Name:     arg.Name,
		Location: arg.Location,
		JobType:  arg.JobType,
		SortBy:   arg.SortBy,
		SortDir:  arg.SortDir,
	}, nil
}

func (m *mockSavedSearchStore) UpdateUserSavedSearch(ctx context.Context, arg db.UpdateUserSavedSearchParams) (db.UserSavedSearch, error) {
	return db.UserSavedSearch{ID: arg.ID, Name: arg.Name}, nil
}

func (m *mockSavedSearchStore) DeleteUserSavedSearch(ctx context.Context, arg db.DeleteUserSavedSearchParams) (int64, error) {
	return 1, nil
}

func ioNopCloser(b *bytes.Reader) *bytesReaderCloser {
	return &bytesReaderCloser{Reader: b}
}

type bytesReaderCloser struct {
	*bytes.Reader
}

func (b *bytesReaderCloser) Close() error { return nil }
