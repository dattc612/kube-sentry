package jobs

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/dattc/go-though-storm/api/automation"
	"github.com/dattc/go-though-storm/api/common"
	"github.com/dattc/go-though-storm/api/models"
	"github.com/go-viper/mapstructure/v2"
)

func RegisterJVMAutoRestart(
	reg *automation.Registry,
	kubeClient kubernetes.Interface,
	metricsClient metricsv1beta1.Interface,
	restConfig *rest.Config,
) {
	registeredAutomationCode := "auto_heal"

	reg.Register(registeredAutomationCode, func() []automation.StepDefinition {
		var payload models.JVMRestartPayload
		var targetPID string

		return []automation.StepDefinition{
			{
				Name: "Parse & Validate Input",
				Exec: func(
					ctx context.Context,
					input map[string]interface{},
					extraConfigs []models.AutomationJobExtraConfig,
				) (map[string]interface{}, string, error) {

					outputMessage := ""

					if err := mapstructure.Decode(input, &payload); err != nil {
						outputMessage += "[FAIL] Decode payload structure... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: Invalid payload structure: %w", outputMessage, err)
					}
					outputMessage += "[PASS] Decode payload structure... => OK\n"

					if payload.Namespace == "" {
						outputMessage += "[FAIL] Check 'namespace' field... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: namespace is required", outputMessage)
					}
					outputMessage += "[PASS] Check 'namespace' field... => OK\n"

					if payload.PodName == "" {
						outputMessage += "[FAIL] Check 'pod_name' field... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: pod_name is required", outputMessage)
					}
					outputMessage += "[PASS] Check 'pod_name' field... => OK\n"

					if payload.ContainerName == "" {
						outputMessage += "[FAIL] Check 'container_name' field... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: container_name is required", outputMessage)
					}
					outputMessage += "[PASS] Check 'container_name' field... => OK\n"

					policyCheckPassed := true
					for _, extraConfig := range extraConfigs {
						if extraConfig.Code != registeredAutomationCode {
							continue
						}

						if !extraConfig.Enabled {
							outputMessage += fmt.Sprintf("[FAIL] Automation job '%s' enabled policy... => FAILED\n", extraConfig.Code)
							policyCheckPassed = false
							return nil, outputMessage, fmt.Errorf("%sError: Automation job '%s' is disabled by configuration", outputMessage, extraConfig.Code)
						}

						isAllowed, err := common.MatchesAnyPattern(
							payload.PodName,
							extraConfig.AllowedPod,
						)
						if err != nil {
							outputMessage += fmt.Sprintf("[FAIL] Evaluate pod patterns for config '%s'... => FAILED\n", extraConfig.Code)
							policyCheckPassed = false
							return nil, outputMessage, fmt.Errorf("%sError evaluating pod patterns for config '%s': %w", outputMessage, extraConfig.Code, err)
						}

						if !isAllowed {
							outputMessage += fmt.Sprintf("[FAIL] Pod '%s' allowed by policy '%s'... => FAILED\n", payload.PodName, extraConfig.Code)
							policyCheckPassed = false
							return nil, outputMessage, fmt.Errorf("%sError: Pod '%s' is not allowed by policy '%s' (patterns: %v)", outputMessage, payload.PodName, extraConfig.Code, extraConfig.AllowedPod)
						}
					}
					if policyCheckPassed {
						outputMessage += "[PASS] Check automation job policies... => OK\n"
					}

					if payload.GracePeriodSeconds <= 0 {
						payload.GracePeriodSeconds = 60
					}

					outputMessage += fmt.Sprintf(
						"[PASS] All validation checks passed (Target: %s/%s, Container: %s, CPU threshold: %.2f%%, Memory threshold: %.2f%%, Grace period: %ds)... => OK\n",
						payload.Namespace,
						payload.PodName,
						payload.ContainerName,
						payload.CPUThresholdPct,
						payload.MemThresholdPct,
						payload.GracePeriodSeconds,
					)

					return input, outputMessage, nil
				},
			},

			{
				Name: "Evaluate Resource Usage Thresholds & Process Age",
				Exec: func(
					ctx context.Context,
					input map[string]interface{},
					extraConfigs []models.AutomationJobExtraConfig,
				) (map[string]interface{}, string, error) {

					outputMessage := ""

					pidDiscoveryScript := `
                        PID=$(pgrep -f -o java 2>/dev/null || ps -ef | grep '[j]ava' | awk '{print $2}' | head -n 1)

                        if [ -z "$PID" ]; then
                        echo "NOT_FOUND"
                        exit 0
                        fi

                        CLK_TCK=$(getconf CLK_TCK 2>/dev/null || echo 100)
                        SYS_UPTIME=$(awk '{print $1}' /proc/uptime)

                        STAT=$(cat /proc/$PID/stat)

                        PPID=$(echo "$STAT" | awk '{print $4}')
                        START_TIME_TICKS=$(echo "$STAT" | awk '{print $22}')

                        START_SEC=$(awk \
                        -v ticks="$START_TIME_TICKS" \
                        -v clk="$CLK_TCK" \
                        'BEGIN { print ticks / clk }'
                        )

                        AGE_SEC=$(awk \
                        -v uptime="$SYS_UPTIME" \
                        -v start="$START_SEC" \
                        'BEGIN { printf "%.0f", uptime - start }'
                        )

                        echo "${PID}:${PPID}:${AGE_SEC}"
                        `

					stdoutPID, stderrPID, err := execContainerCmd(
						ctx,
						kubeClient,
						restConfig,
						payload.Namespace,
						payload.PodName,
						payload.ContainerName,
						[]string{
							"sh",
							"-c",
							pidDiscoveryScript,
						},
					)
					if err != nil {
						outputMessage += "[FAIL] Execute PID discovery script... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: PID discovery command failed: %w; stdout: %s; stderr: %s", outputMessage, err, outputOrNone(stdoutPID), outputOrNone(stderrPID))
					}
					outputMessage += "[PASS] Execute PID discovery script... => OK\n"

					rawPIDRes := strings.TrimSpace(stdoutPID)

					if rawPIDRes == "NOT_FOUND" || rawPIDRes == "" {
						outputMessage += "[FAIL] Active Java process check... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: No active Java process found; stdout: %s; stderr: %s", outputMessage, outputOrNone(stdoutPID), outputOrNone(stderrPID))
					}
					outputMessage += "[PASS] Active Java process check... => OK\n"

					pidParts := strings.Split(rawPIDRes, ":")

					if len(pidParts) != 3 {
						outputMessage += "[FAIL] Parse PID discovery output... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: Unexpected PID discovery output: %s; stderr: %s", outputMessage, outputOrNone(stdoutPID), outputOrNone(stderrPID))
					}

					targetPID = strings.TrimSpace(pidParts[0])

					if targetPID == "" {
						outputMessage += "[FAIL] Target PID validation... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: PID discovery returned an empty PID; stdout: %s\n", outputMessage, outputOrNone(stdoutPID))
					}
					outputMessage += fmt.Sprintf("[PASS] Target PID discovery (PID: %s)... => OK\n", targetPID)

					ageSec, err := strconv.ParseFloat(
						strings.TrimSpace(pidParts[2]),
						64,
					)
					if err != nil {
						outputMessage += "[FAIL] Parse process age... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: Failed to parse Java process age from output %q: %w\n", outputMessage, rawPIDRes, err)
					}

					minLifeTime := 60 * 30
					if payload.MinLifeTime > 0 {
						minLifeTime = payload.MinLifeTime
					}
					if ageSec < float64(minLifeTime) {
						outputMessage += fmt.Sprintf("[FAIL] Process lifetime check (PID %s age %.0fm < min %dm)... => FAILED\n", targetPID, ageSec/60, minLifeTime/60)
						return nil, outputMessage, fmt.Errorf("%sError: Evaluation failed: Java process PID %s is only %.0f minutes old (< %d minutes), restart skipped", outputMessage, targetPID, ageSec/60, minLifeTime/60)
					}
					outputMessage += fmt.Sprintf("[PASS] Process lifetime check (PID %s age %.0fm >= min %dm)... => OK\n", targetPID, ageSec/60, minLifeTime/60)

					metricsScript := `
                        if [ -f /sys/fs/cgroup/memory.current ]; then
                        USAGE=$(cat /sys/fs/cgroup/memory.current)
                        LIMIT=$(cat /sys/fs/cgroup/memory.max)
                        else
                        USAGE=$(cat /sys/fs/cgroup/memory/memory.usage_in_bytes 2>/dev/null || echo 0)
                        LIMIT=$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes 2>/dev/null || echo 0)
                        fi

                        USAGE_MB=$((USAGE / 1024 / 1024))

                        if [ "$LIMIT" = "max" ] || [ "$LIMIT" -gt 100000000000000 2>/dev/null ]; then
                        echo "MEM:${USAGE_MB}:0:100"
                        else
                        LIMIT_MB=$((LIMIT / 1024 / 1024))

                        if [ "$LIMIT_MB" -le 0 ]; then
                            echo "MEM:${USAGE_MB}:0:0"
                        else
                            PCT=$((USAGE_MB * 100 / LIMIT_MB))
                            echo "MEM:${USAGE_MB}:${LIMIT_MB}:${PCT}"
                        fi
                        fi
                    `

					stdoutMem, stderrMem, err := execContainerCmd(
						ctx,
						kubeClient,
						restConfig,
						payload.Namespace,
						payload.PodName,
						payload.ContainerName,
						[]string{
							"sh",
							"-c",
							metricsScript,
						},
					)
					if err != nil {
						outputMessage += "[FAIL] Execute memory metrics script... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: Memory metrics command failed: %w; stdout: %s; stderr: %s", outputMessage, err, outputOrNone(stdoutMem), outputOrNone(stderrMem))
					}
					outputMessage += "[PASS] Execute memory metrics script... => OK\n"

					metricsOutput := strings.TrimSpace(stdoutMem)
					parts := strings.Split(metricsOutput, ":")

					if len(parts) < 4 || parts[0] != "MEM" {
						outputMessage += "[FAIL] Parse memory metrics output... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: Unexpected memory metrics output: %s; stderr: %s", outputMessage, outputOrNone(stdoutMem), outputOrNone(stderrMem))
					}

					memPct, err := strconv.ParseFloat(parts[3], 64)
					if err != nil {
						outputMessage += "[FAIL] Parse memory percentage... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: failed to parse memory percentage %q: %w", outputMessage, parts[3], err)
					}

					memUsageMB := parts[1]
					memLimitMB := parts[2]
					memBreached := memPct >= payload.MemThresholdPct

					if !memBreached {
						outputMessage += fmt.Sprintf("[FAIL] Memory threshold breach check (Usage: %.2f%% < Threshold: %.2f%%)... => FAILED\n", memPct, payload.MemThresholdPct)
						return nil, outputMessage, fmt.Errorf("%sError: Memory thresholds not breached (Usage: %.2f%%, Limit: %sMB/%sMB)", outputMessage, memPct, memUsageMB, memLimitMB)
					}
					outputMessage += fmt.Sprintf("[PASS] Memory threshold breach check (Usage: %.2f%% >= Threshold: %.2f%%)... => OK\n", memPct, payload.MemThresholdPct)

					return input, outputMessage, nil
				},
			},

			{
				Name: "Execute Graceful JVM Shutdown (SIGTERM)",
				Exec: func(
					ctx context.Context,
					input map[string]interface{},
					extraConfigs []models.AutomationJobExtraConfig,
				) (map[string]interface{}, string, error) {

					outputMessage := ""

					if targetPID == "" {
						outputMessage += "[FAIL] Target PID state check... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: target PID not set from evaluation step", outputMessage)
					}
					outputMessage += fmt.Sprintf("[PASS] Target PID state check (PID: %s)... => OK\n", targetPID)

					cmd := []string{
						"sh",
						"-c",
						fmt.Sprintf(
							"kill -15 %s",
							targetPID,
						),
					}

					stdout, stderr, err := execContainerCmd(
						ctx,
						kubeClient,
						restConfig,
						payload.Namespace,
						payload.PodName,
						payload.ContainerName,
						cmd,
					)
					if err != nil {
						outputMessage += fmt.Sprintf("[FAIL] Send SIGTERM command to PID %s... => FAILED\n", targetPID)
						return nil, outputMessage, fmt.Errorf("%sError: SIGTERM command failed for PID %s: %w; stdout: %s; stderr: %s", outputMessage, targetPID, err, outputOrNone(stdout), outputOrNone(stderr))
					}
					outputMessage += fmt.Sprintf("[PASS] Send SIGTERM command to PID %s... => OK\n", targetPID)

					if err := waitForGracePeriod(
						ctx,
						payload.GracePeriodSeconds,
					); err != nil {
						outputMessage += fmt.Sprintf("[FAIL] Wait grace period (%ds)... => FAILED\n", payload.GracePeriodSeconds)
						return nil, outputMessage, fmt.Errorf("%sError: Grace period interrupted after SIGTERM to PID %s: %w", outputMessage, targetPID, err)
					}
					outputMessage += fmt.Sprintf("[PASS] Wait grace period (%ds)... => OK\n", payload.GracePeriodSeconds)

					return input, outputMessage, nil
				},
			},

			{
				Name: "Verify Process Status & Force Kill (SIGKILL fallback)",
				Exec: func(
					ctx context.Context,
					input map[string]interface{},
					extraConfigs []models.AutomationJobExtraConfig,
				) (map[string]interface{}, string, error) {

					outputMessage := ""

					// Check 1: Target PID state
					if targetPID == "" {
						outputMessage += "[FAIL] Target PID state check... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: Target PID not set from evaluation step", outputMessage)
					}
					outputMessage += fmt.Sprintf("[PASS] Target PID state check (PID: %s)... => OK\n", targetPID)

					checkCmd := []string{
						"sh",
						"-c",
						fmt.Sprintf(
							"kill -0 %s 2>/dev/null && echo ALIVE || echo DEAD",
							targetPID,
						),
					}

					stdout, stderr, err := execContainerCmd(
						ctx,
						kubeClient,
						restConfig,
						payload.Namespace,
						payload.PodName,
						payload.ContainerName,
						checkCmd,
					)
					if err != nil {
						outputMessage += "[FAIL] Execute process verification command... => FAILED\n"
						return nil, outputMessage, fmt.Errorf("%sError: Failed checking JVM PID %s: %w; stdout: %s; stderr: %s", outputMessage, targetPID, err, outputOrNone(stdout), outputOrNone(stderr))
					}
					outputMessage += "[PASS] Execute process verification command... => OK\n"

					status := strings.TrimSpace(stdout)

					if status == "DEAD" {
						outputMessage += fmt.Sprintf("[PASS] Process termination check (PID %s is DEAD)... => OK\n", targetPID)
						return input, outputMessage, nil
					}

					outputMessage += fmt.Sprintf("[FAIL] Process termination check (PID %s status: %s)... => FAILED\n", targetPID, status)
					return nil, outputMessage, fmt.Errorf("%sError: Unexpected JVM PID %s status: %q; stdout: %s; stderr: %s", outputMessage, targetPID, status, outputOrNone(stdout), outputOrNone(stderr))
				},
			},
		}
	})
}

func execContainerCmd(
	ctx context.Context,
	client kubernetes.Interface,
	config *rest.Config,
	namespace,
	pod,
	container string,
	command []string,
) (string, string, error) {
	req := client.
		CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec")

	option := &corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}

	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(
		config,
		"POST",
		req.URL(),
	)
	if err != nil {
		return "", "", fmt.Errorf(
			"failed to create Kubernetes exec executor: %w",
			err,
		)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err = exec.StreamWithContext(
		ctx,
		remotecommand.StreamOptions{
			Stdout: &stdout,
			Stderr: &stderr,
		},
	)
	fmt.Printf("stderr raw = %q\n", stderr)

	return stdout.String(), stderr.String(), err
}

func waitForGracePeriod(ctx context.Context, seconds int) error {
	if seconds <= 0 {
		return nil
	}

	timer := time.NewTimer(
		time.Duration(seconds) * time.Second,
	)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

func outputOrNone(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		fmt.Println("RETURNING <empty>")
		return "<empty>"
	}

	return output
}
