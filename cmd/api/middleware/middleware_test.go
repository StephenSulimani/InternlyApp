package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func TestEnsureJSONBody(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("accepts valid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytesReader(`{"ok":true}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		EnsureJSONBody(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("rejects wrong content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytesReader(`{}`)))
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		EnsureJSONBody(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnsupportedMediaType)
		}
	})

	t.Run("rejects unreadable body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", errReadCloser{err: errors.New("read failed")})
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		EnsureJSONBody(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytesReader(`{`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		EnsureJSONBody(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestDatabaseMiddleware(t *testing.T) {
	var gotPool *pgxpool.Pool
	handler := DatabaseMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPool = r.Context().Value("db").(*pgxpool.Pool)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if gotPool != nil {
		t.Fatal("expected nil pool in context")
	}
}

func TestLoggerContext(t *testing.T) {
	log := zap.NewNop().Sugar()
	var gotLog *zap.SugaredLogger
	handler := LoggerContext(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLog = r.Context().Value("log").(*zap.SugaredLogger)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if gotLog != log {
		t.Fatal("expected logger in context")
	}
}

func TestLoggingMiddleware(t *testing.T) {
	log := zap.NewNop().Sugar()
	called := false
	handler := LoggingMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
}

type bytesReader string

func (b bytesReader) Read(p []byte) (int, error) {
	return copy(p, []byte(b)), io.EOF
}

type errReadCloser struct {
	err error
}

func (e errReadCloser) Read([]byte) (int, error) {
	return 0, e.err
}

func (e errReadCloser) Close() error {
	return nil
}
