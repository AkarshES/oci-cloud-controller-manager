package ociclient

import (
	"context"
	"time"

	"github.com/oracle/oci-go-sdk/v65/identity"
)

func (c *client) ListCompartments(ctx context.Context, tenancy string) ([]identity.Compartment, error) {
	var result []identity.Compartment

	req := identity.ListCompartmentsRequest{
		CompartmentId:   &tenancy,
		RequestMetadata: c.requestMetadata,
		Limit:           &listLimit,
	}

	for {
		resp, err := c.identity.ListCompartments(ctx, req)
		incRequestCounter(err, listVerb, compartmentResource)
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

func (c *client) GetCompartment(ctx context.Context, id string) (*identity.Compartment, error) {
	resp, err := c.identity.GetCompartment(ctx, identity.GetCompartmentRequest{
		CompartmentId:   &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, compartmentResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.Compartment, nil
}

func (c *client) ListADs(ctx context.Context, compartment string) ([]identity.AvailabilityDomain, error) {
	resp, err := c.identity.ListAvailabilityDomains(ctx, identity.ListAvailabilityDomainsRequest{
		CompartmentId:   &compartment,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, listVerb, availabilityDomainResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return resp.Items, nil
}

func (c *client) ListFaultDomains(ctx context.Context, tenancyID string, ad string) ([]identity.FaultDomain, error) {
	resp, err := c.identity.ListFaultDomains(ctx, identity.ListFaultDomainsRequest{
		CompartmentId:      &tenancyID,
		AvailabilityDomain: &ad,
		RequestMetadata:    c.requestMetadata,
	})
	incRequestCounter(err, listVerb, faultDomainResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return resp.Items, nil
}

func (c *client) GetTenancy(ctx context.Context, id string) (*identity.Tenancy, error) {
	resp, err := c.identity.GetTenancy(ctx, identity.GetTenancyRequest{
		TenancyId:       &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, tenancyResource)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.Tenancy, nil
}
