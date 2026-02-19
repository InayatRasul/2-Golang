package users

import (
	"fmt"
	"time"

	"golang/internal/repository/_postgres"
	"golang/pkg/modules"
)

type Repository struct {
	db               *_postgres.Dialect
	executionTimeout time.Duration
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:               db,
		executionTimeout: time.Second * 5,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	var users []modules.User

	// Using sqlx.Select to scan multiple rows into the slice
	err := r.db.DB.Select(&users, "SELECT id, name FROM users")
	if err != nil {
		return nil, err
	}

	fmt.Println(users)
	return users, nil
}