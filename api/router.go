package api

import (
	"net/http"
	"slices"
	"strings"

	"github.com/dattc/go-though-storm/api/controller"
	"github.com/dattc/go-though-storm/api/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Router interface {
	SetupRouter(router *gin.Engine)
}

type router struct {
	kubeClient          kubernetes.Interface
	metricsClient       metricsv1beta1.Interface
	restConfig          *rest.Config
	allowedNamespaces   []string
	editableFilesConfig models.EditableFilesConfig
	auth                models.AuthConfig
}

// NewRouter initializes the router dependency injection
func NewRouter(kubeClient kubernetes.Interface,
	metricsClient metricsv1beta1.Interface,
	restConfig *rest.Config,
	allowedNamespaces []string,
	editableFilesConfig models.EditableFilesConfig,
	auth models.AuthConfig) Router {

	return &router{
		kubeClient:          kubeClient,
		metricsClient:       metricsClient,
		restConfig:          restConfig,
		allowedNamespaces:   allowedNamespaces,
		editableFilesConfig: editableFilesConfig,
		auth:                auth,
	}
}

func (r *router) NamespaceGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(r.allowedNamespaces) == 0 {
			c.Next()
			return
		}

		targetNs := c.Param("namespace")
		if targetNs == "" {
			targetNs = c.Query("namespace")
		}

		// If a namespace was supplied, verify it exists inside allowedNamespaces
		if targetNs != "" && !slices.Contains(r.allowedNamespaces, targetNs) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access to namespace '" + targetNs + "' is restricted.",
			})
			return
		}

		c.Next()
	}
}

func (r *router) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid authorization header"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &models.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return r.auth.JwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

// SetupRouter configuration router information
func (r *router) SetupRouter(router *gin.Engine) {
	// Initialize Controllers
	healthController := controller.NewHealthController()
	podController := controller.NewPodController(r.kubeClient, r.metricsClient, r.restConfig, r.allowedNamespaces, r.editableFilesConfig)
	authController := controller.NewAuthController(r.auth)
	// Setup route group for the API
	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// Health check
			v1.GET("/healthcheck", healthController.Check)
			v1.GET("/namespaces", podController.GetNamespaces)
			v1.POST("/login", authController.LoginHandler)

			pods := v1.Group("/pods", r.NamespaceGuardMiddleware())
			{
				pods.GET("", podController.GetPods)
				pods.GET("/:namespace/:pod/logs", podController.GetPodLogs)
				pods.GET("/:namespace/:pod/containers/:container/files", podController.GetContainerFiles)
				pods.PUT("/:namespace/:pod/containers/:container/file", podController.SaveContainerFileContent)
				pods.GET("/:namespace/:pod/containers/:container/fileContent", podController.GetContainerFileContent)
				pods.GET("/:namespace/:pod/containers/:container/download", podController.DownloadContainerFile)
			}
		}
	}
}
