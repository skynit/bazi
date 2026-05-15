package model

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a registered user.
type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email        string `gorm:"type:varchar(128);not null" json:"email"`
	PasswordHash string `gorm:"type:varchar(256);not null" json:"-"`
}

// SetPassword hashes the given password and stores it.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword returns true if the given password matches the stored hash.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
