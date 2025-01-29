package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	cloudControllerManagerConfig "k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	norv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/node-cycling/v1beta1"
	"github.com/oracle/oci-cloud-controller-manager/controllers"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

const norController = "nor-controller"

// NorController represents the Nor controller.
type NorController struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	mgr    ctrl.Manager
}

// NewNorController creates a new Nor controller instance.
func NewNorController(ctx context.Context, logger *zap.SugaredLogger, mgr ctrl.Manager) (Controller, error) {
	return &NorController{
		ctx:    ctx,
		logger: logger,
		mgr:    mgr,
	}, nil
}

// Run starts the Nor controller.
func (c *NorController) Run(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	return c.initializeNORController(mgr, config, options)
}

func (c *NorController) AddHealthCheck(mgr ctrl.Manager, extraArgs ...interface{}) error {
	c.logger.Infof(norController, mgr)

	if err := mgr.AddHealthzCheck(norController, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add healthz check: %v", err)
		client.SetHealthStatusForController(norController, "failed")
		return err
	}
	client.SetHealthStatusForController(norController, "ok")

	if err := mgr.AddReadyzCheck(norController, func(req *http.Request) error {
		return npnControllerReadyCheck(req, mgr)
	}); err != nil {
		c.logger.Errorf("Failed to add readyz check: %v", err)
		client.SetReadinessForController(norController, "failed")
		return err
	}
	client.SetReadinessForController(norController, "ok")

	return nil
}

func (c *NorController) initializeNORController(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	norControllerBvrLimit := oci.GetIntegerFromEnv(c.logger, "NOR_CONTROLLER_BVR_RATE_LIMIT_RPM", 10)
	norControllerRebootLimit := oci.GetIntegerFromEnv(c.logger, "NOR_CONTROLLER_REBOOT_RATE_LIMIT_RPM", 10)
	utilruntime.Must(norv1beta1.AddToScheme(scheme))
	logger := c.logger.With(zap.String("component", "nor-controller"))
	ctrl.SetLogger(zapr.NewLogger(logger.Desugar()))

	controllerManagerSetupLog.Info("starting manager")
	c.logger.Info("NOR controller is enabled.")
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

	norControllerBvrLimiter := generateRateLimiter(norControllerBvrLimit)
	norControllerRebootLimiter := generateRateLimiter(norControllerRebootLimit)
	if err = (&controllers.NodeOperationRuleReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		OCIClient:         ociClient,
		KubeClient:        kubeClient,
		Config:            cfg,
		MetricPusher:      metricPusher,
		BvrRateLimiter:    norControllerBvrLimiter,
		RebootRateLimiter: norControllerRebootLimiter,
	}).SetupWithManager(mgr); err != nil {
		c.logger.With(zap.Error(err)).Error(err, "unable to create controller", "controller", "nor-controller")
		return fmt.Errorf("failed to setup NOR controller: %w", err)
	}

	return nil
}

func getRateLimitValue(rateLimitFromLimits int) int {
	if rateLimitFromLimits <= 0 {
		return norDefaultRateLimit
	}
	return rateLimitFromLimits
}

func generateRateLimiter(rateLimitFromLimits int) *rate.Limiter {
	rateLimit := getRateLimitValue(rateLimitFromLimits)
	return rate.NewLimiter(rate.Every(time.Minute/time.Duration(rateLimit)), norMaxBurstTokens)
}
