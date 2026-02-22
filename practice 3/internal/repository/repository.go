package repository

import (
	"golang/internal/repository/_postgres"
	"golang/pkg/modules"
	"golang/internal/repository/_postgres/users" // Assuming your User repository is in a sub-package
)

type UserRepository interface {
	GetUsers() ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(user *modules.User) (int, error)
	UpdateUser(user *modules.User) error
	DeleteUser(id int) (int64, error)
}

type Repositories struct {
	UserRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		UserRepository: users.NewUserRepository(db),
	}
}