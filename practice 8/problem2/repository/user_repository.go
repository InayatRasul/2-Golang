package repository

import "errors"

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	GetUserByID(id int) (*User, error)
	CreateUser(user *User) error
	GetByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error
}

// Mock implementation for testing (optional, but useful for reference)
type MemoryUserRepository struct {
	users map[int]*User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[int]*User),
	}
}

func (m *MemoryUserRepository) GetUserByID(id int) (*User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MemoryUserRepository) CreateUser(user *User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	m.users[user.ID] = user
	return nil
}

func (m *MemoryUserRepository) GetByEmail(email string) (*User, error) {
	for _, user := range m.users {
		if user.Name == email { // Using Name as email for simplicity
			return user, nil
		}
	}
	return nil, nil // Return nil, nil if not found (as per requirements)
}

func (m *MemoryUserRepository) UpdateUser(user *User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	m.users[user.ID] = user
	return nil
}

func (m *MemoryUserRepository) DeleteUser(id int) error {
	delete(m.users, id)
	return nil
}
