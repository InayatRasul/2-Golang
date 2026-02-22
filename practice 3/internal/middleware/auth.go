package middleware

import (
	"net/http"
)

// Auth is a simple authentication middleware that checks for the presence of
// an Authorization header with a fixed value. In a real application you would
// verify a token, session or call out to an auth service.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
