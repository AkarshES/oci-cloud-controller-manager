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
	nor := norv1beta1.NodeOperationRule{
		Spec: norv1beta1.NodeOperationRuleSpec{
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
	nor := norv1beta1.NodeOperationRule{
		Spec: norv1beta1.NodeOperationRuleSpec{
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

// TestReplaceBootVolumeClusterNode verifies the correctness of the RebootClusterNode function.
func TestReplaceBootVolumeClusterNode(t *testing.T) {
	nor := norv1beta1.NodeOperationRule{
		Spec: norv1beta1.NodeOperationRuleSpec{
			NodeEvictionSettings: norv1beta1.NodeEvictionSettings{
				EvictionGracePeriod:             60,
				IsForceActionAfterGraceDuration: true,
			},
		},
	}

	// Create a mock container engine client.
	mockContainerEngine := &mockContainerEngineClient{
		ReplaceBootVolumeClusterNodeFunc: func(ctx context.Context, req containerengine.ReplaceBootVolumeClusterNodeRequest) (containerengine.ReplaceBootVolumeClusterNodeResponse, error) {
			// Simulate a successful response.
			return containerengine.ReplaceBootVolumeClusterNodeResponse{
				OpcRequestId:     common.String("success-request-id"),
				OpcWorkRequestId: common.String("success-work-request-id"),
			}, nil
		},
	}

	workRequestId, err := mockContainerEngine.ReplaceBootVolumeClusterNode(context.Background(), nodeId, clusterId, nor)
	assert.NoError(t, err)
	assert.Equal(t, "success-work-request-id", workRequestId)
}

// TestReplaceBootVolumeClusterNodeFailure simulates a scenario where the cycling of a node fails due to an error.
func TestReplaceBootVolumeClusterNodeFailure(t *testing.T) {
	nor := norv1beta1.NodeOperationRule{
		Spec: norv1beta1.NodeOperationRuleSpec{
			NodeEvictionSettings: norv1beta1.NodeEvictionSettings{
				EvictionGracePeriod:             60,
				IsForceActionAfterGraceDuration: true,
			},
		},
	}

	mockContainerEngine := &mockContainerEngineClient{
		ReplaceBootVolumeClusterNodeFunc: func(ctx context.Context, req containerengine.ReplaceBootVolumeClusterNodeRequest) (containerengine.ReplaceBootVolumeClusterNodeResponse, error) {
			// Simulate a failure response.
			return containerengine.ReplaceBootVolumeClusterNodeResponse{
				OpcRequestId:     common.String(""),
				OpcWorkRequestId: common.String(""),
			}, apiError
		},
	}

	workRequestId, err := mockContainerEngine.ReplaceBootVolumeClusterNode(context.Background(), nodeId, clusterId, nor)
	assert.Error(t, err)
	assert.Equal(t, "", workRequestId)
}

type mockContainerEngineClient struct {
	RebootClusterNodeFunc            func(ctx context.Context, req containerengine.RebootClusterNodeRequest) (containerengine.RebootClusterNodeResponse, error)
	ReplaceBootVolumeClusterNodeFunc func(ctx context.Context, req containerengine.ReplaceBootVolumeClusterNodeRequest) (containerengine.ReplaceBootVolumeClusterNodeResponse, error)
}

func (m *mockContainerEngineClient) RebootClusterNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) (string, error) {
	req := defaultRebootClusterNodeRequest(nodeId, clusterId, nor)
	response, err := m.RebootClusterNodeFunc(ctx, req)
	return *response.OpcWorkRequestId, err
}

func (m *mockContainerEngineClient) ReplaceBootVolumeClusterNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) (string, error) {
	req := defaultReplaceBootVolumeClusterNodeRequest(nodeId, clusterId, nor)
	response, err := m.ReplaceBootVolumeClusterNodeFunc(ctx, req)
	return *response.OpcWorkRequestId, err
}

func defaultRebootClusterNodeRequest(nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) containerengine.RebootClusterNodeRequest {
	req := containerengine.RebootClusterNodeRequest{
		NodeId:                   common.String(nodeId),
		ClusterId:                common.String(clusterId),
		RebootClusterNodeDetails: *(*containerengine.RebootClusterNodeDetails)(unsafe.Pointer(&nor.Spec.NodeEvictionSettings)),
	}
	return req
}

func defaultReplaceBootVolumeClusterNodeRequest(nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) containerengine.ReplaceBootVolumeClusterNodeRequest {
	req := containerengine.ReplaceBootVolumeClusterNodeRequest{
		NodeId:                              common.String(nodeId),
		ClusterId:                           common.String(clusterId),
		ReplaceBootVolumeClusterNodeDetails: *(*containerengine.ReplaceBootVolumeClusterNodeDetails)(unsafe.Pointer(&nor.Spec.NodeEvictionSettings)),
	}
	return req
}
