package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dattc/go-though-storm/api/db"
	"github.com/dattc/go-though-storm/api/models"
)

type Engine struct {
	repo         *db.Repository
	registry     *Registry
	jobQueue     chan *models.Execution
	wg           sync.WaitGroup
	logger       *slog.Logger
	extraConfigs []models.AutomationJobExtraConfig
}

func NewEngine(repo *db.Repository, bufferSize int, logger *slog.Logger, registry *Registry, extraConfigs []models.AutomationJobExtraConfig) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	if registry == nil {
		registry = DefaultRegistry
	}
	return &Engine{
		registry:     registry,
		repo:         repo,
		jobQueue:     make(chan *models.Execution, bufferSize),
		logger:       logger,
		extraConfigs: extraConfigs,
	}
}

func (e *Engine) Start(ctx context.Context, workers int) {
	e.logger.InfoContext(ctx, "Starting automation engine worker pool", slog.Int("workers", workers))
	for i := 0; i < workers; i++ {
		e.wg.Add(1)
		go e.worker(ctx, i+1)
	}
}

func (e *Engine) Enqueue(exec *models.Execution) {
	e.logger.Debug("Enqueuing execution job",
		slog.Int64("execution_id", exec.ID),
		slog.String("execution_code", exec.ExecutionCode),
	)
	e.jobQueue <- exec
}

func (e *Engine) Stop() {
	e.logger.Info("Stopping automation engine and draining job queue")
	close(e.jobQueue)
	e.wg.Wait()
	e.logger.Info("Automation engine stopped")
}

func (e *Engine) worker(ctx context.Context, workerID int) {
	defer e.wg.Done()

	e.logger.DebugContext(ctx, "Worker started", slog.Int("worker_id", workerID))

	for exec := range e.jobQueue {
		e.logger.InfoContext(ctx, "processing execution job",
			slog.Int("worker_id", workerID),
			slog.Int64("execution_id", exec.ID),
			slog.String("execution_code", exec.ExecutionCode),
		)

		steps, err := e.registry.GetSteps(exec.AutomationCode)
		if err != nil {
			e.logger.ErrorContext(ctx, "failed fetching steps for execution",
				slog.Int64("execution_id", exec.ID),
				slog.String("automation_code", exec.AutomationCode),
				slog.Any("error", err),
			)
			e.failExecutionEarly(ctx, exec, err)
			continue
		}

		e.runExecution(ctx, exec, steps)
	}

	e.logger.DebugContext(ctx, "Worker stopped", slog.Int("worker_id", workerID))
}

