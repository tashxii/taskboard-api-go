package service

import (
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"taskboard-api-go/repository"

	"github.com/jinzhu/gorm"
)

// UserService provides apis for user management.
type UserService struct {
	tx       *gorm.DB
	userRepo *repository.UserRepository
}

// NewUserService return new instance of UserService.
func NewUserService(tx *gorm.DB) *UserService {
	return &UserService{
		tx:       tx,
		userRepo: repository.NewUserRepository(tx),
	}
}

// FindUser returns user matching specified condition
func (s *UserService) FindUser(condition interface{}) (*model.User, error) {
	find, err := s.userRepo.FindFirstUser(condition, []string{"id"})
	if err != nil {
		if err == orm.ErrorRecordNotFound {
			return nil, NewSvcErrorf(ErrorCodeNotFound, err, "User not found")
		}
		return nil, NewSvcError(ErrorCodeDB, err, "Failed to find user")
	}
	return &find, nil
}

// FindUsers finds all users
func (s *UserService) FindUsers(condition interface{}, sortOrders []string) ([]model.User, error) {
	users, err := s.userRepo.FindUsers(condition, 0, orm.NoLimit, sortOrders)
	if err != nil {
		return nil, NewSvcError(ErrorCodeDB, err, "Failed to find users")
	}
	return users, nil
}

// CreateUser creates new user
func (s *UserService) CreateUser(user *model.User) error {
	err := s.userRepo.CreateUser(user)
	if err != nil {
		return NewSvcError(ErrorCodeDB, err, "Failed to create user")
	}
	return nil
}

// UpdateUser updates specifed user
func (s *UserService) UpdateUser(user *model.User) error {
	err := s.userRepo.UpdateUser(user)
	if err != nil {
		return NewSvcErrorf(ErrorCodeDB, err, "Failed to update user. ID:%s", user.ID)
	}
	return nil
}

// DeleteUser deletes specifed user
func (s *UserService) DeleteUser(user *model.User) error {
	err := s.userRepo.DeleteUser(user)
	if err != nil {
		return NewSvcErrorf(ErrorCodeDB, err, "Failed to delete user. ID:%s", user.ID)
	}
	return nil
}

// Login returns valid user or nil
func (s *UserService) Login(name, password string) (*model.User, error) {
	find, err := s.userRepo.FindFirstUser(&model.User{Name: name}, []string{})
	if err != nil {
		// Does not describe details
		return nil, NewSvcError(ErrorCodeUnauthenticated, err, "Login failed")
	}
	if err := find.VerifyPassword(password); err != nil {
		// Does not describe details
		return nil, NewSvcError(ErrorCodeUnauthenticated, err, "Login failed")
	}
	return &find, nil
}
