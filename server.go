package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/dattc/go-though-storm/api/automation"
	"github.com/dattc/go-though-storm/api/automation/jobs"
	"github.com/dattc/go-though-storm/api/common"
	"github.com/dattc/go-though-storm/api/config"
	"github.com/dattc/go-though-storm/api/db"
	"github.com/dattc/go-though-storm/api/logging"
	"github.com/dattc/go-though-storm/api/router"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	// Setup structured logger
	loggers, err := logging.InitLoggers("./logs")
	if err != nil {
		fmt.Printf("Fatal: failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	slog.SetDefault(loggers.App)

	// Create root context for background routines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init db
	dsn := "./data/app.db?_journal_mode=WAL&_busy_timeout=10000&_synchronous=NORMAL"
	database, err := db.Init(dsn, automation.DefaultJobs, loggers.App)
	if err != nil {
		logger.Error("failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}
	defer database.Close()

	loggers.App.Info("database initialized successfully")

	// Init engin and db repo
	repo := db.NewRepository(database)

	// Retrieve ENV Variables
	if err := godotenv.Load(); err != nil {
		loggers.App.Info("Notice: No .env file found, using system environment variables")
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		loggers.App.Error("Failed to load config.yaml", slog.Any("error", err))
		os.Exit(1)
	}

	// Load k8s config
	var k8sConfig *rest.Config
	if cfg.Kube.KubeconfigPath != "" || common.FileExists(cfg.Kube.KubeconfigPath) {
		var overrides *clientcmd.ConfigOverrides
		if cfg.Kube.Context != "" {
			overrides = &clientcmd.ConfigOverrides{
				CurrentContext: cfg.Kube.Context,
			}
		}
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: cfg.Kube.KubeconfigPath},
			overrides,
		)
		k8sConfig, err = clientConfig.ClientConfig()
	} else {
		k8sConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		loggers.App.Error("Failed to establish Kubernetes config", slog.Any("error", err))
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		loggers.App.Error("Failed to initialize Kubernetes client", slog.Any("error", err))
		os.Exit(1)
	}

	metricsClient, err := metricsv1beta1.NewForConfig(k8sConfig)
	if err != nil {
		loggers.App.Warn("Failed to initialize Metrics client", slog.Any("error", err))
	}

	// Register automation jobs
	registry := automation.DefaultRegistry
	if registry == nil {
		registry = automation.NewRegistry()
	}
	jobs.RegisterJVMAutoRestart(registry, kubeClient, metricsClient, k8sConfig)
	engine := automation.NewEngine(repo, 10, loggers.App, registry, cfg.AutomationExtraConfigs)
	engine.Start(ctx, 3)
	loggers.App.Info("Automation engine started", slog.Int("workers", 3))

	router := router.NewRouter(
		kubeClient,
		metricsClient,
		k8sConfig,
		cfg.Namespace.Allowed,
		cfg.EditableFiles,
		cfg.Auth,
		cfg.Jobs,
		loggers,
		repo,
		engine,
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
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: app,
	}

	go func() {
		logger.Info("Server starting", slog.String("address", serverAddr), slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server startup failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server and flushing worker queue...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
	}

	engine.Stop()

	logger.Info("Server exiting gracefully")
}
