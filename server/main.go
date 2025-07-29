package main

import (
	"log"

	"crapp-go/internal/database"
	"crapp-go/internal/models"
	"crapp-go/internal/router"
)

func main() {
	// Initialize Database
	database.Init()

	// Load assessment questions at startup
	assessment, err := models.LoadAssessment("../config/questions.yaml")
	if err != nil {
		log.Fatalf("Failed to load assessment: %v", err)
	}

	// Setup router
	r := router.Setup(assessment)

	// Start the Gin server
	port := ":5050"
	log.Printf("Server listening on http://localhost%s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to run Gin server: %v", err)
	}
}
