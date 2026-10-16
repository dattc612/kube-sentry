package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dattc/go-though-storm/api/models"
)

func SyncAutomations(db *sql.DB, jobs []models.Automation, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("starting automations database synchronization", slog.Int("total_jobs", len(jobs)))

	tx, err := db.Begin()
	if err != nil {
		logger.Error("failed to begin sync transaction", slog.Any("error", err))
		return fmt.Errorf("begin sync transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
        INSERT INTO automations (code, name, description, type, status, sample_input, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(code) DO UPDATE SET
            name = excluded.name,
            description = excluded.description,
            type = excluded.type,
            status = excluded.status,
            sample_input = excluded.sample_input,
            updated_at = CURRENT_TIMESTAMP;
    `)
	if err != nil {
		logger.Error("failed to prepare upsert statement", slog.Any("error", err))
		return fmt.Errorf("prepare upsert statement: %w", err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		sampleInput, err := json.Marshal(job.SampleInput)
		if err != nil {
			logger.Error("failed to marshal sample input", slog.Any("error", err))
			return fmt.Errorf("marshal sample input for job '%s': %w", job.Code, err)
		}
		_, err = stmt.Exec(
			job.Code,
			job.Name,
			job.Description,
			job.Type,
			job.Status,
			sampleInput,
		)
		if err != nil {
			logger.Error("failed to sync automation job",
				slog.String("code", job.Code),
				slog.String("name", job.Name),
				slog.Any("error", err),
			)
			return fmt.Errorf("failed to sync job '%s': %w", job.Code, err)
		}

		logger.Debug("upserted automation job",
			slog.String("code", job.Code),
			slog.String("name", job.Name),
			slog.String("type", string(job.Type)),
		)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("failed to commit sync transaction", slog.Any("error", err))
		return fmt.Errorf("commit sync transaction: %w", err)
	}

	logger.Info("automations database synchronization completed", slog.Int("synced_jobs", len(jobs)))
	return nil
}
