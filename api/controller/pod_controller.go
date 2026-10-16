package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dattc/go-though-storm/api/common"
	"github.com/dattc/go-though-storm/api/models"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodController interface {
	GetPods(c *gin.Context)
	GetPodLogs(c *gin.Context)
	GetContainerFiles(c *gin.Context)
	GetContainerFileContent(c *gin.Context)
	DownloadContainerFile(c *gin.Context)
	GetNamespaces(c *gin.Context)
	SaveContainerFileContent(c *gin.Context)
	GetAvailableJobs(c *gin.Context)
	RunJob(c *gin.Context)
}

type podController struct {
	kubeClient          kubernetes.Interface
	metricsClient       metricsv1beta1.Interface
	restConfig          *rest.Config
	allowedNamespaces   []string
	editableFilesConfig models.EditableFilesConfig
	jobs                []models.JobConfig
	logger              *slog.Logger
}

func NewPodController(
	kubeClient kubernetes.Interface,
	metricsClient metricsv1beta1.Interface,
	restConfig *rest.Config,
	allowedNamespaces []string,
	editableFilesConfig models.EditableFilesConfig,
	jobs []models.JobConfig,
	logger *slog.Logger,
) PodController {
	if logger == nil {
		logger = slog.Default()
	}
	return &podController{
		kubeClient:          kubeClient,
		metricsClient:       metricsClient,
		restConfig:          restConfig,
		allowedNamespaces:   allowedNamespaces,
		editableFilesConfig: editableFilesConfig,
		jobs:                jobs,
		logger:              logger,
	}
}

type ContainerFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"isDirectory"`
	Size        int    `json:"size,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
	Content     string `json:"content,omitempty"`
	Editable    bool   `json:"editable"`
}

type ContainerInfo struct {
	Name     string          `json:"name"`
	Image    string          `json:"image"`
	Restarts int32           `json:"restarts"`
	Ready    bool            `json:"ready"`
	Files    []ContainerFile `json:"files,omitempty"`
}

type PodItem struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Namespace   string          `json:"namespace"`
	Status      string          `json:"status"`
	Node        string          `json:"node"`
	IP          string          `json:"ip"`
	Age         string          `json:"age"`
	CPUUsage    string          `json:"cpuUsage"`
	MemoryUsage string          `json:"memoryUsage"`
	Containers  []ContainerInfo `json:"containers"`
	Logs        string          `json:"logs,omitempty"`
}

type NamespaceItem struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

// GET /:namespace/:pod/containers/:container/jobs
func (p *podController) GetAvailableJobs(c *gin.Context) {
	pod := c.Param("pod")

	var result []models.JobConfig = []models.JobConfig{}

	for _, job := range p.jobs {
		if common.IsAllowedPod(pod, job.AllowedPod) {
			result = append(result, job)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"data":  result,
	})
}

