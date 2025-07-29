// main.go
package main

import (
	"log"
	"net/http"

	"crapp-go/internal/models"
	"crapp-go/views"
	"crapp-go/views/common"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// A simple middleware to check if the user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			// User is not logged in, render the unauthorized page
			err := views.Unauthorized().Render(c, c.Writer)
			if err != nil {
				log.Printf("Error rendering unauthorized page: %v", err)
				c.String(http.StatusInternalServerError, "Error loading page")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

var assessment *models.Assessment

func main() {
	// Load assessment questions at startup
	var err error
	assessment, err = models.LoadAssessment("../config/questions.yaml")
	if err != nil {
		log.Fatalf("Failed to load assessment: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()
	// Setup session middleware
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// Serve static files from the 'assets' directory.
	router.Static("/assets", "./assets")

	// --- Routes using Gin handlers ---

	// Root route
	router.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		err := views.Page("Home Page", user != nil).Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Printf("Error rendering root page: %v", err)
			c.String(http.StatusInternalServerError, "Error loading page")
		}
	})

	router.POST("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "admin" && password == "password" {
			session.Set("user", username)
			err := session.Save()
			if err != nil {
				log.Printf("Error saving session: %v", err)
				c.String(http.StatusInternalServerError, "Failed to login")
				return
			}
			c.Header("HX-Trigger", "login")
			startAssessment(c)
		} else {
			err := views.Login().Render(c, c.Writer)
			if err != nil {
				log.Printf("Error rendering login component: %v", err)
				c.String(http.StatusInternalServerError, "Error loading login content")
			}
		}
	})

	router.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Delete("user")
		err := session.Save()
		if err != nil {
			log.Printf("Error saving session: %v", err)
			c.String(http.StatusInternalServerError, "Failed to logout")
			return
		}
		c.Header("HX-Trigger", "logout")
		err = views.Login().Render(c, c.Writer)
		if err != nil {
			log.Printf("Error rendering login component: %v", err)
			c.String(http.StatusInternalServerError, "Error loading login content")
		}
	})

	router.GET("/nav", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		err := common.Nav(user != nil).Render(c, c.Writer)
		if err != nil {
			log.Printf("Error rendering nav component: %v", err)
			c.String(http.StatusInternalServerError, "Error loading nav content")
		}
	})

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(AuthRequired())
	{
		authorized.GET("/assessment", startAssessment)
		authorized.POST("/assessment/next", nextQuestion)
	}

	// Start the Gin server
	port := ":5050"
	log.Printf("Server listening on http://localhost%s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to run Gin server: %v", err)
	}
}

func startAssessment(c *gin.Context) {
	session := sessions.Default(c)

	// Shuffle questions and store their IDs in the session
	shuffledQuestions := make([]models.Question, len(assessment.Questions))
	copy(shuffledQuestions, assessment.Questions)
	models.ShuffleQuestions(shuffledQuestions)

	var questionOrder []string
	for _, q := range shuffledQuestions {
		questionOrder = append(questionOrder, q.ID)
	}

	session.Set("question_order", questionOrder)
	session.Set("current_question_index", 0)
	session.Set("answers", make(map[string]string))
	session.Save()

	// Render the first question
	firstQuestionID := questionOrder[0]
	var firstQuestion models.Question
	for _, q := range assessment.Questions {
		if q.ID == firstQuestionID {
			firstQuestion = q
			break
		}
	}

	err := views.AssessmentPage(firstQuestion, 0, len(questionOrder)).Render(c, c.Writer)
	if err != nil {
		log.Printf("Error rendering assessment page: %v", err)
		c.String(http.StatusInternalServerError, "Error starting assessment")
	}
}

func nextQuestion(c *gin.Context) {
	session := sessions.Default(c)
	questionOrder := session.Get("question_order").([]string)
	currentIndex := session.Get("current_question_index").(int)
	answers := session.Get("answers").(map[string]string)

	// Save the answer from the previous question
	questionID := c.PostForm("questionId")
	answer := c.PostForm("answer")
	answers[questionID] = answer
	session.Set("answers", answers)

	// Move to the next question
	currentIndex++
	session.Set("current_question_index", currentIndex)
	session.Save()

	if currentIndex >= len(questionOrder) {
		// Assessment is complete
		err := views.AssessmentResults(answers).Render(c, c.Writer)
		if err != nil {
			log.Printf("Error rendering assessment results: %v", err)
			c.String(http.StatusInternalServerError, "Error showing results")
		}
		return
	}

	// Render the next question
	nextQuestionID := questionOrder[currentIndex]
	var nextQuestion models.Question
	for _, q := range assessment.Questions {
		if q.ID == nextQuestionID {
			nextQuestion = q
			break
		}
	}

	err := views.AssessmentPage(nextQuestion, currentIndex, len(questionOrder)).Render(c, c.Writer)
	if err != nil {
		log.Printf("Error rendering next question: %v", err)
		c.String(http.StatusInternalServerError, "Error loading next question")
	}
}
