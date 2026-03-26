package middleware

import (
	"log"
	"net/http"
	"time"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		// Log: timestamp, method. path
		log.Printf("%s %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		
		//check API Key
		if r.Header.Get("X-API-KEY") != "secret12345"{
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)

	})
}