package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email                     string `gorm:"unique;not null"`
	Password                  string `gorm:"not null"`
	FirstName                 string
	LastName                  string
	EmailNotificationsEnabled bool   `gorm:"default:false"`
	ReminderTime              string `gorm:"type:varchar(5);default:'09:00'"` // Default to a common local time
	TimeZone                  string `gorm:"default:'UTC'"`                   // e.g., "America/New_York"
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
