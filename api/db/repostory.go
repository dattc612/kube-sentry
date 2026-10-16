package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dattc/go-though-storm/api/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateExecution(ctx context.Context, exec *models.Execution) error {
	query := `
        INSERT INTO executions (
            execution_code,
            automation_code,
            trigger_id,
            trigger_source,
            status,
            triggered_by,
            input,
            created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	var inputJSON []byte

	if exec.Input != nil {
		var err error

		inputJSON, err = json.Marshal(exec.Input)
		if err != nil {
			return fmt.Errorf(
				"failed to marshal execution input: %w",
				err,
			)
		}
	}

	res, err := r.db.ExecContext(
		ctx,
		query,
		exec.ExecutionCode,
		exec.AutomationCode,
		exec.TriggerID,
		exec.TriggerSource,
		exec.Status,
		exec.TriggeredBy,
		inputJSON,
		exec.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	exec.ID = id

	return nil
}

func (r *Repository) StartExecution(ctx context.Context, id int64, startedAt time.Time) error {
	query := `UPDATE executions SET status = ?, started_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, models.ExecutionRunning, startedAt, id)
	return err
}

func (r *Repository) CompleteExecution(ctx context.Context, exec *models.Execution) error {
	query := `
        UPDATE executions
        SET status = ?, finished_at = ?, duration_ms = ?, output = ?, error = ?
        WHERE id = ?
    `

	outputJSON, err := marshalJSONOrBytes(exec.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal execution output: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		exec.Status, exec.FinishedAt, exec.DurationMS, outputJSON, exec.Error, exec.ID,
	)
	return err
}

func (r *Repository) CreateExecutionStep(ctx context.Context, step *models.ExecutionStep) error {
	query := `
        INSERT INTO execution_steps (
            execution_id,
            step_number,
            name,
            status,
            started_at,
            input
        ) VALUES (?, ?, ?, ?, ?, ?)
    `

	var inputJSON []byte

	if step.Input != nil {
		var err error

		inputJSON, err = json.Marshal(step.Input)
		if err != nil {
			return fmt.Errorf(
				"failed to marshal execution step input: %w",
				err,
			)
		}
	}

	res, err := r.db.ExecContext(
		ctx,
		query,
		step.ExecutionID,
		step.StepNumber,
		step.Name,
		step.Status,
		step.StartedAt,
		inputJSON,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to create execution step: %w",
			err,
		)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf(
			"failed to get execution step ID: %w",
			err,
		)
	}

	step.ID = id

	return nil
}

func (r *Repository) CompleteExecutionStep(ctx context.Context, step *models.ExecutionStep) error {
	query := `
        UPDATE execution_steps
        SET status = ?, finished_at = ?, duration_ms = ?, output = ?, error = ?
        WHERE id = ?
    `

	outputJSON, err := marshalJSONOrBytes(step.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal execution step output: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		step.Status, step.FinishedAt, step.DurationMS, outputJSON, step.Error, step.ID,
	)
	return err
}

