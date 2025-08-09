package services

import (
	"crapp-go/internal/models"
	"crapp-go/internal/repository"
	"time"

	"go.uber.org/zap"
)

type Scheduler struct {
	log          *zap.Logger
	emailService *EmailService
}

func NewScheduler(log *zap.Logger, emailService *EmailService) *Scheduler {
	return &Scheduler{
		log:          log,
		emailService: emailService,
	}
}

// Start runs the scheduler in a goroutine.
func (s *Scheduler) Start() {
	s.log.Info("Starting reminder scheduler...")
	go func() {
		// Ticker will fire on every minute.
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			s.runReminderCheck()
		}
	}()
}

func (s *Scheduler) runReminderCheck() {
	// Get current time in UTC, formatted as HH:MM
	currentTime := time.Now().UTC().Format("15:04")
	s.log.Debug("Running reminder check", zap.String("utc_time", currentTime))

	// This function now correctly queries against UTC times stored in the DB
	users, err := repository.GetUsersForEmailReminder(currentTime)
	if err != nil {
		s.log.Error("Failed to get users for email reminder", zap.Error(err))
		return
	}

	for _, user := range users {
		completed, err := repository.HasCompletedAssessmentToday(user.ID)
		if err != nil {
			s.log.Error("Failed to check assessment completion status", zap.Uint("userID", user.ID), zap.Error(err))
			continue
		}

		if !completed {
			go s.sendReminder(user)
		}
	}
}

func (s *Scheduler) sendReminder(user models.User) {
	s.emailService.SendReminderEmail(user)
}
