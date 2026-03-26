package main

import (
	"log"
	"net/http"
	"task-api/internal/handlers"
	"task-api/internal/middleware"
)

func main(){
	
	mux := http.NewServeMux() 
	// Multiplexer is essentially a router. Its only job is to look at the URL of an incoming request (like /tasks) and decide which function should handle it.

	//register handler
	mux.HandleFunc("/tasks", handlers.TaskHandler)

	//apply middleware to everything
	wrappedMux := middleware.AuthMiddleware(mux)
	// Instead of giving the user direct access to the router, you "wrap" it inside your middleware.
	// It checks for the X-API-KEY before the request ever reaches your actual code.

	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", wrappedMux)
	// starts the server on port 8080.
	// we pass wrappedMux (the router with the security guard) instead of just mux. This ensures every request is checked for an API key and logged before it's processed.
}