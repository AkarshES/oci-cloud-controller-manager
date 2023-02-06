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

package snapshotcontroller

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	snapshotscheme "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned/scheme"
	informers "github.com/kubernetes-csi/external-snapshotter/client/v6/informers/externalversions"
	controller "github.com/kubernetes-csi/external-snapshotter/v6/pkg/common-controller"
	"github.com/kubernetes-csi/external-snapshotter/v6/pkg/metrics"
	coreinformers "k8s.io/client-go/informers"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	csisnapshotter "github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csi-snapshotter"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
)

var (
	// the retryIntervalStart is kept as 1 second
	retryIntervalStart   	= time.Second
	retryIntervalMax     	= 5*time.Minute
	version = "0.0.1"
)

func StartSnapshotController(csioptions csioptions.CSIOptions, stopCh chan struct{}){
	if csioptions.ShowVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	klog.Infof("Version: %s", version)

	config := csisnapshotter.BuildConfig(csioptions)

	kubeClient, snapClient := csisnapshotter.InitializeClients(config)

	factory 	:= informers.NewSharedInformerFactory(snapClient, csioptions.Resync)
	coreFactory := coreinformers.NewSharedInformerFactory(kubeClient, csioptions.Resync)

	// Add Snapshot types to the default Kubernetes so events can be logged for them
	snapshotscheme.AddToScheme(scheme.Scheme)

	metricsManager := metrics.NewMetricsManager()

	var nodeInformer v1.NodeInformer

	ctrl := controller.NewCSISnapshotCommonController(
		snapClient,
		kubeClient,
		factory.Snapshot().V1().VolumeSnapshots(),
		factory.Snapshot().V1().VolumeSnapshotContents(),
		factory.Snapshot().V1().VolumeSnapshotClasses(),
		coreFactory.Core().V1().PersistentVolumeClaims(),
		nodeInformer,
		metricsManager,
		csioptions.Resync,
		workqueue.NewItemExponentialFailureRateLimiter(retryIntervalStart, retryIntervalMax),
		workqueue.NewItemExponentialFailureRateLimiter(retryIntervalStart, retryIntervalMax),
		false,
		false,
	)

	run := func(context.Context) {
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
		lockName := "snapshot-controller-leader"
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
