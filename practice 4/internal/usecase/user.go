package usecase

import(
	"golang/internal/repository"
	"golang/pkg/modules"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(r repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: r}
}

func (u *UserUsecase) CreateUser(user *modules.User) (int, error) {
	return u.repo.CreateUser(user)
}

func (u *UserUsecase) GetUsers() ([]modules.User, error) {
	return u.repo.GetUsers()
}

func (u *UserUsecase) GetUserByID(id int) (*modules.User, error) {
	return u.repo.GetUserByID(id)
}

func (u *UserUsecase) UpdateUser(user *modules.User) error {
	return u.repo.UpdateUser(user)
}

func (u *UserUsecase) DeleteUser(id int) (int64, error) {
	return u.repo.DeleteUser(id)
}
