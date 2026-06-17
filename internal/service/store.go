package service

import (
	"context"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type UserStore interface {
	GetUserCount(ctx context.Context) (int64, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}
