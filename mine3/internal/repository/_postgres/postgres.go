package _postgres

import (
	"context"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"mine3/pkg/modules"
)

type Dialect struct {
	DB *sqlx.DB
}
// Dialect: This is the Name of the box.
// DB: This is a Variable (or field) inside the box.
// *sqlx.DB: This is the Value stored in that variable.
//* (the pointer), it doesn't store the "actual" database inside the box. It stores the memory address (the link) to where the database connection is living.

func NewPGXDialect(ctx context.Context, cfg *modules.PostgreConfig) *Dialect{
	// return type Dialect its address
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
	cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode) // Short Variable Declaration."Create a new variable named dsn and figure out its type based on what I’m putting into it."

	db, err := sqlx.Connect("postgres", dsn)
	
	if err != nil{
		panic(err)
	}

	err = df.Ping()

	if err != nil{
		panic(err)
	}

	AutoMigrate(cfg)

	return Dialect{
		DB: db,
	}
}

func AutoMigrate(cfg *modules.PostgreConfig){
	sourceURL := "file://database/migrations"
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", 
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	m. err := migrate.New(sourceURL, databaseURL)

	if err != nil{
		panic(err)
	}

	err = m.Up()

	if err != nil && err != migrate.ErrNoChange{
		panic(err)
	}
	
}