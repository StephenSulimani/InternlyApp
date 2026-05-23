package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func LoggingMiddleware(log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Infof("%s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}
