package main

import (
	"crapp-go/internal/database"
	logger "crapp-go/internal/logging"
	"crapp-go/internal/models"
	"crapp-go/internal/router"

	"go.uber.org/zap"
)

func main() {
	// Initialize Logger
	log, err := logger.Init()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	// Initialize Database
	database.Init(log)

	// Load assessment questions at startup
	assessment, err := models.LoadAssessment("../config/questions.yaml")
	if err != nil {
		log.Fatal("Failed to load assessment", zap.Error(err))
	}

	// Setup router, passing the logger to it
	r := router.Setup(log, assessment)

	// Start the Gin server
	port := ":5050"
	log.Info("Server listening on http://localhost" + port)
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to run Gin server", zap.Error(err))
	}
}
