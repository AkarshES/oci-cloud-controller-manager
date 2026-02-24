package ociclient

import (
	"context"
	"io"
	"net/http"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/keymanagement"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
	"github.com/pkg/errors"
)

var (
	listLimit   = 500
	ErrNotFound = errors.New("not found")
)

const (
	// OBOHeaderName is the http request header name for passing On-Behalf-of(OBO) Token.
	OBOHeaderName = "opc-obo-token"
)

type Interface interface {
	// compute
	GetImage(ctx context.Context, image string) (*core.Image, error)
	ListImages(ctx context.Context, compartment string, shape *string) ([]core.Image, error)
	GetInstance(ctx context.Context, id string) (*core.Instance, error)
	GetPrimaryVNIC(ctx context.Context, compartmentID string, instanceID string) (*core.Vnic, error)
	LaunchInstance(context.Context, core.LaunchInstanceRequest) (*core.Instance, error)
	ListInstances(ctx context.Context, req core.ListInstancesRequest) ([]core.Instance, error)
	ListShapes(ctx context.Context, compartment string, imageID *string, availabilityDomain *string) ([]core.Shape, error)
	TerminateInstance(ctx context.Context, id string) error
	GetImageShapeCompatibilityEntry(ctx context.Context, imageId string, shapeName string) (*core.GetImageShapeCompatibilityEntryResponse, error)

	// identity
	GetCompartment(ctx context.Context, id string) (*identity.Compartment, error)
	GetTenancy(ctx context.Context, id string) (*identity.Tenancy, error)
	ListADs(ctx context.Context, compartment string) ([]identity.AvailabilityDomain, error)
	ListCompartments(ctx context.Context, tenancy string) ([]identity.Compartment, error)
	ListFaultDomains(ctx context.Context, tenancyID string, ad string) ([]identity.FaultDomain, error)

	// virtnet
	ListSubnets(ctx context.Context, vcnID, compartmentID string) ([]core.Subnet, error)
	GetSubnet(ctx context.Context, id string) (*core.Subnet, error)
	GetVCN(ctx context.Context, id string) (*core.Vcn, error)
	ListNSGs(ctx context.Context, vcnID, compartmentID string) ([]core.NetworkSecurityGroup, error)
	GetNSG(ctx context.Context, nsgId string) (*core.NetworkSecurityGroup, error)

	//lb
	AwaitWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error)
	CreateBackend(ctx context.Context, request loadbalancer.CreateBackendRequest) (string, error)
	CreateBackendSet(ctx context.Context, request loadbalancer.CreateBackendSetRequest) (string, error)
	CreateListener(ctx context.Context, request loadbalancer.CreateListenerRequest) (string, error)
	CreateLoadBalancer(ctx context.Context, details loadbalancer.CreateLoadBalancerDetails) (string, error)
	DeleteLoadBalancer(ctx context.Context, id string) (string, error)
	GetLoadBalancer(ctx context.Context, id string) (*loadbalancer.LoadBalancer, error)
	ListLoadBalancers(ctx context.Context, compartmentID string) ([]loadbalancer.LoadBalancer, error)
	UpdateBackendSet(ctx context.Context, request loadbalancer.UpdateBackendSetRequest) (string, error)

	// object storage
	DeleteObject(ctx context.Context, namespace, bucketName, objectName string) (string, error)
	DeleteBucket(ctx context.Context, namespace, bucketName string) (string, error)
	GetNamespace(ctx context.Context) (string, error)
	ListObjects(ctx context.Context, namespace, bucketName string) ([]objectstorage.ObjectSummary, error)
	GetObject(ctx context.Context, namespace string, bucketName string, objectName string) (objectstorage.GetObjectResponse, error)
	PutObject(ctx context.Context, namespace string, bucketName string, objectName string, contentLength int64, object io.ReadCloser) (string, error)

	// KMS management
	GetKey(ctx context.Context, keyID string) (*keymanagement.Key, error)
	GetKeyVersion(ctx context.Context, keyID string, keyVersionID string) (*keymanagement.KeyVersion, error)

	// Monitoring
	PostMetricData(ctx context.Context, request monitoring.PostMetricDataRequest) (response monitoring.PostMetricDataResponse, err error)
}

