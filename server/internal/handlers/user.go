package handlers

import (
	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"crapp-go/views"
	"crapp-go/views/components"
	"crapp-go/views/profile"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	log *zap.Logger
}

func NewUserHandler(log *zap.Logger) *UserHandler {
	return &UserHandler{log: log}
}

// ShowProfilePage is the single handler for all GET requests to the profile.
func (h *UserHandler) ShowProfilePage(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userModel := user.(*models.User)
	csrfToken, _ := c.Get("csrf_token")
	cspNonce, _ := c.Get("csp_nonce")
	isHTMX := c.GetHeader("HX-Request") == "true"
	activeSection := c.Param("section")
	if activeSection == "" {
		activeSection = "personal"
	}

	userForView := *userModel
	if userModel.TimeZone != "" && userModel.ReminderTime != "" {
		loc, err := time.LoadLocation(userModel.TimeZone)
		if err == nil {
			utcTime, err := time.Parse("15:04", userModel.ReminderTime)
			if err == nil {
				now := time.Now().UTC()
				userReminderInUTC := time.Date(now.Year(), now.Month(), now.Day(), utcTime.Hour(), utcTime.Minute(), 0, 0, time.UTC)
				localReminderTime := userReminderInUTC.In(loc)
				userForView.ReminderTime = localReminderTime.Format("15:04")
			}
		}
	}

	hxTarget := c.GetHeader("HX-Target")
	if isHTMX && hxTarget == "profile-content" {
		profile.ProfileUpdate(&userForView, csrfToken.(string), activeSection).Render(c.Request.Context(), c.Writer)
		return
	}

	profileComponent := views.Profile(&userForView, csrfToken.(string), activeSection)
	if isHTMX {
		profileComponent.Render(c.Request.Context(), c.Writer)
	} else {
		views.Layout("Profile", true, csrfToken.(string), cspNonce.(string)).Render(
			templ.WithChildren(c.Request.Context(), profileComponent),
			c.Writer,
		)
	}
}

func (h *UserHandler) UpdateInfo(c *gin.Context) {
	user, _ := c.Get("user")
	userID := user.(*models.User).ID
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	if err := repository.UpdateUser(c, userID, firstName, lastName); err != nil {
		h.log.Error("Failed to update user info", zap.Error(err), zap.Uint("userID", userID))
		components.Alert("Failed to update profile", "error").Render(c, c.Writer)
		return
	}
	components.Alert("Profile updated successfully!", "success").Render(c, c.Writer)
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(*models.User)
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	if !currentUser.CheckPassword(currentPassword) {
		components.Alert("Incorrect current password", "error").Render(c, c.Writer)
		return
	}
	if newPassword != confirmPassword {
		components.Alert("New passwords do not match", "error").Render(c, c.Writer)
		return
	}
	if err := repository.UpdateUserPassword(c, currentUser.ID, newPassword); err != nil {
		h.log.Error("Failed to update password", zap.Error(err), zap.Uint("userID", currentUser.ID))
		components.Alert("Failed to update password", "error").Render(c, c.Writer)
		return
	}
	components.Alert("Password updated successfully", "success").Render(c, c.Writer)
}

func (h *UserHandler) UpdateNotificationSettings(c *gin.Context) {
	user, _ := c.Get("user")
	userID := user.(*models.User).ID

	enabled := c.PostForm("enable_email_notifications") == "on"
	localReminderTime := c.PostForm("reminder_time")
	userTimezone := c.PostForm("timezone")
	if userTimezone == "" {
		userTimezone = "UTC"
	}

	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		h.log.Error("Invalid timezone identifier", zap.Error(err), zap.String("timezone", userTimezone))
		components.Alert("Invalid timezone provided by your browser.", "error").Render(c, c.Writer)
		return
	}

	// Combine the current date with the user's selected time.
	// This provides the context needed to correctly determine if DST is active.
	now := time.Now()
	dateTimeString := fmt.Sprintf("%s %s", now.Format("2006-01-02"), localReminderTime)

	// Parse the full date and time string in the user's local timezone.
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04", dateTimeString, loc)
	if err != nil {
		components.Alert("Invalid time format. Please use HH:MM.", "error").Render(c, c.Writer)
		return
	}

	// Convert to UTC and format for storage.
	utcReminderTime := parsedTime.UTC().Format("15:04")
	if err := repository.UpdateNotificationPreferences(userID, enabled, utcReminderTime, userTimezone); err != nil {
		h.log.Error("Failed to update notification preferences", zap.Error(err), zap.Uint("userID", userID))
		components.Alert("Failed to save notification settings.", "error").Render(c, c.Writer)
		return
	}
	components.Alert("Notification settings saved successfully!", "success").Render(c, c.Writer)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(*models.User)
	password := c.PostForm("password")
	confirmation := c.PostForm("confirmation")
	if confirmation != "DELETE" {
		components.Alert("Please type DELETE to confirm.", "error").Render(c, c.Writer)
		return
	}
	if !currentUser.CheckPassword(password) {
		components.Alert("Incorrect password.", "error").Render(c, c.Writer)
		return
	}
	if err := repository.DeleteUser(c, currentUser.ID); err != nil {
		h.log.Error("Failed to delete account", zap.Error(err), zap.Uint("userID", currentUser.ID))
		components.Alert("Failed to delete account.", "error").Render(c, c.Writer)
		return
	}
	c.Header("HX-Redirect", "/")
}
