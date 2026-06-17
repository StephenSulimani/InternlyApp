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
	t.Run("accepts valid json and stores body in context", func(t *testing.T) {
		body := []byte(`{"ok":true}`)
		var gotBody []byte
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ok bool
			gotBody, ok = BodyFromContext(r.Context())
			if !ok {
				t.Fatal("expected body in context")
			}
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytesReader(body)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		EnsureJSONBody(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if string(gotBody) != string(body) {
			t.Fatalf("body = %q, want %q", gotBody, body)
		}
	})

	t.Run("rejects wrong content type", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytesReader(`{}`)))
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		EnsureJSONBody(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnsupportedMediaType)
		}
	})

	t.Run("rejects unreadable body", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/", errReadCloser{err: errors.New("read failed")})
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		EnsureJSONBody(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

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
	var gotOK bool
	handler := DatabaseMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPool, gotOK = DBFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if !gotOK {
		t.Fatal("expected db in context")
	}
	if gotPool != nil {
		t.Fatal("expected nil pool in context")
	}
}

func TestLoggerContext(t *testing.T) {
	log := zap.NewNop().Sugar()
	var gotLog *zap.SugaredLogger
	var gotOK bool
	handler := LoggerContext(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLog, gotOK = LoggerFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if !gotOK {
		t.Fatal("expected logger in context")
	}
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
