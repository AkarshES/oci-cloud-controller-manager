package ociclient

import (
	"context"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
)

func (c *client) PostMetricData(ctx context.Context, request monitoring.PostMetricDataRequest) (response monitoring.PostMetricDataResponse, err error) {
	return c.monitoring.PostMetricData(ctx, request)
}
