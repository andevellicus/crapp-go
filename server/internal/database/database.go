package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Init() {
	var err error
	dbHost := "db"
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	DB, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Unable to open database connection: %v\n", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	log.Println("Database connection established successfully.")
	createTables()
}

func createTables() {
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

	activeAssessmentIndex := `
    CREATE UNIQUE INDEX IF NOT EXISTS one_active_assessment_per_user_idx
    ON assessments (user_id)
    WHERE is_complete = false;
    `
	_, err := DB.Exec(usersTable)
	if err != nil {
		log.Fatalf("Unable to create users table: %v\n", err)
	}

	_, err = DB.Exec(assessmentsTable)
	if err != nil {
		log.Fatalf("Unable to create assessments table: %v\n", err)
	}

	_, err = DB.Exec(answersTable)
	if err != nil {
		log.Fatalf("Unable to create answers table: %v\n", err)
	}

	_, err = DB.Exec(activeAssessmentIndex)
	if err != nil {
		log.Fatalf("Unable to create active assessment index: %v\n", err)
	}
}
