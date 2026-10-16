package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/dattc/go-though-storm/api/common"
	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sResourceController interface {
	GetDeployments(c *gin.Context)
	GetDeploymentJSON(c *gin.Context)
	UpdateDeploymentJSON(c *gin.Context)
	DeleteDeployment(c *gin.Context)

	GetServices(c *gin.Context)
	GetServiceJSON(c *gin.Context)
	UpdateServiceJSON(c *gin.Context)
	DeleteService(c *gin.Context)

	GetConfigMaps(c *gin.Context)
	GetConfigMapJSON(c *gin.Context)
	UpdateConfigMapJSON(c *gin.Context)
	DeleteConfigMap(c *gin.Context)

	GetSecrets(c *gin.Context)
	GetSecretJSON(c *gin.Context)
	UpdateSecretJSON(c *gin.Context)
	DeleteSecret(c *gin.Context)

	GetStatefulSets(c *gin.Context)
	GetStatefulSetJSON(c *gin.Context)
	UpdateStatefulSetJSON(c *gin.Context)
	DeleteStatefulSet(c *gin.Context)

	GetDaemonSets(c *gin.Context)
	GetDaemonSetJSON(c *gin.Context)
	UpdateDaemonSetJSON(c *gin.Context)
	DeleteDaemonSet(c *gin.Context)

	GetCronJobs(c *gin.Context)
	GetCronJobJSON(c *gin.Context)
	UpdateCronJobJSON(c *gin.Context)
	DeleteCronJob(c *gin.Context)

	GetIngresses(c *gin.Context)
	GetIngressJSON(c *gin.Context)
	UpdateIngressJSON(c *gin.Context)
	DeleteIngress(c *gin.Context)

	GetPVCs(c *gin.Context)
	GetPVCJSON(c *gin.Context)
	DeletePVC(c *gin.Context)

	GetPVs(c *gin.Context)
	GetPVJSON(c *gin.Context)
	DeletePV(c *gin.Context)

	GetNodes(c *gin.Context)
	GetNodeJSON(c *gin.Context)
}

type k8sResourceController struct {
	kubeClient        kubernetes.Interface
	restConfig        *rest.Config
	allowedNamespaces []string
	logger            *slog.Logger
}

func NewK8sResourceController(
	kubeClient kubernetes.Interface,
	restConfig *rest.Config,
	allowedNamespaces []string,
	logger *slog.Logger,
) K8sResourceController {
	if logger == nil {
		logger = slog.Default()
	}
	return &k8sResourceController{
		kubeClient:        kubeClient,
		restConfig:        restConfig,
		allowedNamespaces: allowedNamespaces,
		logger:            logger,
	}
}

// ==========================================
// SUMMARY STRUCTS FOR LIGHTWEIGHT LIST VIEWS
// ==========================================

type DeploymentSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	UpToDate  int32  `json:"upToDate"`
	Available int32  `json:"available"`
	Age       string `json:"age"`
}

type ServiceSummary struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Type       string `json:"type"`
	ClusterIP  string `json:"clusterIp"`
	ExternalIP string `json:"externalIp"`
	Ports      string `json:"ports"`
	Age        string `json:"age"`
}

type ConfigMapSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	KeysCount int    `json:"keysCount"`
	Age       string `json:"age"`
}

type SecretSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	KeysCount int    `json:"keysCount"`
	Age       string `json:"age"`
}

type NodeSummary struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Roles      string `json:"roles"`
	Version    string `json:"version"`
	InternalIP string `json:"internalIp"`
	Age        string `json:"age"`
}

type StatefulSetSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	Replicas  int32  `json:"replicas"`
	Age       string `json:"age"`
}

type DaemonSetSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Desired   int32  `json:"desired"`
	Current   int32  `json:"current"`
	Ready     string `json:"ready"`
	Available int32  `json:"available"`
	Age       string `json:"age"`
}

type CronJobSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Schedule  string `json:"schedule"`
	Suspend   bool   `json:"suspend"`
	Active    int    `json:"active"`
	Age       string `json:"age"`
}

type IngressSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Class     string `json:"class"`
	Hosts     string `json:"hosts"`
	Address   string `json:"address"`
	Ports     string `json:"ports"`
	Age       string `json:"age"`
}

type PVCSummary struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Status       string `json:"status"`
	Volume       string `json:"volume"`
	Capacity     string `json:"capacity"`
	StorageClass string `json:"storageClass"`
	Age          string `json:"age"`
}

type PVSummary struct {
	Name          string `json:"name"`
	Capacity      string `json:"capacity"`
	AccessModes   string `json:"accessModes"`
	ReclaimPolicy string `json:"reclaimPolicy"`
	Status        string `json:"status"`
	StorageClass  string `json:"storageClass"`
	Age           string `json:"age"`
}

// ==========================================
// DEPLOYMENTS
// ==========================================

func (r *k8sResourceController) GetDeployments(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")

	var rawDeployments []appsv1.Deployment
	if namespace == "" || namespace == "all" {
		nsList, err := r.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var names []string
		for _, ns := range nsList.Items {
			names = append(names, ns.Name)
		}
		validNs := common.Intersect(names, r.allowedNamespaces)
		for _, ns := range validNs {
			if list, err := r.kubeClient.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{}); err == nil {
				rawDeployments = append(rawDeployments, list.Items...)
			}
		}
	} else {
		if len(r.allowedNamespaces) > 0 && !slices.Contains(r.allowedNamespaces, namespace) {
			c.JSON(http.StatusForbidden, gin.H{"error": "namespace not allowed"})
			return
		}
		list, err := r.kubeClient.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rawDeployments = list.Items
	}

	var summaries []DeploymentSummary
	for _, dep := range rawDeployments {
		replicas := int32(1)
		if dep.Spec.Replicas != nil {
			replicas = *dep.Spec.Replicas
		}
		readyStr := fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, replicas)

		summaries = append(summaries, DeploymentSummary{
			Name:      dep.Name,
			Namespace: dep.Namespace,
			Ready:     readyStr,
			UpToDate:  dep.Status.UpdatedReplicas,
			Available: dep.Status.AvailableReplicas,
			Age:       formatAge(dep.CreationTimestamp.Time),
		})
	}

	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetDeploymentJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	obj, err := r.kubeClient.AppsV1().Deployments(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) UpdateDeploymentJSON(c *gin.Context) {
	ctx := c.Request.Context()
	ns, name := c.Param("namespace"), c.Param("name")
	var dep appsv1.Deployment
	if err := c.ShouldBindJSON(&dep); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
		return
	}
	if dep.Name != name || dep.Namespace != ns {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource metadata.name and metadata.namespace must match URL parameters"})
		return
	}
	res, err := r.kubeClient.AppsV1().Deployments(ns).Update(ctx, &dep, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (r *k8sResourceController) DeleteDeployment(c *gin.Context) {
	ctx := c.Request.Context()
	err := r.kubeClient.AppsV1().Deployments(c.Param("namespace")).Delete(ctx, c.Param("name"), metav1.DeleteOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deployment deleted successfully"})
}

// ==========================================
// SERVICES
// ==========================================

func (r *k8sResourceController) GetServices(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")

	var rawServices []corev1.Service
	if namespace == "" || namespace == "all" {
		nsList, err := r.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var names []string
		for _, ns := range nsList.Items {
			names = append(names, ns.Name)
		}
		validNs := common.Intersect(names, r.allowedNamespaces)
		for _, ns := range validNs {
			if list, err := r.kubeClient.CoreV1().Services(ns).List(ctx, metav1.ListOptions{}); err == nil {
				rawServices = append(rawServices, list.Items...)
			}
		}
	} else {
		if len(r.allowedNamespaces) > 0 && !slices.Contains(r.allowedNamespaces, namespace) {
			c.JSON(http.StatusForbidden, gin.H{"error": "namespace not allowed"})
			return
		}
		list, err := r.kubeClient.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rawServices = list.Items
	}

	var summaries []ServiceSummary
	for _, svc := range rawServices {
		var ports []string
		for _, p := range svc.Spec.Ports {
			ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
		}

		extIP := "<none>"
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			if svc.Status.LoadBalancer.Ingress[0].IP != "" {
				extIP = svc.Status.LoadBalancer.Ingress[0].IP
			} else {
				extIP = svc.Status.LoadBalancer.Ingress[0].Hostname
			}
		} else if len(svc.Spec.ExternalIPs) > 0 {
			extIP = strings.Join(svc.Spec.ExternalIPs, ",")
		}

		summaries = append(summaries, ServiceSummary{
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: extIP,
			Ports:      strings.Join(ports, ","),
			Age:        formatAge(svc.CreationTimestamp.Time),
		})
	}

	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetServiceJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	obj, err := r.kubeClient.CoreV1().Services(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) UpdateServiceJSON(c *gin.Context) {
	ctx := c.Request.Context()
	ns, name := c.Param("namespace"), c.Param("name")
	var svc corev1.Service
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
		return
	}
	if svc.Name != name || svc.Namespace != ns {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource metadata.name and metadata.namespace must match URL parameters"})
		return
	}
	res, err := r.kubeClient.CoreV1().Services(ns).Update(ctx, &svc, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (r *k8sResourceController) DeleteService(c *gin.Context) {
	ctx := c.Request.Context()
	err := r.kubeClient.CoreV1().Services(c.Param("namespace")).Delete(ctx, c.Param("name"), metav1.DeleteOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
}

// ==========================================
// CONFIGMAPS & SECRETS
// ==========================================

func (r *k8sResourceController) GetConfigMaps(c *gin.Context) {
	items := r.fetchCoreList(c, "configmaps")
	c.JSON(http.StatusOK, items)
}

func (r *k8sResourceController) GetConfigMapJSON(c *gin.Context) {
	r.getCoreJson(c, "configmaps")
}

func (r *k8sResourceController) UpdateConfigMapJSON(c *gin.Context) {
	r.updateCoreJson(c, "configmaps")
}

func (r *k8sResourceController) DeleteConfigMap(c *gin.Context) {
	r.deleteCoreResource(c, "configmaps")
}

func (r *k8sResourceController) GetSecrets(c *gin.Context) {
	items := r.fetchCoreList(c, "secrets")
	c.JSON(http.StatusOK, items)
}

func (r *k8sResourceController) GetSecretJSON(c *gin.Context) {
	r.getCoreJson(c, "secrets")
}

func (r *k8sResourceController) UpdateSecretJSON(c *gin.Context) {
	r.updateCoreJson(c, "secrets")
}

func (r *k8sResourceController) DeleteSecret(c *gin.Context) {
	r.deleteCoreResource(c, "secrets")
}

// ==========================================
// NODES (Cluster-wide)
// ==========================================

func (r *k8sResourceController) GetNodes(c *gin.Context) {
	ctx := c.Request.Context()
	nodeList, err := r.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var summaries []NodeSummary
	for _, node := range nodeList.Items {
		status := "NotReady"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}

		var roles []string
		for k := range node.Labels {
			if strings.HasPrefix(k, "node-role.kubernetes.io/") {
				roles = append(roles, strings.TrimPrefix(k, "node-role.kubernetes.io/"))
			}
		}
		if len(roles) == 0 {
			roles = []string{"worker"}
		}

		internalIP := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				internalIP = addr.Address
				break
			}
		}

		summaries = append(summaries, NodeSummary{
			Name:       node.Name,
			Status:     status,
			Roles:      strings.Join(roles, ","),
			Version:    node.Status.NodeInfo.KubeletVersion,
			InternalIP: internalIP,
			Age:        formatAge(node.CreationTimestamp.Time),
		})
	}

	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetNodeJSON(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
		return
	}
	obj, err := r.kubeClient.CoreV1().Nodes().Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

// ==========================================
// INTERNAL HELPERS
// ==========================================

func (r *k8sResourceController) fetchCoreList(c *gin.Context, kind string) interface{} {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")

	nsList, err := r.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []interface{}{}
	}
	var names []string
	for _, ns := range nsList.Items {
		names = append(names, ns.Name)
	}
	validNs := common.Intersect(names, r.allowedNamespaces)
	if namespace != "" && namespace != "all" {
		if !slices.Contains(validNs, namespace) {
			return []interface{}{}
		}
		validNs = []string{namespace}
	}

	if kind == "configmaps" {
		var summaries []ConfigMapSummary
		for _, ns := range validNs {
			if list, err := r.kubeClient.CoreV1().ConfigMaps(ns).List(ctx, metav1.ListOptions{}); err == nil {
				for _, cm := range list.Items {
					summaries = append(summaries, ConfigMapSummary{
						Name:      cm.Name,
						Namespace: cm.Namespace,
						KeysCount: len(cm.Data) + len(cm.BinaryData),
						Age:       formatAge(cm.CreationTimestamp.Time),
					})
				}
			}
		}
		return summaries
	} else if kind == "secrets" {
		var summaries []SecretSummary
		for _, ns := range validNs {
			if list, err := r.kubeClient.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{}); err == nil {
				for _, sec := range list.Items {
					summaries = append(summaries, SecretSummary{
						Name:      sec.Name,
						Namespace: sec.Namespace,
						Type:      string(sec.Type),
						KeysCount: len(sec.Data) + len(sec.StringData),
						Age:       formatAge(sec.CreationTimestamp.Time),
					})
				}
			}
		}
		return summaries
	}
	return []interface{}{}
}