func (e *Engine) runExecution(ctx context.Context, exec *models.Execution, steps []StepDefinition) {
	now := time.Now()
	exec.StartedAt = &now
	exec.Status = models.ExecutionRunning
	if err := e.repo.StartExecution(ctx, exec.ID, now); err != nil {
		e.logger.ErrorContext(ctx, "Failed to update execution status to running in db",
			slog.Int64("execution_id", exec.ID),
			slog.Any("error", err),
		)
	}

	e.logger.InfoContext(ctx, "Execution started",
		slog.Int64("execution_id", exec.ID),
		slog.String("execution_code", exec.ExecutionCode),
	)
	e.writeLog(ctx, exec.ID, nil, "info", fmt.Sprintf("Execution %s started", exec.ExecutionCode))

	var hasError bool

	for idx, stepDef := range steps {
		stepNum := idx + 1
		stepStart := time.Now()

		step := &models.ExecutionStep{
			ExecutionID: exec.ID,
			StepNumber:  stepNum,
			Name:        stepDef.Name,
			Status:      models.ExecutionRunning,
			StartedAt:   &stepStart,
			Input:       exec.Input,
		}

		if err := e.repo.CreateExecutionStep(ctx, step); err != nil {
			e.logger.ErrorContext(ctx, "Error recording step start in db",
				slog.Int64("execution_id", exec.ID),
				slog.Int("step_number", stepNum),
				slog.String("step_name", stepDef.Name),
				slog.Any("error", err),
			)
		}

		output, outputLog, err := stepDef.Exec(ctx, step.Input, e.extraConfigs)
		stepEnd := time.Now()
		duration := stepEnd.Sub(stepStart).Milliseconds()

		step.FinishedAt = &stepEnd
		step.DurationMS = &duration

		if err != nil {
			step.Status = models.ExecutionFailed
			errStr := err.Error()
			step.Error = &errStr
			hasError = true

			if dbErr := e.repo.CompleteExecutionStep(ctx, step); dbErr != nil {
				e.logger.ErrorContext(ctx, "Failed to record failed step completion in db", slog.Any("error", dbErr))
			}

			e.logger.ErrorContext(ctx, "Step execution failed",
				slog.Int64("execution_id", exec.ID),
				slog.String("step_name", stepDef.Name),
				slog.Any("error", err),
			)
			e.writeLog(ctx, exec.ID, &step.ID, "error", fmt.Sprintf("Step '%s' failed: %v", stepDef.Name, err))
			break
		}

		step.Status = models.ExecutionSucceeded
		step.Output = output

		if dbErr := e.repo.CompleteExecutionStep(ctx, step); dbErr != nil {
			e.logger.ErrorContext(ctx, "Failed to record successful step completion in db", slog.Any("error", dbErr))
		}

		e.logger.InfoContext(ctx, "Step executed successfully",
			slog.Int64("execution_id", exec.ID),
			slog.String("step_name", stepDef.Name),
			slog.Int64("duration_ms", duration),
		)
		if dbErr := e.repo.CompleteExecutionStep(ctx, step); dbErr != nil {
			e.logger.ErrorContext(ctx, "Failed to record successful step completion in db", slog.Any("error", dbErr))
		}
		b, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		e.writeLog(ctx, exec.ID, &step.ID, "info", fmt.Sprintf("%s\nOuput: %s\nStep '%s' succeeded", outputLog, b, stepDef.Name))
	}

	finishedAt := time.Now()
	totalDuration := finishedAt.Sub(*exec.StartedAt).Milliseconds()
	exec.FinishedAt = &finishedAt
	exec.DurationMS = &totalDuration

	if hasError {
		exec.Status = models.ExecutionFailed
	} else {
		exec.Status = models.ExecutionSucceeded
	}

	if err := e.repo.CompleteExecution(ctx, exec); err != nil {
		e.logger.ErrorContext(ctx, "Failed to record final execution status in db",
			slog.Int64("execution_id", exec.ID),
			slog.Any("error", err),
		)
	}

	e.logger.InfoContext(ctx, "Execution finished",
		slog.Int64("execution_id", exec.ID),
		slog.String("execution_code", exec.ExecutionCode),
		slog.String("status", string(exec.Status)),
		slog.Int64("duration_ms", totalDuration),
	)
	e.writeLog(ctx, exec.ID, nil, "info", fmt.Sprintf("Execution %s finished with status: %s", exec.ExecutionCode, exec.Status))
}

func (e *Engine) writeLog(ctx context.Context, execID int64, stepID *int64, level, message string) {
	logEntry := &models.ExecutionStepLog{
		ExecutionID: execID,
		StepID:      stepID,
		Level:       level,
		Message:     message,
		CreatedAt:   time.Now(),
	}
	if err := e.repo.CreateLog(ctx, logEntry); err != nil {
		e.logger.ErrorContext(ctx, "Failed to write execution log entry to db",
			slog.Int64("execution_id", execID),
			slog.Any("error", err),
		)
	}
}

func (e *Engine) failExecutionEarly(ctx context.Context, exec *models.Execution, err error) {
	now := time.Now()
	exec.StartedAt = &now
	exec.FinishedAt = &now
	exec.Status = models.ExecutionFailed

	_ = e.repo.StartExecution(ctx, exec.ID, now)
	_ = e.repo.CompleteExecution(ctx, exec)
	e.writeLog(ctx, exec.ID, nil, "error", fmt.Sprintf("Workflow initialization failed: %v", err))
}
