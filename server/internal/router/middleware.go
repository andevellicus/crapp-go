package router

import (
	"crapp-go/internal/repository"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserLoaderMiddleware checks for a userID in the session.
func UserLoaderMiddleware(log *zap.Logger) gin.HandlerFunc {
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
			log.Warn("Invalid user ID in session, clearing.",
				zap.Int("userID", userID),
				zap.Error(err),
			)
			// Clear the bad session and treat as a guest.
			session.Clear()
			session.Options(sessions.Options{Path: "/", MaxAge: -1})
			session.Save()
			c.Next()
			return
		}

		// User is valid, store user object in context for other handlers.
		log.Debug("User loaded from session and added to context.", zap.Int("userID", user.ID))
		c.Set("user", user)
		c.Next()
	}
}

// AuthRequired now simply checks if a valid user was loaded into the context.
func AuthRequired(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("user"); !exists {
			log.Warn("Unauthorized access attempt",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
			)
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
