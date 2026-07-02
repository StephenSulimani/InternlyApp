package routes

import (
	"context"
	"net/http"

	"github.com/stephensulimani/internlyapp/internal/auth"
)

type tokenCtxKey int

const tokenCtxKeyIssuer tokenCtxKey = 1

func TokenIssuerMiddleware(tokens *auth.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := withTokenIssuer(r.Context(), tokens)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withTokenIssuer(ctx context.Context, tokens *auth.TokenIssuer) context.Context {
	return context.WithValue(ctx, tokenCtxKeyIssuer, tokens)
}

func tokenIssuerFromContext(ctx context.Context) (*auth.TokenIssuer, bool) {
	tokens, ok := ctx.Value(tokenCtxKeyIssuer).(*auth.TokenIssuer)
	return tokens, ok
}
