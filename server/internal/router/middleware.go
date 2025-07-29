package router

import (
	"crapp-go/internal/repository"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// UserLoaderMiddleware checks for a userID in the session.
// If found, it loads the user from the database and adds it to the context.
// This ensures we don't have "zombie" sessions for users who no longer exist.
func UserLoaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID, ok := session.Get("userID").(int)
		if !ok {
			// No user ID in session, proceed as a guest.
			c.Next()
			return
		}

		user, err := repository.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			// User ID from session is invalid (user was deleted, etc.)
			// Clear the bad session and treat as a guest.
			session.Clear()
			session.Options(sessions.Options{Path: "/", MaxAge: -1})
			session.Save()
			c.Next()
			return
		}

		// User is valid, store user object in context for other handlers.
		c.Set("user", user)
		c.Next()
	}
}

// AuthRequired now simply checks if a valid user was loaded into the context.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("user"); !exists {
			if c.GetHeader("HX-Request") == "true" {
				c.Header("HX-Redirect", "/")
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}