func (r *Repository) CreateLog(ctx context.Context, log *models.ExecutionStepLog) error {
	query := `
        INSERT INTO execution_logs (execution_id, step_id, level, message, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
	res, err := r.db.ExecContext(ctx, query,
		log.ExecutionID, log.StepID, log.Level, log.Message, log.CreatedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = id
	return nil
}

var ErrNotFound = errors.New("record not found")

func (r *Repository) GetExecutionByID(ctx context.Context, id int64) (*models.Execution, error) {
	query := `
        SELECT id, execution_code, automation_code, trigger_id, trigger_source, status,
               started_at, finished_at, duration_ms, triggered_by, input, output, error, created_at
        FROM executions WHERE id = ?
    `
	exec := &models.Execution{}
	var inputJSON []byte
	var outputJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&exec.ID, &exec.ExecutionCode, &exec.AutomationCode, &exec.TriggerID, &exec.TriggerSource, &exec.Status,
		&exec.StartedAt, &exec.FinishedAt, &exec.DurationMS, &exec.TriggeredBy, &inputJSON, &outputJSON, &exec.Error, &exec.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if len(inputJSON) > 0 {
		if err := json.Unmarshal(inputJSON, &exec.Input); err != nil {
			return nil, fmt.Errorf("failed to parse execution input JSON: %w", err)
		}
	}

	if len(outputJSON) > 0 {
		parseOutput(outputJSON, &exec.Output)
	}

	return exec, nil
}

func (r *Repository) GetExecutionSteps(ctx context.Context, executionID int64) ([]models.ExecutionStep, error) {
	query := `
        SELECT id, execution_id, step_number, name, status, started_at, finished_at, duration_ms, input, output, error
        FROM execution_steps WHERE execution_id = ? ORDER BY step_number ASC
    `
	rows, err := r.db.QueryContext(ctx, query, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.ExecutionStep
	for rows.Next() {
		var s models.ExecutionStep
		var inputJSON []byte
		var outputJSON []byte

		if err := rows.Scan(
			&s.ID, &s.ExecutionID, &s.StepNumber, &s.Name, &s.Status,
			&s.StartedAt, &s.FinishedAt, &s.DurationMS, &inputJSON, &outputJSON, &s.Error,
		); err != nil {
			return nil, err
		}

		if len(inputJSON) > 0 {
			if err := json.Unmarshal(inputJSON, &s.Input); err != nil {
				return nil, fmt.Errorf("failed to parse execution step input JSON: %w", err)
			}
		}

		if len(outputJSON) > 0 {
			parseOutput(outputJSON, &s.Output)
		}

		stepLog, err := r.getExecutionStepLogs(ctx, executionID, s.ID)
		if err != nil {
			return nil, err
		}
		s.Log = stepLog
		steps = append(steps, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return steps, nil
}

func (r *Repository) getExecutionStepLogs(ctx context.Context, executionID, stepID int64) (*models.ExecutionStepLog, error) {
	query := `
        SELECT id, execution_id, step_id, level, message, created_at
        FROM execution_logs
        WHERE execution_id = ? AND step_id = ?
        ORDER BY id ASC
        LIMIT 1
    `

	var l models.ExecutionStepLog
	err := r.db.QueryRowContext(ctx, query, executionID, stepID).Scan(
		&l.ID, &l.ExecutionID, &l.StepID, &l.Level, &l.Message, &l.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &l, nil
}

func (r *Repository) ListExecutions(ctx context.Context, limit, offset int, sort_by, order string) ([]models.Execution, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM executions`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
        SELECT id, execution_code, automation_code, trigger_id, trigger_source, status,
               started_at, finished_at, duration_ms, triggered_by, input, output, error, created_at
        FROM executions ORDER BY ` + sort_by + ` ` + order + ` LIMIT ? OFFSET ?
    `
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	executions := make([]models.Execution, 0)
	for rows.Next() {
		var inputJSON []byte
		var outputJSON []byte
		var exec models.Execution

		if err := rows.Scan(
			&exec.ID, &exec.ExecutionCode, &exec.AutomationCode, &exec.TriggerID, &exec.TriggerSource, &exec.Status,
			&exec.StartedAt, &exec.FinishedAt, &exec.DurationMS, &exec.TriggeredBy, &inputJSON, &outputJSON, &exec.Error, &exec.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		if len(inputJSON) > 0 {
			if err := json.Unmarshal(inputJSON, &exec.Input); err != nil {
				return nil, 0, fmt.Errorf("failed to parse execution input JSON: %w", err)
			}
		}

		if len(outputJSON) > 0 {
			parseOutput(outputJSON, &exec.Output)
		}

		executions = append(executions, exec)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

func (r *Repository) GetAutomationJobs(ctx context.Context, limit, offset int) ([]models.Automation, error) {
	query := `
        SELECT id, code, name, description, type, status, sample_input, created_at, updated_at
        FROM automations ORDER BY id ASC LIMIT ? OFFSET ?
    `
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.Automation
	for rows.Next() {
		var job models.Automation
		var sampleInput []byte
		if err := rows.Scan(&job.ID, &job.Code, &job.Name, &job.Description, &job.Type, &job.Status, &sampleInput, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, err
		}

		if len(sampleInput) > 0 {
			if err := json.Unmarshal(sampleInput, &job.SampleInput); err != nil {
				return nil, err
			}
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return jobs, err
	}
	return jobs, nil
}

func (r *Repository) GetAutomationJob(ctx context.Context, id int) (*models.Automation, error) {
	query := `
        SELECT id, code, name, description, type, status, sample_input, created_at, updated_at
        FROM automations WHERE id = ? LIMIT 1
    `
	var job models.Automation
	var sampleInput []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.Code, &job.Name, &job.Description, &job.Type, &job.Status, &sampleInput, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if len(sampleInput) > 0 {
		if err := json.Unmarshal(sampleInput, &job.SampleInput); err != nil {
			return nil, err
		}
	}
	return &job, nil
}

// Helpers for Output handling
func parseOutput(data []byte, target interface{}) {
	if len(data) == 0 {
		return
	}
	if err := json.Unmarshal(data, target); err != nil {
		// Fallback to raw string if value is plain string rather than JSON object
		if ptr, ok := target.(*interface{}); ok {
			*ptr = string(data)
		}
	}
}

func marshalJSONOrBytes(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	default:
		return json.Marshal(val)
	}
}
