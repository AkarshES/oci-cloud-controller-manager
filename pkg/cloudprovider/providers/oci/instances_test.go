// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package oci

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	v1discoverylisters "k8s.io/client-go/listers/discovery/v1"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

var (
	instanceVnics = map[string]*core.Vnic{
		"basic-complete": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("basic-complete"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-external-ip": {
			PrivateIp:     common.String("10.0.0.1"),
			HostnameLabel: common.String("no-external-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-internal-ip": {
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-internal-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"invalid-internal-ip": {
			PrivateIp:     common.String("10.0.0."),
			HostnameLabel: common.String("no-internal-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"invalid-external-ip": {
			PublicIp:      common.String("0.0.0."),
			HostnameLabel: common.String("invalid-external-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-hostname-label": {
			PrivateIp: common.String("10.0.0.1"),
			PublicIp:  common.String("0.0.0.1"),
			SubnetId:  common.String("subnetwithdnslabel"),
		},
		"no-subnet-dns-label": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-subnet-dns-label"),
			SubnetId:      common.String("subnetwithoutdnslabel"),
		},
		"no-vcn-dns-label": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("subnetwithnovcndnslabel"),
		},
	}

	instances = map[string]*core.Instance{
		"basic-complete": {
			CompartmentId: common.String("default"),
		},
		"no-external-ip": {
			CompartmentId: common.String("default"),
		},
		"no-internal-ip": {
			CompartmentId: common.String("default"),
		},
		"invalid-internal-ip": {
			CompartmentId: common.String("default"),
		},
		"invalid-external-ip": {
			CompartmentId: common.String("default"),
		},
		"no-hostname-label": {
			CompartmentId: common.String("default"),
		},
		"no-subnet-dns-label": {
			CompartmentId: common.String("default"),
		},
		"no-vcn-dns-label": {
			CompartmentId: common.String("default"),
		},
		"instance1": {
			CompartmentId: common.String("compartment1"),
			Id:            common.String("instance1"),
			Shape:         common.String("VM.Standard1.2"),
			DisplayName:   common.String("instance1"),
		},
		"instance_zone_test": {
			AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
			CompartmentId:      common.String("compartment1"),
			Id:                 common.String("instance_zone_test"),
			Region:             common.String("PHX"),
			Shape:              common.String("VM.Standard1.2"),
			DisplayName:        common.String("instance_zone_test"),
		},
	}
	subnets = map[string]*core.Subnet{
		"subnetwithdnslabel": {
			Id:       common.String("subnetwithdnslabel"),
			DnsLabel: common.String("subnetwithdnslabel"),
			VcnId:    common.String("vcnwithdnslabel"),
		},
		"subnetwithoutdnslabel": {
			Id:    common.String("subnetwithoutdnslabel"),
			VcnId: common.String("vcnwithdnslabel"),
		},
		"subnetwithnovcndnslabel": {
			Id:       common.String("subnetwithnovcndnslabel"),
			DnsLabel: common.String("subnetwithnovcndnslabel"),
			VcnId:    common.String("vcnwithoutdnslabel"),
		},
		"one": {
			Id:                 common.String("one"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD1"),
		},
		"two": {
			Id:                 common.String("two"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD2"),
		},
		"annotation-one": {
			Id:                 common.String("annotation-one"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD1"),
		},
		"annotation-two": {
			Id:                 common.String("annotation-two"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD2"),
		},
		"regional-subnet": {
			Id:                 common.String("regional-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
		},
	}

	vcns = map[string]*core.Vcn{
		"vcnwithdnslabel": {
			Id:       common.String("vcnwithdnslabel"),
			DnsLabel: common.String("vcnwithdnslabel"),
		},
		"vcnwithoutdnslabel": {
			Id: common.String("vcnwithoutdnslabel"),
		},
	}

	virtualNodes = map[string]*containerengine.VirtualNode{
		"ocid1.virtualnode.oc1.iad.default": {
			Id:                common.String("ocid1.virtualnode.oc1.iad.default"),
			VirtualNodePoolId: common.String("vnpId"),
		},
		"ocid1.virtualnode.oc1.iad.zonetest": {
			Id:                 common.String("ocid1.virtualnode.oc1.iad.zonetest"),
			VirtualNodePoolId:  common.String("vnpId"),
			AvailabilityDomain: common.String("PHX-AD-1"),
		},
	}

	nodeList = map[string]*v1.Node{
		"default": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "default",
			},
		},
		"instance1": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "compartment1",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "instance1",
			},
		},
		"instanceWithAddress1": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "compartment1",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "instanceWithAddress1",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Address: "0.0.0.0",
						Type:    "InternalIP",
					},
				},
			},
		},
		"instanceWithAddress2": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "compartment1",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "instanceWithAddress2",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Address: "0.0.0.1",
						Type:    "InternalIP",
					},
				},
			},
		},
		"virtualNodeDefault": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					VirtualNodePoolIdAnnotation: "vnpId",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.virtualnode.oc1.iad.default",
			},
		},
		"virtualNodeZoneTest": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					VirtualNodePoolIdAnnotation: "vnpId",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.virtualnode.oc1.iad.zonetest",
			},
		},
		"virtualNodeNonCache": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					VirtualNodePoolIdAnnotation: "vnpId",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "ocid1.virtualnode.oc1.iad.noncache",
			},
		},
	}

	podList = map[string]*v1.Pod{
		"virtualPod1": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "virtualPod1",
				Labels: map[string]string{
					"app": "pod1",
				},
			},
			Spec: v1.PodSpec{
				NodeName: "virtualNodeDefault",
			},
		},
		"virtualPod2": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "virtualPod2",
				Labels: map[string]string{
					"app": "pod2",
				},
			},
			Spec: v1.PodSpec{
				NodeName: "virtualNodeDefault",
			},
			Status: v1.PodStatus{
				PodIP: "0.0.0.10",
			},
		},
		"regularPod1": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "regularPod1",
			},
			Spec: v1.PodSpec{
				NodeName: "default",
			},
		},
		"regularPod2": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "regularPod2",
			},
			Spec: v1.PodSpec{
				NodeName: "default",
			},
		},
	}

	ready             = true
	endpointSliceList = map[string]*v1discovery.EndpointSlice{
		"endpointSliceVirtual": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "virtualService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.9"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.10"},
				},
			},
		},
		"endpointSliceRegular": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "regularService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "regularPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.19"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "regularPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.20"},
				},
			},
		},

		"endpointSliceMixed": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "mixedService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.9"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "regularPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.19"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.10"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "regularPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.20"},
				},
			},
		},
		"endpointSliceUnknownPod": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "unknownService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "unknown",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.100"},
				},
			},
		},
		"endpointSliceDuplicate1.1": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "duplicateEndpointsService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.10"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.9"},
				},
			},
		},
		"endpointSliceDuplicate1.2": {
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					v1discovery.LabelServiceName: "duplicateEndpointsService",
				},
			},
			Endpoints: []v1discovery.Endpoint{
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod1",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.10"},
				},
				{
					TargetRef: &v1.ObjectReference{
						Kind: "Pod",
						Name: "virtualPod2",
					},
					Conditions: v1discovery.EndpointConditions{
						Ready: &ready,
					},
					Addresses: []string{"0.0.0.9"},
				},
			},
		},
	}

	loadBalancers = map[string]*client.GenericLoadBalancer{
		"privateLB": {
			Id:          common.String("privateLB"),
			DisplayName: common.String("privateLB"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"privateLB-no-IP": {
			Id:          common.String("privateLB-no-IP"),
			DisplayName: common.String("privateLB-no-IP"),
			IpAddresses: []client.GenericIpAddress{},
		},
		"test-uid": {
			Id:          common.String("test-uid"),
			DisplayName: common.String("test-uid"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"test-uid-delete-err": {
			Id:          common.String("test-uid-delete-err"),
			DisplayName: common.String("test-uid-delete-err"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"test-uid-node-err": {
			Id:          common.String("test-uid-delete-err"),
			DisplayName: common.String("test-uid-delete-err"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
			SubnetIds: []string{*subnets["one"].Id, *subnets["two"].Id},
			Listeners: map[string]client.GenericListener{
				"one": {
					Name:                  common.String("one"),
					DefaultBackendSetName: common.String("one"),
					Port:                  common.Int(5665),
				}},
			BackendSets: map[string]client.GenericBackendSetDetails{
				"one": {
					Backends: []client.GenericBackend{{
						Name:      common.String("one"),
						IpAddress: common.String("10.0.50.5"),
						Port:      common.Int(5665),
					}},
				}},
		},
	}
)

type MockSecurityListManager struct{}

func (MockSecurityListManager) Update(ctx context.Context, lbSubnets []*core.Subnet, _ []*core.Subnet, sourceCIDRs []string, actualPorts *portSpec, desiredPorts portSpec, isPreserveSource bool) error {
	return nil
}

func (MockSecurityListManager) Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, ports portSpec, sourceCIDRs []string, isPreserveSource bool) error {
	return nil
}

type MockSecurityListManagerFactory func(mode string) MockSecurityListManager

type MockOCIClient struct{}

func (MockOCIClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (MockOCIClient) LoadBalancer(string) client.GenericLoadBalancerInterface {
	return &MockLoadBalancerClient{}
}

func (MockOCIClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

func (MockOCIClient) BlockStorage() client.BlockStorageInterface {
	return &MockBlockStorageClient{}
}

func (MockOCIClient) FSS() client.FileStorageInterface {
	return &MockFileStorageClient{}
}

func (MockOCIClient) Identity() client.IdentityInterface {
	return &MockIdentityClient{}
}

func (MockOCIClient) ContainerEngine() client.ContainerEngineInterface {
	return &MockContainerEngineClient{}
}

// MockComputeClient mocks Compute client implementation
type MockComputeClient struct{}

func (MockComputeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	if instance, ok := instances[id]; ok {
		return instance, nil
	}
	return &core.Instance{
		AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
		CompartmentId:      common.String("default"),
		Id:                 &id,
		Region:             common.String("PHX"),
		Shape:              common.String("VM.Standard1.2"),
	}, nil
}

func (MockComputeClient) GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error) {
	if instance, ok := instances[nodeName]; ok {
		return instance, nil
	}
	return &core.Instance{
		AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
		CompartmentId:      &compartmentID,
		Id:                 &nodeName,
		Region:             common.String("PHX"),
		Shape:              common.String("VM.Standard1.2"),
	}, nil
}

func (MockComputeClient) GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	return instanceVnics[instanceID], nil
}

func (c *MockComputeClient) ListVnicAttachments(ctx context.Context, compartmentID, instanceID string) ([]core.VnicAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) AttachVnic(ctx context.Context, instanceID, subnetID *string, nsgIds []*string, skipSourceDestCheck *bool) (response core.VnicAttachment, err error) {
	return core.VnicAttachment{}, nil
}

func (MockComputeClient) FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) AttachParavirtualizedVolume(ctx context.Context, instanceID, volumeID string, isPvEncryptionInTransitEnabled bool) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) AttachVolume(ctx context.Context, instanceID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) WaitForVolumeAttached(ctx context.Context, attachmentID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) DetachVolume(ctx context.Context, id string) error {
	return nil
}

func (MockComputeClient) WaitForVolumeDetached(ctx context.Context, attachmentID string) error {
	return nil
}

func (c *MockComputeClient) FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

func (c *MockVirtualNetworkClient) IsRegionalSubnet(ctx context.Context, id string) (bool, error) {
	return subnets[id].AvailabilityDomain == nil, nil
}

func (c *MockVirtualNetworkClient) GetPrivateIp(ctx context.Context, id string) (*core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) CreatePrivateIp(ctx context.Context, vnicId string) (*core.PrivateIp, error) {
	return &core.PrivateIp{}, nil
}

func (c *MockVirtualNetworkClient) ListPrivateIps(ctx context.Context, id string) ([]core.PrivateIp, error) {
	return []core.PrivateIp{}, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	if subnet, ok := subnets[id]; ok {
		return subnet, nil
	}
	return nil, errors.New("Subnet not found")
}

func (c *MockVirtualNetworkClient) GetVcn(ctx context.Context, id string) (*core.Vcn, error) {
	return vcns[id], nil
}

func (c *MockVirtualNetworkClient) GetVNIC(ctx context.Context, id string) (*core.Vnic, error) {
	return &core.Vnic{}, nil
}

func (c *MockVirtualNetworkClient) GetSubnetFromCacheByIP(ip string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) GetPublicIpByIpAddress(ctx context.Context, id string) (*core.PublicIp, error) {
	return nil, nil
}

// MockLoadBalancerClient mocks LoadBalancer client implementation.
type MockLoadBalancerClient struct{}

func (c *MockLoadBalancerClient) ListWorkRequests(ctx context.Context, compartmentId, lbId string) ([]*client.GenericWorkRequest, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, details *client.GenericCreateLoadBalancerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, id string) (*client.GenericLoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) GetLoadBalancerByName(ctx context.Context, compartmentID string, name string) (*client.GenericLoadBalancer, error) {
	if lb, ok := loadBalancers[name]; ok {
		return lb, nil
	}
	return nil, nil
}

func (c *MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	if id == "test-uid-delete-err" {
		return "workReqId", errors.New("error")
	}
	return "", nil
}

func (c *MockLoadBalancerClient) GetCertificateByName(ctx context.Context, lbID string, name string) (*client.GenericCertificate, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateCertificate(ctx context.Context, lbID string, cert *client.GenericCertificate) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

var awaitLoadbalancerWorkrequestMap = map[string]error{
	"failedToGetUpdateNetworkSecurityGroupsWorkRequest": errors.New("internal server error for get workrequest call"),
}

func (c *MockLoadBalancerClient) AwaitWorkRequest(ctx context.Context, id string) (*client.GenericWorkRequest, error) {
	if err, ok := awaitLoadbalancerWorkrequestMap[id]; ok {
		return nil, err
	}
	return nil, nil
}

func (c *MockLoadBalancerClient) UpdateLoadBalancerShape(context.Context, string, *client.GenericUpdateLoadBalancerShapeDetails) (string, error) {
	return "", nil
}

var updateNetworkSecurityGroupsLBsFailures = map[string]error{
	"":                      errors.New("provided LB ID is empty"),
	"failedToCreateRequest": errors.New("internal server error"),
}
var updateNetworkSecurityGroupsLBsWorkRequests = map[string]string{
	"failedToGetUpdateNetworkSecurityGroupsWorkRequest": "failedToGetUpdateNetworkSecurityGroupsWorkRequest",
}

func (c *MockLoadBalancerClient) UpdateNetworkSecurityGroups(ctx context.Context, lbId string, nsgIds []string) (string, error) {
	if err, ok := updateNetworkSecurityGroupsLBsFailures[lbId]; ok {
		return "", err
	}
	if wrID, ok := updateNetworkSecurityGroupsLBsWorkRequests[lbId]; ok {
		return wrID, nil
	}
	return "", nil
}

// MockBlockStorageClient mocks BlockStorage client implementation
type MockBlockStorageClient struct{}

func (MockBlockStorageClient) AwaitVolumeAvailableORTimeout(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	return nil, nil
}

func (c MockBlockStorageClient) UpdateVolume(ctx context.Context, volumeId string, details core.UpdateVolumeDetails) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) DeleteVolume(ctx context.Context, id string) error {
	return nil
}

// MockFileStorageClient mocks FileStorage client implementation.
type MockFileStorageClient struct{}

func (c MockFileStorageClient) GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (bool, []filestorage.FileSystemSummary, error) {
	return false, nil, nil
}

func (c MockFileStorageClient) FindExport(ctx context.Context, fsID, path, exportSetID string) (*filestorage.ExportSummary, error) {
	return nil, nil
}

func (c MockFileStorageClient) GetMountTarget(ctx context.Context, id string) (*filestorage.MountTarget, error) {
	return nil, nil
}

func (c MockFileStorageClient) GetMountTargetSummaryByDisplayName(ctx context.Context, compartmentID, ad, mountTargetName string) (bool, []filestorage.MountTargetSummary, error) {
	return false, nil, nil
}

func (MockFileStorageClient) AwaitMountTargetActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.MountTarget, error) {
	return nil, nil
}

func (MockFileStorageClient) GetFileSystem(ctx context.Context, id string) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) CreateFileSystem(ctx context.Context, details filestorage.CreateFileSystemDetails) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) DeleteFileSystem(ctx context.Context, id string) error {
	return nil
}

