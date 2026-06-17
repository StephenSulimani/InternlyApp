package routes

import (
	"context"
	"net/http"

	"github.com/stephensulimani/internlyapp/internal/service"
)

type serviceCtxKey int

const serviceCtxKeyUserService serviceCtxKey = 1

func UserServiceMiddleware(users *service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := withUserService(r.Context(), users)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withUserService(ctx context.Context, users *service.UserService) context.Context {
	return context.WithValue(ctx, serviceCtxKeyUserService, users)
}

func userServiceFromContext(ctx context.Context) (*service.UserService, bool) {
	users, ok := ctx.Value(serviceCtxKeyUserService).(*service.UserService)
	return users, ok
}