// POST /:namespace/:pod/containers/:container/job/:id/run
func (p *podController) RunJob(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		p.logger.WarnContext(ctx, "invalid job ID param", slog.String("id", c.Param("id")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	pod := c.Param("pod")
	container := c.Param("container")
	namespace := c.Param("namespace")

	job, err := common.FindItemByID(p.jobs, id)
	if err != nil {
		p.logger.WarnContext(ctx, "failed to retrieve job", slog.Int("job_id", id), slog.Any("error", err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retrieve job: " + err.Error()})
		return
	}

	if !common.IsAllowedPod(pod, job.AllowedPod) {
		p.logger.WarnContext(ctx, "unauthorized job execution attempted",
			slog.String("job_name", job.JobName),
			slog.String("namespace", namespace),
			slog.String("pod", pod),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Can not run %s job in protected pod", job.JobName)})
		return
	}

	p.logger.InfoContext(ctx, "starting job execution",
		slog.String("job_name", job.JobName),
		slog.String("namespace", namespace),
		slog.String("pod", pod),
		slog.String("container", container),
	)

	var fullOutput strings.Builder
	fullOutput.WriteString(fmt.Sprintf("=== Running Job: %s ===\n\n", job.JobName))

	for i, step := range job.Steps {
		p.logger.DebugContext(ctx, "running job step",
			slog.String("job_name", job.JobName),
			slog.Int("step_index", i+1),
			slog.String("step_name", step.Name),
		)

		fullOutput.WriteString(fmt.Sprintf("--- Step %d: %s ---\n", i+1, step.Name))

		cmd := step.Command
		output, err := p.execCommand(ctx, namespace, pod, container, cmd)

		if output != "" {
			fullOutput.WriteString(output)
			if !strings.HasSuffix(output, "\n") {
				fullOutput.WriteString("\n")
			}
		}

		if err != nil {
			p.logger.ErrorContext(ctx, "job step failed",
				slog.String("job_name", job.JobName),
				slog.String("step_name", step.Name),
				slog.Any("error", err),
			)
			fullOutput.WriteString(fmt.Sprintf("\n[ERROR] Step failed: %s\n", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  fmt.Sprintf("Job '%s' failed at step '%s': %s", job.JobName, step.Name, err.Error()),
				"output": fullOutput.String(),
			})
			return
		}

		fullOutput.WriteString("\n")
	}

	p.logger.InfoContext(ctx, "job executed successfully", slog.String("job_name", job.JobName))

	c.JSON(http.StatusOK, gin.H{
		"output": fullOutput.String(),
		"error":  "",
	})
}

// GET /api/v1/namespaces
func (p *podController) GetNamespaces(c *gin.Context) {
	ctx := c.Request.Context()
	nsList, err := p.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to list namespaces from kubernetes", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list namespaces: " + err.Error()})
		return
	}

	var items []NamespaceItem
	for _, ns := range nsList.Items {
		if len(p.allowedNamespaces) > 0 && !slices.Contains(p.allowedNamespaces, ns.Name) {
			continue
		}

		items = append(items, NamespaceItem{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Age:    formatAge(ns.CreationTimestamp.Time),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(items),
		"data":  items,
	})
}

// GET /api/v1/pods
func (p *podController) GetPods(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")

	var finalPods []v1.Pod

	if namespace == "" {
		nsList, err := p.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to list namespaces", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var names []string
		for _, ns := range nsList.Items {
			names = append(names, ns.Name)
		}

		newNsList := common.Intersect(names, p.allowedNamespaces)

		for _, ns := range newNsList {
			pods, err := p.kubeClient.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				p.logger.ErrorContext(ctx, "failed to list pods for namespace", slog.String("namespace", ns), slog.Any("error", err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			finalPods = append(finalPods, pods.Items...)
		}
	} else {
		if !slices.Contains(p.allowedNamespaces, namespace) {
			p.logger.WarnContext(ctx, "access to forbidden namespace attempted", slog.String("namespace", namespace))
			c.JSON(http.StatusForbidden, gin.H{"error": "namespace not allowed"})
			return
		}

		allPods, err := p.kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to list pods for namespace", slog.String("namespace", namespace), slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		finalPods = allPods.Items
	}

	metricsMap := make(map[string]map[string]string)
	if p.metricsClient != nil {
		podMetrics, err := p.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			p.logger.WarnContext(ctx, "failed to fetch pod metrics", slog.String("namespace", namespace), slog.Any("error", err))
		} else {
			for _, pm := range podMetrics.Items {
				var totalCPU, totalMem int64
				for _, container := range pm.Containers {
					totalCPU += container.Usage.Cpu().MilliValue()
					totalMem += container.Usage.Memory().Value() / (1024 * 1024)
				}
				metricsMap[pm.Name] = map[string]string{
					"cpuUsage":    fmt.Sprintf("%dm", totalCPU),
					"memoryUsage": fmt.Sprintf("%dMi", totalMem),
				}
			}
		}
	}

	var response []PodItem
	for _, pod := range finalPods {
		cpu, mem := "0m", "0Mi"
		if m, ok := metricsMap[pod.Name]; ok {
			cpu = m["cpuUsage"]
			mem = m["memoryUsage"]
		}

		var containers []ContainerInfo
		for _, container := range pod.Spec.Containers {
			restarts, ready := int32(0), false
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name == container.Name {
					restarts = cs.RestartCount
					ready = cs.Ready
					break
				}
			}
			containers = append(containers, ContainerInfo{
				Name:     container.Name,
				Image:    container.Image,
				Restarts: restarts,
				Ready:    ready,
			})
		}

		response = append(response, PodItem{
			ID:          string(pod.UID),
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			Status:      string(pod.Status.Phase),
			Node:        pod.Spec.NodeName,
			IP:          pod.Status.PodIP,
			Age:         formatAge(pod.CreationTimestamp.Time),
			CPUUsage:    cpu,
			MemoryUsage: mem,
			Containers:  containers,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GET /api/v1/pods/:namespace/:pod/logs
func (p *podController) GetPodLogs(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	container := c.Query("container")

	opts := &corev1.PodLogOptions{Container: container}
	req := p.kubeClient.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to stream pod logs",
			slog.String("namespace", namespace),
			slog.String("pod", podName),
			slog.String("container", container),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, stream); err != nil {
		p.logger.ErrorContext(ctx, "error reading log stream buffer", slog.Any("error", err))
	}
	c.JSON(http.StatusOK, gin.H{"logs": buf.String()})
}

// GET /api/v1/pods/:namespace/:pod/containers/:container/files
func (p *podController) GetContainerFiles(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	container := c.Param("container")
	targetPath := c.DefaultQuery("path", "/")

	cmd := []string{"ls", "-la", "-L", targetPath}
	output, err := p.execCommand(ctx, namespace, pod, container, cmd)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to list container files",
			slog.String("namespace", namespace),
			slog.String("pod", pod),
			slog.String("container", container),
			slog.String("path", targetPath),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, parseLsOutput(output, targetPath, p.editableFilesConfig))
}

// GET /api/v1/pods/:namespace/:pod/containers/:container/fileContent
func (p *podController) GetContainerFileContent(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	container := c.Param("container")
	targetPath := c.Query("path")

	if targetPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'path' is required"})
		return
	}

	targetPath = filepath.Clean(targetPath)

	script := fmt.Sprintf(
		`TARGET=$(readlink -f %s) && `+
			`SIZE=$(stat -c%%s "$TARGET" 2>/dev/null || wc -c < "$TARGET") && `+
			`if [ "$SIZE" -gt %d ]; then `+
			`  echo "ERROR_FILE_TOO_LARGE: size is $SIZE bytes"; exit 1; `+
			`else `+
			`  cat "$TARGET"; `+
			`fi`,
		targetPath, 5242880,
	)
	cmd := []string{
		"sh",
		"-c",
		script,
	}
	output, err := p.execCommand(ctx, namespace, pod, container, cmd)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to read file content",
			slog.String("namespace", namespace),
			slog.String("pod", pod),
			slog.String("container", container),
			slog.String("path", targetPath),
			slog.Any("error", err),
		)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Failed to read file content: " + err.Error() + "\n" + output})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": output})
}

func (p *podController) GetFileSize(ctx context.Context, namespace, pod, container string, filePath string) (int64, error) {
	cmd := []string{"stat", "-c%s", filePath}
	stdout, err := p.execCommand(ctx, namespace, pod, container, cmd)
	if err != nil {
		return 0, err
	}

	sizeStr := strings.TrimSpace(stdout)
	return strconv.ParseInt(sizeStr, 10, 64)
}

func (p *podController) downloader(c *gin.Context, namespace, pod, container string, filePath string) error {
	req := p.kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{"sh", "-c", fmt.Sprintf("cat $(readlink -f %s)", filePath)},
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(p.restConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	var stderr bytes.Buffer

	err = exec.StreamWithContext(c.Request.Context(), remotecommand.StreamOptions{
		Stdout: c.Writer,
		Stderr: &stderr,
	})

	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}

	return nil
}

func (p *podController) DownloadContainerFile(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	container := c.Param("container")
	filePath := c.Query("path")

	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter required"})
		return
	}

	fileSize, err := p.GetFileSize(ctx, namespace, pod, container, filePath)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to fetch file size for download",
			slog.String("namespace", namespace),
			slog.String("pod", pod),
			slog.String("path", filePath),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file size: " + err.Error()})
		return
	}

	ext := filepath.Ext(filePath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filePath)))
	c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
	c.Header("Access-Control-Expose-Headers", "Content-Length")
	c.Header("Content-Type", contentType)
	c.Header("Transfer-Encoding", "chunked") // ignore content length for streaming

	c.Writer.WriteHeader(http.StatusOK)

	err = p.downloader(c, namespace, pod, container, filePath)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to stream file download",
			slog.String("namespace", namespace),
			slog.String("pod", pod),
			slog.String("path", filePath),
			slog.Any("error", err),
		)
		c.Error(fmt.Errorf("failed to stream file download: %w", err))
		return
	}
}

