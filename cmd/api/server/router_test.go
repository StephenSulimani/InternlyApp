package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stephensulimani/internlyapp/internal/testutil"
	"go.uber.org/zap"
)

func TestNewHandler(t *testing.T) {
	pool, cleanup := testutil.SetupPostgres(t)
	defer cleanup()

	handler := NewHandler(zap.NewNop().Sugar(), pool)

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
