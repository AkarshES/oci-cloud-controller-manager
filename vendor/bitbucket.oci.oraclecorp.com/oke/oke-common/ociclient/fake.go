package ociclient

import (
	"context"
	"io"

	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/keymanagement"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

type FakeClient struct {
	Compartment                             *identity.Compartment
	Compartments                            []identity.Compartment
	Instance                                *core.Instance
	Instances                               []core.Instance
	VNIC                                    *core.Vnic
	Image                                   *core.Image
	Images                                  []core.Image
	Shapes                                  []core.Shape
	Subnet                                  *core.Subnet
	Subnets                                 []core.Subnet
	NSGs                                    []core.NetworkSecurityGroup
	NSG                                     *core.NetworkSecurityGroup
	AvailabilityDomains                     []identity.AvailabilityDomain
	FaultDomains                            []identity.FaultDomain
	Namespace                               string
	Objects                                 []objectstorage.ObjectSummary
	Tenancy                                 *identity.Tenancy
	VCN                                     *core.Vcn
	LoadBalancer                            *loadbalancer.LoadBalancer
	LoadBalancers                           []loadbalancer.LoadBalancer
	LoadBalancerBackend                     *loadbalancer.Backend
	LoadBalancerBackendSet                  *loadbalancer.BackendSet
	LoadBalancerListener                    *loadbalancer.Listener
	LoadBalancerWorkRequest                 *loadbalancer.WorkRequest
	Key                                     *keymanagement.Key
	KeyVersion                              *keymanagement.KeyVersion
	GetImageShapeCompatibilityEntryResponse *core.GetImageShapeCompatibilityEntryResponse
	Err                                     error
}

var _ Interface = &FakeClient{}

func (c *FakeClient) GetCompartment(ctx context.Context, id string) (*identity.Compartment, error) {
	return c.Compartment, c.Err
}

func (c *FakeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	return c.Instance, c.Err
}

func (c *FakeClient) LaunchInstance(ctx context.Context, req core.LaunchInstanceRequest) (*core.Instance, error) {
	return c.Instance, c.Err
}

func (c *FakeClient) ListImages(ctx context.Context, compartment string, shape *string) ([]core.Image, error) {
	return c.Images, c.Err
}

func (c *FakeClient) ListInstances(ctx context.Context, req core.ListInstancesRequest) ([]core.Instance, error) {
	return c.Instances, c.Err
}

func (c *FakeClient) TerminateInstance(ctx context.Context, id string) error {
	return c.Err
}

func (c *FakeClient) GetPrimaryVNIC(ctx context.Context, compartmentID string, instanceID string) (*core.Vnic, error) {
	return c.VNIC, c.Err
}

func (c *FakeClient) GetImage(ctx context.Context, imageID string) (*core.Image, error) {
	return c.Image, c.Err
}

func (c *FakeClient) ListShapes(ctx context.Context, compartment string, imageId *string, availabilityDomain *string) ([]core.Shape, error) {
	return c.Shapes, c.Err
}

func (c *FakeClient) GetTenancy(ctx context.Context, id string) (*identity.Tenancy, error) {
	return c.Tenancy, c.Err
}

func (c *FakeClient) ListCompartments(ctx context.Context, tenancy string) ([]identity.Compartment, error) {
	return c.Compartments, c.Err
}

func (c *FakeClient) ListADs(ctx context.Context, compartment string) ([]identity.AvailabilityDomain, error) {
	return c.AvailabilityDomains, c.Err
}

func (c *FakeClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	return c.Subnet, c.Err
}

func (c *FakeClient) ListSubnets(ctx context.Context, vcnID, compartmentID string) ([]core.Subnet, error) {
	return c.Subnets, c.Err
}

func (c *FakeClient) ListNSGs(ctx context.Context, vcnID, compartmentID string) ([]core.NetworkSecurityGroup, error) {
	return c.NSGs, c.Err
}

func (c *FakeClient) GetNSG(ctx context.Context, nsgId string) (*core.NetworkSecurityGroup, error) {
	return c.NSG, c.Err
}

func (c *FakeClient) GetVCN(ctx context.Context, id string) (*core.Vcn, error) {
	return c.VCN, c.Err
}

func (c *FakeClient) CreateLoadBalancer(ctx context.Context, details loadbalancer.CreateLoadBalancerDetails) (string, error) {
	return "", c.Err
}

func (c *FakeClient) GetLoadBalancer(ctx context.Context, id string) (*loadbalancer.LoadBalancer, error) {
	return c.LoadBalancer, c.Err
}

func (c *FakeClient) ListLoadBalancers(ctx context.Context, compartmentID string) ([]loadbalancer.LoadBalancer, error) {
	return c.LoadBalancers, c.Err
}

func (c *FakeClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	return "", c.Err
}

func (c *FakeClient) ListFaultDomains(ctx context.Context, tenancyID string, ad string) ([]identity.FaultDomain, error) {
	return c.FaultDomains, c.Err
}

func (c *FakeClient) AwaitWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error) {
	return c.LoadBalancerWorkRequest, c.Err
}

func (c *FakeClient) CreateBackend(ctx context.Context, request loadbalancer.CreateBackendRequest) (string, error) {
	return "", c.Err
}

func (c *FakeClient) CreateBackendSet(ctx context.Context, request loadbalancer.CreateBackendSetRequest) (string, error) {
	return "", c.Err
}

func (c *FakeClient) CreateListener(ctx context.Context, request loadbalancer.CreateListenerRequest) (string, error) {
	return "", c.Err
}

func (c *FakeClient) UpdateBackendSet(ctx context.Context, request loadbalancer.UpdateBackendSetRequest) (string, error) {
	return "", c.Err
}

func (c *FakeClient) GetNamespace(ctx context.Context) (string, error) {
	return c.Namespace, c.Err
}

func (c *FakeClient) ListObjects(ctx context.Context, namespace, bucketName string) ([]objectstorage.ObjectSummary, error) {
	return c.Objects, c.Err
}

func (c *FakeClient) DeleteObject(ctx context.Context, namespace, bucketName, objectName string) (string, error) {
	return "", c.Err
}

func (c *FakeClient) DeleteBucket(ctx context.Context, namespace, bucketName string) (string, error) {
	return "", c.Err
}

func (c *FakeClient) GetObject(ctx context.Context, namespace string, bucketName string, objectName string) (objectstorage.GetObjectResponse, error) {
	return objectstorage.GetObjectResponse{}, c.Err
}

func (c *FakeClient) PutObject(ctx context.Context, namespace string, bucketName string,
	objectName string, contentLength int64, object io.ReadCloser) (string, error) {
	return "", c.Err
}

func (c *FakeClient) GetKey(ctx context.Context, keyID string) (*keymanagement.Key, error) {
	return c.Key, c.Err
}

func (c *FakeClient) GetKeyVersion(ctx context.Context, keyID string, keyVersionID string) (*keymanagement.KeyVersion, error) {
	return c.KeyVersion, c.Err
}

func (c *FakeClient) PostMetricData(ctx context.Context, request monitoring.PostMetricDataRequest) (response monitoring.PostMetricDataResponse, err error) {
	return monitoring.PostMetricDataResponse{}, nil
}

func (c *FakeClient) GetImageShapeCompatibilityEntry(ctx context.Context, imageId string, shapeName string) (*core.GetImageShapeCompatibilityEntryResponse, error) {
	return c.GetImageShapeCompatibilityEntryResponse, c.Err
}
