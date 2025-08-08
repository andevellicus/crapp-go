// server/internal/handlers/user.go
package handlers

import (
	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"crapp-go/views"
	"crapp-go/views/components"
	"crapp-go/views/profile"
	"net/http"

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

	csrfToken, _ := c.Get("csrf_token")
	cspNonce, _ := c.Get("csp_nonce")
	isHTMX := c.GetHeader("HX-Request") == "true"

	activeSection := c.Param("section")
	if activeSection == "" {
		activeSection = "personal"
	}

	// This is the key: we check which element the HTMX request is targeting.
	// The profile nav targets "#profile-content", so the header value will be "profile-content".
	hxTarget := c.GetHeader("HX-Target")

	// Case 1: An HTMX request to swap ONLY the inner content.
	// This happens when clicking links inside the profile nav menu.
	if isHTMX && hxTarget == "profile-content" {
		// Instead of rendering just one part, render the multi-swap component.
		profile.ProfileUpdate(user.(*models.User), csrfToken.(string), activeSection).Render(c.Request.Context(), c.Writer)
		return
	}

	// Case 2: A full page load OR the initial HTMX request from the main navigation.
	// Both of these need to render the entire profile component shell.
	profileComponent := views.Profile(user.(*models.User), csrfToken.(string), activeSection)

	if isHTMX {
		// This is the initial HTMX request from the main nav (which targets #content)
		profileComponent.Render(c.Request.Context(), c.Writer)
	} else {
		// This is a full browser page load or refresh
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
		h.log.Error("Failed to update user info", zap.Error(err), zap.Int("userID", int(userID)))
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
		h.log.Error("Failed to update password", zap.Error(err), zap.Int("userID", int(currentUser.ID)))
		components.Alert("Failed to update password", "error").Render(c, c.Writer)
		return
	}
	components.Alert("Password updated successfully", "success").Render(c, c.Writer)
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
		h.log.Error("Failed to delete account", zap.Error(err), zap.Int("userID", int(currentUser.ID)))
		components.Alert("Failed to delete account.", "error").Render(c, c.Writer)
		return
	}

	c.Header("HX-Redirect", "/")
}
