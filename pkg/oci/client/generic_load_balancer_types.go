package client

import (
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
)

const (
	GenericIPv4        GenericIpVersion = "IPv4"
	GenericIPv6        GenericIpVersion = "IPv6"
	GenericIPv4AndIPv6 GenericIpVersion = "IPv4_AND_IPv6"
)

type GenericBackendSetDetails struct {
	Name                            *string
	HealthChecker                   *GenericHealthChecker
	Policy                          *string
	Backends                        []GenericBackend
	SessionPersistenceConfiguration *GenericSessionPersistenceConfiguration
	// Only needed for LB
	BackendMaxConnections *int
	SslConfiguration      *GenericSslConfigurationDetails
	// Only needed for NLB
	IsPreserveSource *bool
	IpVersion        *GenericIpVersion
}

type GenericIpVersion string

type GenericSessionPersistenceConfiguration struct {
	CookieName      *string
	DisableFallback *bool
}

type GenericHealthChecker struct {
	Protocol          string  `json:"protocol,required"`
	IsForcePlainText  *bool   `json:"isForcePlainText,omitempty"`
	Port              *int    `json:"port,required"`
	UrlPath           *string `json:"urlPath,omitempty"`
	Retries           *int    `json:"retries,omitempty"`
	TimeoutInMillis   *int    `json:"timeoutInMillis,omitempty"`
	IntervalInMillis  *int    `json:"intervalInMillis,omitempty"`
	ResponseBodyRegex *string `json:"responseBodyRegex,omitempty"`

	// Only needed for NLB & when using Pods as Backends mode
	ReturnCode *int `json:"returnCode,omitempty"`
	// Base64 encoded pattern to be sent as UDP or TCP health check probe.
	RequestData []byte `mandatory:"false" json:"requestData,omitempty"`
	// Base64 encoded pattern to be validated as UDP or TCP health check probe response.
	ResponseData []byte                                       `mandatory:"false" json:"responseData,omitempty"`
	Dns          *networkloadbalancer.DnsHealthCheckerDetails `mandatory:"false" json:"dns"`
}

type GenericBackend struct {
	Port           *int
	Name           *string
	IpAddress      *string
	TargetId       *string
	Weight         *int
	Backup         *bool
	Drain          *bool
	Offline        *bool
	MaxConnections *int
}

type GenericSslConfigurationDetails struct {
	VerifyDepth                    *int     `json:"verifyDepth"`
	VerifyPeerCertificate          *bool    `json:"verifyPeerCertificate"`
	HasSessionResumption           *bool    `json:"hasSessionResumption"`
	TrustedCertificateAuthorityIds []string `json:"trustedCertificateAuthorityIds"`
	CertificateIds                 []string `json:"certificateIds"`
	CertificateName                *string  `json:"certificateName"`
	Protocols                      []string `json:"protocols"`
	CipherSuiteName                *string  `json:"cipherSuiteName"`
	ServerOrderPreference          string   `json:"serverOrderPreference"`
}

type GenericListener struct {
	Name                    *string
	DefaultBackendSetName   *string
	Port                    *int
	Protocol                *string
	HostnameNames           []string
	PathRouteSetName        *string
	SslConfiguration        *GenericSslConfigurationDetails
	ConnectionConfiguration *GenericConnectionConfiguration
	RoutingPolicyName       *string
	RuleSetNames            []string
	IpVersion               *GenericIpVersion
	IsPpv2Enabled           *bool
}

type GenericConnectionConfiguration struct {
	IdleTimeout                    *int64
	BackendTcpProxyProtocolVersion *int
	BackendTcpProxyProtocolOptions []string
}

type GenericCreateLoadBalancerDetails struct {
	CompartmentId               *string
	DisplayName                 *string
	ShapeName                   *string
	SubnetIds                   []string
	ShapeDetails                *GenericShapeDetails
	IsPrivate                   *bool
	IsPreserveSourceDestination *bool
	ReservedIps                 []GenericReservedIp
	Listeners                   map[string]GenericListener
	BackendSets                 map[string]GenericBackendSetDetails
	NetworkSecurityGroupIds     []string
	FreeformTags                map[string]string
	DefinedTags                 map[string]map[string]interface{}
	IpVersion                   *GenericIpVersion

	// Only needed for LB
	Certificates map[string]GenericCertificate

	// Internal. Only supported by NLB
	CpgId *string

	RuleSets map[string]loadbalancer.RuleSetDetails
	// Supported only in NLB
	AssignedPrivateIpv4 *string
	AssignedIpv6        *string
	*IpVersionTranslationConfig
}

type GenericShapeDetails struct {
	MinimumBandwidthInMbps *int
	MaximumBandwidthInMbps *int
}

type GenericCertificate struct {
	CertificateName   *string
	Passphrase        *string
	PrivateKey        *string
	PublicCertificate *string
	CaCertificate     *string
}

type GenericReservedIp struct {
	Id *string
}

type GenericIpAddress struct {
	IpAddress  *string
	IsPublic   *bool
	ReservedIp *GenericReservedIp
}

type GenericUpdateLoadBalancerShapeDetails struct {
	ShapeName    *string
	ShapeDetails *GenericShapeDetails
}

type GenericLoadBalancer struct {
	Id                      *string
	CompartmentId           *string
	DisplayName             *string
	LifecycleState          *string
	ShapeName               *string
	IpAddresses             []GenericIpAddress
	ShapeDetails            *GenericShapeDetails
	IsPrivate               *bool
	SubnetIds               []string
	NetworkSecurityGroupIds []string
	Listeners               map[string]GenericListener
	Certificates            map[string]GenericCertificate
	BackendSets             map[string]GenericBackendSetDetails
	RuleSets                map[string]loadbalancer.RuleSetDetails
	IpVersion               *GenericIpVersion

	FreeformTags map[string]string
	DefinedTags  map[string]map[string]interface{}
	SystemTags   map[string]map[string]interface{}

	// Supported only in NLB
	*IpVersionTranslationConfig
}

type IpVersionTranslationConfig struct {
	IpVersionTranslationMode networkloadbalancer.NetworkLoadBalancerIpVersionTranslationEnum
	Nat46Ipv6CidrPrefix      *string
}

type GenericWorkRequest struct {
	Id             *string
	LoadBalancerId *string
	Type           *string
	LifecycleState *string
	Message        *string
	CompartmentId  *string
	OperationType  string
	Status         string
}

type GenericUpdateNetworkSecurityGroupsDetails struct {
	NetworkSecurityGroupIds []string
}

type GenericBackendSetHealth struct {
	Status                    string
	WarningStateBackendNames  []string
	CriticalStateBackendNames []string
	UnknownStateBackendNames  []string
	BackendCount              *int
}

type GenericUpdateLoadBalancerDetails struct {
	IpVersion    *GenericIpVersion
	FreeformTags map[string]string
	DefinedTags  map[string]map[string]interface{}
	*IpVersionTranslationConfig
}
