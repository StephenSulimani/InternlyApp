package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephensulimani/internlyapp/internal/testutil"
	"go.uber.org/zap"
)

func TestRegisterUser_errors(t *testing.T) {
	pool, cleanup := testutil.SetupPostgres(t)
	defer cleanup()

	log := zap.NewNop().Sugar()

	t.Run("handler json unmarshal error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := registerRequest(t, []byte(`{"first_name":123,"last_name":"Lovelace","email":"a@example.com","password":"pw"}`), pool, log)
		RegisterUser(rec, req)

		assertRegisterError(t, rec, http.StatusBadRequest, "Error parsing request body")
	})

	t.Run("get user count database error", func(t *testing.T) {
		testutil.TruncateUsers(t, pool)

		_, err := pool.Exec(context.Background(), `ALTER TABLE users RENAME TO users_count_test`)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `ALTER TABLE users_count_test RENAME TO users`)
		})

		rec := httptest.NewRecorder()
		req := registerRequest(t, validRegisterJSON(), pool, log)
		RegisterUser(rec, req)

		assertRegisterError(t, rec, http.StatusInternalServerError, "Error querying the database")
	})

	t.Run("hash password error", func(t *testing.T) {
		testutil.TruncateUsers(t, pool)

		original := hashPassword
		hashPassword = func(string) (string, error) {
			return "", errors.New("hash failed")
		}
		t.Cleanup(func() { hashPassword = original })

		rec := httptest.NewRecorder()
		req := registerRequest(t, validRegisterJSON(), pool, log)
		RegisterUser(rec, req)

		assertRegisterError(t, rec, http.StatusInternalServerError, "Error hashing password")
	})

	t.Run("create user database error", func(t *testing.T) {
		testutil.TruncateUsers(t, pool)
		testutil.InstallRejectUserInsertTrigger(t, pool)
		t.Cleanup(func() { testutil.RemoveRejectUserInsertTrigger(t, pool) })

		rec := httptest.NewRecorder()
		req := registerRequest(t, validRegisterJSON(), pool, log)
		RegisterUser(rec, req)

		assertRegisterError(t, rec, http.StatusInternalServerError, "Error creating user")
	})
}

func registerRequest(t *testing.T, body []byte, pool *pgxpool.Pool, log *zap.SugaredLogger) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/register", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, "body", body)
	ctx = context.WithValue(ctx, "log", log)
	ctx = context.WithValue(ctx, "db", pool)
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

type errorResponse struct {
	Success int    `json:"success"`
	Message string `json:"message"`
}

func assertRegisterError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, wantStatus, rec.Body.String())
	}

	var res errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res.Message != wantMessage {
		t.Fatalf("message = %q, want %q", res.Message, wantMessage)
	}
}
