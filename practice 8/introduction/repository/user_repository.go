package repository

type User struct {
	ID int
	Name string
}

type UserRepository interface {
	GetUserByID(id int) (*User, error)
	CreateUser(user *User) error
}