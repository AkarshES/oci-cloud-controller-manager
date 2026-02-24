package ociclient

import (
	"context"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	lifecycleWaitTimeout = 6 * time.Minute
)

func WaitForLoadBalancerLifecycle(ctx context.Context, client Interface, id string, lifecycle loadbalancer.LoadBalancerLifecycleStateEnum) error {
	ctx, cancel := context.WithTimeout(ctx, lifecycleWaitTimeout)
	defer cancel()
	return wait.PollImmediateUntil(30*time.Second, func() (bool, error) {
		lb, err := client.GetLoadBalancer(ctx, id)
		if IsNotFound(err) {
			return true, nil
		}

		if IsRetryable(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return lb.LifecycleState == lifecycle, nil
	}, ctx.Done())
}

func WaitForInstanceLifecycle(ctx context.Context, client Interface, id string, lifecycle core.InstanceLifecycleStateEnum) error {
	ctx, cancel := context.WithTimeout(ctx, lifecycleWaitTimeout)
	defer cancel()
	return wait.PollImmediateUntil(30*time.Second, func() (bool, error) {
		instance, err := client.GetInstance(ctx, id)
		if IsNotFound(err) {
			return true, nil
		}

		if IsRetryable(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return instance.LifecycleState == lifecycle, nil
	}, ctx.Done())
}

func WaitForInstanceTerminated(ctx context.Context, client Interface, id string) error {
	return WaitForInstanceLifecycle(ctx, client, id, core.InstanceLifecycleStateTerminated)
}
