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
	// --- Base Tables ---
	usersTable := `CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, email TEXT NOT NULL UNIQUE, password TEXT NOT NULL);`
	assessmentsTable := `CREATE TABLE IF NOT EXISTS assessments (id SERIAL PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, is_complete BOOLEAN NOT NULL DEFAULT false, question_order INTEGER[] NOT NULL, current_question_index INTEGER NOT NULL DEFAULT 0, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);`

	answersTable := `
	CREATE TABLE IF NOT EXISTS answers (
		id SERIAL PRIMARY KEY,
		assessment_id INTEGER NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
		question_id VARCHAR(255) NOT NULL,
		answer_value TEXT NOT NULL,
		UNIQUE(assessment_id, question_id)
	);`

	metricsTable := `CREATE TABLE IF NOT EXISTS metrics (id SERIAL PRIMARY KEY, assessment_id INTEGER REFERENCES assessments(id) ON DELETE CASCADE, question_id VARCHAR(255), metric_key VARCHAR(255) NOT NULL, metric_value FLOAT NOT NULL, sample_size INTEGER, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);`
	activeAssessmentIndex := `CREATE UNIQUE INDEX IF NOT EXISTS one_active_assessment_per_user_idx ON assessments (user_id) WHERE is_complete = false;`

	// --- Summary Tables ---
	dstResultsTable := `CREATE TABLE IF NOT EXISTS dst_results (id SERIAL PRIMARY KEY, assessment_id INTEGER NOT NULL REFERENCES assessments(id) ON DELETE CASCADE, highest_span_achieved INTEGER NOT NULL, total_trials INTEGER NOT NULL, correct_trials INTEGER NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);`
	cptResultsTable := `CREATE TABLE IF NOT EXISTS cpt_results (id SERIAL PRIMARY KEY, assessment_id INTEGER NOT NULL REFERENCES assessments(id) ON DELETE CASCADE, correct_detections INTEGER NOT NULL, commission_errors INTEGER NOT NULL, omission_errors INTEGER NOT NULL, average_reaction_time FLOAT NOT NULL, reaction_time_sd FLOAT NOT NULL, detection_rate FLOAT NOT NULL, omission_error_rate FLOAT NOT NULL, commission_error_rate FLOAT NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);`
	tmtResultsTable := `CREATE TABLE IF NOT EXISTS tmt_results (id SERIAL PRIMARY KEY, assessment_id INTEGER NOT NULL REFERENCES assessments(id) ON DELETE CASCADE, part_a_completion_time FLOAT NOT NULL, part_a_errors INTEGER NOT NULL, part_b_completion_time FLOAT NOT NULL, part_b_errors INTEGER NOT NULL, b_to_a_ratio FLOAT NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);`

	// --- Granular Event Tables ---
	dstAttemptsTable := `CREATE TABLE IF NOT EXISTS dst_attempts (id SERIAL PRIMARY KEY, result_id INTEGER NOT NULL REFERENCES dst_results(id) ON DELETE CASCADE, span INTEGER NOT NULL, trial INTEGER NOT NULL, sequence TEXT NOT NULL, input TEXT NOT NULL, is_correct BOOLEAN NOT NULL, "timestamp" FLOAT NOT NULL);`
	cptEventsTable := `CREATE TABLE IF NOT EXISTS cpt_events (id SERIAL PRIMARY KEY, result_id INTEGER NOT NULL REFERENCES cpt_results(id) ON DELETE CASCADE, event_type VARCHAR(10) NOT NULL, stimulus_value VARCHAR(1), is_target BOOLEAN, presented_at FLOAT, response_time FLOAT, stimulus_index INTEGER);`
	tmtClicksTable := `CREATE TABLE IF NOT EXISTS tmt_clicks (id SERIAL PRIMARY KEY, result_id INTEGER NOT NULL REFERENCES tmt_results(id) ON DELETE CASCADE, x FLOAT NOT NULL, y FLOAT NOT NULL, "time" FLOAT NOT NULL, target_item INTEGER NOT NULL, current_part VARCHAR(10) NOT NULL);`

	execSQL(log, usersTable, "users")
	execSQL(log, assessmentsTable, "assessments")
	execSQL(log, answersTable, "answers")
	execSQL(log, metricsTable, "metrics")
	execSQL(log, activeAssessmentIndex, "active_assessment_index")
	execSQL(log, dstResultsTable, "dst_results")
	execSQL(log, cptResultsTable, "cpt_results")
	execSQL(log, tmtResultsTable, "tmt_results")
	execSQL(log, dstAttemptsTable, "dst_attempts")
	execSQL(log, cptEventsTable, "cpt_events")
	execSQL(log, tmtClicksTable, "tmt_clicks")
}

func execSQL(log *zap.Logger, sql, tableName string) {
	_, err := DB.Exec(sql)
	if err != nil {
		log.Fatal("Unable to create table/index", zap.String("name", tableName), zap.Error(err))
	}
}
