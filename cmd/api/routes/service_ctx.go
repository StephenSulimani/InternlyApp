package routes

import (
	"context"

	"github.com/stephensulimani/internlyapp/internal/service"
)

type serviceCtxKey int

const serviceCtxKeyUserService serviceCtxKey = 1

func withUserService(ctx context.Context, users *service.UserService) context.Context {
	return context.WithValue(ctx, serviceCtxKeyUserService, users)
}

func userServiceFromContext(ctx context.Context) (*service.UserService, bool) {
	users, ok := ctx.Value(serviceCtxKeyUserService).(*service.UserService)
	return users, ok
}
