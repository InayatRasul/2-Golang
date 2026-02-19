package app

import (
	"context"
	"fmt"
	"time"

	"golang/internal/repository/_postgres" // Replace with your actual internal path
	"golang/pkg/modules"
	"golang/internal/repository" // Replace with your actual internal path

)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Configuration
	dbConfig := initPostgreConfig()

	// 2. Connect to Database (Dialect)
	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)

	// 3. Initialize Repositories
	repositories := repository.NewRepositories(_postgre)

	// 4. Execute Business Logic
	users, err := repositories.GetUsers()
	if err != nil {
		fmt.Printf("Error fetching users: %v\n", err)
		return
	}

	fmt.Printf("Users: %+v\n", users)
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