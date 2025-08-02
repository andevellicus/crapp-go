package database

import (
	"crapp-go/internal/config"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var DB *sql.DB

func Init(log *zap.Logger) {
	var err error
	/*
		dbHost := "db"
		dbPort := os.Getenv("POSTGRES_PORT")
		dbUser := os.Getenv("POSTGRES_USER")
		dbPassword := os.Getenv("POSTGRES_PASSWORD")
		dbName := os.Getenv("POSTGRES_DB")
	*/

	// Use the configuration from Viper
	dbConf := config.Conf.Database

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConf.Host, dbConf.Port, dbConf.User, dbConf.Password, dbConf.DBName)

	DB, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("Unable to open database connection", zap.Error(err))
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Unable to ping database", zap.Error(err))
	}

	log.Info("Database connection established successfully.")
	createTables(log)
}

func createTables(log *zap.Logger) {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);`

	assessmentsTable := `
	CREATE TABLE IF NOT EXISTS assessments (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		is_complete BOOLEAN NOT NULL DEFAULT false,
		question_order INTEGER[] NOT NULL,
		current_question_index INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	answersTable := `
	CREATE TABLE IF NOT EXISTS answers (
		id SERIAL PRIMARY KEY,
		assessment_id INTEGER NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
		question_id VARCHAR(255) NOT NULL,
		answer_value TEXT NOT NULL
	);`

	metricsTable := `
	CREATE TABLE IF NOT EXISTS metrics (
		id SERIAL PRIMARY KEY,
		assessment_id INTEGER REFERENCES assessments(id) ON DELETE CASCADE,
		question_id VARCHAR(255),
		metric_key VARCHAR(255) NOT NULL,
		metric_value FLOAT NOT NULL,
		sample_size INTEGER,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	activeAssessmentIndex := `
    CREATE UNIQUE INDEX IF NOT EXISTS one_active_assessment_per_user_idx
    ON assessments (user_id)
    WHERE is_complete = false;
    `
	_, err := DB.Exec(usersTable)
	if err != nil {
		log.Fatal("Unable to create users table", zap.Error(err))
	}

	_, err = DB.Exec(assessmentsTable)
	if err != nil {
		log.Fatal("Unable to create assessments table", zap.Error(err))
	}

	_, err = DB.Exec(answersTable)
	if err != nil {
		log.Fatal("Unable to create answers table", zap.Error(err))
	}

	_, err = DB.Exec(metricsTable)
	if err != nil {
		log.Fatal("Unable to create assessment_metrics table", zap.Error(err))
	}

	_, err = DB.Exec(activeAssessmentIndex)
	if err != nil {
		log.Fatal("Unable to create active assessment index", zap.Error(err))
	}
}
