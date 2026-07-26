package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dattc/go-though-storm/api"
	"github.com/dattc/go-though-storm/api/common"
	"github.com/dattc/go-though-storm/api/config"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: No .env file found, using system environment variables")
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config.yaml: %v", err)
	}

	var k8sConfig *rest.Config

	if cfg.Kube.KubeconfigPath != "" || common.FileExists(cfg.Kube.KubeconfigPath) {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kube.KubeconfigPath)
	} else {
		k8sConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatalf("Failed to establish Kubernetes config: %v", err)
	}

	// 4. Create Kubernetes clients
	kubeClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	metricsClient, err := metricsv1beta1.NewForConfig(k8sConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize Metrics client: %v", err)
	}

	router := api.NewRouter(
		kubeClient,
		metricsClient,
		k8sConfig,
		cfg.Namespace.Allowed,
		cfg.EditableFiles,
		cfg.Auth,
	)

	gin.SetMode(cfg.Server.Mode)
	app := gin.Default()

	app.Use(static.Serve("/", static.LocalFile("./static", true)))

	router.SetupRouter(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", cfg.Server.Port)
	}

	serverAddr := "0.0.0.0:" + port
	log.Printf("Server starting on %s", serverAddr)

	if err := app.Run(serverAddr); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
