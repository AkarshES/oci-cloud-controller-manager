// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cloudprovider "k8s.io/cloud-provider"
	cloudControllerManager "k8s.io/cloud-provider/app"
	cloudControllerManagerConfig "k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/names"
	"k8s.io/cloud-provider/options"
	cliflag "k8s.io/component-base/cli/flag"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/term"
	"k8s.io/component-base/version/verflag"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	npnv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/v1beta1"
	csicontroller "github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csi-controller"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
	"github.com/oracle/oci-cloud-controller-manager/controllers"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	_ "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client" // for oci client metric registration
	provisioner "github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
)

var (
	logLevel                                                                                  int8
	minVolumeSize, resourcePrincipalFile, metricsEndpoint, logfilePath                        string
	enableCSI, enableVolumeProvisioning, volumeRoundingEnabled, useResourcePrincipal, logJSON bool
	enableNPNController                                                                       bool
	enableOCIServiceController                                                                bool
	resourcePrincipalInitialTimeout                                                           time.Duration
)

var csioption = csioptions.CSIOptions{}

var (
	scheme         = runtime.NewScheme()
	npnSetupLog    = ctrl.Log.WithName("npn-controller-setup")
	configFilePath = "/etc/oci/config.yaml"
)

const (
	defaultFssAddress  = "/var/run/shared-tmpfs/csi-fss.sock"
	defaultFssEndpoint = "unix:///var/run/shared-tmpfs/csi-fss.sock"
)

