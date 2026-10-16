package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/dattc/go-though-storm/api/automation"
	"github.com/dattc/go-though-storm/api/db"
	"github.com/dattc/go-though-storm/api/models"
	"github.com/gin-gonic/gin"
	"github.com/go-viper/mapstructure/v2"
)

type JobController interface {
	WebhooksHandler(c *gin.Context)
	GetExecution(c *gin.Context)
	GetExecutionStatus(c *gin.Context)
	GetExecutionLogs(c *gin.Context)
	ListExecutions(c *gin.Context)
	GetAutomationJobs(c *gin.Context)
	GetAutomationJob(c *gin.Context)
	GetExecutionSteps(c *gin.Context)
	ManualTriggerJob(c *gin.Context)
}

type jobController struct {
	repo   *db.Repository
	engine *automation.Engine
	logger *slog.Logger
}

func NewJobController(repo *db.Repository, engine *automation.Engine, logger *slog.Logger) JobController {
	if logger == nil {
		logger = slog.Default()
	}
	return &jobController{
		repo:   repo,
		engine: engine,
		logger: logger,
	}
}

// TriggerJob POST /api/v1/jobs/webhooks/alermanager
func (ctl *jobController) WebhooksHandler(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request payload"})
		return
	}

	var payload models.AlertmanagerWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json"})
		return
	}

	var automationCode string
	var jobInput *models.JVMRestartPayload

	for _, alert := range payload.Alerts {
		if alert.Labels["alertname"] == "KubePodMemoryUnderPressure" && alert.Status == "firing" {
			automationCode = "auto_heal"
			jobInput = &models.JVMRestartPayload{
				ContainerName:   alert.Labels["container"],
				PodName:         alert.Labels["pod"],
				Namespace:       alert.Labels["namespace"],
				MemThresholdPct: 90,
				MinLifeTime:     60 * 30,
			}
			break
		}
	}

	if jobInput == nil || automationCode == "" {
		ctl.logger.Info(fmt.Sprintf("No matching firing alerts found in webhook payload: %v", payload))
		ctl.logger.Debug(fmt.Sprintf("Webhook raw payload: %v", body))
		c.JSON(http.StatusOK, gin.H{"message": "No actionable alerts found"})
		return
	}
	var jobInputMap map[string]interface{}

	if err := mapstructure.Decode(jobInput, &jobInputMap); err != nil {
		ctl.logger.Error("Failed to convert job input to map: %w", err)
		c.JSON(http.StatusOK, gin.H{"message": fmt.Errorf("Failed to convert job input to map: %w", err)})
		return
	}

	exec := &models.Execution{
		ExecutionCode:  fmt.Sprintf("EXEC-%d", time.Now().UnixNano()),
		AutomationCode: automationCode,
		TriggerSource:  string(models.TriggerWebhook),
		Status:         models.ExecutionQueued,
		TriggeredBy:    c.ClientIP(),
		Input:          jobInputMap,
		CreatedAt:      time.Now(),
	}

	if err := ctl.repo.CreateExecution(c.Request.Context(), exec); err != nil {
		ctl.logger.Error("Failed to create execution record", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue job execution"})
		return
	}

	ctl.engine.Enqueue(exec)

	ctl.logger.Info("Job execution queued successfully",
		slog.Int64("execution_id", exec.ID),
		slog.String("code", exec.ExecutionCode),
	)

	c.JSON(http.StatusAccepted, gin.H{
		"data": gin.H{
			"execution_id":   exec.ID,
			"execution_code": exec.ExecutionCode,
			"status":         exec.Status,
			"created_at":     exec.CreatedAt,
		},
	})
}

// ManualTriggerJob POST /api/v1/jobs/:automation_code/run
func (ctl *jobController) ManualTriggerJob(c *gin.Context) {
	automationCode := c.Param("automation_code")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request payload"})
		return
	}

	var input map[string]interface{}

	if err := json.Unmarshal(body, &input); err != nil {
		ctl.logger.Error("Failed to create execution record", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue job execution"})
	}

	exec := &models.Execution{
		ExecutionCode:  fmt.Sprintf("EXEC-%d", time.Now().UnixNano()),
		AutomationCode: automationCode,
		TriggerSource:  string(models.TriggerManual),
		Status:         models.ExecutionQueued,
		TriggeredBy:    c.ClientIP(),
		Input:          input,
		CreatedAt:      time.Now(),
	}

	if err := ctl.repo.CreateExecution(c.Request.Context(), exec); err != nil {
		ctl.logger.Error("Failed to create execution record", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue job execution"})
		return
	}

	ctl.engine.Enqueue(exec)

	ctl.logger.Info("Job execution queued successfully",
		slog.Int64("execution_id", exec.ID),
		slog.String("code", exec.ExecutionCode),
	)

	c.JSON(http.StatusAccepted, gin.H{
		"data": gin.H{
			"execution_id":   exec.ID,
			"execution_code": exec.ExecutionCode,
			"status":         exec.Status,
			"created_at":     exec.CreatedAt,
		},
	})
}

// GetExecution GET /api/v1/executions/:id
func (ctl *jobController) GetExecution(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution id"})
		return
	}

	ctx := c.Request.Context()
	exec, err := ctl.repo.GetExecutionByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve execution"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": exec,
	})
}

// GetExecutionStatus GET /api/v1/executions/:id/status
func (ctl *jobController) GetExecutionStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution id"})
		return
	}

	exec, err := ctl.repo.GetExecutionByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query execution status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             exec.ID,
		"execution_code": exec.ExecutionCode,
		"status":         exec.Status,
		"duration_ms":    exec.DurationMS,
		"started_at":     exec.StartedAt,
		"finished_at":    exec.FinishedAt,
		"error":          exec.Error,
	})
}

func (ctl *jobController) GetExecutionSteps(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution id"})
		return
	}

	executionStep, err := ctl.repo.GetExecutionSteps(c.Request.Context(), id)
	if err != nil {
		ctl.logger.ErrorContext(c.Request.Context(), "Failed to fetch execution steps", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch execution steps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": executionStep,
	})
}

// GetExecutionLogs GET /api/v1/executions/:id/logs
func (ctl *jobController) GetExecutionLogs(c *gin.Context) {
	// id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution id"})
	// 	return
	// }

	// logs, err := ctl.repo.GetExecutionLogs(c.Request.Context(), id)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"execution_id": id,
	// 	"logs":         logs,
	// })
}

// ListExecutions GET /api/v1/executions
func (ctl *jobController) ListExecutions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sort_by := c.DefaultQuery("sort_by", "id")
	order := c.DefaultQuery("order", "desc")
	executions, total, err := ctl.repo.ListExecutions(c.Request.Context(), limit, offset, sort_by, order)
	if err != nil {
		ctl.logger.ErrorContext(c.Request.Context(), "Failed to fetch executions list", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch executions list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   executions,
		"limit":  limit,
		"offset": offset,
		"total":  total,
	})
}

// ListAutomationJob GET /api/v1/jobs
func (ctl *jobController) GetAutomationJobs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	executions, err := ctl.repo.GetAutomationJobs(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch jobs list: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   executions,
		"limit":  limit,
		"offset": offset,
	})
}

// ListAutomationJob GET /api/v1/jobs
func (ctl *jobController) GetAutomationJob(c *gin.Context) {
	jobsId, _ := strconv.Atoi(c.Query("automation_code"))

	executions, err := ctl.repo.GetAutomationJob(c.Request.Context(), jobsId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch job %s: %s", jobsId, err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": executions,
	})
}
