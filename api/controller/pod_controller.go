package controller

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
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
}

type podController struct {
	kubeClient          kubernetes.Interface
	metricsClient       metricsv1beta1.Interface
	restConfig          *rest.Config
	allowedNamespaces   []string
	editableFilesConfig models.EditableFilesConfig
}

func NewPodController(kubeClient kubernetes.Interface, metricsClient metricsv1beta1.Interface, restConfig *rest.Config, allowedNamespaces []string, editableFilesConfig models.EditableFilesConfig) PodController {
	return &podController{
		kubeClient:          kubeClient,
		metricsClient:       metricsClient,
		restConfig:          restConfig,
		allowedNamespaces:   allowedNamespaces,
		editableFilesConfig: editableFilesConfig,
	}
}

type ContainerFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"isDirectory"`
	Size        string `json:"size,omitempty"`
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

// GET /api/v1/namespaces
func (p *podController) GetNamespaces(c *gin.Context) {

	nsList, err := p.kubeClient.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list namespaces: " + err.Error()})
		return
	}

	var items []NamespaceItem
	count := 0

	for _, ns := range nsList.Items {
		if len(p.allowedNamespaces) > 0 && !slices.Contains(p.allowedNamespaces, ns.Name) {
			continue
		}

		items = append(items, NamespaceItem{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Age:    formatAge(ns.CreationTimestamp.Time),
		})
		count++
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(items),
		"data":  items,
	})
}

// GET /api/v1/pods
func (p *podController) GetPods(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "")

	var finalPods []v1.Pod

	if namespace == "" {
		nsList, err := p.kubeClient.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var names []string
		for _, ns := range nsList.Items {
			names = append(names, ns.Name)
		}

		newNsList := common.Intersect(names, p.allowedNamespaces)

		for _, ns := range newNsList {
			pods, err := p.kubeClient.CoreV1().Pods(ns).List(c.Request.Context(), metav1.ListOptions{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			finalPods = append(finalPods, pods.Items...)
		}
	} else {
		if !slices.Contains(p.allowedNamespaces, namespace) {
			c.JSON(http.StatusForbidden, gin.H{"error": "namespace not allowed"})
			return
		}

		allPods, err := p.kubeClient.CoreV1().Pods(namespace).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		finalPods = allPods.Items
	}

	metricsMap := make(map[string]map[string]string)
	if p.metricsClient != nil {
		podMetrics, err := p.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(c.Request.Context(), metav1.ListOptions{})
		if err == nil {
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
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	container := c.Query("container")

	opts := &corev1.PodLogOptions{Container: container}
	req := p.kubeClient.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, stream)
	c.JSON(http.StatusOK, gin.H{"logs": buf.String()})
}

// GET /api/v1/pods/:namespace/:pod/containers/:container/files
func (p *podController) GetContainerFiles(c *gin.Context) {
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	container := c.Param("container")
	targetPath := c.DefaultQuery("path", "/")

	cmd := []string{"ls", "-la", "-h", "-L", targetPath}
	output, err := p.execCommand(namespace, pod, container, cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, parseLsOutput(output, targetPath, p.editableFilesConfig))
}

// GET /api/v1/pods/:namespace/:pod/containers/:container/fileContent
func (p *podController) GetContainerFileContent(c *gin.Context) {
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	container := c.Param("container")
	targetPath := c.Query("path")

	if targetPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'path' is required"})
		return
	}

	targetPath = filepath.Clean(targetPath)

	cmd := []string{
		"sh",
		"-c",
		fmt.Sprintf("head -n 1000 $(readlink -f %s)", targetPath),
	}
	output, err := p.execCommand(namespace, pod, container, cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content: " + err.Error()})
		return
	}

	c.String(http.StatusOK, output)
}

func (p *podController) DownloadContainerFile(c *gin.Context) {
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	container := c.Param("container")
	filePath := c.Query("path")

	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter required"})
		return
	}

	cmd := []string{
		"sh",
		"-c",
		fmt.Sprintf("tar chf - $(readlink -f %s)", filePath),
	}

	outputBytes, err := p.execCommandBytes(namespace, pod, container, cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file: " + err.Error()})
		return
	}

	tarReader := tar.NewReader(bytes.NewReader(outputBytes))
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed parsing tar file"})
			return
		}

		if !header.FileInfo().IsDir() {
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
			c.Header("Content-Type", "application/octet-stream")
			_, _ = io.Copy(c.Writer, tarReader)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
}

func (p *podController) execCommand(namespace, pod, container string, command []string) (string, error) {
	bytes, err := p.execCommandBytes(namespace, pod, container, command)
	return string(bytes), err
}

func (p *podController) execCommandBytes(namespace, pod, container string, command []string) ([]byte, error) {
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
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, stderr.String())
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
		files = append(files, ContainerFile{
			Name:        name,
			Path:        fullFilePath,
			IsDirectory: isDir,
			Size:        fields[4] + " B",
			Permissions: perms,
			UpdatedAt:   fmt.Sprintf("%s %s %s", fields[5], fields[6], fields[7]),
			Editable:    common.IsFileEditable(fullFilePath, editableFilesConfig),
		})
	}

	return files
}

// Put /api/v1/pods/:namespace/:pod/containers/:container/files
func (p *podController) SaveContainerFileContent(c *gin.Context) {
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	container := c.Param("container")
	targetPath := filepath.Clean(c.Query("path"))

	if !common.IsFileEditable(targetPath, p.editableFilesConfig) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	encodedContent := base64.StdEncoding.EncodeToString([]byte(req.Content))
	cmd := []string{
		"sh",
		"-c",
		fmt.Sprintf("echo '%s' | base64 -d > $(readlink -f %s)", encodedContent, targetPath),
	}

	_, err := p.execCommand(namespace, pod, container, cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File saved successfully"})
}
