package router

import (
	"crapp-go/internal/utils"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Define keys for storing the token in the session and context.
const (
	csrfTokenSessionKey = "csrf_token"
	csrfTokenFormKey    = "_csrf"
	csrfTokenContextKey = "csrf_token"
	csrfTokenHeaderKey  = "X-CSRF-Token"
)

// GenerateSecureToken creates a cryptographically secure random token.
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CSRFProtection is a custom middleware to protect against CSRF attacks.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		// 1. Get or create the real CSRF token for the session.
		var token string
		sessionToken := session.Get(csrfTokenSessionKey)

		if sessionToken == nil {
			// Generate a new token if one doesn't exist.
			newToken, err := utils.GenerateSecureToken(32)
			if err != nil {
				// Handle the unlikely event of a token generation failure.
				c.AbortWithError(http.StatusInternalServerError, errors.New("failed to generate CSRF token"))
				return
			}
			token = newToken
			session.Set(csrfTokenSessionKey, token)
			if err := session.Save(); err != nil {
				c.AbortWithError(http.StatusInternalServerError, errors.New("failed to save session"))
				return
			}
		} else {
			token = sessionToken.(string)
		}

		// 2. Make the token available for the templates.
		c.Set(csrfTokenContextKey, token)

		// 3. Validate the token on unsafe methods (POST, etc.).
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			realToken := session.Get(csrfTokenSessionKey)
			if realToken == nil {
				c.AbortWithError(http.StatusForbidden, errors.New("CSRF token not found in session"))
				return
			}

			// Get token from the form data first.
			submittedToken := c.PostForm(csrfTokenFormKey)
			// If it's not in the form, check the header (for fetch requests).
			if submittedToken == "" {
				submittedToken = c.GetHeader(csrfTokenHeaderKey)
			}

			if submittedToken == "" || submittedToken != realToken {
				// If CSRF validation fails, check if it's an HTMX request.
				if c.GetHeader("HX-Request") == "true" {
					// If so, send the redirect header along with the 403 status.
					c.Header("HX-Redirect", "/")
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
				// Otherwise, just send the standard 403.
				c.AbortWithError(http.StatusForbidden, errors.New("invalid CSRF token"))
				return
			}
		}

		// If everything is okay, proceed to the next handler.
		c.Next()
	}
}
