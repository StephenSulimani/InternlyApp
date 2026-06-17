package middleware

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/stephensulimani/internlyapp/cmd/api/types"
)

const maxBodyBytes = 1 << 20 // 1 MiB

func EnsureJSONBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			types.WriteError(w, http.StatusUnsupportedMediaType, "Expected application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				types.WriteError(w, http.StatusRequestEntityTooLarge, "Request body too large")
				return
			}
			types.WriteError(w, http.StatusBadRequest, "Error reading request body")
			return
		}

		ctx := WithBody(r.Context(), body)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
