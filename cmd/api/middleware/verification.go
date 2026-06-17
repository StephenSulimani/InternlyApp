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
		w.Header().Set("Content-Type", "application/json")

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			http.Error(w, types.ErrorResponse("Expected application/json"), http.StatusUnsupportedMediaType)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(w, types.ErrorResponse("Request body too large"), http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, types.ErrorResponse("Error reading request body"), http.StatusBadRequest)
			return
		}

		ctx := WithBody(r.Context(), body)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
