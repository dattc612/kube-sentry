package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/dattc/go-though-storm/api/models"
	_ "modernc.org/sqlite"
)

func Init(path string, predefinedJobs []models.Automation, logger *slog.Logger) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := database.Exec(`
		PRAGMA foreign_keys = ON;
		PRAGMA journal_mode = WAL;
	`); err != nil {
		database.Close()
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}

	if err := createTables(database); err != nil {
		database.Close()
		return nil, err
	}

	if err := SyncAutomations(database, predefinedJobs, logger); err != nil {
		database.Close()
		return nil, fmt.Errorf("sync automations: %w", err)
	}
	return database, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS automations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		type TEXT NOT NULL CHECK (
			type IN ('action', 'rule', 'workflow')
		),
		status TEXT NOT NULL DEFAULT 'active' CHECK (
			status IN ('active', 'disabled')
		),
		sample_input TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_automations_type
		ON automations(type);
	CREATE INDEX IF NOT EXISTS idx_automations_status
		ON automations(status);


	CREATE TABLE IF NOT EXISTS automation_triggers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		automation_code TEXT NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL CHECK (
			type IN (
				'manual',
				'webhook',
				'schedule',
				'event',
				'rule'
			)
		),
		config TEXT,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (automation_code)
			REFERENCES automations(code)
			ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_triggers_automation
		ON automation_triggers(automation_code);



	CREATE TABLE IF NOT EXISTS executions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		execution_code TEXT NOT NULL UNIQUE,
		automation_code TEXT NOT NULL,
		trigger_id INTEGER,
		trigger_source TEXT,
		status TEXT NOT NULL DEFAULT 'queued' CHECK (
			status IN (
				'queued',
				'running',
				'succeeded',
				'failed',
				'cancelled'
			)
    ),
    started_at DATETIME,
    finished_at DATETIME,
    duration_ms INTEGER,
    triggered_by TEXT,
    input TEXT,
    output TEXT,
    error TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (automation_code)
        REFERENCES automations(code),
    FOREIGN KEY (trigger_id)
        REFERENCES automation_triggers(id)
        ON DELETE SET NULL
	);
	CREATE INDEX IF NOT EXISTS idx_executions_automation
		ON executions(automation_code);
	CREATE INDEX IF NOT EXISTS idx_executions_status
		ON executions(status);
	CREATE INDEX IF NOT EXISTS idx_executions_started
		ON executions(started_at);
	CREATE INDEX IF NOT EXISTS idx_executions_created
		ON executions(created_at);
 

	CREATE TABLE IF NOT EXISTS execution_steps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,

		execution_id INTEGER NOT NULL,

		step_number INTEGER NOT NULL,

		name TEXT NOT NULL,

		status TEXT NOT NULL DEFAULT 'queued' CHECK (
			status IN (
				'queued',
				'running',
				'succeeded',
				'failed',
				'skipped'
			)
		),
		started_at DATETIME,
		finished_at DATETIME,
		duration_ms INTEGER,
		input TEXT,
		output TEXT,
		error TEXT,
		FOREIGN KEY (execution_id)
			REFERENCES executions(id)
			ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_execution_steps_execution
		ON execution_steps(execution_id);
	CREATE TABLE IF NOT EXISTS execution_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		execution_id INTEGER NOT NULL,
		step_id INTEGER,
		level TEXT NOT NULL DEFAULT 'info' CHECK (
			level IN (
				'debug',
				'info',
				'warn',
				'error'
			)
		),
		message TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (execution_id)
			REFERENCES executions(id)
			ON DELETE CASCADE,
		FOREIGN KEY (step_id)
			REFERENCES execution_steps(id)
			ON DELETE SET NULL
	);
	CREATE INDEX IF NOT EXISTS idx_execution_logs_execution
		ON execution_logs(execution_id);
	CREATE INDEX IF NOT EXISTS idx_execution_logs_created
		ON execution_logs(created_at);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	return nil
}