func (r *k8sResourceController) getCoreJson(c *gin.Context, kind string) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource name is required"})
		return
	}

	var obj interface{}
	var err error
	switch kind {
	case "configmaps":
		obj, err = r.kubeClient.CoreV1().ConfigMaps(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "secrets":
		obj, err = r.kubeClient.CoreV1().Secrets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported resource type"})
		return
	}
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	switch resource := obj.(type) {
	case *corev1.ConfigMap:
		resource.ManagedFields = nil
	case *corev1.Secret:
		resource.ManagedFields = nil
	}
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) updateCoreJson(c *gin.Context, kind string) {
	ctx := c.Request.Context()
	ns, name := c.Param("namespace"), c.Param("name")

	if kind == "configmaps" {
		var cm corev1.ConfigMap
		if err := c.ShouldBindJSON(&cm); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
			return
		}
		if cm.Name != name || cm.Namespace != ns {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource metadata.name and metadata.namespace must match URL parameters"})
			return
		}
		res, err := r.kubeClient.CoreV1().ConfigMaps(ns).Update(ctx, &cm, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	} else {
		var sec corev1.Secret
		if err := c.ShouldBindJSON(&sec); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
			return
		}
		if sec.Name != name || sec.Namespace != ns {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource metadata.name and metadata.namespace must match URL parameters"})
			return
		}
		res, err := r.kubeClient.CoreV1().Secrets(ns).Update(ctx, &sec, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	}
}

