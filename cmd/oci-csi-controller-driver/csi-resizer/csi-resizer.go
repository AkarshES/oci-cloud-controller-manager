// Copyright 2022 Oracle and/or its affiliates. All rights reserved.
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

package csiresizer

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/external-resizer/pkg/controller"
	"github.com/kubernetes-csi/external-resizer/pkg/csi"
	"github.com/kubernetes-csi/external-resizer/pkg/features"
	"github.com/kubernetes-csi/external-resizer/pkg/modifier"
	"github.com/kubernetes-csi/external-resizer/pkg/modifycontroller"
	"github.com/kubernetes-csi/external-resizer/pkg/resizer"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"

	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

var (
	kubeAPIQPS   = flag.Float64("kube-api-qps", 5, "QPS to use while communicating with the kubernetes apiserver. Defaults to 5.0.")
	kubeAPIBurst = flag.Int("kube-api-burst", 10, "Burst to use while communicating with the kubernetes apiserver. Defaults to 10.")

	handleVolumeInUseError = flag.Bool("handle-volume-inuse-error", true, "Flag to turn on/off capability to handle volume in use error in resizer controller. Defaults to true if not set.")
	extraModifyMetadata = flag.Bool("extra-modify-metadata", false, "If set, add pv/pvc metadata to plugin modify requests as parameters.")


	version = "unknown"
)

func StartCSIResizer(csioptions csioptions.CSIOptions) {
	if csioptions.ShowVersion {
		fmt.Println(os.Args[0], version)
		os.Exit(0)
	}
	klog.Infof("Version : %s", version)
	if csioptions.MetricsAddress != "" && csioptions.HttpEndpoint != "" {
		klog.ErrorS(nil, "Only one of `--metrics-address` and `--http-endpoint` can be set.")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	addr := csioptions.MetricsAddress
	if addr == "" {
		addr = csioptions.HttpEndpoint
	}
	klog.Infof("Adding feature gate: ")
	bs, _ := json.Marshal(csioptions.FeatureGates)
	klog.Infof("Adding feature gate : %s", string(bs))
	if err := utilfeature.DefaultMutableFeatureGate.SetFromMap(csioptions.FeatureGates); err != nil {
		klog.Fatal(err)
	}

	var config *rest.Config
	var err error
	if csioptions.Master != "" || csioptions.Kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(csioptions.Master, csioptions.Kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.ErrorS(err, "Failed to create cluster config")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	config.QPS = float32(*kubeAPIQPS)
	config.Burst = *kubeAPIBurst

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.ErrorS(err, "Failed to create kube client")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, csioptions.Resync)

	mux := http.NewServeMux()

	metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)

	ctx := context.Background()
	csiClient, err := csi.New(ctx, csioptions.CsiAddress, csioptions.Timeout, metricsManager)
	if err != nil {
		klog.ErrorS(err, "Failed to create CSI client")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	driverName, err := getDriverName(csiClient, csioptions.Timeout)
	if err != nil {
		klog.ErrorS(err, "Get driver name failed")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	klog.V(2).InfoS("CSI driver name", "driverName", driverName)

	csiResizer, err := resizer.NewResizerFromClient(
		csiClient,
		csioptions.Timeout,
		kubeClient,
		driverName)
	if err != nil {
		klog.ErrorS(err, "Failed to create CSI resizer")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	csiModifier, err := modifier.NewModifierFromClient(
		csiClient,
		csioptions.Timeout,
		kubeClient,
		informerFactory,
		*extraModifyMetadata,
		driverName)
	if err != nil {
		klog.ErrorS(err, "Failed to create CSI modifier")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	// Start HTTP server for metrics + leader election healthz
	if addr != "" {
		metricsManager.RegisterToServer(mux, csioptions.MetricsPath)
		metricsManager.SetDriverName(driverName)
		go func() {
			klog.InfoS("ServeMux listening", "address", addr)
			err := http.ListenAndServe(addr, mux)
			if err != nil {
				klog.ErrorS(err, "Failed to start HTTP server", "address", addr, "metricsPath", csioptions.MetricsPath)
				klog.FlushAndExit(klog.ExitFlushTimeout, 1)
			}
		}()
	}

	resizerName := csiResizer.Name()
	rc := controller.NewResizeController(resizerName, csiResizer, kubeClient, csioptions.Resync, informerFactory,
		workqueue.NewTypedItemExponentialFailureRateLimiter[string](csioptions.RetryIntervalStart, csioptions.RetryIntervalMax),
		*handleVolumeInUseError, csioptions.RetryIntervalMax)
	modifierName := csiModifier.Name()
	var mc modifycontroller.ModifyController
	// Add modify controller only if the feature gate is enabled
	if utilfeature.DefaultFeatureGate.Enabled(features.VolumeAttributesClass) {
		mc = modifycontroller.NewModifyController(modifierName, csiModifier, kubeClient, csioptions.Resync,csioptions.RetryIntervalMax,*extraModifyMetadata, informerFactory,
			workqueue.NewTypedItemExponentialFailureRateLimiter[string](csioptions.RetryIntervalStart, csioptions.RetryIntervalMax))
	}

	run := func(ctx context.Context) {
		informerFactory.Start(wait.NeverStop)
		go rc.Run(int(csioptions.WorkerThreads), ctx)
		if utilfeature.DefaultFeatureGate.Enabled(features.VolumeAttributesClass) {
			go mc.Run(int(csioptions.WorkerThreads), ctx)
		}
		<-ctx.Done()
	}

	if !csioptions.EnableLeaderElection {
		run(context.TODO())
	} else {
		lockName := "external-resizer-" + util.SanitizeName(resizerName)
		leKubeClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			klog.ErrorS(err, "Failed to create leKubeClient")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		le := leaderelection.NewLeaderElection(leKubeClient, lockName, run)

		if csioptions.LeaderElectionNamespace != "" {
			le.WithNamespace(csioptions.LeaderElectionNamespace)
		}

		if err := le.Run(); err != nil {
			klog.ErrorS(err, "Error initializing leader election")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}
}

func getDriverName(client csi.Client, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return client.GetDriverName(ctx)
}
