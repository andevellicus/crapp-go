package repository

import (
	"context"
	"crapp-go/internal/database"
	"crapp-go/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(email, password, firstName, lastName string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
	}
	result := database.DB.Create(user)
	return user, result.Error
}

func GetUserByEmail(c context.Context, email string) (*models.User, error) {
	var user models.User
	result := database.DB.WithContext(c).First(&user, "email = ?", email)
	return &user, result.Error
}

func GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	result := database.DB.WithContext(ctx).First(&user, id)
	return &user, result.Error
}

func UpdateUser(ctx context.Context, userID uint, firstName, lastName string) error {
	return database.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{"first_name": firstName, "last_name": lastName}).Error
}

func UpdateUserPassword(ctx context.Context, userID uint, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return database.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password", string(hashedPassword)).Error
}

func DeleteUser(ctx context.Context, userID uint) error {
	return database.DB.WithContext(ctx).Delete(&models.User{}, userID).Error
}
