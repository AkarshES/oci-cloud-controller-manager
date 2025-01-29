package app

import (
	"context"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	provisioner "github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"go.uber.org/zap"
	cloudControllerManagerConfig "k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

const volumeProvisioner = "volume-provisioner"

var minVolumeSize string
var volumeRoundingEnabled bool

type VolumeProvisionerController struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	mgr    ctrl.Manager
}

// NewVolumeProvisionerController creates a new VolumeProvisionerController instance.
func NewVolumeProvisionerController(ctx context.Context, logger *zap.SugaredLogger, mgr ctrl.Manager) (Controller, error) {
	return &VolumeProvisionerController{
		ctx:    ctx,
		logger: logger,
		mgr:    mgr,
	}, nil
}

func (c *VolumeProvisionerController) Run(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	return c.RunVolumeProvisioner(config, options)
}

func (c *VolumeProvisionerController) AddHealthCheck(mgr ctrl.Manager, extraArgs ...interface{}) error {
	c.logger.Infof(volumeProvisioner, mgr)

	if err := mgr.AddHealthzCheck(volumeProvisioner, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add healthz check: %v", err)
		client.SetHealthStatusForController(volumeProvisioner, "failed")
		return err
	}
	client.SetHealthStatusForController(volumeProvisioner, "ok")

	if err := mgr.AddReadyzCheck(volumeProvisioner, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add readyz check: %v", err)
		client.SetReadinessForController(volumeProvisioner, "failed")
		return err
	}
	client.SetReadinessForController(volumeProvisioner, "ok")

	return nil
}

func (c *VolumeProvisionerController) RunVolumeProvisioner(config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	if err := provisioner.Run(c.logger, options.Generic.ClientConnection.Kubeconfig, options.Master, minVolumeSize, volumeRoundingEnabled, c.ctx.Done()); err != nil {
		c.logger.With(zap.Error(err)).Error("Error running volume provisioner")
	}
	return nil
}
