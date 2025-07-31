package handlers

import (
	"net/http"

	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"crapp-go/views"

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
		c.String(http.StatusUnauthorized, "Invalid email or password.")
		return
	}

	session.Set("userID", user.ID)
	if err := session.Save(); err != nil {
		h.log.Error("Failed to save session", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to login")
		return
	}

	c.Header("HX-Trigger", "login")

	h.log.Info("User logged in successfully", zap.Int("userID", user.ID))
	state, err := repository.GetOrCreateAssessmentState(user.ID, len(h.Assessment.Questions))
	if err != nil {
		h.log.Error("Failed to get or create assessment state", zap.Error(err))
		c.String(http.StatusInternalServerError, "Could not start assessment")
		return
	}

	questionIndex := state.QuestionOrder[state.CurrentQuestionIndex]
	currentQuestion := h.Assessment.Questions[questionIndex]

	// Render the assessment component directly
	views.AssessmentPage(currentQuestion, state.CurrentQuestionIndex, len(state.QuestionOrder)).Render(c, c.Writer)
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

	if _, err := repository.CreateUser(email, password); err != nil {
		h.log.Error("Failed to create user", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to register")
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
	userID, _ := session.Get("userID").(int)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	if err := session.Save(); err != nil {
		h.log.Error("Failed to save session on logout", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to logout")
		return
	}

	// Trigger the nav bar to re-render itself as logged-out
	c.Header("HX-Trigger", "logout")
	// Directly render the login view. HTMX will swap this into the #content div.
	err := views.Login().Render(c, c.Writer)
	if err != nil {
		h.log.Error("Error rendering login view on logout", zap.Error(err))
	}
	h.log.Info("User logged out successfully", zap.Int("userID", userID))
}
