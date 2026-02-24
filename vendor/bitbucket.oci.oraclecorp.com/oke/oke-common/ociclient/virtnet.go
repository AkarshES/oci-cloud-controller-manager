package ociclient

import (
	"context"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
)

func (c *client) GetVCN(ctx context.Context, vcnID string) (*core.Vcn, error) {
	resp, err := c.virtnet.GetVcn(ctx, core.GetVcnRequest{
		VcnId: &vcnID,
	})
	incRequestCounter(err, getVerb, vcnResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.Vcn, nil
}

func (c *client) GetSubnet(ctx context.Context, subnetID string) (*core.Subnet, error) {
	resp, err := c.virtnet.GetSubnet(ctx, core.GetSubnetRequest{
		SubnetId:        &subnetID,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, subnetResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.Subnet, nil
}

func (c *client) ListSubnets(ctx context.Context, vcnID, compartmentID string) ([]core.Subnet, error) {
	var result []core.Subnet

	req := core.ListSubnetsRequest{
		CompartmentId:   &compartmentID,
		VcnId:           &vcnID,
		RequestMetadata: c.requestMetadata,
		Limit:           &listLimit,
	}

	for {
		resp, err := c.virtnet.ListSubnets(ctx, req)
		incRequestCounter(err, listVerb, subnetResource)
		if err != nil {
			return nil, NewOCIClientError(resp.OpcRequestId, err)
		}

		result = append(result, resp.Items...)
		if resp.OpcNextPage == nil {
			break
		}

		req.Page = resp.OpcNextPage
		time.Sleep(50 * time.Millisecond)
	}

	return result, nil
}

func (c *client) ListNSGs(ctx context.Context, vcnID, compartmentID string) ([]core.NetworkSecurityGroup, error) {
	var result []core.NetworkSecurityGroup

	req := core.ListNetworkSecurityGroupsRequest{
		CompartmentId:   &compartmentID,
		VcnId:           &vcnID,
		RequestMetadata: c.requestMetadata,
		Limit:           &listLimit,
	}

	for {
		resp, err := c.virtnet.ListNetworkSecurityGroups(ctx, req)
		incRequestCounter(err, listVerb, nsgResource)
		if err != nil {
			return nil, NewOCIClientError(resp.OpcRequestId, err)
		}

		result = append(result, resp.Items...)
		if resp.OpcNextPage == nil {
			break
		}

		req.Page = resp.OpcNextPage
		time.Sleep(50 * time.Millisecond)
	}

	return result, nil
}

func (c *client) GetNSG(ctx context.Context, nsgId string) (*core.NetworkSecurityGroup, error) {
	resp, err := c.virtnet.GetNetworkSecurityGroup(ctx, core.GetNetworkSecurityGroupRequest{
		NetworkSecurityGroupId: &nsgId,
		RequestMetadata:        c.requestMetadata,
	})
	incRequestCounter(err, getVerb, nsgResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.NetworkSecurityGroup, nil
}
