package models

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string         `gorm:"unique;uniqueIndex;not null"`
	Password string         `gorm:"not null"`
	Name     string         `gorm:"not null"`
	Roles    pq.StringArray `gorm:"type:text[];default:'{}'"`
}

func (user *User) HashPassword() error {
	user.Password = strings.TrimSpace(user.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = string(hashedPassword)
	return nil

}

func (user *User) VerifyPassword(password string) error {
	password = strings.TrimSpace(password)
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}
