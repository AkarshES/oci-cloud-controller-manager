package ociclient

import (
	"strconv"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ociRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oci_requests_total",
			Help: "OCI API requests total.",
		},
		[]string{"resource", "code", "verb"},
	)
)

type resource string

const (
	instanceResource                = "instance"
	imageResource                   = "image"
	shapeResource                   = "shape"
	vnicAttachmentResource          = "vnic_attachment"
	vnicResource                    = "vnic"
	compartmentResource             = "compartment"
	availabilityDomainResource      = "availability_domain"
	faultDomainResource             = "fault_domain"
	vcnResource                     = "vcn"
	subnetResource                  = "subnet"
	nsgResource                     = "nsg"
	tenancyResource                 = "tenancy"
	loadBalancerResource            = "load_balancer"
	loadBalancerWorkRequestResource = "load_balancer_work_request"
	loadBalancerBackendResource     = "load_balancer_backend"
	loadBalancerBackendSetResource  = "load_balancer_backend_set"
	loadBalancerlistenerResource    = "load_balancer_listener"
	objectStorageResource           = "object_storage"
)

type verb string

const (
	getVerb    verb = "get"
	listVerb   verb = "list"
	createVerb verb = "create"
	updateVerb verb = "update"
	deleteVerb verb = "delete"
)

func incRequestCounter(err error, v verb, r resource) {
	statusCode := 200
	if err != nil {
		if serviceErr, ok := err.(common.ServiceError); ok {
			statusCode = serviceErr.GetHTTPStatusCode()
		} else {
			statusCode = 500
		}
	}

	ociRequestCounter.With(prometheus.Labels{
		"resource": string(r),
		"verb":     string(v),
		"code":     strconv.Itoa(statusCode),
	}).Inc()
}

func init() {
	prometheus.MustRegister(ociRequestCounter)
}
