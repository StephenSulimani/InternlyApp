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

type loginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token"`
	User    struct {
		Email string `json:"email"`
	} `json:"user"`
}

func TestLoginRoute(t *testing.T) {
	hash, err := auth.HashPassword("secure-password")
	if err != nil {
		t.Fatal(err)
	}

	var id pgtype.UUID
	if err := id.Scan("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Fatal(err)
	}

	store := &mockUserStore{
		getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
			return db.User{
				ID:        id,
				Email:     email,
				Password:  hash,
				IsActive:  true,
				FirstName: "Ada",
				LastName:  "Lovelace",
			}, nil
		},
	}

	handler := testLoginHandler(store)
	rec := postLogin(t, handler, loginBody{
		Email:    "ada@example.com",
		Password: "secure-password",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var res loginResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !res.Success || res.Token == "" {
		t.Fatalf("response = %+v", res)
	}
	if res.User.Email != "ada@example.com" {
		t.Fatalf("email = %q", res.User.Email)
	}

	issuer := auth.NewTokenIssuer("test-secret", time.Hour)
	claims, err := issuer.Parse(res.Token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != id.String() {
		t.Fatalf("user_id = %q", claims.UserID)
	}
}

func TestLoginRoute_errors(t *testing.T) {
	t.Run("invalid credentials", func(t *testing.T) {
		handler := testLoginHandler(&mockUserStore{})
		rec := postLogin(t, handler, loginBody{
			Email:    "missing@example.com",
			Password: "secure-password",
		})
		assertAPIError(t, rec, http.StatusUnauthorized, "Invalid email or password")
	})

	t.Run("inactive account", func(t *testing.T) {
		hash, err := auth.HashPassword("secure-password")
		if err != nil {
			t.Fatal(err)
		}
		store := &mockUserStore{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return db.User{
					Email:    email,
					Password: hash,
					IsActive: false,
				}, nil
			},
		}
		handler := testLoginHandler(store)
		rec := postLogin(t, handler, loginBody{
			Email:    "ada@example.com",
			Password: "secure-password",
		})
		assertAPIError(t, rec, http.StatusForbidden, "Account pending activation")
	})
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func testLoginHandler(store service.UserStore) http.Handler {
	log := zap.NewNop().Sugar()
	tokens := auth.NewTokenIssuer("test-secret", time.Hour)
	router := mux.NewRouter()
	router.Use(middleware.LoggerContext(log))
	router.Use(UserServiceMiddleware(service.NewUserService(store, nil)))
	router.Use(TokenIssuerMiddleware(tokens))
	router.Use(middleware.EnsureJSONBody)
	router.HandleFunc("/login", LoginUser).Methods("POST")
	return router
}

func postLogin(t *testing.T, handler http.Handler, body loginBody) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", &buf)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
