package main

import (
	"log"
	"net/http" // Still useful for HTTP status codes

	// For the counter component

	"crapp-go/views"
	"crapp-go/views/components"

	"github.com/gin-gonic/gin" // Import Gin
)

var count = 0 // Simple in-memory counter for demonstration

func main() {
	// Initialize Gin router
	router := gin.Default()

	// Serve static files from the 'assets' directory.
	// This will serve your generated assets/css/style.css, and any other static assets.
	router.Static("/assets", "./assets") // Changed from /static to /assets

	// --- Routes using Gin handlers ---

	// Root route - serves the full layout with the home content
	router.GET("/", func(c *gin.Context) {
		err := views.Page("Home Page").Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Printf("Error rendering root page: %v", err)
			c.String(http.StatusInternalServerError, "Error loading page")
		}
	})

	// HTMX endpoint for the home content only
	router.GET("/home", func(c *gin.Context) {
		err := views.Home().Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Printf("Error rendering home component: %v", err)
			c.String(http.StatusInternalServerError, "Error loading home content")
		}
	})

	// HTMX endpoint for the counter component
	router.GET("/counter", func(c *gin.Context) {
		err := components.Counter(count).Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Printf("Error rendering counter component: %v", err)
			c.String(http.StatusInternalServerError, "Error loading counter")
		}
	})

	// HTMX endpoint to increment the counter
	router.POST("/increment", func(c *gin.Context) {
		count++
		err := components.Counter(count).Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Printf("Error rendering incremented counter: %v", err)
			c.String(http.StatusInternalServerError, "Error incrementing counter")
		}
	})

	// HTMX endpoint to decrement the counter
	router.POST("/decrement", func(c *gin.Context) {
		count--
		err := components.Counter(count).Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Printf("Error rendering decremented counter: %v", err)
			c.String(http.StatusInternalServerError, "Error decrementing counter")
		}
	})

	// Start the Gin server
	port := ":8080"
	log.Printf("Server listening on http://localhost%s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to run Gin server: %v", err)
	}
}
