package routes

import (
	"net/http"

	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/internal/auth"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

type requestDeps struct {
	body   []byte
	log    *zap.SugaredLogger
	users  *service.UserService
	tokens *auth.TokenIssuer
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

	users, ok := userServiceFromContext(r.Context())
	if !ok {
		return nil, false
	}

	tokens, ok := tokenIssuerFromContext(r.Context())
	if !ok {
		return nil, false
	}

	return &requestDeps{
		body:   body,
		log:    log,
		users:  users,
		tokens: tokens,
	}, true
}
