package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/dattc/go-though-storm/api/automation"
	"github.com/dattc/go-though-storm/api/controller"
	"github.com/dattc/go-though-storm/api/db"
	"github.com/dattc/go-though-storm/api/logging"
	middleware "github.com/dattc/go-though-storm/api/middleware"
	"github.com/dattc/go-though-storm/api/models"
)

type Router interface {
	SetupRouter(router *gin.Engine)
}

type Config struct {
	KubeClient          kubernetes.Interface
	MetricsClient       metricsv1beta1.Interface
	RestConfig          *rest.Config
	AllowedNamespaces   []string
	EditableFilesConfig models.EditableFilesConfig
	Auth                models.AuthConfig
	Jobs                []models.JobConfig
	Loggers             logging.Loggers
	Repo                *db.Repository
	Engine              *automation.Engine
}

type router struct {
	kubeClient          kubernetes.Interface
	metricsClient       metricsv1beta1.Interface
	restConfig          *rest.Config
	allowedNamespaces   []string
	editableFilesConfig models.EditableFilesConfig
	auth                models.AuthConfig
	jobs                []models.JobConfig
	loggers             logging.Loggers
	repo                *db.Repository
	engine              *automation.Engine
}

func NewRouter(
	kubeClient kubernetes.Interface,
	metricsClient metricsv1beta1.Interface,
	restConfig *rest.Config,
	allowedNamespaces []string,
	editableFilesConfig models.EditableFilesConfig,
	auth models.AuthConfig,
	jobs []models.JobConfig,
	loggers *logging.Loggers,
	repo *db.Repository,
	engine *automation.Engine,
) Router {
	return &router{
		kubeClient:          kubeClient,
		metricsClient:       metricsClient,
		restConfig:          restConfig,
		allowedNamespaces:   allowedNamespaces,
		editableFilesConfig: editableFilesConfig,
		auth:                auth,
		jobs:                jobs,
		loggers:             *loggers,
		repo:                repo,
		engine:              engine,
	}
}

type Route struct {
	Method  string
	Path    string
	Action  string
	Handler gin.HandlerFunc
}

func registerSecured(rg *gin.RouterGroup, routes []Route) {
	for _, r := range routes {
		rg.Handle(r.Method, r.Path, middleware.RequireAction(r.Action, ActionRolePermissions), r.Handler)
	}
}

