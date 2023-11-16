/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"go.uber.org/zap"
	authv1 "k8s.io/api/authentication/v1"
	"reflect"
	"testing"

	"github.com/oracle/oci-cloud-controller-manager/api/v1beta1"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/core"
)

func TestComputeAveragesByReturnCode(t *testing.T) {
	testCases := []struct {
		name     string
		metrics  []ErrorMetric
		expected map[string]float64
	}{
		{
			name:     "base case",
			metrics:  nil,
			expected: map[string]float64{},
		},
		{
			name:     "base case 2",
			metrics:  endToEndLatencySlice{}.ErrorMetric(),
			expected: map[string]float64{},
		},
		{
			name: "base case e2e time",
			metrics: endToEndLatencySlice{
				endToEndLatency{timeTaken: 5.0},
				endToEndLatency{timeTaken: 10.0},
			}.ErrorMetric(),
			expected: map[string]float64{
				util.Success: 7.5,
			},
		},
		{
			name: "base case vnic attachment time",
			metrics: VnicAttachmentResponseSlice{
				VnicAttachmentResponse{timeTaken: 8.5},
				VnicAttachmentResponse{timeTaken: 6.5},
			}.ErrorMetric(),
			expected: map[string]float64{
				util.Success: 7.5,
			},
		},
		{
			name: "base case ip application time",
			metrics: IPAllocationSlice{
				IPAllocation{timeTaken: 5.0},
			}.ErrorMetric(),
			expected: map[string]float64{
				util.Success: 5.0,
			},
		},
		{
			name: "ip application failures",
			metrics: IPAllocationSlice{
				IPAllocation{timeTaken: 1.0, err: errors.New("http status code: 500")},
				IPAllocation{timeTaken: 2.0, err: errors.New("http status code: 500")},
				IPAllocation{timeTaken: 3.0, err: errors.New("http status code: 429")},
				IPAllocation{timeTaken: 4.0, err: errors.New("http status code: 401")},
				IPAllocation{timeTaken: 5.0},
				IPAllocation{timeTaken: 6.0},
			}.ErrorMetric(),
			expected: map[string]float64{
				util.Err5XX:  1.5,
				util.Err429:  3.0,
				util.Err4XX:  4.0,
				util.Success: 5.5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			averages := computeAveragesByReturnCode(tc.metrics)
			if !reflect.DeepEqual(averages, tc.expected) {
				t.Errorf("expected metrics:\n%+v\nbut got:\n%+v", tc.expected, averages)
			}
		})
	}
}

var (
	trueVal      = true
	falseVal     = false
	testAddress1 = "1.1.1.1"
	testAddress2 = "2.2.2.2"
)

func TestFilterPrivateIp(t *testing.T) {
	testCases := []struct {
		name     string
		ips      []core.PrivateIp
		expected []core.PrivateIp
	}{
		{
			name:     "base case",
			ips:      []core.PrivateIp{},
			expected: []core.PrivateIp{},
		},
		{
			name: "only primary ip",
			ips: []core.PrivateIp{
				{IsPrimary: &trueVal},
			},
			expected: []core.PrivateIp{},
		},
		{
			name: "primary and secondary ip",
			ips: []core.PrivateIp{
				{IsPrimary: &trueVal},
				{IsPrimary: &falseVal, IpAddress: &testAddress1},
			},
			expected: []core.PrivateIp{
				{IsPrimary: &falseVal, IpAddress: &testAddress1},
			},
		},
		{
			name: "only secondary ip",
			ips: []core.PrivateIp{
				{IsPrimary: &falseVal, IpAddress: &testAddress1},
			},
			expected: []core.PrivateIp{
				{IsPrimary: &falseVal, IpAddress: &testAddress1},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered := filterPrivateIp(tc.ips)
			if !reflect.DeepEqual(filtered, tc.expected) {
				t.Errorf("expected ips:\n%+v\nbut got:\n%+v", tc.expected, filtered)
			}
		})
	}
}

