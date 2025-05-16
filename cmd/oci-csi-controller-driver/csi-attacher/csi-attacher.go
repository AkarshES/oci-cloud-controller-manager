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

package csiattacher

import (
	"context"
	"fmt"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo/leaderelection" // register leader election in the default legacy registry
	"net/http"
	"os"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	_ "k8s.io/component-base/metrics/prometheus/workqueue" // register work queues in the default legacy registry
	"k8s.io/klog/v2"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/csi-lib-utils/rpc"
	"github.com/kubernetes-csi/external-attacher/pkg/attacher"
	"github.com/kubernetes-csi/external-attacher/pkg/controller"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
	"google.golang.org/grpc"
	csitranslationlib "k8s.io/csi-translation-lib"
)

const (

	// Default timeout of short CSI calls like GetPluginInfo
	csiTimeout = 30 * time.Second

	leaderElectionTypeLeases     = "leases"
	leaderElectionTypeConfigMaps = "configmaps"
)

var (
	version = "unknown"
)

type leaderElection interface {
	Run() error
	WithNamespace(namespace string)
}

// StartCSIAttacher main function to start CSI attacher
func StartCSIAttacher(csioptions csioptions.CSIOptions) {

	if csioptions.ShowVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	klog.Infof("Version: %s", version)

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(csioptions.Kubeconfig)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	if csioptions.WorkerThreads == 0 {
		klog.Error("option -worker-threads must be greater than zero")
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	factory := informers.NewSharedInformerFactory(clientset, csioptions.Resync)
	var handler controller.Handler

	metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)

	ctx := context.Background()

	// Connect to CSI.
	csiConn, err := connection.Connect(ctx, csioptions.CsiAddress, metricsManager, connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	err = rpc.ProbeForever(ctx, csiConn, csioptions.Timeout)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Find driver name.
	cancelationCtx, cancel := context.WithTimeout(context.Background(), csiTimeout)
	defer cancel()
	csiAttacher, err := rpc.GetDriverName(cancelationCtx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
	logger := klog.LoggerWithValues(klog.Background(), "driver", csiAttacher)
	klog.V(2).Infof("CSI driver name: %q", csiAttacher)

	// Add default legacy registry so that metrics manager serves Go runtime and process metrics.

	// Also registers the `k8s.io/component-base/` work queue and leader election metrics we anonymously import.

	metricsManager.WithAdditionalRegistry(legacyregistry.DefaultGatherer)

	mux := http.NewServeMux()
	addr := csioptions.MetricsAddress
	// Start HTTP server for metrics + leader election healthz
	if addr != "" {
		metricsManager.RegisterToServer(mux, csioptions.MetricsPath)
		metricsManager.SetDriverName(csiAttacher)
		go func() {
			klog.Infof("ServeMux listening at %q", addr)
			err := http.ListenAndServe(addr, mux)
			if err != nil {
				klog.Fatalf("Failed to start HTTP server at specified address (%q) and metrics path (%q): %s", addr, csioptions.MetricsPath, err)
			}
		}()
	}

	cancelationCtx, cancel = context.WithTimeout(context.Background(), csiTimeout)
	defer cancel()
	supportsService, err := supportsPluginControllerService(cancelationCtx, csiConn)

	var (
		supportsAttach                    bool
		supportsReadOnly                  bool
		supportsListVolumesPublishedNodes bool
		supportsSingleNodeMultiWriter     bool
	)

	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
	if !supportsService {
		handler = controller.NewTrivialHandler(clientset)
		klog.V(2).Infof("CSI driver does not support Plugin Controller Service, using trivial handler")
	} else {
		// Find out if the driver supports attach/detach.
		supportsAttach, supportsReadOnly, supportsListVolumesPublishedNodes, supportsSingleNodeMultiWriter, err = supportsControllerCapabilities(cancelationCtx, csiConn)
		if err != nil {
			klog.Error(err.Error())
			os.Exit(1)
		}
		if supportsAttach {
			pvLister := factory.Core().V1().PersistentVolumes().Lister()
			vaLister := factory.Storage().V1().VolumeAttachments().Lister()
			csiNodeLister := factory.Storage().V1().CSINodes().Lister()
			volAttacher := attacher.NewAttacher(csiConn)
			// Max entries per each page in volume lister call, 0 means no limit.
			maxEntries := 0
			CSIVolumeLister := attacher.NewVolumeLister(csiConn, maxEntries)
			handler = controller.NewCSIHandler(clientset, csiAttacher, volAttacher, CSIVolumeLister, pvLister, csiNodeLister, vaLister, &csioptions.Timeout, supportsReadOnly, supportsSingleNodeMultiWriter, csitranslationlib.New(), csioptions.DefaultFSType)
			klog.V(2).Infof("CSI driver supports ControllerPublishUnpublish, using real CSI handler")
		} else {
			handler = controller.NewTrivialHandler(clientset)
			klog.V(2).Infof("CSI driver does not support ControllerPublishUnpublish, using trivial handler")
		}
	}

	if supportsListVolumesPublishedNodes {
		klog.V(2).Infof("CSI driver supports list volumes published nodes. Using capability to reconcile volume attachment objects with actual backend state")
	}
	if err != nil {
		klog.Errorf("Failed to check if driver supports ListVolumesPublishedNodes, assuming it does not: %v", err)
	}

	ctrl := controller.NewCSIAttachController(
		logger,
		clientset,
		csiAttacher,
		handler,
		factory.Storage().V1().VolumeAttachments(),
		factory.Core().V1().PersistentVolumes(),
		workqueue.NewItemExponentialFailureRateLimiter(csioptions.RetryIntervalStart, csioptions.RetryIntervalMax),
		workqueue.NewItemExponentialFailureRateLimiter(csioptions.RetryIntervalStart, csioptions.RetryIntervalMax),
		supportsListVolumesPublishedNodes,
		csioptions.ReconcileSync,
	)

	run := func(ctx context.Context) {
		stopCh := ctx.Done()
		factory.Start(stopCh)
		ctrl.Run(ctx, int(csioptions.WorkerThreads))
	}

	if !csioptions.EnableLeaderElection {
		run(context.TODO())
	} else {
		// Name of config map with leader election lock
		lockName := "external-attacher-leader-" + csiAttacher
		le := leaderelection.NewLeaderElection(clientset, lockName, run)

		if csioptions.LeaderElectionNamespace != "" {
			le.WithNamespace(csioptions.LeaderElectionNamespace)
		}

		if err := le.Run(); err != nil {
			klog.Fatalf("failed to initialize leader election: %v", err)
		}
	}
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func supportsControllerCapabilities(ctx context.Context, csiConn *grpc.ClientConn) (bool, bool, bool, bool, error) {
	caps, err := rpc.GetControllerCapabilities(ctx, csiConn)
	if err != nil {
		return false, false, false, false, err
	}

	supportsControllerPublish := caps[csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME]
	supportsPublishReadOnly := caps[csi.ControllerServiceCapability_RPC_PUBLISH_READONLY]
	supportsListVolumesPublishedNodes := caps[csi.ControllerServiceCapability_RPC_LIST_VOLUMES] && caps[csi.ControllerServiceCapability_RPC_LIST_VOLUMES_PUBLISHED_NODES]
	supportsSingleNodeMultiWriter := caps[csi.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER]
	return supportsControllerPublish, supportsPublishReadOnly, supportsListVolumesPublishedNodes, supportsSingleNodeMultiWriter, nil
}

func supportsPluginControllerService(ctx context.Context, csiConn *grpc.ClientConn) (bool, error) {
	caps, err := rpc.GetPluginCapabilities(ctx, csiConn)
	if err != nil {
		return false, err
	}

	return caps[csi.PluginCapability_Service_CONTROLLER_SERVICE], nil
}
