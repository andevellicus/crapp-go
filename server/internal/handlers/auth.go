package handlers

import (
	"net/http"
	"strings"

	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"crapp-go/internal/utils"
	"crapp-go/views"
	"crapp-go/views/components"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthHandler struct {
	log        *zap.Logger
	Assessment *models.Assessment
}

func NewAuthHandler(log *zap.Logger, assessment *models.Assessment) *AuthHandler {
	return &AuthHandler{log: log, Assessment: assessment}
}

func (h *AuthHandler) ShowLoginPage(c *gin.Context) {
	// Retrieve the token from the context set by our middleware.
	// The key "csrf_token" must match the one used in CSRFProtection middleware.
	csrfToken, exists := c.Get("csrf_token")
	if !exists {
		// This should not happen if the middleware is set up correctly.
		// Abort or handle the error appropriately.
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Pass the CSRF token to the template.
	views.Login(csrfToken.(string)).Render(c, c.Writer)
}

func (h *AuthHandler) Login(c *gin.Context) {
	session := sessions.Default(c)
	email := c.PostForm("email")
	password := c.PostForm("password")

	user, err := repository.GetUserByEmail(c, email)
	if err != nil || !user.CheckPassword(password) {
		h.log.Warn("Invalid login attempt", zap.String("email", email))
		// Set status
		c.Status(http.StatusUnauthorized)
		// Render the form with error
		components.Alert("Invalid email or password.", "error").Render(c, c.Writer)
		return
	}

	// Best practice: Regenerate tokens upon a privilege change (e.g., login).
	newToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		h.log.Error("Failed to generate new CSRF token on login", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		components.Alert("Internal server error.", "error").Render(c, c.Writer)
		return
	}
	session.Set("csrf_token", newToken)

	session.Set("userID", int(user.ID))
	if err := session.Save(); err != nil {
		h.log.Error("Failed to save session", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		components.Alert("Internal server error.", "error").Render(c, c.Writer)
		return
	}

	c.Header("HX-Redirect", "/assessment") // Redirect user to the assessment page on successful login
}

func (h *AuthHandler) ShowRegisterPage(c *gin.Context) {
	csrfToken, exists := c.Get("csrf_token")
	if !exists {
		// This should not happen if the middleware is set up correctly.
		// Abort or handle the error appropriately.
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Pass the CSRF token to the template.
	err := views.Register(csrfToken.(string)).Render(c, c.Writer)
	if err != nil {
		h.log.Error("Error rendering register page", zap.Error(err))
		c.String(http.StatusInternalServerError, "Error rendering page")
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirmPassword")
	firstName := strings.TrimSpace(c.PostForm("first_name"))
	lastName := strings.TrimSpace(c.PostForm("last_name"))
	timezone := c.PostForm("timezone") // Get timezone from form

	// 1. Check for empty fields
	if email == "" || password == "" || firstName == "" || lastName == "" {
		c.Status(http.StatusBadRequest)
		components.Alert("All fields are required.", "error").Render(c, c.Writer)
		return
	}

	// 2. Validate Email Format
	if !utils.IsValidEmail(email) {
		c.Status(http.StatusBadRequest)
		components.Alert("Please enter a valid email address.", "error").Render(c, c.Writer)
		return
	}

	// 3. Validate Password Complexity
	if !utils.IsComplexPassword(password) {
		c.Status(http.StatusBadRequest)
		// This error will target the #password-error-container
		components.Alert("Password does not meet complexity requirements.", "error").Render(c, c.Writer)
		return
	}

	// 4. Check if passwords match
	if password != confirmPassword {
		c.Status(http.StatusBadRequest)
		components.Alert("Passwords do not match.", "error").Render(c, c.Writer)
		return
	}

	// 5. Check if email already exists
	_, err := repository.GetUserByEmail(c, email)
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			h.log.Warn("Registration attempt with existing email", zap.String("email", email))
			c.Status(http.StatusBadRequest)
			components.Alert("A user with this email address already exists.", "error").Render(c, c.Writer)
			return
		}
		h.log.Error("Database error during registration check", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		components.Alert("Internal server error.", "error").Render(c, c.Writer)
		return
	}

	if _, err := repository.CreateUser(email, password, firstName, lastName, timezone); err != nil {
		h.log.Error("Failed to create user", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		components.Alert("Internal server error.", "error").Render(c, c.Writer)
		return
	}

	h.log.Info("User registered successfully", zap.String("email", email))
	csrfToken, exists := c.Get("csrf_token")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	// On success, render the login form so the user can sign in
	err = views.Login(csrfToken.(string)).Render(c, c.Writer)
	if err != nil {
		h.log.Error("Error rendering login component after registration", zap.Error(err))
	}
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	if err := session.Save(); err != nil {
		h.log.Error("Failed to save session on logout", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to logout")
		return
	}

	// Trigger the nav bar to re-render itself as logged-out
	c.Header("HX-Trigger", "logout")
	c.Header("HX-Redirect", "/")
}