func TestTotalAllocatedSecondaryIpsForInstance(t *testing.T) {
	testCases := []struct {
		name     string
		ips      map[string][]core.PrivateIp
		expected int
	}{
		{
			name:     "base case",
			ips:      map[string][]core.PrivateIp{},
			expected: 0,
		},
		{
			name: "one vnic, two ips",
			ips: map[string][]core.PrivateIp{
				"one": {{IpAddress: &testAddress1}, {IpAddress: &testAddress2}},
			},
			expected: 2,
		},
		{
			name: "two vnic, 1/2 ips ",
			ips: map[string][]core.PrivateIp{
				"one": {{IpAddress: &testAddress1}, {IpAddress: &testAddress2}},
				"two": {{IpAddress: &testAddress2}},
			},
			expected: 3,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allocated := totalAllocatedSecondaryIpsForInstance(tc.ips)
			if !reflect.DeepEqual(allocated, tc.expected) {
				t.Errorf("expected ip count:\n%+v\nbut got:\n%+v", tc.expected, allocated)
			}
		})
	}
}

func TestGetAdditionalSecondaryIPsNeededPerVNIC(t *testing.T) {
	testCases := []struct {
		name                   string
		existingIpsByVnic      map[string][]core.PrivateIp
		additionalSecondaryIps int
		expected               []VnicIPAllocations
		err                    error
	}{
		{
			name:                   "base case",
			existingIpsByVnic:      map[string][]core.PrivateIp{},
			additionalSecondaryIps: 0,
			expected:               []VnicIPAllocations{},
			err:                    nil,
		},
		{
			name: "one vnic with one additional IP required",
			existingIpsByVnic: map[string][]core.PrivateIp{
				"one": {{IpAddress: &testAddress1}},
			},
			additionalSecondaryIps: 1,
			expected:               []VnicIPAllocations{{"one", 1}},
			err:                    nil,
		},
		{
			name: "one vnic with space for required IPs",
			existingIpsByVnic: map[string][]core.PrivateIp{
				"one": {{IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1},
					{IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1},
					{IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}},
			},
			additionalSecondaryIps: 13,
			expected:               []VnicIPAllocations{{"one", 13}},
			err:                    nil,
		},
		{
			name: "one vnic without space for required IPs",
			existingIpsByVnic: map[string][]core.PrivateIp{
				"one": {{IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1},
					{IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1},
					{IpAddress: &testAddress1}, {IpAddress: &testAddress1}, {IpAddress: &testAddress1}},
			},
			additionalSecondaryIps: 31,
			expected:               nil,
			err:                    errors.New("failed to allocate the required number of IPs with existing VNICs"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allocation, err := getAdditionalSecondaryIPsNeededPerVNIC(tc.existingIpsByVnic, tc.additionalSecondaryIps)
			if (err == nil && tc.err != nil) || err != nil && tc.err == nil {
				t.Errorf("expected err:\n%+v\nbut got err:\n%+v", tc.err, err)
				t.FailNow()
			}
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("expected err:\n%+v\nbut got err:\n%+v", tc.expected, allocation)
			}
			if !reflect.DeepEqual(allocation, tc.expected) {
				t.Errorf("expected ip allocation:\n%+v\nbut got:\n%+v", tc.expected, allocation)
			}
		})
	}
}

var (
	one         = "one"
	mac1        = "11.bb.cc.dd.ee.66"
	routerIP1   = "192.168.1.1"
	cidr1       = "10.0.0.0/64"
	subnetVnic1 = SubnetVnic{
		Vnic:   &core.Vnic{Id: &one, MacAddress: &mac1},
		Subnet: &core.Subnet{VirtualRouterIp: &routerIP1, CidrBlock: &cidr1},
	}
	npnVnic1 = v1beta1.VNICAddress{
		VNICID:     &one,
		MACAddress: &mac1,
		RouterIP:   &routerIP1,
		Addresses:  []*string{&testAddress1, &testAddress2},
		SubnetCidr: &cidr1,
	}
)

func TestConvertCoreVNICtoNPNStatus(t *testing.T) {
	testCases := []struct {
		name                   string
		existingSecondaryVNICs []SubnetVnic
		additionalSecondaryIps map[string][]core.PrivateIp
		expected               []v1beta1.VNICAddress
	}{
		{
			name:                   "base case",
			existingSecondaryVNICs: []SubnetVnic{},
			additionalSecondaryIps: map[string][]core.PrivateIp{},
			expected:               []v1beta1.VNICAddress{},
		},
		{
			name:                   "base case",
			existingSecondaryVNICs: []SubnetVnic{subnetVnic1},
			additionalSecondaryIps: map[string][]core.PrivateIp{
				one: {{IpAddress: &testAddress1}, {IpAddress: &testAddress2}},
			},
			expected: []v1beta1.VNICAddress{npnVnic1},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vnics := convertCoreVNICtoNPNStatus(tc.existingSecondaryVNICs, tc.additionalSecondaryIps)
			if !reflect.DeepEqual(vnics, tc.expected) {
				t.Errorf("expected npnVNIC to be:\n%+v\nbut got:\n%+v", tc.expected, vnics)
			}
		})
	}
}

