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

package csisnapshotter

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	csirpc "github.com/kubernetes-csi/csi-lib-utils/rpc"
	"github.com/kubernetes-csi/external-resizer/pkg/util"
	clientset "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	snapshotscheme "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned/scheme"
	informers "github.com/kubernetes-csi/external-snapshotter/client/v6/informers/externalversions"
	controller "github.com/kubernetes-csi/external-snapshotter/v6/pkg/sidecar-controller"
	"github.com/kubernetes-csi/external-snapshotter/v6/pkg/snapshotter"
	"google.golang.org/grpc"
	coreinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
)

var (
	// the csiTimeout is kept as 1 minute
	csiTimeout   			= time.Minute
	snapshotNamePrefix		= "snapshot"
	snapshotNameUUIDLength 	= -1
	extraCreateMetadata    	= false
	// the retryIntervalStart is kept as 1 second
	retryIntervalStart   	= time.Second
	retryIntervalMax     	= 5*time.Minute

	kubeAPIQPS   			= 5
	kubeAPIBurst 			= 10
	version = "0.0.1"
)

func StartCSISnapshotter(csioptions csioptions.CSIOptions, stopCh chan struct{}) {
	if csioptions.ShowVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	klog.Infof("Version: %s", version)

	config := BuildConfig(csioptions)

	kubeClient, snapClient := InitializeClients(config)

	factory 	:= informers.NewSharedInformerFactory(snapClient, csioptions.Resync)
	coreFactory := coreinformers.NewSharedInformerFactory(kubeClient, csioptions.Resync)

	// Add Snapshot types to the default Kubernetes so events can be logged for them
	snapshotscheme.AddToScheme(scheme.Scheme)

	metricsManager := metrics.NewCSIMetricsManager("" /* driverName */)

	conn, err := connection.Connect(csioptions.CsiAddress, metricsManager, connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Pass a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), csiTimeout)
	defer cancel()

	// Find driver name
	driverName, err := csirpc.GetDriverName(ctx, conn)
	if err != nil {
		klog.Errorf("error getting CSI driver name: %v", err)
		os.Exit(1)
	}
	klog.V(2).Infof("CSI driver name: %q", driverName)

	// Find out if the driver supports create/delete snapshot.
	supportsCreateSnapshot, err := supportsControllerCreateDeleteSnapshot(ctx, conn)
	if err != nil {
		klog.Errorf("error determining if driver supports create/delete snapshot operations: %v", err)
		os.Exit(1)
	}
	if !supportsCreateSnapshot {
		klog.Errorf("CSI driver %s does not support ControllerCreateSnapshot", driverName)
		os.Exit(1)
	}

	snapShotter := snapshotter.NewSnapshotter(conn)
	ctrl := controller.NewCSISnapshotSideCarController(
		snapClient,
		kubeClient,
		driverName,
		factory.Snapshot().V1().VolumeSnapshotContents(),
		factory.Snapshot().V1().VolumeSnapshotClasses(),
		snapShotter,
		csiTimeout,
		csioptions.Resync,
		snapshotNamePrefix,
		snapshotNameUUIDLength,
		extraCreateMetadata,
		workqueue.NewItemExponentialFailureRateLimiter(retryIntervalStart, retryIntervalMax),
	)

	run := func(ctx context.Context) {
		// run...
		factory.Start(stopCh)
		coreFactory.Start(stopCh)
		go ctrl.Run(int(csioptions.WorkerThreads), stopCh)

		// ...until SIGINT
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		close(stopCh)
	}

	if !csioptions.EnableLeaderElection {
		run(context.TODO())
	} else {
		lockName := "external-snapshotter-" + util.SanitizeName(driverName)
		leKubeClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			klog.Fatal(err.Error())
		}
		le := leaderelection.NewLeaderElection(leKubeClient, lockName, run)

		if csioptions.LeaderElectionNamespace != "" {
			le.WithNamespace(csioptions.LeaderElectionNamespace)
		}

		if err := le.Run(); err != nil {
			klog.Fatalf("error initializing leader election: %v", err)
		}
	}
}

func supportsControllerCreateDeleteSnapshot(ctx context.Context, conn *grpc.ClientConn) (bool, error) {
	capabilities, err := csirpc.GetControllerCapabilities(ctx, conn)
	if err != nil {
		return false, err
	}
	return capabilities[csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT], nil
}

func BuildConfig(csioptions csioptions.CSIOptions) *rest.Config {
	var config *rest.Config
	var err error
	if csioptions.Master != "" || csioptions.Kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(csioptions.Master, csioptions.Kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.Fatal(err.Error())
	}

	config.QPS = (float32)(kubeAPIQPS)
	config.Burst = kubeAPIBurst

	return config
}

func InitializeClients(config *rest.Config) (*kubernetes.Clientset, *clientset.Clientset){
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	snapClient, err := clientset.NewForConfig(config)
	if err != nil {
		klog.Errorf("Error building snapshot clientset: %s", err.Error())
		os.Exit(1)
	}

	return kubeClient, snapClient
}
