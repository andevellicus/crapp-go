package repository

import (
	"context"
	"crapp-go/internal/database"
	"crapp-go/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
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

func GetUserByEmail(c context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := "SELECT id, email, password FROM users WHERE email = $1"
	// Pass the context to the database call
	err := database.DB.QueryRowContext(c, query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user := &models.User{}
	query := "SELECT id, email, password FROM users WHERE id = $1"
	err := database.DB.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}