func (MockFileStorageClient) CreateExport(ctx context.Context, details filestorage.CreateExportDetails) (*filestorage.Export, error) {
	return nil, nil
}

func (MockFileStorageClient) AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.Export, error) {
	return nil, nil
}

func (MockFileStorageClient) DeleteExport(ctx context.Context, id string) error {
	return nil
}

func (MockFileStorageClient) CreateMountTarget(ctx context.Context, details filestorage.CreateMountTargetDetails) (*filestorage.MountTarget, error) {
	return nil, nil
}

func (MockFileStorageClient) DeleteMountTarget(ctx context.Context, id string) error {
	return nil
}

// MockIdentityClient mocks Identity client implementaion
type MockIdentityClient struct{}

func (MockIdentityClient) GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error) {
	return nil, nil
}

func (MockIdentityClient) ListAvailabilityDomains(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error) {
	return nil, nil
}

type mockInstanceCache struct{}

func (m mockInstanceCache) Add(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) Update(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) Delete(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) List() []interface{} {
	return nil
}

func (m mockInstanceCache) ListKeys() []string {
	return nil
}

func (m mockInstanceCache) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return instances["default"], true, nil
}

func (m mockInstanceCache) GetByKey(key string) (item interface{}, exists bool, err error) {
	if instance, ok := instances[key]; ok {
		return instance, true, nil
	}
	return nil, false, nil
}

