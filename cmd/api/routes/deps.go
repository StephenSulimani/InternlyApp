package routes

import (
	"net/http"

	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

type requestDeps struct {
	body  []byte
	log   *zap.SugaredLogger
	users *service.UserService
}

func depsFromRequest(w http.ResponseWriter, r *http.Request) (*requestDeps, bool) {
	body, ok := middleware.BodyFromContext(r.Context())
	if !ok {
		return nil, false
	}

	log, ok := middleware.LoggerFromContext(r.Context())
	if !ok {
		return nil, false
	}

	if users, ok := userServiceFromContext(r.Context()); ok {
		return &requestDeps{
			body:  body,
			log:   log,
			users: users,
		}, true
	}

	pool, ok := middleware.DBFromContext(r.Context())
	if !ok {
		return nil, false
	}

	return &requestDeps{
		body:  body,
		log:   log,
		users: service.NewUserService(db.New(pool), nil),
	}, true
}
