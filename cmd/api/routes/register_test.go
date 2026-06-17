package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/utils"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

type apiResponse struct {
	Success int    `json:"success"`
	Message string `json:"message"`
}

type mockUserStore struct {
	count      int64
	countErr   error
	createErr  error
	createUser func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	createCalls []db.CreateUserParams
}

func (m *mockUserStore) GetUserCount(ctx context.Context) (int64, error) {
	return m.count, m.countErr
}

func (m *mockUserStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	m.createCalls = append(m.createCalls, arg)
	if m.createUser != nil {
		return m.createUser(ctx, arg)
	}
	if m.createErr != nil {
		return db.User{}, m.createErr
	}
	return db.User{}, nil
}

func TestRegisterRoute(t *testing.T) {
	t.Run("creates first user with bootstrap privileges", func(t *testing.T) {
		store := &mockUserStore{count: 0}
		handler := testRegisterHandler(store)

		rec := postRegister(t, handler, registerBody{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "ada@example.com",
			Password:  "secure-password",
		})

		assertStatus(t, rec, http.StatusOK)
		res := decodeAPIResponse(t, rec.Body)
		if res.Message != "User successfully registered" {
			t.Fatalf("message = %q", res.Message)
		}

		if len(store.createCalls) != 1 {
			t.Fatalf("create calls = %d, want 1", len(store.createCalls))
		}
		created := store.createCalls[0]
		if created.FirstName != "Ada" || created.LastName != "Lovelace" || created.Email != "ada@example.com" {
			t.Fatalf("unexpected create params: %+v", created)
		}
		if !boolVal(created.IsAdmin) || !boolVal(created.IsActive) || !boolVal(created.IsPremium) {
			t.Fatal("expected first user to be admin, active, and premium")
		}
		if !utils.CheckPasswordHash("secure-password", created.Password) {
			t.Fatal("expected password to be stored as a bcrypt hash")
		}
	})

	t.Run("creates subsequent user without bootstrap privileges", func(t *testing.T) {
		store := &mockUserStore{count: 1}
		handler := testRegisterHandler(store)

		rec := postRegister(t, handler, registerBody{
			FirstName: "Grace",
			LastName:  "Hopper",
			Email:     "grace@example.com",
			Password:  "another-password",
		})
		assertStatus(t, rec, http.StatusOK)

		created := store.createCalls[0]
		if boolVal(created.IsAdmin) || boolVal(created.IsActive) || boolVal(created.IsPremium) {
			t.Fatal("expected non-first user to have default privilege flags")
		}
	})

	t.Run("rejects missing required fields", func(t *testing.T) {
		store := &mockUserStore{}
		handler := testRegisterHandler(store)

		cases := []struct {
			name string
			body registerBody
		}{
			{"missing first_name", registerBody{LastName: "Lovelace", Email: "a@example.com", Password: "password1"}},
			{"missing last_name", registerBody{FirstName: "Ada", Email: "a@example.com", Password: "password1"}},
			{"missing email", registerBody{FirstName: "Ada", LastName: "Lovelace", Password: "password1"}},
			{"missing password", registerBody{FirstName: "Ada", LastName: "Lovelace", Email: "a@example.com"}},
			{"empty body", registerBody{}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				rec := postRegister(t, handler, tc.body)
				assertAPIError(t, rec, http.StatusBadRequest, "Missing required fields")
				if len(store.createCalls) != 0 {
					t.Fatal("expected store not to be called")
				}
			})
		}
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		store := &mockUserStore{
			count: 1,
			createUser: func(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
				return db.User{}, &pgconn.PgError{Code: "23505"}
			},
		}
		handler := testRegisterHandler(store)

		rec := postRegister(t, handler, registerBody{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "duplicate@example.com",
			Password:  "secure-password",
		})
		assertAPIError(t, rec, http.StatusBadRequest, "User already exists")
	})

	t.Run("rejects invalid email", func(t *testing.T) {
		handler := testRegisterHandler(&mockUserStore{})

		rec := postRegister(t, handler, registerBody{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "not-an-email",
			Password:  "secure-password",
		})
		assertAPIError(t, rec, http.StatusBadRequest, "Invalid email address")
	})

	t.Run("rejects weak password", func(t *testing.T) {
		handler := testRegisterHandler(&mockUserStore{})

		rec := postRegister(t, handler, registerBody{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "weak@example.com",
			Password:  "short",
		})
		assertAPIError(t, rec, http.StatusBadRequest, "Password must be at least 8 characters")
	})

	t.Run("rejects json with wrong field types", func(t *testing.T) {
		handler := testRegisterHandler(&mockUserStore{})

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(
			`{"first_name":123,"last_name":"Lovelace","email":"types@example.com","password":"pw"}`,
		))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assertAPIError(t, rec, http.StatusBadRequest, "Error parsing request body")
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		handler := testRegisterHandler(&mockUserStore{})

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{`))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assertAPIError(t, rec, http.StatusBadRequest, "Error parsing request body")
	})

	t.Run("rejects non-json content type", func(t *testing.T) {
		handler := testRegisterHandler(&mockUserStore{})

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "text/plain")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assertAPIError(t, rec, http.StatusUnsupportedMediaType, "Expected application/json")
	})

	t.Run("rejects non-post method", func(t *testing.T) {
		handler := testRegisterHandler(&mockUserStore{})

		req := httptest.NewRequest(http.MethodGet, "/register", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestRegisterUser_errors(t *testing.T) {
	log := zap.NewNop().Sugar()

	t.Run("missing request dependencies", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/register", nil)
		RegisterUser(rec, req)

		assertAPIError(t, rec, http.StatusInternalServerError, "Error getting request dependencies")
	})

	t.Run("handler json unmarshal error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := registerRequest(t, []byte(`{"first_name":123,"last_name":"Lovelace","email":"a@example.com","password":"pw"}`), service.NewUserService(&mockUserStore{}, nil), log)
		RegisterUser(rec, req)

		assertAPIError(t, rec, http.StatusBadRequest, "Error parsing request body")
	})

	t.Run("get user count database error", func(t *testing.T) {
		store := &mockUserStore{countErr: errors.New("db unavailable")}
		rec := httptest.NewRecorder()
		req := registerRequest(t, validRegisterJSON(), service.NewUserService(store, nil), log)
		RegisterUser(rec, req)

		assertAPIError(t, rec, http.StatusInternalServerError, "Error querying the database")
	})

	t.Run("hash password error", func(t *testing.T) {
		users := service.NewUserService(&mockUserStore{}, func(string) (string, error) {
			return "", errors.New("hash failed")
		})

		rec := httptest.NewRecorder()
		req := registerRequest(t, validRegisterJSON(), users, log)
		RegisterUser(rec, req)

		assertAPIError(t, rec, http.StatusInternalServerError, "Error hashing password")
	})

	t.Run("create user database error", func(t *testing.T) {
		store := &mockUserStore{createErr: errors.New("insert failed")}
		rec := httptest.NewRecorder()
		req := registerRequest(t, validRegisterJSON(), service.NewUserService(store, nil), log)
		RegisterUser(rec, req)

		assertAPIError(t, rec, http.StatusInternalServerError, "Error creating user")
	})
}

type registerBody struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
}

func testRegisterHandler(store service.UserStore) http.Handler {
	return testRegisterHandlerWithService(service.NewUserService(store, nil))
}

func testRegisterHandlerWithService(users *service.UserService) http.Handler {
	log := zap.NewNop().Sugar()
	router := mux.NewRouter()
	router.Use(middleware.LoggerContext(log))
	router.Use(UserServiceMiddleware(users))
	router.Use(middleware.EnsureJSONBody)
	router.HandleFunc("/register", RegisterUser).Methods("POST")
	return router
}

func registerRequest(t *testing.T, body []byte, users *service.UserService, log *zap.SugaredLogger) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/register", nil)
	ctx := middleware.WithBody(req.Context(), body)
	ctx = middleware.WithLogger(ctx, log)
	ctx = withUserService(ctx, users)
	return req.WithContext(ctx)
}

func validRegisterJSON() []byte {
	body, err := json.Marshal(RegisterUserBody{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "hash-error@example.com",
		Password:  "secure-password",
	})
	if err != nil {
		panic(err)
	}
	return body
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

func boolVal(v *bool) bool {
	return v != nil && *v
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}

func assertAPIError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
	t.Helper()
	assertStatus(t, rec, wantStatus)

	res := decodeAPIResponse(t, rec.Body)
	if res.Message != wantMessage {
		t.Fatalf("message = %q, want %q", res.Message, wantMessage)
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