func (m mockInstanceCache) Replace(i []interface{}, s string) error {
	return nil
}

func (m mockInstanceCache) Resync() error {
	return nil
}

type mockVirtualNodeCache struct{}

func (m mockVirtualNodeCache) Add(obj interface{}) error {
	return nil
}

func (m mockVirtualNodeCache) Update(obj interface{}) error {
	return nil
}

func (m mockVirtualNodeCache) Delete(obj interface{}) error {
	return nil
}

func (m mockVirtualNodeCache) List() []interface{} {
	return nil
}

func (m mockVirtualNodeCache) ListKeys() []string {
	return nil
}

func (m mockVirtualNodeCache) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return virtualNodes["default"], true, nil
}

func (m mockVirtualNodeCache) GetByKey(key string) (item interface{}, exists bool, err error) {
	if virtualNode, ok := virtualNodes[key]; ok {
		return virtualNode, true, nil
	}
	return nil, false, nil
}

func (m mockVirtualNodeCache) Replace(i []interface{}, s string) error {
	return nil
}

func (m mockVirtualNodeCache) Resync() error {
	return nil
}

type MockContainerEngineClient struct{}

func (m MockContainerEngineClient) GetVirtualNode(ctx context.Context, vnId, vnpId string) (*containerengine.VirtualNode, error) {
	if virtualNode, ok := virtualNodes[vnId]; ok {
		return virtualNode, nil
	}
	return &containerengine.VirtualNode{
		Id:                 &vnId,
		VirtualNodePoolId:  &vnpId,
		AvailabilityDomain: common.String("PHX-AD-1"),
	}, nil
}

