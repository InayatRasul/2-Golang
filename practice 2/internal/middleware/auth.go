package middleware

import (
	"log"
	"net/http"
	"time"
)

//Handler is just a piece of code that responds to an HTTP request.
func AuthMiddleware(next http.Handler) http.Handler {
	// This function takes your actual task logic (next) and wraps it inside a new function.
	// It returns a "protected" version of your handler.

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		
		// Log: timestamp, method. path
		log.Printf("%s %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		// It writes down the current time, the HTTP method (GET, POST, etc.), and the URL path (like /tasks) to your console.
		
		//check API Key
		if r.Header.Get("X-API-KEY") != "secret12345"{
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)

		// If the API key is correct, the guard steps aside.
		// This line tells Go to finally run the actual handler code (like listing or creating tasks) that you defined in task.go.
	})
}