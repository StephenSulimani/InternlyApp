package middleware

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/stephensulimani/internlyapp/cmd/api/types"
)

func EnsureJSONBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))

		if err != nil || mediaType != "application/json" {
			http.Error(w, types.ErrorResponse("Expected application/json"), http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, types.ErrorResponse("Error reading request body"), http.StatusBadRequest)
			return
		}

		jsonBody := map[string]any{}

		err = json.Unmarshal(body, &jsonBody)
		if err != nil {
			http.Error(w, types.ErrorResponse("Error parsing request body"), http.StatusBadRequest)
			return
		}

		ctx := WithBody(r.Context(), body)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