type Config struct {
	RequestDispatcher  common.HTTPRequestDispatcher
	RequestSigner      common.HTTPRequestSigner
	RequestInterceptor common.RequestInterceptor
	RequestHost        string
	RetryPolicy        *common.RetryPolicy
}

type client struct {
	identity        *identity.IdentityClient
	compute         *core.ComputeClient
	virtnet         *core.VirtualNetworkClient
	loadbalancer    *loadbalancer.LoadBalancerClient
	objectstorage   *objectstorage.ObjectStorageClient
	monitoring      *monitoring.MonitoringClient
	requestMetadata common.RequestMetadata
	provider        common.ConfigurationProvider
	config          *Config
}

func newClient(provider common.ConfigurationProvider, config *Config) (*client, error) {

	common.EnableInstanceMetadataServiceLookup()

	identityClient, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	computeClient, err := core.NewComputeClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	virtnetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	lbClient, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	objectstorageClient, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	monitoringClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	if config.RequestSigner != nil || config.RequestInterceptor != nil || config.RequestDispatcher != nil {
		configureSDKBaseClient(&identityClient.BaseClient, config.RequestSigner, config.RequestInterceptor, config.RequestDispatcher, "")
		configureSDKBaseClient(&computeClient.BaseClient, config.RequestSigner, config.RequestInterceptor, config.RequestDispatcher, "")
		configureSDKBaseClient(&virtnetClient.BaseClient, config.RequestSigner, config.RequestInterceptor, config.RequestDispatcher, "")
		configureSDKBaseClient(&lbClient.BaseClient, config.RequestSigner, config.RequestInterceptor, config.RequestDispatcher, "")
		configureSDKBaseClient(&objectstorageClient.BaseClient, config.RequestSigner, config.RequestInterceptor, config.RequestDispatcher, "")
		configureSDKBaseClient(&monitoringClient.BaseClient, config.RequestSigner, config.RequestInterceptor, config.RequestDispatcher, config.RequestHost)
	}

	var retryPolicy *common.RetryPolicy
	if config.RetryPolicy != nil {
		retryPolicy = config.RetryPolicy
	} else {
		retryPolicy = newRetryPolicy()
	}

	c := &client{
		identity:      &identityClient,
		virtnet:       &virtnetClient,
		compute:       &computeClient,
		loadbalancer:  &lbClient,
		objectstorage: &objectstorageClient,
		monitoring:    &monitoringClient,
		requestMetadata: common.RequestMetadata{
			RetryPolicy: retryPolicy,
		},
		provider: provider,
		config:   config,
	}

	return c, nil
}

func configureSDKBaseClient(client *common.BaseClient, signer common.HTTPRequestSigner, interceptor common.RequestInterceptor, dispatcher common.HTTPRequestDispatcher, host string) {
	if signer != nil {
		client.Signer = signer
	}

	if interceptor != nil {
		client.Interceptor = interceptor
	}

	if dispatcher != nil {
		client.HTTPClient = dispatcher
	}

	if host != "" {
		client.Host = host
	}
}

func New(provider common.ConfigurationProvider) (Interface, error) {
	return newClient(provider, &Config{})
}

func NewServicePrincipalClientFromInstancePrincipal() (Interface, error) {
	common.EnableInstanceMetadataServiceLookup()
	cp, err := auth.NewServicePrincipalWithInstancePrincipalConfigurationProvider("")
	if err != nil {
		return nil, err
	}
	return newClient(cp, &Config{})
}

// NewWithConfig returns an ociclient.Interface configured based on the Config param.
func NewWithConfig(cp common.ConfigurationProvider, config *Config) (Interface, error) {
	return newClient(cp, config)
}

// NewOBOInterceptor returns a oci-go-sdk request interceptor that adds provided OBO token in the request header.
func NewOBOInterceptor(oboToken string) common.RequestInterceptor {
	return func(r *http.Request) error {
		r.Header.Set(OBOHeaderName, oboToken)
		return nil
	}
}
