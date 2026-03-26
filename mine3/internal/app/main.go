package app

import(
	"context" //for dealing with path sensitivity: golang is my module but there exist std golang, so problem
	"fmt"
	"log"
	"net/http"
	"mine3/pkg/modules"
	"mine3/internal/repository/_postgres"
	// "golang/internal/repository/_postgres"
)

func Run(){
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := initPostgreConfig()

	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)

	fmt.Println(_postgre) //printing _postgre variable which is *Dialect pointer 
}

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host: "localhost",
		Port: "5432",
		Username: "appuser",
		Password: "appuser",
		DBName: "golangdb",
		SSLMode: "disable",
		ExecTimeout: 5 * time.Second,
	}
}