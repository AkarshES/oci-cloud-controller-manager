package app

import (
	"context"
	"net/http"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	cloudControllerManagerConfig "k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

const metricsController = "metrics-controller"

var metricsEndpoint string

type MetricsController struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	mgr    ctrl.Manager
}

// NewMetricsController creates a new MetricsController instance.
func NewMetricsController(ctx context.Context, logger *zap.SugaredLogger, mgr ctrl.Manager) (Controller, error) {
	return &MetricsController{
		ctx:    ctx,
		logger: logger,
	}, nil
}

func (c *MetricsController) Run(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	if err := c.RunMetricsController(); err != nil {
		return err
	}
	return nil
}

func (c *MetricsController) AddHealthCheck(mgr ctrl.Manager, extraArgs ...interface{}) error {
	c.logger.Infof(metricsController, mgr)

	// Add Liveness Probe (Ensures CCM process is running)
	if err := mgr.AddHealthzCheck(metricsController, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add healthz check: %v", err)
		client.SetHealthStatusForController(metricsController, "failed")
		return err
	}
	client.SetHealthStatusForController(metricsController, "ok")

	if err := mgr.AddReadyzCheck(metricsController, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add readyz check: %v", err)
		client.SetReadinessForController(metricsController, "failed")
		return err
	}
	client.SetReadinessForController(metricsController, "ok")

	return nil
}

func (c *MetricsController) RunMetricsController() error {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(metricsEndpoint, nil); err != nil {
		c.logger.With(zap.Error(err)).Errorf("Error exposing metrics at %s/metrics", metricsEndpoint)
		return err
	}
	return nil
}