func (p *podController) execCommand(ctx context.Context, namespace, pod, container string, command []string) (string, error) {
	bytes, err := p.execCommandBytes(ctx, namespace, pod, container, command)
	return string(bytes), err
}

func (p *podController) execCommandBytes(ctx context.Context, namespace, pod, container string, command []string) ([]byte, error) {
	req := p.kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(p.restConfig, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return stdout.Bytes(), fmt.Errorf("%v: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

func formatAge(t time.Time) string {
	duration := time.Since(t)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	minutes := int(duration.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func parseLsOutput(output, parentPath string, editableFilesConfig models.EditableFilesConfig) []ContainerFile {
	var files []ContainerFile
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 9 || fields[0] == "total" {
			continue
		}

		perms := fields[0]
		name := strings.Join(fields[8:], " ")

		if name == "." || name == ".." {
			continue
		}

		isDir := strings.HasPrefix(perms, "d")
		fullFilePath := filepath.Join(parentPath, name)
		size, err := strconv.Atoi(fields[4])
		if err != nil {
			size = 0
		}
		files = append(files, ContainerFile{
			Name:        name,
			Path:        fullFilePath,
			IsDirectory: isDir,
			Size:        size,
			Permissions: perms,
			UpdatedAt:   fmt.Sprintf("%s %s %s", fields[5], fields[6], fields[7]),
			Editable:    common.IsFileEditable(fullFilePath, editableFilesConfig),
		})
	}

	return files
}

// PUT /api/v1/pods/:namespace/:pod/containers/:container/files
func (p *podController) SaveContainerFileContent(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	container := c.Param("container")
	targetPath := filepath.Clean(c.Query("path"))

	if !common.IsFileEditable(targetPath, p.editableFilesConfig) {
		p.logger.WarnContext(ctx, "unauthorized file edit attempt",
			slog.String("namespace", namespace),
			slog.String("pod", pod),
			slog.String("path", targetPath),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can not edit this file"})
		return
	}

	if targetPath == "" || targetPath == "." || targetPath == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid target file path is required"})
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		p.logger.WarnContext(ctx, "invalid request body for saving container file", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	encodedContent := base64.StdEncoding.EncodeToString([]byte(req.Content))
	cmd := []string{
		"sh",
		"-c",
		fmt.Sprintf("echo '%s' | base64 -d > $(readlink -f %s)", encodedContent, targetPath),
	}

	_, err := p.execCommand(ctx, namespace, pod, container, cmd)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to save container file content",
			slog.String("namespace", namespace),
			slog.String("pod", pod),
			slog.String("container", container),
			slog.String("path", targetPath),
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	p.logger.InfoContext(ctx, "file saved successfully",
		slog.String("namespace", namespace),
		slog.String("pod", pod),
		slog.String("container", container),
		slog.String("path", targetPath),
	)

	c.JSON(http.StatusOK, gin.H{"message": "File saved successfully"})
}
