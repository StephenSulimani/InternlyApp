package routes

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/stephensulimani/internlyapp/cmd/api/types"
	"github.com/stephensulimani/internlyapp/internal/auth"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
)

type authCtxKey int

const (
	authCtxKeyClaims authCtxKey = 1
	authCtxKeyUser   authCtxKey = 2
)

// RequireAuth authenticates the Bearer JWT and authorizes active members.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens, ok := tokenIssuerFromContext(r.Context())
		if !ok {
			types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
			return
		}

		users, ok := userServiceFromContext(r.Context())
		if !ok {
			types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
			return
		}

		raw, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			types.WriteError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		claims, err := tokens.Parse(raw)
		if err != nil {
			types.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		user, err := users.GetActiveByEmail(r.Context(), claims.Email)
		if err != nil {
			writeAuthzError(w, err)
			return
		}

		ctx := withAuthClaims(r.Context(), claims)
		ctx = withAuthUser(ctx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if header == "" || !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

func withAuthClaims(ctx context.Context, claims auth.Claims) context.Context {
	return context.WithValue(ctx, authCtxKeyClaims, claims)
}

func AuthClaimsFromContext(ctx context.Context) (auth.Claims, bool) {
	claims, ok := ctx.Value(authCtxKeyClaims).(auth.Claims)
	return claims, ok
}

func withAuthUser(ctx context.Context, user db.User) context.Context {
	return context.WithValue(ctx, authCtxKeyUser, user)
}

func AuthUserFromContext(ctx context.Context) (db.User, bool) {
	user, ok := ctx.Value(authCtxKeyUser).(db.User)
	return user, ok
}

func writeAuthzError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserInactive):
		types.WriteError(w, http.StatusForbidden, "Account pending activation")
	case errors.Is(err, service.ErrInvalidEmailOrPassword):
		types.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
	case errors.Is(err, service.ErrGetUser):
		types.WriteError(w, http.StatusInternalServerError, "Error authorizing request")
	default:
		types.WriteError(w, http.StatusInternalServerError, "Error authorizing request")
	}
}
