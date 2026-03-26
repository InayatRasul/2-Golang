package users

import (
	"fmt"
	"time"
	"database/sql"

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

// CRUD

// CreateUser 
func (r *Repository) CreateUser(user *modules.User) (int, error) {
	query := `
	INSERT INTO users (name, email, age)
	VALUES ($1, $2, $3)
	RETURNING id;
	`

	var id int
	err := r.db.DB.QueryRow(
		query,
		user.Name,
		user.Email,
		user.Age,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	return id, nil
}
// GetAll

func (r *Repository) GetUsers() ([]modules.User, error) {
	var users []modules.User

	err := r.db.DB.Select(&users, `
		SELECT id, name, email, age, created_at
		FROM users
	`)

	if err != nil {
		return nil, err
	}

	return users, nil
}

// Get by ID

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var user modules.User

	err := r.db.DB.Get(&user, `
		SELECT id, name, email, age, created_at
		FROM users WHERE id=$1
	`, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, err
	}

	return &user, nil
}

//Update
func (r *Repository) UpdateUser(user *modules.User) error {
	result, err := r.db.DB.Exec(`
		UPDATE users
		SET name=$1, email=$2, age=$3
		WHERE id=$4
	`,
		user.Name,
		user.Email,
		user.Age,
		user.ID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("user with id %d does not exist", user.ID)
	}

	return nil
}

// Delete
func (r *Repository) DeleteUser(id int) (int64, error) {
	result, err := r.db.DB.Exec(`
		DELETE FROM users WHERE id=$1
	`, id)

	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rows == 0 {
		return 0, fmt.Errorf("user not found")
	}

	return rows, nil
}
