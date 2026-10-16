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
	// engine.Use(middleware.AuditMiddleware(r.loggers.Audit))

	healthCtrl := controller.NewHealthController()
	authCtrl := controller.NewAuthController(r.auth.Local, r.loggers.App)
	podCtrl := controller.NewPodController(
		r.kubeClient, r.metricsClient, r.restConfig,
		r.allowedNamespaces, r.editableFilesConfig, r.jobs, r.loggers.App,
	)
	jobCtrl := controller.NewJobController(r.repo, r.engine, r.loggers.App)

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
		}
	}
}
