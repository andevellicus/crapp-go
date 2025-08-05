package utils

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// GenerateSecureToken creates a cryptographically secure random token.
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
