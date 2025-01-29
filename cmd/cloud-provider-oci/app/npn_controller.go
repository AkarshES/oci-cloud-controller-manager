package app

import (
	"context"
	"fmt"
	"github.com/go-logr/zapr"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	cloudControllerManagerConfig "k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	npnv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/v1beta1"
	"github.com/oracle/oci-cloud-controller-manager/controllers"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

const npnController = "npn-controller"

// NpnController represents the NPN controller.
type NpnController struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	mgr    ctrl.Manager
}

func NewNpnController(ctx context.Context, logger *zap.SugaredLogger, mgr ctrl.Manager) (Controller, error) {
	return &NpnController{
		ctx:    ctx,
		logger: logger,
		mgr:    mgr,
	}, nil
}

// Run sets up the NPN controller.
func (c *NpnController) Run(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {

	return c.initializeNPNController(mgr)
}

func (c *NpnController) AddHealthCheck(mgr ctrl.Manager, extraArgs ...interface{}) error {
	c.logger.Infof(npnController, mgr)

	if err := mgr.AddHealthzCheck(npnController, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add healthz check: %v", err)
		client.SetHealthStatusForController(npnController, "failed")
		return err
	}
	client.SetHealthStatusForController(npnController, "ok")

	if err := mgr.AddReadyzCheck(npnController, func(req *http.Request) error {
		return npnControllerReadyCheck(req, mgr)
	}); err != nil {
		c.logger.Errorf("Failed to add readyz check: %v", err)
		client.SetReadinessForController(npnController, "failed")
		return err
	}
	client.SetReadinessForController(npnController, "ok")

	return nil
}

func (c *NpnController) initializeNPNController(mgr ctrl.Manager) error {
	c.logger.Info("Initializing NPN Controller...")

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

	c.logger.Info("NPN controller is enabled.")
	c.logger = c.logger.With(zap.String("component", "npn-controller"))
	ctrl.SetLogger(zapr.NewLogger(c.logger.Desugar()))

	utilruntime.Must(npnv1beta1.AddToScheme(scheme))
	if err = (&controllers.NativePodNetworkReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		MetricPusher: metricPusher,
		OCIClient:    ociClient,
		Config:       cfg,
	}).SetupWithManager(mgr); err != nil {
		c.logger.With(zap.Error(err)).Error("unable to create controller", "controller", "NativePodNetwork")
		return fmt.Errorf("failed to setup NPN controller: %w", err)
	}
	return nil
}

// npnControllerReadyCheck ensures npn is functional and talks to api-server
func npnControllerReadyCheck(req *http.Request, mgr ctrl.Manager) error {
	config := mgr.GetConfig()

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return errors.Errorf("failed to create discovery client: %v", err)
	}

	_, err = discoveryClient.ServerGroups()
	if err != nil {
		return errors.Errorf("failed to query API server: %v", err)
	}

	return nil
}
