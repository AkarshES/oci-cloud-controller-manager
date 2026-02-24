package ociclient

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"github.com/pkg/errors"
)

const workRequestPollInterval = 15 * time.Second

func (c *client) CreateLoadBalancer(ctx context.Context, details loadbalancer.CreateLoadBalancerDetails) (string, error) {
	resp, err := c.loadbalancer.CreateLoadBalancer(ctx, loadbalancer.CreateLoadBalancerRequest{
		CreateLoadBalancerDetails: details,
		RequestMetadata:           c.requestMetadata,
	})

	incRequestCounter(err, createVerb, loadBalancerResource)
	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) GetLoadBalancer(ctx context.Context, id string) (*loadbalancer.LoadBalancer, error) {
	resp, err := c.loadbalancer.GetLoadBalancer(ctx, loadbalancer.GetLoadBalancerRequest{
		LoadBalancerId:  &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, loadBalancerResource)

	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.LoadBalancer, nil
}

func (c *client) ListLoadBalancers(ctx context.Context, compartmentID string) ([]loadbalancer.LoadBalancer, error) {

	limit := int64(listLimit)
	req := loadbalancer.ListLoadBalancersRequest{
		CompartmentId:   &compartmentID,
		RequestMetadata: c.requestMetadata,
		Limit:           &limit,
	}

	var result []loadbalancer.LoadBalancer
	for {
		resp, err := c.loadbalancer.ListLoadBalancers(ctx, req)
		if err != nil {
			return nil, NewOCIClientError(resp.OpcRequestId, err)
		}
		incRequestCounter(err, listVerb, loadBalancerResource)

		result = append(result, resp.Items...)
		if resp.OpcNextPage == nil {
			break
		}

		req.Page = resp.OpcNextPage
		time.Sleep(50 * time.Millisecond)
	}

	return result, nil
}

func (c *client) AwaitWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error) {
	var wr *loadbalancer.WorkRequest
	err := wait.PollUntil(workRequestPollInterval, func() (done bool, err error) {
		twr, err := c.GetWorkRequest(ctx, id)
		if err != nil {
			if IsRetryable(err) {
				return false, nil
			}
			return true, errors.WithStack(err)
		}
		switch twr.LifecycleState {
		case loadbalancer.WorkRequestLifecycleStateSucceeded:
			wr = twr
			return true, nil
		case loadbalancer.WorkRequestLifecycleStateFailed:
			return false, errors.Errorf("WorkRequest %q failed: %s", id, *twr.Message)
		}
		return false, nil
	}, ctx.Done())
	return wr, err
}

func (c *client) GetWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error) {
	resp, err := c.loadbalancer.GetWorkRequest(ctx, loadbalancer.GetWorkRequestRequest{
		WorkRequestId:   &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, loadBalancerWorkRequestResource)

	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.WorkRequest, nil
}

func (c *client) CreateBackend(ctx context.Context, request loadbalancer.CreateBackendRequest) (string, error) {
	if request.RequestMetadata.RetryPolicy == nil {
		request.RequestMetadata.RetryPolicy = c.requestMetadata.RetryPolicy
	}

	resp, err := c.loadbalancer.CreateBackend(ctx, request)
	incRequestCounter(err, createVerb, loadBalancerBackendResource)

	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) CreateBackendSet(ctx context.Context, request loadbalancer.CreateBackendSetRequest) (string, error) {
	if request.RequestMetadata.RetryPolicy == nil {
		request.RequestMetadata.RetryPolicy = c.requestMetadata.RetryPolicy
	}

	resp, err := c.loadbalancer.CreateBackendSet(ctx, request)
	incRequestCounter(err, createVerb, loadBalancerBackendSetResource)

	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) CreateListener(ctx context.Context, request loadbalancer.CreateListenerRequest) (string, error) {
	if request.RequestMetadata.RetryPolicy == nil {
		request.RequestMetadata.RetryPolicy = c.requestMetadata.RetryPolicy
	}

	resp, err := c.loadbalancer.CreateListener(ctx, request)
	incRequestCounter(err, createVerb, loadBalancerlistenerResource)

	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	resp, err := c.loadbalancer.DeleteLoadBalancer(ctx, loadbalancer.DeleteLoadBalancerRequest{
		LoadBalancerId: &id,
	})
	incRequestCounter(err, deleteVerb, loadBalancerResource)

	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) UpdateBackendSet(ctx context.Context, request loadbalancer.UpdateBackendSetRequest) (string, error) {
	if request.RequestMetadata.RetryPolicy == nil {
		request.RequestMetadata.RetryPolicy = c.requestMetadata.RetryPolicy
	}

	resp, err := c.loadbalancer.UpdateBackendSet(ctx, request)
	incRequestCounter(err, updateVerb, loadBalancerBackendSetResource)

	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}

	return *resp.OpcWorkRequestId, nil
}