// NewCloudProviderOCICommand creates a *cobra.Command object with default parameters
func NewCloudProviderOCICommand(logger *zap.SugaredLogger) *cobra.Command {

	// FIXME Create CLoudProviderOCIOptions struct that shall contain options for all the components
	s, err := options.NewCloudControllerManagerOptions()
	if err != nil {
		logger.With(zap.Error(err)).Fatalf("unable to initialize command options")
	}

	command := &cobra.Command{
		Use: "cloud-provider-oci",
		Long: `The cloud provider oci daemon is a agglomeration of oci cloud controller
manager and oci volume provisioner. It embeds the cloud specific control loops shipped with Kubernetes.`,
		Run: func(cmd *cobra.Command, args []string) {
			log := logging.Logger()
			defer log.Sync()
			zap.ReplaceGlobals(log)
			logger = log.Sugar()
			verflag.PrintAndExitIfRequested()
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				logger.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})

			c, err := s.Config(cloudControllerManager.ControllerNames(cloudControllerManager.DefaultInitFuncConstructors),
				cloudControllerManager.ControllersDisabledByDefault.List(), names.CCMControllerAliases(), cloudControllerManager.AllWebhooks, cloudControllerManager.DisabledByDefaultWebhooks)
			if err != nil {
				logger.With(zap.Error(err)).Fatalf("Unable to create cloud controller manager config")
			}

			run(logger, c.Complete(), s)

		},
	}

	namedFlagSets := s.Flags(cloudControllerManager.ControllerNames(cloudControllerManager.DefaultInitFuncConstructors),
		cloudControllerManager.ControllersDisabledByDefault.List(), names.CCMControllerAliases(), cloudControllerManager.AllWebhooks, cloudControllerManager.DisabledByDefaultWebhooks)

	// logging parameters flagset
	loggingFlagSet := namedFlagSets.FlagSet("logging variables")
	loggingFlagSet.Int8Var(&logLevel, "log-level", int8(zapcore.InfoLevel), "Adjusts the level of the logs that will be omitted.")
	loggingFlagSet.BoolVar(&logJSON, "log-json", false, "Log in json format.")
	loggingFlagSet.StringVar(&logfilePath, "logfile-path", "", "If specified, write log messages to a file at this path.")

	// prometheus metrics endpoint flagset
	metricsFlagSet := namedFlagSets.FlagSet("metrics endpoint")
	metricsFlagSet.StringVar(&metricsEndpoint, "metrics-endpoint", "0.0.0.0:8080", "The endpoint where to expose metrics")

	// volume provisioner flag set
	vpFlagSet := namedFlagSets.FlagSet("volume provisioner")
	vpFlagSet.BoolVar(&enableVolumeProvisioning, "enable-volume-provisioning", true, "When enabled volumes will be provisioned/deleted by cloud controller manager")
	vpFlagSet.BoolVar(&volumeRoundingEnabled, "rounding-enabled", true, "When enabled volumes will be rounded up if less than 'minVolumeSizeMB'")
	vpFlagSet.StringVar(&minVolumeSize, "min-volume-size", "50Gi", "The minimum size for a block volume. By default OCI only supports block volumes > 50GB")

	// oci authentication mode flag set
	ociAuthFlagSet := namedFlagSets.FlagSet("oci authentication modes")
	ociAuthFlagSet.BoolVar(&useResourcePrincipal, "use-resource-principal", false, "If true use resource principal as authentication mode else use service principal as authentication mode")
	ociAuthFlagSet.StringVar(&resourcePrincipalFile, "resource-principal-file", "", "The filesystem path at which the serialized Resource Principal is stored")
	ociAuthFlagSet.DurationVar(&resourcePrincipalInitialTimeout, "resource-principal-initial-timeout", 1*time.Minute, "How long to wait for an initial Resource Principal before terminating with an error if one is not supplied")

	// csi flag set.
	csiFlagSet := namedFlagSets.FlagSet("CSI Controller")
	csiFlagSet.BoolVar(&enableCSI, "csi-enabled", false, "Whether to enable CSI feature in OKE")
	csiFlagSet.StringVar(&csioption.CsiAddress, "csi-address", "/run/csi/socket", "Address of the CSI Block Volume driver socket.")
	csiFlagSet.StringVar(&csioption.Endpoint, "csi-endpoint", "unix://tmp/csi.sock", "CSI Block Volume endpoint")
	csiFlagSet.StringVar(&csioption.VolumeNamePrefix, "csi-volume-name-prefix", "pvc", "Prefix to apply to the name of a created volume.")
	csiFlagSet.IntVar(&csioption.VolumeNameUUIDLength, "csi-volume-name-uuid-length", -1, "Truncates generated UUID of a created volume to this length. Defaults behavior is to NOT truncate.")
	csiFlagSet.BoolVar(&csioption.ShowVersion, "csi-version", false, "Show version.")
	csiFlagSet.DurationVar(&csioption.RetryIntervalStart, "csi-retry-interval-start", time.Second, "Initial retry interval of failed provisioning or deletion. It doubles with each failure, up to retry-interval-max.")
	csiFlagSet.DurationVar(&csioption.RetryIntervalMax, "csi-retry-interval-max", 5*time.Minute, "Maximum retry interval of failed provisioning or deletion.")
	csiFlagSet.UintVar(&csioption.WorkerThreads, "csi-worker-threads", 100, "Number of provisioner worker threads, in other words nr. of simultaneous CSI calls.")
	csiFlagSet.DurationVar(&csioption.OperationTimeout, "csi-op-timeout", 120*time.Second, "Timeout for waiting for creation or deletion of a volume")
	csiFlagSet.BoolVar(&csioption.EnableLeaderElection, "csi-enable-leader-election", false, "Enables leader election. If leader election is enabled, additional RBAC rules are required. Please refer to the Kubernetes CSI documentation for instructions on setting up these RBAC rules.")
	csiFlagSet.StringVar(&csioption.LeaderElectionType, "csi-leader-election-type", "endpoints", "the type of leader election, options are 'endpoints' (default) or 'leases' (strongly recommended). The 'endpoints' option is deprecated in favor of 'leases'.")
	csiFlagSet.StringVar(&csioption.LeaderElectionNamespace, "csi-leader-election-namespace", "", "Namespace where the leader election resource lives. Defaults to the pod namespace if not set.")
	csiFlagSet.BoolVar(&csioption.StrictTopology, "csi-strict-topology", false, "Passes only selected node topology to CreateVolume Request, unlike default behavior of passing aggregated cluster topologies that match with topology keys of the selected node.")
	csiFlagSet.BoolVar(&csioption.ImmediateTopology, "csi-immediate-topology", true, "Immediate binding: pass aggregated cluster topologies for all nodes where the CSI driver is available (enabled, the default) or no topology requirements (if disabled)")
	csiFlagSet.DurationVar(&csioption.Resync, "csi-resync", 10*time.Minute, "Resync interval of the controller.")
	csiFlagSet.DurationVar(&csioption.Timeout, "csi-timeout", 15*time.Second, "Timeout for waiting for attaching or detaching the volume.")
	csiFlagSet.BoolVar(&csioption.EnableResizer, "csi-bv-expansion-enabled", false, "Enables go routine csi-resizer.")
	csiFlagSet.UintVar(&csioption.FinalizerThreads, "cloning-protection-threads", 1, "Number of simultaneously running threads, handling cloning finalizer removal")
	csiFlagSet.Var(utilflag.NewMapStringBool(&csioption.FeatureGates), "csi-feature-gates", "A set of key=value pairs that describe feature gates for alpha/experimental features. ")
	csiFlagSet.StringVar(&csioption.GroupSnapshotNamePrefix, "groupsnapshot-name-prefix", "groupsnapshot", "Prefix to apply to the name of a created group snapshot.")
	csiFlagSet.IntVar(&csioption.GroupSnapshotNameUUIDLength, "groupsnapshot-name-uuid-length", -1, "Length in characters for the generated uuid of a created group snapshot. Defaults behavior is to NOT truncate.")

	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), command.Name())

	npnFlagSet := namedFlagSets.FlagSet("NPN Controller")
	npnFlagSet.BoolVar(&enableNPNController, "enable-npn-controller", false, "Whether to enable Native Pod Network controller")

	ociSvcCtrlFlagSet := namedFlagSets.FlagSet("OCI Service Controller")
	ociSvcCtrlFlagSet.BoolVar(&enableOCIServiceController, "enable-oci-service-controller", false, "Whether to enable OCI service controller instead of Kubernetes Cloud Provider service controller")

	if flag.CommandLine.Lookup("cloud-provider-gce-lb-src-cidrs") != nil {
		// hoist this flag from the global flagset to preserve the commandline until
		// the gce cloudprovider is removed.
		globalflag.Register(namedFlagSets.FlagSet("generic"), "cloud-provider-gce-lb-src-cidrs")
	}
	for _, f := range namedFlagSets.FlagSets {
		command.Flags().AddFlagSet(f)
	}
	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(command.OutOrStdout())
	command.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
		return nil
	})
	command.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	viper.BindPFlags(command.Flags())

	return command
}

