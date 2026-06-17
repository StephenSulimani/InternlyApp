package middleware

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func DatabaseMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithDB(r.Context(), pool)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func LoggerContext(log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithLogger(r.Context(), log)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
