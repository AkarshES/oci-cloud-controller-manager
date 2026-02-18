package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cloudControllerManagerConfig "k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const MetricsPort = "8080"
const HealthPort = "8081"

type Controller interface {
	Run(mgr ctrl.Manager, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) error
	AddHealthCheck(mgr ctrl.Manager, extraArgs ...interface{}) error
}

var controllerFactories = map[string]ControllerFactory{
	metricsController: NewMetricsController,
	volumeProvisioner: NewVolumeProvisionerController,
	ccmController:     NewCloudControllerManager,
	csiController:     NewCSIController,
	npnController:     NewNpnController,
	norController:     NewNorController,
	narController:     NewNarController,
}

type ControllerFactory func(ctx context.Context, logger *zap.SugaredLogger, mgr ctrl.Manager) (Controller, error)

func GetController(name string, ctx context.Context, logger *zap.SugaredLogger, mgr ctrl.Manager) (Controller, error) {
	factory, ok := controllerFactories[name]
	if !ok {
		return nil, fmt.Errorf("unknown controller type: %s", name)
	}
	return factory(ctx, logger, mgr)
}

var (
	sharedMgr ctrl.Manager
	once      sync.Once
	onceErr   error
)

func InitSharedManager(scheme *runtime.Scheme, options *options.CloudControllerManagerOptions) error {
	if scheme == nil {
		return errors.New("InitSharedManager called with nil scheme")
	}

	once.Do(func() {
		sharedMgr, onceErr = initializeManager(scheme, options)
	})
	return onceErr
}

// GetSharedManager returns the manager or the error captured during init.
func GetSharedManager() (ctrl.Manager, error) {
	if sharedMgr == nil {
		if onceErr != nil {
			return nil, onceErr
		}
		return nil, errors.New("manager not initialised")
	}
	return sharedMgr, nil
}

func initializeManager(scheme *runtime.Scheme, options *options.CloudControllerManagerOptions) (ctrl.Manager, error) {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("fail to add client-go scheme: %w", err)
	}
	cfg, err := buildRESTConfig(options.Generic.ClientConnection.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build rest config: %w", err)
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: ":" + MetricsPort,
		},
		HealthProbeBindAddress:  ":" + HealthPort,
		LeaderElection:          true,
		LeaderElectionID:        "cpo.oci.oraclecloud.com",
		LeaderElectionNamespace: "kube-system",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialise manager: %w", err)
	}

	return mgr, nil
}