func (r *k8sResourceController) deleteCoreResource(c *gin.Context, kind string) {
	ctx := c.Request.Context()
	ns := c.Param("namespace")
	name := c.Param("name")

	var err error
	if kind == "configmaps" {
		err = r.kubeClient.CoreV1().ConfigMaps(ns).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		err = r.kubeClient.CoreV1().Secrets(ns).Delete(ctx, name, metav1.DeleteOptions{})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource deleted successfully"})
}

func (r *k8sResourceController) ensureNamespaceAllowed(c *gin.Context, namespace string) bool {
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return false
	}
	if len(r.allowedNamespaces) > 0 && !slices.Contains(r.allowedNamespaces, namespace) {
		c.JSON(http.StatusForbidden, gin.H{"error": "namespace not allowed"})
		return false
	}
	return true
}

func writeJsonResponse(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, obj)
}

// ==========================================
// STATEFULSETS
// ==========================================

func (r *k8sResourceController) GetStatefulSets(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")
	var items []appsv1.StatefulSet
	if namespace != "" && namespace != "all" {
		if !r.ensureNamespaceAllowed(c, namespace) {
			return
		}
		list, err := r.kubeClient.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = list.Items
	} else {
		nsList, err := r.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, ns := range nsList.Items {
			if len(r.allowedNamespaces) > 0 && !slices.Contains(r.allowedNamespaces, ns.Name) {
				continue
			}
			list, err := r.kubeClient.AppsV1().StatefulSets(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, list.Items...)
		}
	}

	var summaries []StatefulSetSummary
	for _, ss := range items {
		replicas := int32(1)
		if ss.Spec.Replicas != nil {
			replicas = *ss.Spec.Replicas
		}
		readyStr := fmt.Sprintf("%d/%d", ss.Status.ReadyReplicas, replicas)
		summaries = append(summaries, StatefulSetSummary{
			Name:      ss.Name,
			Namespace: ss.Namespace,
			Ready:     readyStr,
			Replicas:  replicas,
			Age:       formatAge(ss.CreationTimestamp.Time),
		})
	}
	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetStatefulSetJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	obj, err := r.kubeClient.AppsV1().StatefulSets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) UpdateStatefulSetJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	var obj appsv1.StatefulSet
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
		return
	}
	if obj.Name != name || obj.Namespace != ns {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource metadata.name and metadata.namespace must match URL parameters"})
		return
	}
	res, err := r.kubeClient.AppsV1().StatefulSets(ns).Update(c.Request.Context(), &obj, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (r *k8sResourceController) DeleteStatefulSet(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	if err := r.kubeClient.AppsV1().StatefulSets(ns).Delete(c.Request.Context(), name, metav1.DeleteOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "StatefulSet deleted successfully"})
}

