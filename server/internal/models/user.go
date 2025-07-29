package models

import (
	"crapp-go/internal/database"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Email    string
	Password string
}

func CreateUser(email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:    email,
		Password: string(hashedPassword),
	}

	query := "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id"
	err = database.DB.QueryRow(query, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := "SELECT id, email, password FROM users WHERE email = $1"
	err := database.DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
