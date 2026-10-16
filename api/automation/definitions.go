package automation

import (
	"github.com/dattc/go-though-storm/api/models"
)

var DefaultJobs = []models.Automation{
	{
		Code:        "auto_heal",
		Name:        "Auto Heal",
		Description: "Detects High memory or CPU usage pods and executes container cleanup and restart.",
		Type:        models.AutomationAction,
		Status:      models.AutomationActive,
		SampleInput: map[string]any{
			"namespace":            "default",
			"pod_name":             "example-pod",
			"container_name":       "fluentbit",
			"cpu_threshold_pct":    80,
			"mem_threshold_pct":    90,
			"grace_period_seconds": 30,
			"min_life_time":        300,
		},
	},
	{
		Code:        "dump_java_process",
		Name:        "Dump Java Process",
		Description: "Captures a thread or heap dump from a target Java process for diagnostics.",
		Type:        models.AutomationAction,
		SampleInput: map[string]any{
			"pid": 12345,
		},
		Status: models.AutomationDisabled,
	},
	// {
	// 	Code:        "SCALE-DEPLOYMENT",
	// 	Name:        "scale_deployment",
	// 	Description: "Adjusts horizontal pod autoscaler bounds or manually scales replica count.",
	// 	Type:        models.AutomationAction,
	// 	Pack:        "k8s_ops",
	// 	Status:      models.AutomationActive,
	// 	Definition:  `{"steps": ["validate_replicas", "apply_scale"]}`,
	// },
	// {
	// 	Code:        "COLLECT-POD-DUMP",
	// 	Name:        "collect_pod_dump",
	// 	Description: "Archives heap dump, thread dump, and standard console logs into S3 storage.",
	// 	Type:        models.AutomationWorkflow,
	// 	Pack:        "diagnostics",
	// 	Status:      models.AutomationActive,
	// 	Definition:  `{"steps": ["generate_heap_dump", "fetch_logs", "upload_s3"]}`,
	// },
	// {
	// 	Code:        "SLACK-INCIDENT-ALERT",
	// 	Name:        "slack_incident_alert",
	// 	Description: "Dispatches formatted alert payload to On-Call Slack channel on pod failure.",
	// 	Type:        models.AutomationAction,
	// 	Pack:        "notifications",
	// 	Status:      models.AutomationActive,
	// 	Definition:  `{"steps": ["format_payload", "send_webhook"]}`,
	// },
	// {
	// 	Code:        "ROTATE-SERVICE-ACCOUNT",
	// 	Name:        "rotate_service_account",
	// 	Description: "Rotates and re-applies Kubernetes service account tokens automatically.",
	// 	Type:        models.AutomationAction,
	// 	Pack:        "security",
	// 	Status:      models.AutomationDisabled,
	// 	Definition:  `{"steps": ["generate_token", "update_secret", "restart_deployments"]}`,
	// },
}
