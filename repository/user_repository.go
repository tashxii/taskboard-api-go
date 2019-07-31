package repository

import (
	"sync"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"

	"github.com/jinzhu/gorm"
)

var lockUser = &sync.Mutex{}

// UserRepository is repository of user table
type UserRepository struct {
	tx *gorm.DB
}

// NewUserRepository returns new instance of UserRepository
func NewUserRepository(tx *gorm.DB) *UserRepository {
	if tx == nil {
		// Programing error!!
		panic("tx must be set")
	}
	return &UserRepository{
		tx: tx,
	}
}

// FindFirstUser returns first User matching with specified condition
func (repo *UserRepository) FindFirstUser(condition interface{}, sortOrders []string) (result model.User, err error) {
	query := repo.tx.Where(condition)
	if sortOrders == nil {
		sortOrders = []string{}
	}

	for _, sortOrder := range sortOrders {
		query = query.Order(sortOrder)
	}
	err = query.First(&result).Error
	return
}

// FindUsers returns Users matching with specified condition
func (repo *UserRepository) FindUsers(condition interface{}, offset int, limit int, sortOrders []string) (result []model.User, err error) {
	query := repo.tx.Where(condition)
	if offset >= 0 {
		query = query.Offset(offset)
	}
	if limit >= 0 {
		query = query.Limit(limit)
	}

	if sortOrders == nil {
		sortOrders = []string{}
	}
	for _, user := range sortOrders {
		query = query.Order(user)
	}

	err = query.Find(&result).Error
	return
}

// CountUsers returns the number of Users matching specfied condition
func (repo *UserRepository) CountUsers(condition interface{}) (count int, err error) {
	var users []model.User
	err = repo.tx.Where(condition).Find(&users).Count(&count).Error
	return
}

// CreateUser inserts new User record
func (repo *UserRepository) CreateUser(user *model.User) error {
	return repo.CreateUsers([]*model.User{user})
}

// UpdateUser updates User record
func (repo *UserRepository) UpdateUser(user *model.User) error {
	return repo.UpdateUsers([]*model.User{user})
}

// DeleteUser deletes User record
func (repo *UserRepository) DeleteUser(user *model.User) error {
	return repo.DeleteUsers([]*model.User{user})
}

// CreateUsers inserts new User records.
func (repo *UserRepository) CreateUsers(users []*model.User) (err error) {
	for _, user := range users {
		err = repo.tx.Create(user).Error
		if err != nil {
			return
		}
	}
	return
}

// UpdateUsers updates user records
func (repo *UserRepository) UpdateUsers(users []*model.User) (err error) {
	lockUser.Lock()
	defer lockUser.Unlock()

	for _, user := range users {
		oldVersion := user.Version
		user.Version++
		db := repo.tx.Model(&model.User{}).Where("version = ?", oldVersion).Save(user)
		count := db.RowsAffected
		err = db.Error
		// return ErrorRecordNotFoud as optimistic lock error
		if err == nil && count == 0 {
			return orm.ErrorRecordNotFound
		}
		if err != nil {
			return
		}
	}
	return
}

// DeleteUsers deletes User records
func (repo *UserRepository) DeleteUsers(users []*model.User) (err error) {
	for _, user := range users {
		if user.ID == "" {
			continue // To avoid deleting all due to gorm warning, continue here.
		}
		err = repo.tx.Delete(user).Error
		if err != nil {
			return
		}
	}
	return
}
