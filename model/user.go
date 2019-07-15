package model

import (
	"taskboard-api-go/common"

	"golang.org/x/crypto/bcrypt"
)

// User is user of the app.
type User struct {
	ID           string `gorm:"primary_key;size:32"`
	Name         string `gorm:"not null;size:255;unique"`
	PasswordHash string `gorm:"not null;size:255"`
	Avatar       string `gorm:"size:255"`
	Version      int    `gorm:"not null"` // Version for optimistic lock
}

// NewUser returns created new user
func NewUser(name, rawpassword, avatar string) *User {
	result := &User{
		ID:      "user_" + common.GenerateID(),
		Name:    name,
		Avatar:  avatar,
		Version: 1,
	}
	result.SetPassword(rawpassword)
	return result
}

// SetPassword sets the hash of specified password to user
func (user *User) SetPassword(password string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
}

// VerifyPassword checks whether specified password matches PasswordHash in database.
func (user *User) VerifyPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}
