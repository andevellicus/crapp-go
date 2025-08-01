package handlers

import (
	"net/http"

	"crapp-go/internal/models"
	"crapp-go/internal/repository"
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

func (h *AuthHandler) RegisterPage(c *gin.Context) {
	err := views.Register().Render(c, c.Writer)
	if err != nil {
		h.log.Error("Error rendering register page", zap.Error(err))
		c.String(http.StatusInternalServerError, "Error rendering page")
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.Status(http.StatusBadRequest)
		components.Alert("Email and password are required.", "error").Render(c, c.Writer)
		return
	}

	if _, err := repository.CreateUser(email, password); err != nil {
		h.log.Error("Failed to create user", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		components.Alert("Internal server error.", "error").Render(c, c.Writer)
		return
	}

	h.log.Info("User registered successfully", zap.String("email", email))

	err := views.Login().Render(c, c.Writer)
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
