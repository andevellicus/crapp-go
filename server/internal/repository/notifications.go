package repository

import (
	"crapp-go/internal/database"
	"crapp-go/internal/models"
	"time"
)

// GetUsersForEmailReminder finds users who have email reminders enabled for a specific time.
func GetUsersForEmailReminder(reminderTime string) ([]models.User, error) {
	var users []models.User
	err := database.DB.Where("email_notifications_enabled = ? AND reminder_time = ?", true, reminderTime).Find(&users).Error
	return users, err
}

// HasCompletedAssessmentToday checks if a user has completed an assessment on the current day.
func HasCompletedAssessmentToday(userID uint) (bool, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := database.DB.Model(&models.AssessmentState{}).
		Where("user_id = ? AND is_complete = ? AND updated_at >= ? AND updated_at < ?", userID, true, today, tomorrow).
		Count(&count).Error

	return count > 0, err
}

// UpdateNotificationPreferences updates a user's notification settings.
func UpdateNotificationPreferences(userID uint, enabled bool, reminderTime, timezone string) error {
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"email_notifications_enabled": enabled,
		"reminder_time":               reminderTime,
		"time_zone":                   timezone,
	}).Error
}