type MockOCIClient struct {
}

func (c MockOCIClient) LoadBalancer(*zap.SugaredLogger, string, string, *authv1.TokenRequest) client.GenericLoadBalancerInterface {
	return nil
}

func (c MockOCIClient) BlockStorage() client.BlockStorageInterface {
	return nil
}

func (c MockOCIClient) FSS() client.FileStorageInterface {
	return nil
}

func (c MockOCIClient) Identity() client.IdentityInterface {
	return nil
}

func (c MockOCIClient) ContainerEngine() client.ContainerEngineInterface {
	return nil
}

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

func (c *MockVirtualNetworkClient) CreateNetworkSecurityGroup(ctx context.Context, compartmentId, vcnId, displayName, lbId string) (*core.NetworkSecurityGroup, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) UpdateNetworkSecurityGroup(ctx context.Context, id, etag string, freeformTags map[string]string) (*core.NetworkSecurityGroup, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetNetworkSecurityGroup(ctx context.Context, id string) (*core.NetworkSecurityGroup, *string, error) {
	return nil, nil, nil
}

func (c *MockVirtualNetworkClient) ListNetworkSecurityGroups(ctx context.Context, displayName, compartmentId, vcnId string) ([]core.NetworkSecurityGroup, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) DeleteNetworkSecurityGroup(ctx context.Context, id, etag string) (*string, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) AddNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.AddNetworkSecurityGroupSecurityRulesDetails) (*core.AddNetworkSecurityGroupSecurityRulesResponse, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) RemoveNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.RemoveNetworkSecurityGroupSecurityRulesDetails) (*core.RemoveNetworkSecurityGroupSecurityRulesResponse, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) ListNetworkSecurityGroupSecurityRules(ctx context.Context, id string, direction core.ListNetworkSecurityGroupSecurityRulesDirectionEnum) ([]core.SecurityRule, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) UpdateNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.UpdateNetworkSecurityGroupSecurityRulesDetails) (*core.UpdateNetworkSecurityGroupSecurityRulesResponse, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSubnetFromCacheByIP(ip string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) IsRegionalSubnet(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (c *MockVirtualNetworkClient) GetVcn(ctx context.Context, id string) (*core.Vcn, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) ListPrivateIps(ctx context.Context, vnicId string) ([]core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetPrivateIp(ctx context.Context, id string) (*core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) CreatePrivateIp(ctx context.Context, vnicID string) (*core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetPublicIpByIpAddress(ctx context.Context, id string) (*core.PublicIp, error) {
	return nil, nil
}

// MockComputeClient mocks Compute client implementation
type MockComputeClient struct{}

func (c *MockComputeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	return nil, nil
}

func (c *MockComputeClient) GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error) {
	return nil, nil
}

func (c *MockComputeClient) GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	return nil, nil
}

func (c *MockComputeClient) AttachVnic(ctx context.Context, instanceID, subnetId *string, nsgIds []*string, skipSourceDestCheck *bool) (response core.VnicAttachment, err error) {
	return core.VnicAttachment{}, nil
}

func (c *MockComputeClient) FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) AttachVolume(ctx context.Context, instanceID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) AttachParavirtualizedVolume(ctx context.Context, instanceID, volumeID string, isPvEncryptionInTransitEnabled bool) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) WaitForVolumeAttached(ctx context.Context, attachmentID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) DetachVolume(ctx context.Context, id string) error {
	return nil
}

func (c *MockComputeClient) WaitForVolumeDetached(ctx context.Context, attachmentID string) error {
	return nil
}

