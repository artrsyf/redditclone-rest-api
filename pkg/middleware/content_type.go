package middleware

import (
	"net/http"
	"redditclone/tools"
)

func ValidateContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			tools.JSONError(w, http.StatusBadRequest, "unknown payload content type", "middleware.ValidateContentType")
			return
		}

		next.ServeHTTP(w, r)
	})
}
