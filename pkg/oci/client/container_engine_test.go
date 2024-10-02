package client

import (
	"context"
	"errors"
	norv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/node-cycling/v1beta1"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

var (
	nodeId    = "test-node"
	clusterId = "test-cluster"
	apiError  = errors.New("unexpected API error")
)

// TestRebootClusterNode verifies the correctness of the RebootClusterNode function.
func TestRebootClusterNode(t *testing.T) {
	nor := norv1beta1.NodeOperationRequest{
		Spec: norv1beta1.NodeOperationRequestSpec{
			NodeEvictionSettings: norv1beta1.NodeEvictionSettings{},
		},
	}

	// Create a mock container engine client.
	mockContainerEngine := &mockContainerEngineClient{
		RebootClusterNodeFunc: func(ctx context.Context, req containerengine.RebootClusterNodeRequest) (containerengine.RebootClusterNodeResponse, error) {
			// Simulate a successful response.
			return containerengine.RebootClusterNodeResponse{
				OpcRequestId:     common.String("success-request-id"),
				OpcWorkRequestId: common.String("success-work-request-id"),
			}, nil
		},
	}

	workRequestId, err := mockContainerEngine.RebootClusterNode(context.Background(), nodeId, clusterId, nor)
	assert.NoError(t, err)
	assert.Equal(t, "success-work-request-id", workRequestId)
}

// TestRebootClusterNodeFailure simulates a scenario where the reboot of a node fails due to an error.
func TestRebootClusterNodeFailure(t *testing.T) {
	nor := norv1beta1.NodeOperationRequest{
		Spec: norv1beta1.NodeOperationRequestSpec{
			NodeEvictionSettings: norv1beta1.NodeEvictionSettings{
				EvictionGracePeriod: 60,
			},
		},
	}

	mockContainerEngine := &mockContainerEngineClient{
		RebootClusterNodeFunc: func(ctx context.Context, req containerengine.RebootClusterNodeRequest) (containerengine.RebootClusterNodeResponse, error) {
			// Simulate a failure response.
			return containerengine.RebootClusterNodeResponse{
				OpcRequestId:     common.String(""),
				OpcWorkRequestId: common.String(""),
			}, apiError
		},
	}

	workRequestId, err := mockContainerEngine.RebootClusterNode(context.Background(), nodeId, clusterId, nor)
	assert.Error(t, err)
	assert.Equal(t, "", workRequestId)
}

// TestCycleClusterNode verifies the correctness of the RebootClusterNode function.
func TestCycleClusterNode(t *testing.T) {
	nor := norv1beta1.NodeOperationRequest{
		Spec: norv1beta1.NodeOperationRequestSpec{
			NodeEvictionSettings: norv1beta1.NodeEvictionSettings{
				EvictionGracePeriod: 60,
			},
			CyclingActionDetails: norv1beta1.CyclingActionDetails{
				KubernetesVersion: "v1.30.0",
				NodeMetaData:      map[string]string{"key": "value"},
				SshPublicKey:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQD...",
				CycleMode:         "bootVolumeReplace",
				IsCycleInSyncNode: true,
			},
		},
	}

	// Create a mock container engine client.
	mockContainerEngine := &mockContainerEngineClient{
		CycleClusterNodeFunc: func(ctx context.Context, req containerengine.CycleClusterNodeRequest) (containerengine.CycleClusterNodeResponse, error) {
			// Simulate a successful response.
			return containerengine.CycleClusterNodeResponse{
				OpcRequestId:     common.String("success-request-id"),
				OpcWorkRequestId: common.String("success-work-request-id"),
			}, nil
		},
	}

	workRequestId, err := mockContainerEngine.CycleClusterNode(context.Background(), nodeId, clusterId, nor)
	assert.NoError(t, err)
	assert.Equal(t, "success-work-request-id", workRequestId)
}

// TestCycleClusterNodeFailure simulates a scenario where the cycling of a node fails due to an error.
func TestCycleClusterNodeFailure(t *testing.T) {
	nor := norv1beta1.NodeOperationRequest{
		Spec: norv1beta1.NodeOperationRequestSpec{
			NodeEvictionSettings: norv1beta1.NodeEvictionSettings{
				EvictionGracePeriod: 60,
			},
			CyclingActionDetails: norv1beta1.CyclingActionDetails{
				KubernetesVersion: "v1.30.0",
				NodeMetaData:      map[string]string{"key": "value"},
				SshPublicKey:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQD...",
				CycleMode:         "bootVolumeReplace",
				IsCycleInSyncNode: true,
			},
		},
	}

	mockContainerEngine := &mockContainerEngineClient{
		CycleClusterNodeFunc: func(ctx context.Context, req containerengine.CycleClusterNodeRequest) (containerengine.CycleClusterNodeResponse, error) {
			// Simulate a failure response.
			return containerengine.CycleClusterNodeResponse{
				OpcRequestId:     common.String(""),
				OpcWorkRequestId: common.String(""),
			}, apiError
		},
	}

	workRequestId, err := mockContainerEngine.CycleClusterNode(context.Background(), nodeId, clusterId, nor)
	assert.Error(t, err)
	assert.Equal(t, "", workRequestId)
}

type mockContainerEngineClient struct {
	RebootClusterNodeFunc func(ctx context.Context, req containerengine.RebootClusterNodeRequest) (containerengine.RebootClusterNodeResponse, error)
	CycleClusterNodeFunc  func(ctx context.Context, req containerengine.CycleClusterNodeRequest) (containerengine.CycleClusterNodeResponse, error)
}

func (m *mockContainerEngineClient) RebootClusterNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) (string, error) {
	req := defaultRebootClusterNodeRequest(nodeId, clusterId, nor)
	response, err := m.RebootClusterNodeFunc(ctx, req)
	return *response.OpcWorkRequestId, err
}

func (m *mockContainerEngineClient) CycleClusterNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) (string, error) {
	req := defaultCycleClusterNodeRequest(nodeId, clusterId, nor)
	response, err := m.CycleClusterNodeFunc(ctx, req)
	return *response.OpcWorkRequestId, err
}

func defaultRebootClusterNodeRequest(nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) containerengine.RebootClusterNodeRequest {
	req := containerengine.RebootClusterNodeRequest{
		NodeId:                   common.String(nodeId),
		ClusterId:                common.String(clusterId),
		RebootClusterNodeDetails: *(*containerengine.RebootClusterNodeDetails)(unsafe.Pointer(&nor.Spec.NodeEvictionSettings)),
	}
	return req
}

func defaultCycleClusterNodeRequest(nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) containerengine.CycleClusterNodeRequest {
	req := containerengine.CycleClusterNodeRequest{
		NodeId:                  common.String(nodeId),
		ClusterId:               common.String(clusterId),
		CycleClusterNodeDetails: *(*containerengine.CycleClusterNodeDetails)(unsafe.Pointer(&nor.Spec.CyclingActionDetails)),
	}
	return req
}
