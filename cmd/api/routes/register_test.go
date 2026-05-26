package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/routes"
	"github.com/stephensulimani/internlyapp/cmd/api/utils"
	"github.com/stephensulimani/internlyapp/internal/testutil"
	"go.uber.org/zap"
)

type apiResponse struct {
	Success int    `json:"success"`
	Message string `json:"message"`
}

type storedUser struct {
	FirstName string
	LastName  string
	Password  string
	IsAdmin   bool
	IsActive  bool
	IsPremium bool
}

func TestRegisterRoute(t *testing.T) {
	pool, cleanup := testutil.SetupPostgres(t)
	defer cleanup()

	handler := registerHandler(pool)

	t.Run("creates first user with bootstrap privileges", func(t *testing.T) {
		testutil.TruncateUsers(t, pool)

		rec := postRegister(t, handler, registerBody{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "ada@example.com",
			Password:  "secure-password",
		})

		assertStatus(t, rec, http.StatusOK)
		res := decodeAPIResponse(t, rec.Body)
		if res.Success != 1 {
			t.Fatalf("success = %d, want 1", res.Success)
		}
		if res.Message != "User successfully registered" {
			t.Fatalf("message = %q", res.Message)
		}

		user := getUserByEmail(t, pool, "ada@example.com")
		if user.FirstName != "Ada" || user.LastName != "Lovelace" {
			t.Fatalf("stored name = %q %q", user.FirstName, user.LastName)
		}
		if !user.IsAdmin || !user.IsActive || !user.IsPremium {
			t.Fatal("expected first user to be admin, active, and premium")
		}
		if !utils.CheckPasswordHash("secure-password", user.Password) {
			t.Fatal("expected password to be stored as a bcrypt hash")
		}
	})

	t.Run("creates subsequent user without bootstrap privileges", func(t *testing.T) {
		testutil.TruncateUsers(t, pool)

		first := postRegister(t, handler, registerBody{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "first@example.com",
			Password:  "secure-password",
		})
		assertStatus(t, first, http.StatusOK)

		second := postRegister(t, handler, registerBody{
			FirstName: "Grace",
			LastName:  "Hopper",
			Email:     "grace@example.com",
			Password:  "another-password",
		})
		assertStatus(t, second, http.StatusOK)

		user := getUserByEmail(t, pool, "grace@example.com")
		if user.IsAdmin || user.IsActive || user.IsPremium {
			t.Fatal("expected non-first user to have default privilege flags")
		}
	})

	t.Run("rejects missing required fields", func(t *testing.T) {
		testutil.TruncateUsers(t, pool)

		cases := []struct {
			name string
			body registerBody
		}{
			{"missing first_name", registerBody{LastName: "Lovelace", Email: "a@example.com", Password: "pw"}},
			{"missing last_name", registerBody{FirstName: "Ada", Email: "a@example.com", Password: "pw"}},
			{"missing email", registerBody{FirstName: "Ada", LastName: "Lovelace", Password: "pw"}},
			{"missing password", registerBody{FirstName: "Ada", LastName: "Lovelace", Email: "a@example.com"}},
			{"empty body", registerBody{}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				rec := postRegister(t, handler, tc.body)
				assertStatus(t, rec, http.StatusBadRequest)

				res := decodeAPIResponse(t, rec.Body)
				if res.Message != "Missing required fields" {
					t.Fatalf("message = %q, want %q", res.Message, "Missing required fields")
				}
			})
		}
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		testutil.TruncateUsers(t, pool)

		body := registerBody{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "duplicate@example.com",
			Password:  "secure-password",
		}

		assertStatus(t, postRegister(t, handler, body), http.StatusOK)

		rec := postRegister(t, handler, body)
		assertStatus(t, rec, http.StatusBadRequest)

		res := decodeAPIResponse(t, rec.Body)
		if res.Message != "User already exists" {
			t.Fatalf("message = %q, want %q", res.Message, "User already exists")
		}
	})

	t.Run("rejects json with wrong field types", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(
			`{"first_name":123,"last_name":"Lovelace","email":"types@example.com","password":"pw"}`,
		))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)

		res := decodeAPIResponse(t, rec.Body)
		if res.Message != "Error parsing request body" {
			t.Fatalf("message = %q, want %q", res.Message, "Error parsing request body")
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{`))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusBadRequest)

		res := decodeAPIResponse(t, rec.Body)
		if res.Message != "Error parsing request body" {
			t.Fatalf("message = %q, want %q", res.Message, "Error parsing request body")
		}
	})

	t.Run("rejects non-json content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "text/plain")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusUnsupportedMediaType)

		res := decodeAPIResponse(t, rec.Body)
		if res.Message != "Expected application/json" {
			t.Fatalf("message = %q, want %q", res.Message, "Expected application/json")
		}
	})

	t.Run("rejects non-post method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/register", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
}

type registerBody struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
}

func registerHandler(pool *pgxpool.Pool) http.Handler {
	log := zap.NewNop().Sugar()
	router := mux.NewRouter()
	router.Use(middleware.DatabaseMiddleware(pool))
	router.Use(middleware.LoggerContext(log))
	router.PathPrefix("/").Handler(routes.UserRouter())
	return router
}

func postRegister(t *testing.T, handler http.Handler, body registerBody) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/register", &buf)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func getUserByEmail(t *testing.T, pool *pgxpool.Pool, email string) storedUser {
	t.Helper()

	var user storedUser
	err := pool.QueryRow(context.Background(), `
		SELECT first_name, last_name, password, is_admin, is_active, is_premium
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.IsAdmin,
		&user.IsActive,
		&user.IsPremium,
	)
	if err != nil {
		t.Fatalf("query user by email: %v", err)
	}
	return user
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}

func decodeAPIResponse(t *testing.T, body io.Reader) apiResponse {
	t.Helper()

	var res apiResponse
	if err := json.NewDecoder(body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return res
}
