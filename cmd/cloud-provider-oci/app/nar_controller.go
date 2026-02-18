package app

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	cloudControllerManagerConfig "k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/oracle/oci-cloud-controller-manager/controllers"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

// Node auto repair controller
const narController = "nar-controller"

// NorController represents the Nor controller.
type NarController struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	mgr    ctrl.Manager
}

// NewNorController creates a new Nor controller instance.
func NewNarController(ctx context.Context, logger *zap.SugaredLogger, mgr ctrl.Manager) (Controller, error) {
	return &NarController{
		ctx:    ctx,
		logger: logger,
		mgr:    mgr,
	}, nil
}

// Run starts the Nor controller.
func (c *NarController) Run(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	return c.initializeNARController(mgr, config, options)
}

func (c *NarController) AddHealthCheck(mgr ctrl.Manager, extraArgs ...interface{}) error {
	c.logger.Infof(narController, mgr)

	if err := mgr.AddHealthzCheck(narController, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add healthz check: %v", err)
		client.SetHealthStatusForController(narController, "failed")
		return err
	}
	client.SetHealthStatusForController(narController, "ok")

	if err := mgr.AddReadyzCheck(narController, func(req *http.Request) error {
		return npnControllerReadyCheck(req, mgr)
	}); err != nil {
		c.logger.Errorf("Failed to add readyz check: %v", err)
		client.SetReadinessForController(narController, "failed")
		return err
	}
	client.SetReadinessForController(narController, "ok")

	return nil
}

func (c *NarController) initializeNARController(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	logger := c.logger.With(zap.String("component", "nar-controller"))
	ctrl.SetLogger(zapr.NewLogger(logger.Desugar()))

	controllerManagerSetupLog.Info("Starting manager")
	c.logger.Info("NAR controller is enabled.")
	master := options.Master
	kubeConfigPath := options.Generic.ClientConnection.Kubeconfig
	c.logger.Info("master is ", master, "kubeconfig is", kubeConfigPath)

	kubeClient := config.Client
	configPath, ok := os.LookupEnv("CONFIG_YAML_FILENAME")
	if !ok {
		configPath = configFilePath
	}
	cfg := providercfg.GetConfig(c.logger, configPath)
	ociClient := getOCIClient(c.logger, cfg)

	metricPusher, err := metrics.NewMetricPusher(c.logger)
	if err != nil {
		c.logger.With("error", err).Error("metrics collection could not be enabled")
		// disable metrics
		metricPusher = nil
	}

	if err = (&controllers.NodeAutoRepairReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		OCIClient:    ociClient,
		KubeClient:   kubeClient,
		Config:       cfg,
		MetricPusher: metricPusher,
	}).SetupWithManager(mgr); err != nil {
		c.logger.With(zap.Error(err)).Error(err, "unable to create controller", "controller", "nor-controller")
		return fmt.Errorf("failed to setup NOR controller: %w", err)
	}

	return nil
}
