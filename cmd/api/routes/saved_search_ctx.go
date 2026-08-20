package routes

import (
	"context"
	"net/http"

	"github.com/stephensulimani/internlyapp/internal/service"
)

type savedSearchServiceCtxKey int

const savedSearchServiceCtxKeyValue savedSearchServiceCtxKey = 1

func SavedSearchServiceMiddleware(searches *service.SavedSearchService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := withSavedSearchService(r.Context(), searches)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withSavedSearchService(ctx context.Context, searches *service.SavedSearchService) context.Context {
	return context.WithValue(ctx, savedSearchServiceCtxKeyValue, searches)
}

func savedSearchServiceFromContext(ctx context.Context) (*service.SavedSearchService, bool) {
	searches, ok := ctx.Value(savedSearchServiceCtxKeyValue).(*service.SavedSearchService)
	return searches, ok
}
