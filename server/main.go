package main

import (
	"crapp-go/internal/config"
	"crapp-go/internal/database"
	logger "crapp-go/internal/logging"
	"crapp-go/internal/models"
	"crapp-go/internal/router"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// ProjectRoot is the absolute path to the project's root directory.
// This is set at build time using ldflags.
var ProjectRoot string

func main() {
	// Initialize Logger
	projectRoot, err := GetProjectRoot()
	if err != nil {
		fmt.Println("Error getting project root:", err)
		return
	}

	log, err := logger.Init(projectRoot)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	// Initialize Configuration
	if err := config.Init(projectRoot, log); err != nil {
		log.Fatal("Failed to initialize configuration", zap.Error(err))
	}

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
	port := ":" + config.Conf.Server.Port
	log.Info("Server listening on http://localhost" + port)
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to run Gin server", zap.Error(err))
	}
}

func GetProjectRoot() (string, error) {
	if ProjectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("could not get current working directory: %w", err)
		}
		if filepath.Base(cwd) == "server" {
			ProjectRoot = filepath.Dir(cwd)
		} else {
			ProjectRoot = cwd
		}
	}
	return ProjectRoot, nil
}
