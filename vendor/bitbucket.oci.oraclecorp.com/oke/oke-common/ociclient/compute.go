package ociclient

import (
	"context"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
)

func (c *client) GetImage(ctx context.Context, image string) (*core.Image, error) {
	resp, err := c.compute.GetImage(ctx, core.GetImageRequest{
		ImageId: &image,
	})
	incRequestCounter(err, getVerb, imageResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.Image, nil
}

func (c *client) ListImages(ctx context.Context, compartment string, shape *string) ([]core.Image, error) {
	var result []core.Image

	req := core.ListImagesRequest{
		CompartmentId: &compartment,
		Shape:         shape,
		Limit:         &listLimit,
	}

	for {
		resp, err := c.compute.ListImages(ctx, req)
		incRequestCounter(err, listVerb, imageResource)
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

func (c *client) ListShapes(ctx context.Context, compartment string, imageID *string, ad *string) ([]core.Shape, error) {
	var result []core.Shape

	req := core.ListShapesRequest{
		CompartmentId:      &compartment,
		ImageId:            imageID,
		AvailabilityDomain: ad,
		RequestMetadata:    c.requestMetadata,
		Limit:              &listLimit,
	}

	for {
		resp, err := c.compute.ListShapes(ctx, req)
		incRequestCounter(err, listVerb, shapeResource)
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

func (c *client) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	resp, err := c.compute.GetInstance(ctx, core.GetInstanceRequest{
		InstanceId:      &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, instanceResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.Instance, nil
}

func (c *client) LaunchInstance(ctx context.Context, req core.LaunchInstanceRequest) (*core.Instance, error) {
	if req.RequestMetadata.RetryPolicy == nil {
		req.RequestMetadata.RetryPolicy = c.requestMetadata.RetryPolicy
	}

	resp, err := c.compute.LaunchInstance(ctx, req)
	incRequestCounter(err, createVerb, instanceResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.Instance, nil
}

func (c *client) ListInstances(ctx context.Context, req core.ListInstancesRequest) ([]core.Instance, error) {
	if req.RequestMetadata.RetryPolicy == nil {
		req.RequestMetadata.RetryPolicy = c.requestMetadata.RetryPolicy
	}

	var result []core.Instance

	for {
		req.Limit = &listLimit
		resp, err := c.compute.ListInstances(ctx, req)
		incRequestCounter(err, listVerb, instanceResource)
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

func (c *client) TerminateInstance(ctx context.Context, id string) error {
	resp, err := c.compute.TerminateInstance(ctx, core.TerminateInstanceRequest{
		InstanceId:      &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, instanceResource)
	if err != nil {
		return NewOCIClientError(resp.OpcRequestId, err)
	}

	return nil
}

func (c *client) GetPrimaryVNIC(ctx context.Context, compartmentID string, instanceID string) (*core.Vnic, error) {
	opts := core.ListVnicAttachmentsRequest{
		CompartmentId:   &compartmentID,
		InstanceId:      &instanceID,
		Limit:           &listLimit,
		RequestMetadata: c.requestMetadata,
	}

	for {
		resp, err := c.compute.ListVnicAttachments(ctx, opts)
		incRequestCounter(err, listVerb, vnicAttachmentResource)
		if err != nil {
			return nil, NewOCIClientError(resp.OpcRequestId, err)
		}

		for _, attachment := range resp.Items {
			vnicResp, err := c.virtnet.GetVnic(ctx, core.GetVnicRequest{
				VnicId: attachment.VnicId,
			})
			incRequestCounter(err, getVerb, vnicResource)
			if err != nil {
				return nil, NewOCIClientError(resp.OpcRequestId, err)
			}

			if vnicResp.IsPrimary != nil && *vnicResp.IsPrimary {
				return &vnicResp.Vnic, nil
			}
		}

		if resp.OpcNextPage == nil {
			break
		}

		opts.Page = resp.OpcNextPage
		time.Sleep(50 * time.Millisecond)
	}

	return nil, ErrNotFound
}

func (c *client) GetImageShapeCompatibilityEntry(ctx context.Context, imageId string, shapeName string) (*core.GetImageShapeCompatibilityEntryResponse, error) {
	req := core.GetImageShapeCompatibilityEntryRequest{
		ImageId:         &imageId,
		ShapeName:       &shapeName,
		RequestMetadata: c.requestMetadata,
	}

	resp, err := c.compute.GetImageShapeCompatibilityEntry(ctx, req)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}
	return &resp, nil
}
