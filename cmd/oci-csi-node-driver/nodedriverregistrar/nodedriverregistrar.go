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

package nodedriverregistrar

import (
	"context"
	"os"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	csirpc "github.com/kubernetes-csi/csi-lib-utils/rpc"
	"k8s.io/klog"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

const (
	// Name of node annotation that contains JSON map of driver names to node
	// names
	annotationKey = "csi.volume.kubernetes.io/nodeid"

	// Default timeout of short CSI calls like GetPluginInfo
	csiTimeout = 30 * time.Second

	// Verify (and update, if needed) the node ID at this freqeuency.
	sleepDuration = 2 * time.Minute
)

// registrationServer is a sample plugin to work with plugin watcher
type registrationServer struct {
	driverName string
	endpoint   string
	version    []string
}

var _ registerapi.RegistrationServer = registrationServer{}

// NewRegistrationServer returns an initialized registrationServer instance
func newRegistrationServer(driverName string, endpoint string, versions []string) registerapi.RegistrationServer {
	return &registrationServer{
		driverName: driverName,
		endpoint:   endpoint,
		version:    versions,
	}
}

// GetInfo is the RPC invoked by plugin watcher
func (e registrationServer) GetInfo(ctx context.Context, req *registerapi.InfoRequest) (*registerapi.PluginInfo, error) {
	klog.Infof("Received GetInfo call: %+v", req)
	return &registerapi.PluginInfo{
		Type:              registerapi.CSIPlugin,
		Name:              e.driverName,
		Endpoint:          e.endpoint,
		SupportedVersions: e.version,
	}, nil
}

func (e registrationServer) NotifyRegistrationStatus(ctx context.Context, status *registerapi.RegistrationStatus) (*registerapi.RegistrationStatusResponse, error) {
	klog.Infof("Received NotifyRegistrationStatus call: %+v", status)
	if !status.PluginRegistered {
		klog.Errorf("Registration process failed with error: %+v, restarting registration container.", status.Error)
		os.Exit(1)
	}

	return &registerapi.RegistrationStatusResponse{}, nil
}

// RunNodeRegistrar is the main method to start run node register
func RunNodeRegistrar(driverType, csiAddress, registrationPath string, connectionTimeout time.Duration) {
	if registrationPath == "" {
		klog.Errorf("Kubelet Registration Path required for driver: %s", driverType)
		os.Exit(1)
	}

	if connectionTimeout != 0 {
		klog.Warning("--connection-timeout is deprecated and will have no effect")
	}

	// Once https://github.com/container-storage-interface/spec/issues/159 is
	// resolved, if plugin does not support PUBLISH_UNPUBLISH_VOLUME, then we
	// can skip adding mapping to "csi.volume.kubernetes.io/nodeid" annotation.

	metricsManager := metrics.NewCSIMetricsManager(driverType /* driverName */)

	RunCSINodeRegistrar(driverType, csiAddress, registrationPath, metricsManager)
}

func RunCSINodeRegistrar(driverType, csiAddress, registrationPath string, metricsManager metrics.CSIMetricsManager) {
	klog.V(1).Infof("Attempting to open a gRPC connection with: %q", csiAddress)

	csiConn, err := connection.Connect(context.Background(), csiAddress, metricsManager)
	if err != nil {
		klog.Errorf("error connecting to CSI %s driver: %v", driverType, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), csiTimeout)
	defer cancel()

	klog.V(1).Infof("Calling CSI %s driver to discover driver name", driverType)

	csiDriverName, err := csirpc.GetDriverName(ctx, csiConn)
	if err != nil {
		klog.Errorf("error retreiving CSI %s driver name: %v", driverType, err)
		os.Exit(1)
	}

	klog.V(2).Infof("CSI %s driver name: %q", driverType, csiDriverName)

	// Run forever
	nodeRegister(csiDriverName, registrationPath)
}
