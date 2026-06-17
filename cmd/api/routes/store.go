package routes

import (
	"context"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type userStore interface {
	GetUserCount(ctx context.Context) (int64, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}

type storeCtxKey int

const storeCtxKeyUserStore storeCtxKey = 1

func withUserStore(ctx context.Context, store userStore) context.Context {
	return context.WithValue(ctx, storeCtxKeyUserStore, store)
}

func userStoreFromContext(ctx context.Context) (userStore, bool) {
	store, ok := ctx.Value(storeCtxKeyUserStore).(userStore)
	return store, ok
}