// ==========================================
// DAEMONSETS
// ==========================================

func (r *k8sResourceController) GetDaemonSets(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")
	var items []appsv1.DaemonSet
	if namespace != "" && namespace != "all" {
		if !r.ensureNamespaceAllowed(c, namespace) {
			return
		}
		list, err := r.kubeClient.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = list.Items
	} else {
		nsList, err := r.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, ns := range nsList.Items {
			if len(r.allowedNamespaces) > 0 && !slices.Contains(r.allowedNamespaces, ns.Name) {
				continue
			}
			list, err := r.kubeClient.AppsV1().DaemonSets(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, list.Items...)
		}
	}

	var summaries []DaemonSetSummary
	for _, ds := range items {
		readyStr := fmt.Sprintf("%d/%d", ds.Status.NumberReady, ds.Status.DesiredNumberScheduled)
		summaries = append(summaries, DaemonSetSummary{
			Name:      ds.Name,
			Namespace: ds.Namespace,
			Desired:   ds.Status.DesiredNumberScheduled,
			Current:   ds.Status.CurrentNumberScheduled,
			Ready:     readyStr,
			Available: ds.Status.NumberAvailable,
			Age:       formatAge(ds.CreationTimestamp.Time),
		})
	}
	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetDaemonSetJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	obj, err := r.kubeClient.AppsV1().DaemonSets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) UpdateDaemonSetJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	var obj appsv1.DaemonSet
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
		return
	}
	if obj.Name != name || obj.Namespace != ns {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource metadata.name and metadata.namespace must match URL parameters"})
		return
	}
	res, err := r.kubeClient.AppsV1().DaemonSets(ns).Update(c.Request.Context(), &obj, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (r *k8sResourceController) DeleteDaemonSet(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	if err := r.kubeClient.AppsV1().DaemonSets(ns).Delete(c.Request.Context(), name, metav1.DeleteOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "DaemonSet deleted successfully"})
}

// ==========================================
// CRONJOBS
// ==========================================

