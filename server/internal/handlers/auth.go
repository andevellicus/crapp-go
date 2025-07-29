package handlers

import (
	"log"
	"net/http"

	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"crapp-go/views"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Assessment *models.Assessment
}

func NewAuthHandler(assessment *models.Assessment) *AuthHandler {
	return &AuthHandler{Assessment: assessment}
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
		c.String(http.StatusUnauthorized, "Invalid email or password.")
		return
	}

	session.Set("userID", user.ID)
	if err := session.Save(); err != nil {
		c.String(http.StatusInternalServerError, "Failed to login")
		return
	}

	c.Header("HX-Trigger", "login")

	// Get or create the assessment state for the user
	state, err := repository.GetOrCreateAssessmentState(user.ID, len(h.Assessment.Questions))
	if err != nil {
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
		log.Printf("Error rendering register page: %v", err)
		c.String(http.StatusInternalServerError, "Error loading page")
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if _, err := repository.CreateUser(email, password); err != nil {
		log.Printf("Error creating user: %v", err)
		c.String(http.StatusInternalServerError, "Failed to register")
		return
	}

	err := views.Login().Render(c, c.Writer)
	if err != nil {
		log.Printf("Error rendering login component after registration: %v", err)
	}
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	if err := session.Save(); err != nil {
		c.String(http.StatusInternalServerError, "Failed to logout")
		return
	}

	// Trigger the nav bar to re-render itself as logged-out
	c.Header("HX-Trigger", "logout")

	// Directly render the login view. HTMX will swap this into the #content div.
	err := views.Login().Render(c, c.Writer)
	if err != nil {
		log.Printf("Error rendering login view on logout: %v", err)
	}
}
