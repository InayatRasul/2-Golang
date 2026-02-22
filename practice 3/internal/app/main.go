package app

import (
	"context"
	// "fmt"
	"time"
	"net/http"
	"log"


	"golang/internal/repository/_postgres" // Replace with your actual internal path
	"golang/pkg/modules"
	"golang/internal/repository" // Replace with your actual internal path
	"golang/internal/usecase"
	"golang/internal/handler"

)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1️⃣ Initialize config
	dbConfig := initPostgreConfig()

	// 2️⃣ Connect DB
	db := _postgres.NewPGXDialect(ctx, dbConfig)

	// 3️⃣ Initialize repository
	repositories := repository.NewRepositories(db)

	// 4️⃣ Initialize usecase
	userUsecase := usecase.NewUserUsecase(repositories.UserRepository)

	// 5️⃣ Initialize handler
	userHandler := handler.NewUserHandler(userUsecase)

	// 6️⃣ Register routes
	// http.HandleFunc("/users", userHandler.GetUsers)
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUserByID(w, r)
		case http.MethodPut:
			userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})


	// 7️⃣ Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}


func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        "localhost",
		Port:        "5432",
		Username:    "appuser",
		Password:    "appuser",
		DBName:      "golangdb",
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}