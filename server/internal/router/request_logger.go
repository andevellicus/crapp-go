package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger creates a gin middleware for logging requests using zap.
func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		status := c.Writer.Status()
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		}

		switch {
		case status >= 500:
			log.Error("Server error", fields...)
		case status >= 400:
			log.Warn("Client error", fields...)
		default:
			// Log successful requests at the Debug level to reduce noise
			log.Debug("Request processed", fields...)
		}
	}
}
