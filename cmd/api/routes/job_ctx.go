package routes

import (
	"context"
	"net/http"

	"github.com/stephensulimani/internlyapp/internal/service"
)

type jobServiceCtxKey int

const jobServiceCtxKeyValue jobServiceCtxKey = 1

func JobServiceMiddleware(jobs *service.JobService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := withJobService(r.Context(), jobs)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withJobService(ctx context.Context, jobs *service.JobService) context.Context {
	return context.WithValue(ctx, jobServiceCtxKeyValue, jobs)
}

func jobServiceFromContext(ctx context.Context) (*service.JobService, bool) {
	jobs, ok := ctx.Value(jobServiceCtxKeyValue).(*service.JobService)
	return jobs, ok
}