func (r *k8sResourceController) GetCronJobs(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")
	var items []batchv1.CronJob
	if namespace != "" && namespace != "all" {
		if !r.ensureNamespaceAllowed(c, namespace) {
			return
		}
		list, err := r.kubeClient.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = list.Items
	} else {
		nsList, err := r.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, ns := range nsList.Items {
			if len(r.allowedNamespaces) > 0 && !slices.Contains(r.allowedNamespaces, ns.Name) {
				continue
			}
			list, err := r.kubeClient.BatchV1().CronJobs(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, list.Items...)
		}
	}

	var summaries []CronJobSummary
	for _, cj := range items {
		active := 0
		if cj.Status.Active != nil {
			active = len(cj.Status.Active)
		}
		suspend := false
		if cj.Spec.Suspend != nil {
			suspend = *cj.Spec.Suspend
		}
		summaries = append(summaries, CronJobSummary{
			Name:      cj.Name,
			Namespace: cj.Namespace,
			Schedule:  cj.Spec.Schedule,
			Suspend:   suspend,
			Active:    active,
			Age:       formatAge(cj.CreationTimestamp.Time),
		})
	}
	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetCronJobJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	obj, err := r.kubeClient.BatchV1().CronJobs(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) UpdateCronJobJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	var obj batchv1.CronJob
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
		return
	}
	if obj.Name != name || obj.Namespace != ns {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource metadata.name and metadata.namespace must match URL parameters"})
		return
	}
	res, err := r.kubeClient.BatchV1().CronJobs(ns).Update(c.Request.Context(), &obj, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (r *k8sResourceController) DeleteCronJob(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	if err := r.kubeClient.BatchV1().CronJobs(ns).Delete(c.Request.Context(), name, metav1.DeleteOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "CronJob deleted successfully"})
}

// ==========================================
// INGRESSES
// ==========================================