func run(logger *zap.SugaredLogger, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) {
	var wg sync.WaitGroup
	ctx, cancelFunc := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 2)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancelFunc()
		<-sigs
		os.Exit(1)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(metricsEndpoint, nil); err != nil {
			logger.With(zap.Error(err)).Errorf("Error exposing metrics at %s/metrics", metricsEndpoint)
		}
		cancelFunc()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger := logger.With(zap.String("component", "volume-provisioner"))
		if !enableVolumeProvisioning {
			logger.Info("Volume provisioning is disabled")
			return
		}
		// TODO Pass an options/config struct instead of config variables
		if err := provisioner.Run(logger, options.Generic.ClientConnection.Kubeconfig, options.Master, minVolumeSize, volumeRoundingEnabled, ctx.Done()); err != nil {
			logger.With(zap.Error(err)).Error("Error running volume provisioner")
		}
		cancelFunc()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run starts all the cloud controller manager control loops.
		cloudProvider := cloudInitializer(logger, config)

		controllerInitializers := cloudControllerManager.ConstructControllerInitializers(getInitFuncConstructors(logger), config, cloudProvider)
		// TODO move to newer cloudControllerManager dependency that provides a way to pass channel/context
		if err := cloudControllerManager.Run(config, cloudProvider, controllerInitializers, make(map[string]cloudControllerManager.WebhookHandler), ctx.Done()); err != nil {
			logger.With(zap.Error(err)).Error("Error running cloud controller manager")
		}
		cancelFunc()
	}()

	if enableCSI == true {
		wg.Add(1)
		logger := logger.With(zap.String("component", "csi-controller"))
		logger.Info("CSI is enabled.")
		go func() {
			defer wg.Done()
			csioption.Master = options.Master
			csioption.Kubeconfig = options.Generic.ClientConnection.Kubeconfig
			csioption.FssCsiAddress = csioptions.GetFssAddress(csioption.CsiAddress, defaultFssAddress)
			csioption.FssEndpoint = csioptions.GetFssAddress(csioption.Endpoint, defaultFssEndpoint)
			csioption.FssVolumeNamePrefix = csioptions.GetFssVolumeNamePrefix(csioption.VolumeNamePrefix)
			// Check and update feature gate for CrossNamespaceDataSource
			csioption.FeatureGates = csioptions.UpdateFeatureGates(csioption.FeatureGates)
			csioption.RuntimeSchemeMutex = new(sync.Mutex)
			err := csicontroller.Run(csioption, ctx.Done())
			if err != nil {
				logger.With(zap.Error(err)).Error("Error running csi-controller")
			}
			cancelFunc()
		}()
	} else {
		logger := logger.With(zap.String("component", "csi-controller"))
		logger.Info("CSI is disabled.")
	}

	enableNPN := oci.GetIsFeatureEnabledFromEnv(logger, "ENABLE_NPN_CONTROLLER", false)

	if enableNPN || enableNPNController {
		wg.Add(1)
		logger = logger.With(zap.String("component", "npn-controller"))
		ctrl.SetLogger(zapr.NewLogger(logger.Desugar()))
		logger.Info("NPN controller is enabled.")
		go func() {
			defer wg.Done()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))
			utilruntime.Must(npnv1beta1.AddToScheme(scheme))

			configPath, ok := os.LookupEnv("CONFIG_YAML_FILENAME")
			if !ok {
				configPath = configFilePath
			}
			cfg := providercfg.GetConfig(logger, configPath)
			ociClient := getOCIClient(logger, cfg)

			mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
				Scheme: scheme,
				Metrics: metricsserver.Options{
					BindAddress: ":8080",
				},
				HealthProbeBindAddress:  ":8081",
				LeaderElection:          true,
				LeaderElectionID:        "npn.oci.oraclecloud.com",
				LeaderElectionNamespace: "kube-system",
			})
			if err != nil {
				npnSetupLog.Error(err, "unable to start manager")
				os.Exit(1)
			}

			metricPusher, err := metrics.NewMetricPusher(logger)
			if err != nil {
				logger.With("error", err).Error("metrics collection could not be enabled")
				// disable metrics
				metricPusher = nil
			}

			if err = (&controllers.NativePodNetworkReconciler{
				Client:       mgr.GetClient(),
				Scheme:       mgr.GetScheme(),
				MetricPusher: metricPusher,
				OCIClient:    ociClient,
			}).SetupWithManager(mgr); err != nil {
				npnSetupLog.Error(err, "unable to create controller", "controller", "NativePodNetwork")
				os.Exit(1)
			}

			if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
				npnSetupLog.Error(err, "unable to set up health check")
				os.Exit(1)
			}
			if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
				npnSetupLog.Error(err, "unable to set up ready check")
				os.Exit(1)
			}

			npnSetupLog.Info("starting manager")
			if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
				npnSetupLog.Error(err, "problem running manager")
				// TODO: Handle the case of NPN controller not running more gracefully
				os.Exit(1)
			}
		}()
	}

	// wait for all the go routines to finish.
	wg.Wait()
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

func getOCIClient(logger *zap.SugaredLogger, config *providercfg.Config) client.Interface {
	c, err := client.GetClient(logger, config)

	if err != nil {
		logger.With(zap.Error(err)).Fatal("client can not be generated.")
	}
	return c
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
