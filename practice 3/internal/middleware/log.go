package middleware

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code written
// by the handler. We default to 200 in case the handler never explicitly
// writes a header.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging is a middleware that logs every request and response. The log entry
// includes a timestamp, the HTTP method, the URL path and the response status.
// It uses the standard library's log package as required by the assignment.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, r)

		ts := start.Format(time.RFC3339)
		log.Printf("%s %s %s %d %v", ts, r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
