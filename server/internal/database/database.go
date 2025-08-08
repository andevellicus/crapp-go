package database

import (
	"crapp-go/internal/config"
	logging "crapp-go/internal/logging"
	"crapp-go/internal/models"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(log *zap.Logger) {
	var err error
	dbConf := config.Conf.Database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbConf.Host, dbConf.User, dbConf.Password, dbConf.DBName, dbConf.Port)

	// Create our custom GORM logger
	gormLogger := logging.NewGormZapLogger(log)
	gormLogger.LogLevel = logger.Info // Set the desired log level

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	log.Info("Database connection established successfully.")
	runMigrations(log)
}

func runMigrations(log *zap.Logger) {
	// GORM's AutoMigrate will create tables, columns, and foreign keys.
	// It will NOT create custom indexes, so we handle that separately.
	err := DB.AutoMigrate(
		&models.User{},
		&models.AssessmentState{},
		&models.Answer{},
		&models.AssessmentMetric{},
		&models.DSTResult{},
		&models.CPTResult{},
		&models.TMTResult{},
		&models.DSTAttempt{},
		&models.CPTEvent{},
		&models.TMTClick{},
	)
	if err != nil {
		log.Fatal("Failed to run database migrations", zap.Error(err))
	}
	log.Info("Database migrations completed successfully.")

	metricsIndex := `CREATE INDEX IF NOT EXISTS idx_metrics_query ON assessment_metrics (assessment_id, question_id, metric_key, created_at DESC);`
	if err := DB.Exec(metricsIndex).Error; err != nil {
		log.Fatal("Failed to create custom index on metrics table", zap.Error(err))
	}
	log.Info("Custom indexes ensured successfully.")
}