func (r *k8sResourceController) GetIngresses(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")
	var items []networkingv1.Ingress
	if namespace != "" && namespace != "all" {
		if !r.ensureNamespaceAllowed(c, namespace) {
			return
		}
		list, err := r.kubeClient.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = list.Items
	} else {
		nsList, err := r.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, ns := range nsList.Items {
			if len(r.allowedNamespaces) > 0 && !slices.Contains(r.allowedNamespaces, ns.Name) {
				continue
			}
			list, err := r.kubeClient.NetworkingV1().Ingresses(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, list.Items...)
		}
	}

	var summaries []IngressSummary
	for _, ing := range items {
		var hosts []string
		for _, rule := range ing.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}

		var ports []string
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil {
						if path.Backend.Service.Port.Number > 0 {
							ports = append(ports, fmt.Sprintf("%d", path.Backend.Service.Port.Number))
						} else if path.Backend.Service.Port.Name != "" {
							ports = append(ports, path.Backend.Service.Port.Name)
						}
					}
				}
			}
		}

		address := "-"
		for _, lb := range ing.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				address = lb.IP
				break
			} else if lb.Hostname != "" {
				address = lb.Hostname
				break
			}
		}

		class := "-"
		if ing.Spec.IngressClassName != nil {
			class = *ing.Spec.IngressClassName
		}

		hostsStr := strings.Join(hosts, ", ")
		if hostsStr == "" {
			hostsStr = "*"
		}
		portsStr := strings.Join(ports, ", ")
		if portsStr == "" {
			portsStr = "-"
		}

		summaries = append(summaries, IngressSummary{
			Name:      ing.Name,
			Namespace: ing.Namespace,
			Class:     class,
			Hosts:     hostsStr,
			Address:   address,
			Ports:     portsStr,
			Age:       formatAge(ing.CreationTimestamp.Time),
		})
	}
	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetIngressJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	obj, err := r.kubeClient.NetworkingV1().Ingresses(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) UpdateIngressJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	var obj networkingv1.Ingress
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request body"})
		return
	}
	if obj.Name != name || obj.Namespace != ns {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource metadata.name and metadata.namespace must match URL parameters"})
		return
	}
	res, err := r.kubeClient.NetworkingV1().Ingresses(ns).Update(c.Request.Context(), &obj, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (r *k8sResourceController) DeleteIngress(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	if err := r.kubeClient.NetworkingV1().Ingresses(ns).Delete(c.Request.Context(), name, metav1.DeleteOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Ingress deleted successfully"})
}

// ==========================================
// PERSISTENT VOLUME CLAIMS
// ==========================================

func (r *k8sResourceController) GetPVCs(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.DefaultQuery("namespace", "")
	var items []corev1.PersistentVolumeClaim
	if namespace != "" && namespace != "all" {
		if !r.ensureNamespaceAllowed(c, namespace) {
			return
		}
		list, err := r.kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = list.Items
	} else {
		nsList, err := r.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, ns := range nsList.Items {
			if len(r.allowedNamespaces) > 0 && !slices.Contains(r.allowedNamespaces, ns.Name) {
				continue
			}
			list, err := r.kubeClient.CoreV1().PersistentVolumeClaims(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, list.Items...)
		}
	}

	var summaries []PVCSummary
	for _, pvc := range items {
		capacity := "-"
		if pvc.Status.Capacity != nil {
			if val, ok := pvc.Status.Capacity["storage"]; ok {
				capacity = val.String()
			}
		}
		storageClass := "-"
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		}
		summaries = append(summaries, PVCSummary{
			Name:         pvc.Name,
			Namespace:    pvc.Namespace,
			Status:       string(pvc.Status.Phase),
			Volume:       pvc.Spec.VolumeName,
			Capacity:     capacity,
			StorageClass: storageClass,
			Age:          formatAge(pvc.CreationTimestamp.Time),
		})
	}
	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetPVCJSON(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	obj, err := r.kubeClient.CoreV1().PersistentVolumeClaims(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) DeletePVC(c *gin.Context) {
	ns, name := c.Param("namespace"), c.Param("name")
	if !r.ensureNamespaceAllowed(c, ns) {
		return
	}
	if err := r.kubeClient.CoreV1().PersistentVolumeClaims(ns).Delete(c.Request.Context(), name, metav1.DeleteOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "PVC deleted successfully"})
}

// ==========================================
// PERSISTENT VOLUMES (cluster-wide)
// ==========================================

func (r *k8sResourceController) GetPVs(c *gin.Context) {
	list, err := r.kubeClient.CoreV1().PersistentVolumes().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var summaries []PVSummary
	for _, pv := range list.Items {
		capacity := "-"
		if pv.Spec.Capacity != nil {
			if val, ok := pv.Spec.Capacity["storage"]; ok {
				capacity = val.String()
			}
		}
		accessModes := "-"
		if len(pv.Spec.AccessModes) > 0 {
			var modes []string
			for _, m := range pv.Spec.AccessModes {
				modes = append(modes, string(m))
			}
			accessModes = strings.Join(modes, ", ")
		}
		storageClass := "-"
		if pv.Spec.StorageClassName != "" {
			storageClass = pv.Spec.StorageClassName
		}
		summaries = append(summaries, PVSummary{
			Name:          pv.Name,
			Capacity:      capacity,
			AccessModes:   accessModes,
			ReclaimPolicy: string(pv.Spec.PersistentVolumeReclaimPolicy),
			Status:        string(pv.Status.Phase),
			StorageClass:  storageClass,
			Age:           formatAge(pv.CreationTimestamp.Time),
		})
	}
	c.JSON(http.StatusOK, summaries)
}

func (r *k8sResourceController) GetPVJSON(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PersistentVolume name is required"})
		return
	}
	obj, err := r.kubeClient.CoreV1().PersistentVolumes().Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "PersistentVolume not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	obj.ManagedFields = nil
	writeJsonResponse(c, obj)
}

func (r *k8sResourceController) DeletePV(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "PersistentVolume name is required",
		})
		return
	}

	_, err := r.kubeClient.CoreV1().PersistentVolumes().Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "PersistentVolume not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = r.kubeClient.CoreV1().PersistentVolumes().Delete(
		ctx,
		name,
		metav1.DeleteOptions{},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "PersistentVolume not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "PersistentVolume deletion initiated successfully",
		"name":    name,
	})
}
