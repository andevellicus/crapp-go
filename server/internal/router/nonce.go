package router

import (
	"crapp-go/internal/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const CspNonceContextKey = "csp_nonce"

// NonceMiddleware creates a new cryptographic nonce for each request
// and adds it to the Gin context for use in headers and templates.
func NonceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		var nonce string
		// Check if a nonce already exists in the session
		sessionNonce := session.Get(CspNonceContextKey)
		if sessionNonce == nil {
			var err error
			nonce, err = utils.GenerateSecureToken(32)
			if err != nil {
				panic("failed to generate CSP nonce")
			}
			session.Set(CspNonceContextKey, nonce)
			if err := session.Save(); err != nil {
				panic("failed to save session")
			}
		} else {
			nonce = sessionNonce.(string)
		}

		c.Set(CspNonceContextKey, nonce)
		c.Next()
	}
}