func (c *MockComputeClient) FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockOCIClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (MockOCIClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

func (c *MockVirtualNetworkClient) GetVNIC(ctx context.Context, id string) (*core.Vnic, error) {
	vnicCounter++
	if vnics[id].LifecycleState == core.VnicLifecycleStateProvisioning && vnicCounter%3 == 0 {
		copy := vnics[id]
		copy.LifecycleState = core.VnicLifecycleStateAvailable
		return copy, nil // Available
	}
	return vnics[id], nil
}

func (c *MockComputeClient) ListVnicAttachments(ctx context.Context, compartmentID, instanceID string) ([]core.VnicAttachment, error) {
	return attachedVnicsList[compartmentID], nil
}

func (c *MockComputeClient) GetVnicAttachment(ctx context.Context, vnicAttachmentId *string) (response *core.VnicAttachment, err error) {
	attachmentCounter++
	resp := vnicAttachments[*vnicAttachmentId]
	if *resp.Id == "attachmentid5" {
		resp.LifecycleState = core.VnicAttachmentLifecycleStateDetached // Detached
	}
	if attachmentCounter%3 == 0 {
		resp.LifecycleState = core.VnicAttachmentLifecycleStateAttached // Attached
	}
	return &resp, nil
}

var (
	vnicAttachments = map[string]core.VnicAttachment{
		"attachmentid1": {
			Id:             common.String("attachmentid1"),
			VnicId:         common.String("vnic1"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttached,
		},
		"attachmentid2": {
			Id:             common.String("attachmentid2"),
			VnicId:         common.String("vnic2"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttached,
		},
		"attachmentid3": {
			Id:             common.String("attachmentid3"),
			VnicId:         common.String("vnic3"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttached,
		},
		"attachmentid4": {
			Id:             common.String("attachmentid4"),
			VnicId:         common.String("vnic4"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttached,
		},
		"attachmentid5": {
			Id:             common.String("attachmentid5"),
			VnicId:         common.String("vnic5"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttached,
		},
		"attachmentid6": {
			Id:             common.String("attachmentid6"),
			VnicId:         common.String("vnic6"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttached,
		},
		"attachmentid7": {
			Id:             common.String("attachmentid7"),
			VnicId:         common.String("vnic7"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttached,
		},
		"attachmentid8": {
			Id:             common.String("attachmentid8"),
			VnicId:         common.String("vnic8"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttaching,
		},
		"attachmentid9": {
			Id:             common.String("attachmentid9"),
			VnicId:         common.String("vnic9"),
			LifecycleState: core.VnicAttachmentLifecycleStateDetached,
		},
		"attachmentid10": {
			Id:             common.String("attachmentid10"),
			VnicId:         common.String("vnic10"),
			LifecycleState: core.VnicAttachmentLifecycleStateDetached,
		},
		"attachmentid11": {
			Id:             common.String("attachmentid11"),
			VnicId:         common.String("vnic11"),
			LifecycleState: core.VnicAttachmentLifecycleStateDetached,
		},
		"attachmentid12": {
			Id:             common.String("attachmentid12"),
			VnicId:         common.String("vnic12"),
			LifecycleState: core.VnicAttachmentLifecycleStateAttaching,
		},
	}
	False  = false
	Subnet = "test-subnet"
	vnics  = map[string]*core.Vnic{
		"vnic1": {
			Id:             common.String("vnic1"),
			LifecycleState: core.VnicLifecycleStateAvailable,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic2": {
			Id:             common.String("vnic2"),
			LifecycleState: core.VnicLifecycleStateAvailable,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic3": {
			Id:             common.String("vnic3"),
			LifecycleState: core.VnicLifecycleStateAvailable,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic4": {
			Id:             common.String("vnic4"),
			LifecycleState: core.VnicLifecycleStateAvailable,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic5": {
			Id:             common.String("vnic5"),
			LifecycleState: core.VnicLifecycleStateAvailable,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic6": {
			Id:             common.String("vnic6"),
			LifecycleState: core.VnicLifecycleStateProvisioning,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic7": {
			Id:             common.String("vnic7"),
			LifecycleState: core.VnicLifecycleStateAvailable,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic8": {
			Id:             common.String("vnic8"),
			LifecycleState: core.VnicLifecycleStateAvailable,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic9": {
			Id:             common.String("vnic9"),
			LifecycleState: core.VnicLifecycleStateProvisioning,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic10": {
			Id:             common.String("vnic10"),
			LifecycleState: core.VnicLifecycleStateTerminating,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic11": {
			Id:             common.String("vnic11"),
			LifecycleState: core.VnicLifecycleStateTerminated,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
		"vnic12": {
			Id:             nil,
			LifecycleState: core.VnicLifecycleStateTerminated,
			IsPrimary:      &False,
			SubnetId:       &Subnet,
		},
	}

	attachedVnicsList = map[string][]core.VnicAttachment{
		"vnics attached": {
			{
				Id:             common.String("attachmentid1"),
				VnicId:         common.String("vnic1"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttached,
			},
			{
				Id:             common.String("attachmentid2"),
				VnicId:         common.String("vnic2"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttached,
			},
			{
				Id:             common.String("attachmentid3"),
				VnicId:         common.String("vnic3"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttached,
			},
			{
				Id:             common.String("attachmentid4"),
				VnicId:         common.String("vnic4"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttached,
			},
		},
		"single vnic not attached": {
			{
				Id:             common.String("attachmentid6"),
				VnicId:         common.String("vnic6"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttached,
			},
			{
				Id:             common.String("attachmentid7"),
				VnicId:         common.String("vnic7"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttached,
			},
			{
				Id:             common.String("attachmentid8"),
				VnicId:         common.String("vnic8"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttaching,
			},
		},
		"vnic in detaching or detached after a while": {
			{
				Id:             common.String("attachmentid1"),
				VnicId:         common.String("vnic1"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttached,
			},
			{
				Id:             common.String("attachmentid5"),
				VnicId:         common.String("vnic5"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttaching,
			},
		},
		"vnic not available": {
			{
				Id:             common.String("attachmentid11"),
				VnicId:         common.String("vnic11"),
				LifecycleState: core.VnicAttachmentLifecycleStateDetached,
			},
			{
				Id:             common.String("attachmentid12"),
				VnicId:         common.String("vnic12"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttaching,
			},
		},
		"vnic becomes available eventually": {
			{
				Id:             common.String("attachmentid6"),
				VnicId:         common.String("vnic6"),
				LifecycleState: core.VnicAttachmentLifecycleStateAttached,
			},
		},
	}
	attachmentCounter = 1
	vnicCounter       = 1
)

func TestValidateVnicAttachmentsAreInAttachedState(t *testing.T) {
	testCases := []struct {
		name              string
		in                string
		compartmentid     string
		output            bool
		requiredVnicCount int
		err               error
		counter           int
	}{
		{
			name:              "all vnics attached",
			in:                "instanceid",
			compartmentid:     "vnics attached",
			output:            true,
			requiredVnicCount: 4,
			err:               nil,
		},
		{
			name:              "one vnic stuck in attaching",
			in:                "instanceid",
			compartmentid:     "single vnic not attached",
			output:            true,
			requiredVnicCount: 3,
			err:               nil,
		},
		{
			name:              "vnics in other lifecycle states",
			in:                "instanceid",
			compartmentid:     "vnic in detaching or detached after a while",
			output:            false,
			requiredVnicCount: 2,
			err:               errors.New("vnic attachment is in detaching/detached state"),
		},
		{
			name:              "not enough vnic attached",
			in:                "instanceid",
			compartmentid:     "vnic not available",
			output:            false,
			requiredVnicCount: 2,
			err:               errNotEnoughVnicsAttached,
		},
		{
			name:              "vnic becomes available eventually",
			in:                "instanceid",
			compartmentid:     "vnic becomes available eventually",
			output:            true,
			requiredVnicCount: 1,
			err:               nil,
		},
	}

	npn := &NativePodNetworkReconciler{
		OCIClient: MockOCIClient{},
	}

	t.Parallel()
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, existingSecondaryIpsbyVNIC, _ := npn.getPrimaryAndSecondaryVNICs(context.Background(), tt.compartmentid, tt.in)
			result, err := npn.validateVnicAttachmentsAreInAttachedState(context.Background(), tt.in, tt.requiredVnicCount, existingSecondaryIpsbyVNIC)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("validateVnicAttachmentsAreInAttachedState(%s) got error %s, expected %s", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.output) {
				t.Errorf("validateVnicAttachmentsAreInAttachedState(%s) => %t, want %t", tt.in, result, tt.output)
			}
		})
	}
}
