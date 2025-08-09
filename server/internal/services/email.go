package services

import (
	"crapp-go/internal/models"
	"fmt"

	"go.uber.org/zap"
)

// EmailService is a placeholder for a real email sending service.
type EmailService struct {
	log *zap.Logger
}

func NewEmailService(log *zap.Logger) *EmailService {
	return &EmailService{log: log}
}

// SendReminderEmail simulates sending a reminder email.
func (s *EmailService) SendReminderEmail(user models.User) {
	s.log.Info("Sending reminder email",
		zap.String("to", user.Email),
		zap.String("name", user.FirstName),
	)
	// In a real application, you would use an SMTP client like go-mail
	// to send a templated HTML email here. // TODO
	fmt.Printf("--- SIMULATING EMAIL ---\nTo: %s\nSubject: Reminder to complete your daily assessment\nHi %s,\nThis is a friendly reminder to complete your daily CRAPP assessment.\n\n", user.Email, user.FirstName)
}
