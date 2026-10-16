package models

import (
	"time"
)

type AutomationType string

const (
	AutomationAction   AutomationType = "action"
	AutomationRule     AutomationType = "rule"
	AutomationWorkflow AutomationType = "workflow"
)

type AutomationStatus string

const (
	AutomationActive   AutomationStatus = "active"
	AutomationDisabled AutomationStatus = "disabled"
)

type Automation struct {
	ID          int64            `json:"id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        AutomationType   `json:"type"`
	Status      AutomationStatus `json:"status"`
	SampleInput map[string]any   `json:"sample_input,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type TriggerType string

const (
	TriggerManual   TriggerType = "manual"
	TriggerWebhook  TriggerType = "webhook"
	TriggerSchedule TriggerType = "schedule"
	TriggerEvent    TriggerType = "event"
	TriggerRule     TriggerType = "rule"
)

type AutomationTrigger struct {
	ID             int64       `json:"id"`
	AutomationCode int64       `json:"automation_code"`
	Name           string      `json:"name"`
	Type           TriggerType `json:"type"`
	Config         string      `json:"config,omitempty"`
	Enabled        bool        `json:"enabled"`
	CreatedAt      time.Time   `json:"created_at"`
}

type ExecutionStatus string

const (
	ExecutionQueued    ExecutionStatus = "queued"
	ExecutionRunning   ExecutionStatus = "running"
	ExecutionSucceeded ExecutionStatus = "succeeded"
	ExecutionFailed    ExecutionStatus = "failed"
	ExecutionCancelled ExecutionStatus = "cancelled"
)

type Execution struct {
	ID             int64                       `json:"id"`
	Name           string                      `json:"name"`
	ExecutionCode  string                      `json:"execution_code"`
	AutomationCode string                      `json:"automation_code"`
	TriggerID      *int64                      `json:"trigger_id,omitempty"`
	TriggerSource  string                      `json:"trigger_source"`
	Status         ExecutionStatus             `json:"status"`
	StartedAt      *time.Time                  `json:"started_at,omitempty"`
	FinishedAt     *time.Time                  `json:"finished_at,omitempty"`
	DurationMS     *int64                      `json:"duration_ms,omitempty"`
	TriggeredBy    string                      `json:"triggered_by,omitempty"`
	Input          map[string]interface{}      `json:"input,omitempty"`
	Output         *string                     `json:"output,omitempty"`
	Error          *string                     `json:"error,omitempty"`
	CreatedAt      time.Time                   `json:"created_at"`
	ExtraConfig    *[]AutomationJobExtraConfig `json:"extra_config,omitempty"`
}

type ExecutionStep struct {
	ID          int64                  `json:"id"`
	ExecutionID int64                  `json:"execution_id"`
	StepNumber  int                    `json:"step_number"`
	Name        string                 `json:"name"`
	Status      ExecutionStatus        `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	FinishedAt  *time.Time             `json:"finished_at,omitempty"`
	DurationMS  *int64                 `json:"duration_ms,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       *string                `json:"error,omitempty"`
	Log         *ExecutionStepLog      `json:"log,omitempty"`
}

type ExecutionStepLog struct {
	ID          int64     `json:"id"`
	ExecutionID int64     `json:"execution_id"`
	StepID      *int64    `json:"step_id,omitempty"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

type JVMRestartPayload struct {
	Namespace          string  `json:"namespace" mapstructure:"namespace"`
	PodName            string  `json:"pod_name" mapstructure:"pod_name"`
	ContainerName      string  `json:"container_name" mapstructure:"container_name"`
	CPUThresholdPct    float64 `json:"cpu_threshold_pct" mapstructure:"cpu_threshold_pct"`
	MemThresholdPct    float64 `json:"mem_threshold_pct" mapstructure:"mem_threshold_pct"`
	GracePeriodSeconds int     `json:"grace_period_seconds" mapstructure:"grace_period_seconds"`
	MinLifeTime        int     `json:"min_life_time" mapstructure:"min_life_time"`
}