func (r *router) SetupRouter(engine *gin.Engine) {
	healthCtrl := controller.NewHealthController()
	authCtrl := controller.NewAuthController(r.auth.Local, r.loggers.App)
	podCtrl := controller.NewPodController(
		r.kubeClient, r.metricsClient, r.restConfig,
		r.allowedNamespaces, r.editableFilesConfig, r.jobs, r.loggers.App,
	)
	jobCtrl := controller.NewJobController(r.repo, r.engine, r.loggers.App)

	resourceCtrl := controller.NewK8sResourceController(
		r.kubeClient, r.restConfig, r.allowedNamespaces, r.loggers.App,
	)

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/healthcheck", healthCtrl.Check)
		v1.POST("/login", authCtrl.LoginHandler)

		authGroup := v1.Group("", middleware.AuditMiddleware(r.loggers.Audit))
		authGroup.Use(middleware.AuthMiddleware(r.auth, r.loggers.App))
		{
			authGroup.GET("/namespaces", middleware.RequireAction(ActionNamespaceList, ActionRolePermissions), podCtrl.GetNamespaces)

			registerSecured(authGroup.Group("/automations"), []Route{
				{http.MethodGet, "", ActionAutomationList, jobCtrl.GetAutomationJobs},
				{http.MethodGet, "/:automation_code", ActionAutomationRead, jobCtrl.GetAutomationJob},
				{http.MethodPost, "/:automation_code/run", ActionAutomationRun, jobCtrl.ManualTriggerJob},
				{http.MethodPost, "/webhooks/alermanager", ActionAutomationWebhook, jobCtrl.WebhooksHandler},
			})

			registerSecured(authGroup.Group("/executions"), []Route{
				{http.MethodGet, "", ActionExecutionList, jobCtrl.ListExecutions},
				{http.MethodGet, "/:id/steps", ActionExecutionStepsRead, jobCtrl.GetExecutionSteps},
				{http.MethodGet, "/:id", ActionExecutionRead, jobCtrl.GetExecution},
				{http.MethodGet, "/:id/status", ActionExecutionStatusRead, jobCtrl.GetExecutionStatus},
				{http.MethodGet, "/:id/logs", ActionExecutionLogsRead, jobCtrl.GetExecutionLogs},
			})

			// Pod Routes
			pods := authGroup.Group("/pods", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				pods.GET("", middleware.RequireAction(ActionPodList, ActionRolePermissions), podCtrl.GetPods)
				pods.GET("/:namespace/:pod/logs", middleware.RequireAction(ActionPodLogsRead, ActionRolePermissions), podCtrl.GetPodLogs)

				registerSecured(pods.Group("/:namespace/:pod/containers/:container"), []Route{
					{http.MethodGet, "/jobs", ActionContainerJobsList, podCtrl.GetAvailableJobs},
					{http.MethodPost, "/job/:id/run", ActionContainerJobRun, podCtrl.RunJob},
					{http.MethodGet, "/files", ActionContainerFilesList, podCtrl.GetContainerFiles},
					{http.MethodPut, "/file", ActionContainerFileWrite, podCtrl.SaveContainerFileContent},
					{http.MethodGet, "/fileContent", ActionContainerFileRead, podCtrl.GetContainerFileContent},
					{http.MethodGet, "/download", ActionContainerFileDownload, podCtrl.DownloadContainerFile},
				})
			}

			// 1. Deployments Routes
			deployments := authGroup.Group("/deployments", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				deployments.GET("", middleware.RequireAction(ActionDeploymentList, ActionRolePermissions), resourceCtrl.GetDeployments)
				deployments.GET("/:namespace/:name/json", middleware.RequireAction(ActionDeploymentJsonRead, ActionRolePermissions), resourceCtrl.GetDeploymentJSON)
				deployments.PUT("/:namespace/:name/json", middleware.RequireAction(ActionDeploymentJsonUpdate, ActionRolePermissions), resourceCtrl.UpdateDeploymentJSON)
				deployments.DELETE("/:namespace/:name", middleware.RequireAction(ActionDeploymentDelete, ActionRolePermissions), resourceCtrl.DeleteDeployment)
			}

			// 2. StatefulSets Routes
			statefulsets := authGroup.Group("/statefulsets", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				statefulsets.GET("", middleware.RequireAction("statefulset:list", ActionRolePermissions), resourceCtrl.GetStatefulSets)
				statefulsets.GET("/:namespace/:name/json", middleware.RequireAction("statefulset:json:read", ActionRolePermissions), resourceCtrl.GetStatefulSetJSON)
				statefulsets.PUT("/:namespace/:name/json", middleware.RequireAction("statefulset:json:update", ActionRolePermissions), resourceCtrl.UpdateStatefulSetJSON)
				statefulsets.DELETE("/:namespace/:name", middleware.RequireAction("statefulset:delete", ActionRolePermissions), resourceCtrl.DeleteStatefulSet)
			}

			// 3. DaemonSets Routes
			daemonsets := authGroup.Group("/daemonsets", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				daemonsets.GET("", middleware.RequireAction("daemonset:list", ActionRolePermissions), resourceCtrl.GetDaemonSets)
				daemonsets.GET("/:namespace/:name/json", middleware.RequireAction("daemonset:json:read", ActionRolePermissions), resourceCtrl.GetDaemonSetJSON)
				daemonsets.PUT("/:namespace/:name/json", middleware.RequireAction("daemonset:json:update", ActionRolePermissions), resourceCtrl.UpdateDaemonSetJSON)
				daemonsets.DELETE("/:namespace/:name", middleware.RequireAction("daemonset:delete", ActionRolePermissions), resourceCtrl.DeleteDaemonSet)
			}

			// 4. CronJobs Routes
			cronjobs := authGroup.Group("/cronjobs", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				cronjobs.GET("", middleware.RequireAction("cronjob:list", ActionRolePermissions), resourceCtrl.GetCronJobs)
				cronjobs.GET("/:namespace/:name/json", middleware.RequireAction("cronjob:json:read", ActionRolePermissions), resourceCtrl.GetCronJobJSON)
				cronjobs.PUT("/:namespace/:name/json", middleware.RequireAction("cronjob:json:update", ActionRolePermissions), resourceCtrl.UpdateCronJobJSON)
				cronjobs.DELETE("/:namespace/:name", middleware.RequireAction("cronjob:delete", ActionRolePermissions), resourceCtrl.DeleteCronJob)
			}

			// 5. Services Routes
			services := authGroup.Group("/services", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				services.GET("", middleware.RequireAction(ActionServiceList, ActionRolePermissions), resourceCtrl.GetServices)
				services.GET("/:namespace/:name/json", middleware.RequireAction(ActionServiceJsonRead, ActionRolePermissions), resourceCtrl.GetServiceJSON)
				services.PUT("/:namespace/:name/json", middleware.RequireAction("service:json:update", ActionRolePermissions), resourceCtrl.UpdateServiceJSON)
				services.DELETE("/:namespace/:name", middleware.RequireAction(ActionServiceDelete, ActionRolePermissions), resourceCtrl.DeleteService)
			}

			// 6. Ingresses Routes
			ingresses := authGroup.Group("/ingresses", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				ingresses.GET("", middleware.RequireAction("ingress:list", ActionRolePermissions), resourceCtrl.GetIngresses)
				ingresses.GET("/:namespace/:name/json", middleware.RequireAction("ingress:json:read", ActionRolePermissions), resourceCtrl.GetIngressJSON)
				ingresses.PUT("/:namespace/:name/json", middleware.RequireAction("ingress:json:update", ActionRolePermissions), resourceCtrl.UpdateIngressJSON)
				ingresses.DELETE("/:namespace/:name", middleware.RequireAction("ingress:delete", ActionRolePermissions), resourceCtrl.DeleteIngress)
			}

			// 7. PersistentVolumeClaims (PVC) Routes
			pvcs := authGroup.Group("/pvc", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				pvcs.GET("", middleware.RequireAction("pvc:list", ActionRolePermissions), resourceCtrl.GetPVCs)
				pvcs.GET("/:namespace/:name/json", middleware.RequireAction("pvc:json:read", ActionRolePermissions), resourceCtrl.GetPVCJSON)
				pvcs.DELETE("/:namespace/:name", middleware.RequireAction("pvc:delete", ActionRolePermissions), resourceCtrl.DeletePVC)
			}

			// 8. PersistentVolumes (PV) Routes (Cluster-wide)
			pvs := authGroup.Group("/pv")
			{
				pvs.GET("", middleware.RequireAction("pv:list", ActionRolePermissions), resourceCtrl.GetPVs)
				pvs.GET("/:name/json", middleware.RequireAction("pv:json:read", ActionRolePermissions), resourceCtrl.GetPVJSON)
				pvs.DELETE("/:name", middleware.RequireAction("pv:delete", ActionRolePermissions), resourceCtrl.DeletePV)
			}

			// 9. ConfigMaps Routes
			configmaps := authGroup.Group("/configmaps", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				configmaps.GET("", middleware.RequireAction(ActionConfigMapList, ActionRolePermissions), resourceCtrl.GetConfigMaps)
				configmaps.GET("/:namespace/:name/json", middleware.RequireAction(ActionConfigMapJsonRead, ActionRolePermissions), resourceCtrl.GetConfigMapJSON)
				configmaps.PUT("/:namespace/:name/json", middleware.RequireAction(ActionConfigMapJsonUpdate, ActionRolePermissions), resourceCtrl.UpdateConfigMapJSON)
				configmaps.DELETE("/:namespace/:name", middleware.RequireAction(ActionConfigMapDelete, ActionRolePermissions), resourceCtrl.DeleteConfigMap)
			}

			// 10. Secrets Routes
			secrets := authGroup.Group("/secrets", middleware.NamespaceGuardMiddleware(r.allowedNamespaces))
			{
				secrets.GET("", middleware.RequireAction(ActionSecretList, ActionRolePermissions), resourceCtrl.GetSecrets)
				secrets.GET("/:namespace/:name/json", middleware.RequireAction(ActionSecretJsonRead, ActionRolePermissions), resourceCtrl.GetSecretJSON)
				secrets.PUT("/:namespace/:name/json", middleware.RequireAction(ActionSecretJsonUpdate, ActionRolePermissions), resourceCtrl.UpdateSecretJSON)
				secrets.DELETE("/:namespace/:name", middleware.RequireAction(ActionSecretDelete, ActionRolePermissions), resourceCtrl.DeleteSecret)
			}

			// 11. Cluster-wide Nodes Routes
			nodes := authGroup.Group("/nodes")
			{
				nodes.GET("", middleware.RequireAction(ActionNodeList, ActionRolePermissions), resourceCtrl.GetNodes)
				nodes.GET("/:name/json", middleware.RequireAction(ActionNodeJsonRead, ActionRolePermissions), resourceCtrl.GetNodeJSON)
			}
		}
	}
}
