package app

import (
	"context"

	"go.uber.org/zap"
	cloudprovider "k8s.io/cloud-provider"
	cloudControllerManager "k8s.io/cloud-provider/app"
	cloudControllerManagerConfig "k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

const ccmController = "ccm-controller"

// CloudControllerManager manages cloud resources.
type CloudControllerManager struct {
	ctx    context.Context
	logger *zap.SugaredLogger
}

// NewCloudControllerManager creates a new CloudControllerManager instance.
func NewCloudControllerManager(ctx context.Context, logger *zap.SugaredLogger, mgr ctrl.Manager) (Controller, error) {
	return &CloudControllerManager{
		ctx:    ctx,
		logger: logger,
	}, nil
}

func (c *CloudControllerManager) Run(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	if err := c.RunCloudControllerManager(config, options); err != nil {
		return err
	}
	return nil
}

func (c *CloudControllerManager) AddHealthCheck(mgr ctrl.Manager, extraArgs ...interface{}) error {
	c.logger.Infof(ccmController, mgr)

	if err := mgr.AddHealthzCheck(ccmController, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add healthz check: %v", err)
		client.SetHealthStatusForController(ccmController, "failed")
		return err
	}
	client.SetHealthStatusForController(ccmController, "ok")

	if err := mgr.AddReadyzCheck(ccmController, healthz.Ping); err != nil {
		c.logger.Errorf("Failed to add readyz check: %v", err)
		client.SetReadinessForController(ccmController, "failed")
		return err
	}
	client.SetReadinessForController(ccmController, "ok")

	return nil
}

func (c *CloudControllerManager) RunCloudControllerManager(config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error {
	// Run starts all the cloud controller manager control loops.
	cloudProvider := cloudInitializer(c.logger, config)

	controllerInitializers := cloudControllerManager.ConstructControllerInitializers(getInitFuncConstructors(c.logger), config, cloudProvider)
	// TODO move to newer cloudControllerManager dependency that provides a way to pass channel/context
	if err := cloudControllerManager.Run(config, cloudProvider, controllerInitializers, make(map[string]cloudControllerManager.WebhookHandler), c.ctx.Done()); err != nil {
		c.logger.With(zap.Error(err)).Error("Error running cloud controller manager")
		return err
	}
	return nil
}

func cloudInitializer(logger *zap.SugaredLogger, config *cloudControllerManagerConfig.CompletedConfig) cloudprovider.Interface {
	cloudConfig := config.ComponentConfig.KubeCloudShared.CloudProvider
	// initialize cloud provider with the cloud provider name and config file provided
	cloud, err := cloudprovider.InitCloudProvider(cloudConfig.Name, cloudConfig.CloudConfigFile)
	if err != nil {
		logger.With(zap.Error(err)).Fatalf("Cloud provider could not be initialized: %v", err)
	}
	if cloud == nil {
		logger.With(zap.Error(err)).Fatalf("Cloud provider is nil")
	}

	if !cloud.HasClusterID() {
		if config.ComponentConfig.KubeCloudShared.AllowUntaggedCloud {
			logger.With(zap.Error(err)).Info("detected a cluster without a ClusterID.  A ClusterID will be required in the future.  Please tag your cluster to avoid any future issues")
		} else {
			logger.With(zap.Error(err)).Fatalf("no ClusterID found.  A ClusterID is required for the cloud provider to function properly.  This check can be bypassed by setting the allow-untagged-cloud option")
		}
	}

	return cloud
}

func getInitFuncConstructors(logger *zap.SugaredLogger) map[string]cloudControllerManager.ControllerInitFuncConstructor {
	initConstructors := cloudControllerManager.DefaultInitFuncConstructors

	isOciSvcCtrlEnvEnabled := oci.GetIsFeatureEnabledFromEnv(logger, "ENABLE_OCI_SERVICE_CONTROLLER", false)
	if isOciSvcCtrlEnvEnabled || enableOCIServiceController {
		// Disable default Kubernetes Cloud Provider service controller
		cloudControllerManager.ControllersDisabledByDefault.Insert("service")

		// Add OCI service controller init func
		initConstructors["oci-service"] = cloudControllerManager.ControllerInitFuncConstructor{
			InitContext: cloudControllerManager.ControllerInitContext{
				ClientName: "service-controller",
			},
			Constructor: oci.StartOciServiceControllerWrapper,
		}
	}

	return initConstructors
}