func TestExtractNodeAddresses(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  []v1.NodeAddress
		err  error
	}{
		{
			name: "basic-complete",
			in:   "basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "basic-complete.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "basic-complete.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "no-external-ip",
			in:   "no-external-ip",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "no-external-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "no-external-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "no-internal-ip",
			in:   "no-internal-ip",
			out: []v1.NodeAddress{
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "no-internal-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "no-internal-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "invalid-external-ip",
			in:   "invalid-external-ip",
			out:  nil,
			err:  errors.New(`instance has invalid public address: "0.0.0."`),
		},
		{
			name: "invalid-internal-ip",
			in:   "invalid-internal-ip",
			out:  nil,
			err:  errors.New(`instance has invalid private address: "10.0.0."`),
		},
		{
			name: "no-hostname-label",
			in:   "no-hostname-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-subnet-dns-label",
			in:   "no-subnet-dns-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-vcn-dns-label",
			in:   "no-vcn-dns-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
	}

	cp := &CloudProvider{
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister:    &mockNodeLister{},
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.extractNodeAddresses(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("extractNodeAddresses(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("extractNodeAddresses(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceID(t *testing.T) {
	testCases := []struct {
		name string
		in   types.NodeName
		out  string
		err  error
	}{
		{
			name: "get instance id from instance in the cache",
			in:   "instance1",
			out:  "instance1",
			err:  nil,
		},
		{
			name: "get instance id from instance not in the cache",
			in:   "default",
			out:  "default",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceType(t *testing.T) {
	testCases := []struct {
		name string
		in   types.NodeName
		out  string
		err  error
	}{
		{
			name: "check node shape of instance in cache",
			in:   "instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "check node shape of instance not in cache",
			in:   "default",
			out:  "VM.Standard1.2",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceType(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceType(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceType(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceTypeByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  string
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "noncacheinstance",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "provider id for virtual node",
			in:   "ocid1.virtualnode.oc1.iad.default",
			out:  "",
			err:  nil,
		},
		{
			name: "provider id with provider prefix for virtual node",
			in:   providerPrefix + "ocid1.virtualnode.oc1.iad.default",
			out:  "",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceTypeByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceTypeByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceTypeByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestNodeAddressesByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  []v1.NodeAddress
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "provider id for virtual node",
			in:   "ocid1.virtualnode.oc1.iad.default",
			out:  []v1.NodeAddress{},
			err:  nil,
		},
		{
			name: "provider id with provider prefix for virtual node",
			in:   providerPrefix + "ocid1.virtualnode.oc1.iad.default",
			out:  []v1.NodeAddress{},
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.NodeAddressesByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("NodeAddressesByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("NodeAddressesByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceExistsByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  bool
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "instance1",
			out:  true,
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "instance1",
			out:  true,
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "noncacheinstance",
			out:  true,
			err:  nil,
		},
		{
			name: "provider id for virtual node and in cache",
			in:   "ocid1.virtualnode.oc1.iad.default",
			out:  true,
			err:  nil,
		},
		{
			name: "provider id for virtual node with provider prefix and not in cache",
			in:   providerPrefix + "ocid1.virtualnode.oc1.iad.noncache",
			out:  true,
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:       &mockNodeLister{},
		client:           MockOCIClient{},
		config:           &providercfg.Config{CompartmentID: "testCompartment"},
		logger:           zap.S(),
		instanceCache:    &mockInstanceCache{},
		virtualNodeCache: &mockVirtualNodeCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceExistsByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceExistsByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceExistsByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceShutdownByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  bool
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "instance1",
			out:  false,
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "instance1",
			out:  false,
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "noncacheinstance",
			out:  false,
			err:  nil,
		},
		{
			name: "provider id for virtual node",
			in:   "ocid1.virtualnode.oc1.iad.default",
			out:  false,
			err:  nil,
		},
		{
			name: "provider id with provider prefix for virtual node",
			in:   providerPrefix + "ocid1.virtualnode.oc1.iad.default",
			out:  false,
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceShutdownByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceShutdownByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceShutdownByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestGetCompartmentIDByInstanceID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  string
		err  error
	}{
		{
			name: "instance found in cache",
			in:   "instance1",
			out:  "compartment1",
			err:  nil,
		},
		{
			name: "instance found in node lister",
			in:   "default",
			out:  "default",
			err:  nil,
		},
		{
			name: "instance neither found in cache nor node lister",
			in:   "instancex",
			out:  "",
			err:  errors.New("compartmentID annotation missing in the node. Would retry"),
		},
	}

	cp := &CloudProvider{
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister:    &mockNodeLister{},
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.getCompartmentIDByInstanceID(tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("getCompartmentIDByInstanceID(%s) got error %s, expected %s", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("getCompartmentIDByInstanceID(%s) => %s, want %s", tt.in, result, tt.out)
			}
		})
	}
}

type mockNodeLister struct{}

func (s *mockNodeLister) List(selector labels.Selector) (ret []*v1.Node, err error) {
	var nodes []*v1.Node
	for _, n := range nodeList {
		if selector != nil {
			if selector.Matches(labels.Set(n.ObjectMeta.GetLabels())) {
				nodes = append(nodes, n)
			}
		} else {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func (s *mockNodeLister) Get(name string) (*v1.Node, error) {
	if node, ok := nodeList[name]; ok {
		return node, nil
	}
	return nil, errors.New("get node error")
}

func (s *mockNodeLister) ListWithPredicate() ([]*v1.Node, error) {
	return nil, nil
}

type mockEndpointSliceLister struct{}

func (s *mockEndpointSliceLister) List(selector labels.Selector) (ret []*v1discovery.EndpointSlice, err error) {
	return []*v1discovery.EndpointSlice{}, nil
}

func (s *mockEndpointSliceLister) EndpointSlices(namespace string) v1discoverylisters.EndpointSliceNamespaceLister {
	return &mockEndpointSliceNamespaceLister{}
}

type mockEndpointSliceNamespaceLister struct{}

func (s *mockEndpointSliceNamespaceLister) List(selector labels.Selector) (ret []*v1discovery.EndpointSlice, err error) {
	var endpointSlices []*v1discovery.EndpointSlice
	for _, es := range endpointSliceList {
		if selector != nil {
			if selector.Matches(labels.Set(es.ObjectMeta.GetLabels())) {
				endpointSlices = append(endpointSlices, es)
			}
		} else {
			endpointSlices = append(endpointSlices, es)
		}
	}
	return endpointSlices, nil
}

func (s *mockEndpointSliceNamespaceLister) Get(name string) (ret *v1discovery.EndpointSlice, err error) {
	if es, ok := endpointSliceList[name]; ok {
		return es, nil
	}
	return nil, errors.New("get endpointSlice error")
}
