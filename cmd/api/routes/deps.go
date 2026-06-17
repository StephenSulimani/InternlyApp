package routes

import (
	"net/http"

	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/internal/db"
	"go.uber.org/zap"
)

type requestDeps struct {
	body  []byte
	log   *zap.SugaredLogger
	users userStore
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

	if store, ok := userStoreFromContext(r.Context()); ok {
		return &requestDeps{
			body:  body,
			log:   log,
			users: store,
		}, true
	}

	pool, ok := middleware.DBFromContext(r.Context())
	if !ok {
		return nil, false
	}

	return &requestDeps{
		body:  body,
		log:   log,
		users: db.New(pool),
	}, true
}
