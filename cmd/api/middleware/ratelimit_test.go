package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimit(t *testing.T) {
	handler := RateLimit(rate.Every(time.Hour), 1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", rec.Code, http.StatusOK)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestClientIP(t *testing.T) {
	t.Run("uses first x-forwarded-for address", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", " 1.2.3.4 , 5.6.7.8")
		if got := clientIP(req); got != "1.2.3.4" {
			t.Fatalf("clientIP = %q, want %q", got, "1.2.3.4")
		}
	})

	t.Run("falls back to remote addr", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:4321"
		if got := clientIP(req); got != "9.9.9.9" {
			t.Fatalf("clientIP = %q, want %q", got, "9.9.9.9")
		}
	})
}
