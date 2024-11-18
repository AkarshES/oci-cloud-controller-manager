package client

import (
	"context"
	norv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/node-cycling/v1beta1"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/pkg/errors"
	"strconv"
)

type ContainerEngineInterface interface {
	GetVirtualNode(ctx context.Context, virtualNodeId, virtualNodePoolId string) (*containerengine.VirtualNode, error)
	RebootClusterNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) (string, error)
	ReplaceBootVolumeClusterNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) (string, error)
}

func (c *client) GetVirtualNode(ctx context.Context, virtualNodeId, virtualNodePoolId string) (*containerengine.VirtualNode, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetVirtualNode")
	}

	resp, err := c.containerEngine.GetVirtualNode(ctx, containerengine.GetVirtualNodeRequest{
		VirtualNodeId:     common.String(virtualNodeId),
		VirtualNodePoolId: common.String(virtualNodePoolId),
		RequestMetadata:   c.requestMetadata,
	})
	incRequestCounter(err, getVerb, virtualNodeResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.VirtualNode, nil
}

// IsVirtualNodeInTerminalState returns true if the virtual node is in a terminal state, false otherwise.
func IsVirtualNodeInTerminalState(virtualNode *containerengine.VirtualNode) bool {
	return virtualNode.LifecycleState == containerengine.VirtualNodeLifecycleStateDeleted ||
		virtualNode.LifecycleState == containerengine.VirtualNodeLifecycleStateFailed
}

// RebootClusterNode initiates a reboot operation for a specified node within a cluster.
// It takes the node ID, cluster ID, and a NodeOperationRequest object as input.
// The function returns the work request ID associated with the reboot operation.
//
// Parameters:
// - ctx: A context.Context object for managing cancellations and timeouts.
// - nodeId: A string representing the unique identifier of the node to be rebooted.
// - clusterId: A string representing the unique identifier of the cluster where the node resides.
// - nor: An instance of norv1beta1.NodeOperationRequest containing additional details for the reboot operation.
//
// Returns:
// - A string representing the work request ID associated with the reboot operation.
// - An error indicating any issues encountered during the reboot operation; otherwise, returns nil.
func (c *client) RebootClusterNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) (string, error) {
	evictionGracePeriod := strconv.Itoa(nor.Spec.NodeEvictionSettings.EvictionGracePeriod)
	rebootClusterNodeDetails := &containerengine.RebootClusterNodeDetails{
		NodeEvictionSettings: &containerengine.NodeEvictionSettings{
			EvictionGraceDuration:           &evictionGracePeriod,
			IsForceActionAfterGraceDuration: &nor.Spec.NodeEvictionSettings.IsForceActionAfterGraceDuration,
		},
	}
	resp, err := c.containerEngine.RebootClusterNode(ctx, containerengine.RebootClusterNodeRequest{
		NodeId:                   common.String(nodeId),
		ClusterId:                common.String(clusterId),
		RebootClusterNodeDetails: *rebootClusterNodeDetails,
		RequestMetadata:          c.requestMetadata,
	})
	incRequestCounter(err, createVerb, rebootNodeWorkRequestResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcRequestId, nil
}

// ReplaceBootVolumeClusterNode initiates a cycling operation for a specified node within a cluster.
// It takes the node ID, cluster ID, and a NodeOperationRequest object as input.
// The function returns the work request ID associated with the cycling operation.
//
// Parameters:
// - ctx: A context.Context object for managing cancellations and timeouts.
// - nodeId: A string representing the unique identifier of the node to be cycled.
// - clusterId: A string representing the unique identifier of the cluster where the node resides.
// - nor: An instance of norv1beta1.NodeOperationRequest containing additional details for the cycling operation.
//
// Returns:
// - A string representing the work request ID associated with the cycling operation.
// - An error indicating any issues encountered during the cycling operation; otherwise, returns nil.
func (c *client) ReplaceBootVolumeClusterNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) (string, error) {

	evictionGracePeriod := strconv.Itoa(nor.Spec.NodeEvictionSettings.EvictionGracePeriod)
	replaceBootVolumeClusterNodeDetails := &containerengine.ReplaceBootVolumeClusterNodeDetails{
		NodeEvictionSettings: &containerengine.NodeEvictionSettings{
			EvictionGraceDuration:           &evictionGracePeriod,
			IsForceActionAfterGraceDuration: &nor.Spec.NodeEvictionSettings.IsForceActionAfterGraceDuration,
		},
	}

	resp, err := c.containerEngine.ReplaceBootVolumeClusterNode(ctx, containerengine.ReplaceBootVolumeClusterNodeRequest{
		NodeId:                              common.String(nodeId),
		ClusterId:                           common.String(clusterId),
		ReplaceBootVolumeClusterNodeDetails: *replaceBootVolumeClusterNodeDetails,
		RequestMetadata:                     c.requestMetadata,
	})
	incRequestCounter(err, createVerb, cycleNodeWorkRequestResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcRequestId, nil
}
