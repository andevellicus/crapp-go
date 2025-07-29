package main

import (
	"log"
	"net/http" // Still useful for HTTP status codes

	// For the counter component

	"crapp-go/views"
	"crapp-go/views/common"
	"crapp-go/views/components"

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

func main() {
	// Initialize Gin router
	router := gin.Default()
	// Setup session middleware
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// Serve static files from the 'assets' directory.
	// This will serve your generated assets/css/style.css, and any other static assets.
	router.Static("/assets", "./assets") // Changed from /static to /assets

	// --- Routes using Gin handlers ---

	// Root route - serves the full layout with the home content
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

		// Dummy authentication
		if username == "admin" && password == "password" {
			session.Set("user", username)
			err := session.Save()
			if err != nil {
				log.Printf("Error saving session: %v", err)
				c.String(http.StatusInternalServerError, "Failed to login")
				return
			}
			// Trigger a 'login' event on the client
			c.Header("HX-Trigger", "login")

			// On successful login, redirect to the counter page
			err = components.Counter(0).Render(c, c.Writer)
			if err != nil {
				log.Printf("Error rendering counter component: %v", err)
				c.String(http.StatusInternalServerError, "Error loading counter")
			}

		} else {
			// On failed login, re-render the login form with an error
			// (for simplicity, we're just re-rendering the form for now)
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

		// Trigger a 'logout' event on the client
		c.Header("HX-Trigger", "logout")

		// Re-render the login form after logout
		err = views.Login().Render(c, c.Writer)
		if err != nil {
			log.Printf("Error rendering login component: %v", err)
			c.String(http.StatusInternalServerError, "Error loading login content")
		}
	})

	// Add this new route to render the nav component
	router.GET("/nav", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		// Render the nav component based on login state
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
	}

	// Start the Gin server
	port := ":8080"
	log.Printf("Server listening on http://localhost%s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to run Gin server: %v", err)
	}
}
