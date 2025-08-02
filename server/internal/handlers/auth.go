package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	util "crapp-go/internal/utils"
	"crapp-go/views"
	"crapp-go/views/components"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	log        *zap.Logger
	Assessment *models.Assessment
}

func NewAuthHandler(log *zap.Logger, assessment *models.Assessment) *AuthHandler {
	return &AuthHandler{log: log, Assessment: assessment}
}

func (h *AuthHandler) ShowLoginPage(c *gin.Context) {
	views.Login().Render(c, c.Writer)
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

	session.Set("userID", user.ID)
	if err := session.Save(); err != nil {
		h.log.Error("Failed to save session", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		components.Alert("Internal server error.", "error").Render(c, c.Writer)
		return
	}

	c.Header("HX-Redirect", "/assessment") // Redirect user to the assessment page on successful login
}

func (h *AuthHandler) ShowRegisterPage(c *gin.Context) {
	err := views.Register().Render(c, c.Writer)
	if err != nil {
		h.log.Error("Error rendering register page", zap.Error(err))
		c.String(http.StatusInternalServerError, "Error rendering page")
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirmPassword")

	// 1. Check for empty fields
	if email == "" || password == "" {
		c.Status(http.StatusBadRequest)
		components.Alert("Email and password are required.", "error").Render(c, c.Writer)
		return
	}

	// 2. Validate Email Format
	if !util.IsValidEmail(email) {
		c.Status(http.StatusBadRequest)
		components.Alert("Please enter a valid email address.", "error").Render(c, c.Writer)
		return
	}

	// 3. Validate Password Complexity
	if !util.IsComplexPassword(password) {
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
	if err != sql.ErrNoRows {
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

	if _, err := repository.CreateUser(email, password); err != nil {
		h.log.Error("Failed to create user", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		components.Alert("Internal server error.", "error").Render(c, c.Writer)
		return
	}

	h.log.Info("User registered successfully", zap.String("email", email))

	// On success, render the login form so the user can sign in
	err = views.Login().Render(c, c.Writer)
	if err != nil {
		h.log.Error("Error rendering login component after registration", zap.Error(err))
	}
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	//userID, _ := session.Get("userID").(int)
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
