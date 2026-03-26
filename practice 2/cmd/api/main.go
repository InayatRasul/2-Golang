package main

import (
	"log"
	"net/http"
	"task-api/internal/handlers"
	"task-api/internal/middleware"
)

func main(){
	mux := http.NewServeMux()

	//register handler
	mux.HandleFunc("/tasks", handlers.TaskHandler)

	//apply middleware to everything
	wrappedMux := middleware.AuthMiddleware(mux)

	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", wrappedMux)
}