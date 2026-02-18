/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/netip"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	npnv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/v1beta1"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	utils "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	errors2 "github.com/pkg/errors"
)

const (
	CREATE_PRIVATE_IP   = "CREATE_PRIVATE_IP"
	CREATE_IPV6         = "CREATE_IPV6"
	ATTACH_VNIC         = "ATTACH_VNIC"
	INITIALIZE_NPN_NODE = "INITIALIZE_NPN_NODE"
	// GVA_DEFAULT_IP_COUNT is set to 32 at OKE-CP
	GVA_DEFAULT_IP_COUNT = 32

	// The max value of ipCount is an arbitrary value of 256. This corresponds to IP CIDR address of block size 256
	GVA_MAX_IP_COUNT = 256
	// maxSecondaryIpsPerVNIC
	// max allocatable IPs per vnic is 32 where one IP will be used as host address
	// NPN requires one additional IP (from secondary vnic) as host address, it is used to talk with primary VNIC address for the host namespace interface.
	maxSecondaryIpsPerVNIC = 32
	// maxPodIpsPerVNIC maximum number of pod-ips created per vnic
	maxPodIpsPerVNIC = 31
	IPv4             = "IPv4"
	IPv6             = "IPv6"
	// GetNodeTimeout is the timeout for the node object to be created in Kubernetes
	GetNodeTimeout                             = 20 * time.Minute
	ensureVnicAttachedAndAvailablePollDuration = 2 * time.Minute
	// RunningInstanceTimeout is the timeout for the instance to reach running state
	// before we try to attach VNIC(s) to them
	RunningInstanceTimeout                          = 5 * time.Minute
	FetchingInstance                                = "Fetching OCI compute instance"
	FetchingExistingSecondaryVNICsForInstance       = "Fetching existingSecondaryVNICs for instance"
	FetchedExistingSecondaryVNICsForInstance        = "Fetched existingSecondaryVNICs for instance"
	FetchingPrivateIPsForSecondaryVNICs             = "Fetching private IPs for existing secondary VNICs"
	FetchedPrivateIPsForSecondaryVNICs              = "Fetched existingSecondaryIp for VNICs of the instance"
	AllocateAdditionalVNICsToInstance               = "Need to attach additional secondary VNICs to the instance"
	AllocatedAdditionalVNICsToInstance              = "Successfully allocated the required additional VNICs for instance"
	SecondFetchingExistingSecondaryVNICsForInstance = "Fetching existingSecondaryVNICs for instance once again"
	SecondFetchedExistingSecondaryVNICsForInstance  = "Fetched existingSecondaryVNICs for instance once again"
	AllocatingAdditionalPrivateIPsForSecondaryVNICs = "Started allocating additional private IPs for secondary VNICs"
	ComputingAdditionalIpsByVnic                    = "Computing required additionalIpsByVnic"
	ComputedAdditionalIpsByVnic                     = "Computed required additionalIpsByVnic"
	FetchingSecondaryVNICsAndIPsForInstance         = "Fetching secondary VNICs & attached private IPs for instance once again"

	TaintKeyApplicationResourceOnly = "oci.oraclecloud.com/application-resource-only"
)

var (
	STATE_SUCCESS     = "SUCCESS"
	STATE_IN_PROGRESS = "IN_PROGRESS"
	STATE_BACKOFF     = "BACKOFF"
	COMPLETED         = "COMPLETED"

	SKIP_SOURCE_DEST_CHECK    = true
	errPrimaryVnicNotFound    = errors.New("failed to get primary vnic for instance")
	errInstanceNotRunning     = errors.New("instance is not in running state")
	errVnicNotAttached        = errors.New("vnic(s) not in attached state yet")
	errNotEnoughVnicsAttached = errors.New("number of VNICs attached is not equal to required number of VNICs")
	errVnicNotAvailable       = errors.New("vnic is not available")
)

// NativePodNetworkReconciler reconciles a NativePodNetwork object
type NativePodNetworkReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	MetricPusher     *metrics.MetricPusher
	OCIClient        ociclient.Interface
	TimeTakenTracker sync.Map
	Recorder         record.EventRecorder
	Config           *providercfg.Config
}

// VnicAttachmentResponse is used to store the response for attach VNIC
type VnicAttachmentResponse struct {
	VnicAttachment core.VnicAttachment
	err            error
	timeTaken      float64
}

type IpAddressCountByVersion struct {
	V4 int
	V6 int
}

type IpAllocations struct {
	V4 []IPAllocation
	V6 []IPAllocation
}

type VnicIPAllocations struct {
	vnicId     string
	ipFamilies []string
	ips        IpAddressCountByVersion
}

type VnicIPAllocationResponse struct {
	vnicId        string
	errIPv4       error
	errIPv6       error
	ipFamilies    []string
	ipAllocations IpAllocations
}
type VnicAttachmentResponseSlice []VnicAttachmentResponse

type IPAllocation struct {
	err       error
	timeTaken float64
}
type IPAllocationSlice []IPAllocation

type endToEndLatency struct {
	timeTaken float64
}
type endToEndLatencySlice []endToEndLatency

// SubnetVnic is a struct used to pass around information about a VNIC
// and the subnet it belongs to
type SubnetVnic struct {
	Vnic       *core.Vnic
	Subnet     *core.Subnet
	Attachment *core.VnicAttachment
}

type vnicSecondaryAddresses struct {
	V4       []core.PrivateIp
	V6       []core.Ipv6
	hostIpv4 *string
	hostIpv6 *string

	ipFamilies []string
}

type GvaNics struct {
	vnicId            *string
	SecondaryVnicSpec *npnv1beta1.SecondaryVnic
}

type ErrorMetric interface {
	GetMetricName(IpVersion string) string
	GetTimeTaken() float64
	GetError() error
}
type ConvertToErrorMetric interface {
	ErrorMetric() []ErrorMetric
}

func (r *NativePodNetworkReconciler) PushMetric(errorArray []ErrorMetric, ipVersion string) {
	averageByReturnCode := computeAveragesByReturnCode(errorArray)
	if len(errorArray) == 0 {
		return
	}
	metricName := errorArray[0].GetMetricName(ipVersion)
	for k, v := range averageByReturnCode {
		dimensions := map[string]string{"component": k}
		metrics.SendMetricData(r.MetricPusher, metricName, v, dimensions)
	}
}

func (v IPAllocation) GetTimeTaken() float64 {
	return v.timeTaken
}
func (v IPAllocation) GetMetricName(ipVersion string) string {
	switch ipVersion {
	case IPv6:
		return CREATE_IPV6
	default:
		return CREATE_PRIVATE_IP
	}
}
func (v IPAllocation) GetError() error {
	return v.err
}

func (v VnicAttachmentResponse) GetTimeTaken() float64 {
	return v.timeTaken
}
func (v VnicAttachmentResponse) GetMetricName(ipVersion string) string {
	return ATTACH_VNIC
}
func (v VnicAttachmentResponse) GetError() error {
	return v.err
}

func (v endToEndLatency) GetTimeTaken() float64 {
	return v.timeTaken
}
func (v endToEndLatency) GetMetricName(ipVersion string) string {
	return INITIALIZE_NPN_NODE
}
func (v endToEndLatency) GetError() error {
	return nil
}

func (v VnicAttachmentResponseSlice) ErrorMetric() []ErrorMetric {
	ret := make([]ErrorMetric, len(v))
	for i, ele := range v {
		ret[i] = ele
	}
	return ret
}

func (v IPAllocationSlice) ErrorMetric() []ErrorMetric {
	ret := make([]ErrorMetric, len(v))
	for i, ele := range v {
		ret[i] = ele
	}
	return ret
}

func (v endToEndLatencySlice) ErrorMetric() []ErrorMetric {
	ret := make([]ErrorMetric, len(v))
	for i, ele := range v {
		ret[i] = ele
	}
	return ret
}

// TODO: write a unit test
func computeAveragesByReturnCode(errorArray []ErrorMetric) map[string]float64 {
	totalByReturnCode := make(map[string][]float64)
	for _, val := range errorArray {
		if val.GetError() == nil {
			if _, ok := totalByReturnCode[util.Success]; !ok {
				totalByReturnCode[util.Success] = make([]float64, 0)
			}
			totalByReturnCode[util.Success] = append(totalByReturnCode[util.Success], val.GetTimeTaken())
			continue
		}

		returnCode := util.GetError(val.GetError())
		if _, ok := totalByReturnCode[returnCode]; !ok {
			totalByReturnCode[returnCode] = make([]float64, 0)
		}
		totalByReturnCode[returnCode] = append(totalByReturnCode[returnCode], val.GetTimeTaken())
	}

	averageByReturnCode := make(map[string]float64)
	for key, arr := range totalByReturnCode {
		total := 0.0

		for _, val := range arr {
			total += val
		}
		averageByReturnCode[key] = total / float64(len(arr))
	}
	return averageByReturnCode
}

//+kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nativepodnetworkings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nativepodnetworkings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nativepodnetworkings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NativePodNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	var failReason, failMessage = "NPNReconcileFailed", ""
	var panicOccurred = false
	var npn npnv1beta1.NativePodNetwork
	defer func() {
		if rec := recover(); rec != nil {
			panicOccurred = true
			err = fmt.Errorf("panic recovered %v stack is %s", rec, string(debug.Stack()))
			log.FromContext(ctx).
				WithValues("component", "npn-controller").
				Error(err, "Recovered from panic in NPN Reconcile")
			dimensionsMap := make(map[string]string)
			dimensionsMap[metrics.ClusterOCID] = r.Config.ClusterID
			metrics.SendMetricData(r.MetricPusher, "NPN_PANIC", 1, dimensionsMap)
			return
		}
		if panicOccurred {
			r.Recorder.Event(&npn, v1.EventTypeWarning, "NPNReconcileFailed", "Fatal error occurred")
		} else if failMessage != "" && err != nil {
			r.Recorder.Event(&npn, v1.EventTypeWarning, failReason, failMessage+": "+err.Error())
		} else if failMessage != "" {
			r.Recorder.Event(&npn, v1.EventTypeWarning, failReason, failMessage)
		} else if err != nil {
			r.Recorder.Event(&npn, v1.EventTypeWarning, failReason, err.Error())
		}
	}()

	log := log.FromContext(ctx)
	startTime, _ := r.TimeTakenTracker.LoadOrStore(req.Name, time.Now())
	mutex := sync.Mutex{}
	if err := r.Get(ctx, req.NamespacedName, &npn); err != nil {
		log.Error(err, "unable to fetch NativePodNetwork")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if npn.Status.State != nil && *npn.Status.State == STATE_SUCCESS {
		log.Info("NativePodNetwork CR has reached state SUCCESS, nothing to do")
		return ctrl.Result{}, nil
	}
	log.Info("Processing NativePodNetwork CR")
	npn.Status.State = &STATE_IN_PROGRESS
	npn.Status.Reason = &STATE_IN_PROGRESS
	r.Recorder.Event(&npn, v1.EventTypeNormal, "StartNPNReconcile", "Starting NativePodNetwork reconciliation")
	err = r.Status().Update(context.Background(), &npn)
	if err != nil {
		failReason, failMessage = "UpdateNPNStatusFailed", "failed to update status of NPN CR to InProgress"
		log.Error(err, failMessage)
		return ctrl.Result{}, err
	}

	log.WithValues("instanceId", *npn.Spec.Id).Info(FetchingInstance)
	instance, err := r.OCIClient.Compute().GetInstance(ctx, *npn.Spec.Id)
	if err != nil || instance.Id == nil {
		failReason, failMessage = "GetInstanceFailed", "failed to get OCI compute instance"
		log.WithValues("instanceId", *npn.Spec.Id).Error(err, failMessage)
		r.handleError(ctx, req, err, "GetInstance")
		return ctrl.Result{}, err
	}
	log = log.WithValues("instanceId", *instance.Id)

	// remove the CR in case the node never joined the cluster and the instance is terminated
	if instance.LifecycleState == core.InstanceLifecycleStateTerminated ||
		instance.LifecycleState == core.InstanceLifecycleStateTerminating {
		err = r.Client.Delete(ctx, &npn)
		if err != nil {
			failReason, failMessage = "InstanceTerminated", "failed to delete NPN CR for terminated compute instance"
			log.Error(err, failMessage)
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		log.Info("Deleted the CR for terminated compute instance")
		return ctrl.Result{}, nil
	}

	if instance.LifecycleState != core.InstanceLifecycleStateRunning {
		err = r.waitForInstanceToReachRunningState(ctx, npn)
		if err != nil {
			failReason, failMessage = "InstanceNotRunning", errInstanceNotRunning.Error()
			r.handleError(ctx, req, errInstanceNotRunning, "GetRunningInstance")
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
	}

	log.Info(FetchingExistingSecondaryVNICsForInstance)
	primaryVnic, existingSecondaryVNICs, err := r.getPrimaryAndSecondaryVNICs(ctx, *instance.CompartmentId, *instance.Id)
	if err != nil {
		failReason = "GetVNICsFailed"
		r.handleError(ctx, req, err, "GetVNIC")
		return ctrl.Result{}, err
	}
	if primaryVnic == nil {
		failReason, failMessage = "PrimaryVNICNotFound", errPrimaryVnicNotFound.Error()
		r.handleError(ctx, req, errPrimaryVnicNotFound, "GetPrimaryVNIC")
		return ctrl.Result{}, errPrimaryVnicNotFound
	}
	log.WithValues("existingSecondaryVNICs", existingSecondaryVNICs).
		WithValues("countOfExistingSecondaryVNICs", len(existingSecondaryVNICs)).
		Info(FetchedExistingSecondaryVNICsForInstance)

	var requiredSecondaryVNICs int
	var gvaNics []GvaNics
	provisionRequiredAdditionalSecondaryVNICs := 0
	if isGvaNode(&npn) {
		log.WithValues("Reconciling user-specified SecondaryVnics")
		if len(npn.Spec.SecondaryVnics) > len(existingSecondaryVNICs) {
			gvaNics, err = r.attachUserSpecifiedSecondaryVnics(ctx, npn)
			if err != nil {
				failReason, failMessage = "FailedToCreateVNIC", errVnicNotAttached.Error()
				r.handleError(ctx, req, err, "AttachVnic")
				return ctrl.Result{RequeueAfter: time.Second * 10}, err
			}
		} else if len(npn.Spec.SecondaryVnics) == len(existingSecondaryVNICs) {
			log.Info("getting the user-specified SecondaryVnics")
			gvaNics, err = r.getGvaNic(ctx, &npn, existingSecondaryVNICs)
		}
		requiredSecondaryVNICs = len(npn.Spec.SecondaryVnics)
		log.WithValues("gvaNics", gvaNics).Info("attachUserSpecifiedSecondaryVnics returned GvaNics")
	} else {
		requiredSecondaryVNICs = int(math.Ceil(float64(*npn.Spec.MaxPodCount) / maxPodIpsPerVNIC))
		provisionRequiredAdditionalSecondaryVNICs = requiredSecondaryVNICs - len(existingSecondaryVNICs)
	}

	if provisionRequiredAdditionalSecondaryVNICs > 0 {
		log.WithValues("provisionRequiredAdditionalSecondaryVNICs", provisionRequiredAdditionalSecondaryVNICs).Info(AllocateAdditionalVNICsToInstance)
		additionalVNICAttachments := make([]VnicAttachmentResponse, provisionRequiredAdditionalSecondaryVNICs)
		for index := 0; index < provisionRequiredAdditionalSecondaryVNICs; index++ {
			startTime := time.Now()
			opts := ociclient.AttachVnicOptions{
				InstanceID:          npn.Spec.Id,
				SubnetID:            npn.Spec.PodSubnetIds[0],
				NsgIds:              stringPointerToStringSlice(npn.Spec.NetworkSecurityGroupIds),
				SkipSourceDestCheck: &SKIP_SOURCE_DEST_CHECK,
			}
			vnicAttachment, err := r.OCIClient.Compute().AttachVnic(ctx, opts)
			additionalVNICAttachments[index].VnicAttachment, additionalVNICAttachments[index].err = vnicAttachment, err
			if additionalVNICAttachments[index].err != nil {
				failReason, failMessage = "AttachAdditionalVNICsFailed", "failed to attach VNIC to instance: "+additionalVNICAttachments[index].err.Error()
				log.Error(additionalVNICAttachments[index].err, "failed to attach VNIC to instance")
				r.handleError(ctx, req, err, "AttachVNIC")
				r.PushMetric(VnicAttachmentResponseSlice(additionalVNICAttachments).ErrorMetric(), "")
				return ctrl.Result{}, err
			}
			additionalVNICAttachments[index].timeTaken = float64(time.Since(startTime).Seconds())
			log.WithValues("vnic", additionalVNICAttachments[index].VnicAttachment).Info("VNIC attached to instance")

			if _, ensured, err := r.ensureVnicAttachedAndAvailable(ctx, &vnicAttachment); !ensured {
				failReason, failMessage = "AttachAdditionalVNICsFailed", "failed to ensure required additional VNICs"
				log.WithValues("provisionRequiredAdditionalSecondaryVNICs", provisionRequiredAdditionalSecondaryVNICs).
					Error(err, failMessage)
				r.handleError(ctx, req, err, "AttachVNIC")
				r.PushMetric(VnicAttachmentResponseSlice(additionalVNICAttachments).ErrorMetric(), "")
				if errors.Is(err, wait.ErrWaitTimeout) {
					return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
				}
				return ctrl.Result{}, err
			}
		}
		r.PushMetric(VnicAttachmentResponseSlice(additionalVNICAttachments).ErrorMetric(), "")
		log.WithValues("provisionRequiredAdditionalSecondaryVNICs", provisionRequiredAdditionalSecondaryVNICs).Info(AllocatedAdditionalVNICsToInstance)
	}

	log.Info(SecondFetchingExistingSecondaryVNICsForInstance)
	_, existingSecondaryVNICs, err = r.getPrimaryAndSecondaryVNICs(ctx, *instance.CompartmentId, *instance.Id)
	if err != nil {
		failReason = "GetVNICsFailed"
		r.handleError(ctx, req, err, "GetVNIC")
		return ctrl.Result{}, err
	}
	log.WithValues("existingSecondaryVNICs", existingSecondaryVNICs).
		WithValues("countOfExistingSecondaryVNICs", len(existingSecondaryVNICs)).
		Info(SecondFetchedExistingSecondaryVNICsForInstance)

	if provisionRequiredAdditionalSecondaryVNICs > 0 {
		vnicAttached, err := r.validateVnicAttachmentsAreInAttachedState(ctx, *instance.Id, requiredSecondaryVNICs, existingSecondaryVNICs)
		if vnicAttached == false || err != nil {
			failReason, failMessage = "AttachAdditionalVNICsFailed", "failed to validate required VNICs"
			log.Error(err, failMessage)
			r.handleError(ctx, req, err, "AttachVNIC")
			return ctrl.Result{}, err
		}
	}
	nodeIpFamilies, err := getIpFamilies(ctx, npn)
	if err != nil {
		log.Error(err, "failed to get IpFamilies from NPN CR")
		r.handleError(ctx, req, err, "GetNPN_IPFamilies")
		return ctrl.Result{}, err
	}

	nodeName := getNodeNameFromPrimaryVnic(primaryVnic, nodeIpFamilies)

	log.Info(FetchingPrivateIPsForSecondaryVNICs)
	existingSecondaryIpsbyVNIC, err := r.getSecondaryIpsByVNICs(ctx, existingSecondaryVNICs, nodeIpFamilies, &npn)
	if err != nil {
		failReason = "ListPrivateIPsFailed"
		r.handleError(ctx, req, err, "ListPrivateIP")
		return ctrl.Result{}, err
	}
	totalAllocatedSecondaryIPs := totalAllocatedSecondaryIpsForInstance(existingSecondaryIpsbyVNIC)
	log.WithValues("countOfExistingSecondaryIps", totalAllocatedSecondaryIPs).Info(FetchedPrivateIPsForSecondaryVNICs)

	log.Info(ComputingAdditionalIpsByVnic)
	additionalIpsByVnic, err := getAdditionalSecondaryIPsNeededPerVNIC(existingSecondaryIpsbyVNIC, &npn, totalAllocatedSecondaryIPs, nodeIpFamilies, gvaNics, log)
	if err != nil {
		failReason, failMessage = "AllocatePrivateIPsFailed", "failed to allocate the required IP addresses"
		log.WithValues("maxPodCount", *npn.Spec.MaxPodCount).Error(err, failMessage)
		log.WithValues("totalAllocatedSecondaryIPs", totalAllocatedSecondaryIPs).Error(err, failMessage)
		r.handleError(ctx, req, err, "AllocatePrivateIP")
		return ctrl.Result{}, err
	}
	log.WithValues("additionalIpsByVnic", additionalIpsByVnic).Info(ComputedAdditionalIpsByVnic)

	log.Info(AllocatingAdditionalPrivateIPsForSecondaryVNICs)
	vnicAdditionalIpAllocations := make([]VnicIPAllocationResponse, requiredSecondaryVNICs)
	workqueue.ParallelizeUntil(ctx, requiredSecondaryVNICs, requiredSecondaryVNICs, func(outerIndex int) {
		parallelLog := log.WithValues("vnicId", additionalIpsByVnic[outerIndex].vnicId).
			WithValues("requiredIPs", additionalIpsByVnic[outerIndex].ips).
			WithValues("vnicIpFamilies", additionalIpsByVnic[outerIndex].ipFamilies)
		var errIPv4 error = nil
		var errIPv6 error = nil
		allocations := IpAllocations{}
		vnicIpFamilies := additionalIpsByVnic[outerIndex].ipFamilies
		if len(vnicIpFamilies) == 0 || contains(vnicIpFamilies, IPv4) {
			if additionalIpsByVnic[outerIndex].ips.V4 > 0 {
				parallelLog.Info("Need to allocate secondary IPv4 for VNIC")
				ipv4Allocations := make([]IPAllocation, additionalIpsByVnic[outerIndex].ips.V4)
				if isGvaNode(&npn) {
					startTime := time.Now()
					ipsToAttach := additionalIpsByVnic[outerIndex].ips.V4

					//assign v4 ip cidr address (ex: 10.0.1.0/30) to accommodate ipCount addresses
					if ipsToAttach > 0 {
						cidrPrefixLength, err := getCidrPrefixLengthForBlockSize(ipsToAttach, IPv4)
						if err != nil {
							parallelLog.Error(err, "failed to compute CIDR prefix length")
						} else {
							_, err = r.OCIClient.Networking(nil).CreatePrivateIp(ctx, additionalIpsByVnic[outerIndex].vnicId, cidrPrefixLength)
							if err != nil {
								parallelLog.Error(err, "failed to create IPv4")
							}
						}
						ipv4Allocations[0].err = err
						ipv4Allocations[0].timeTaken = float64(time.Since(startTime).Seconds())

					}
				} else {
					// assign individual IP v4 addresses for non-gva nodes.
					for innerIndex := 0; innerIndex < additionalIpsByVnic[outerIndex].ips.V4; innerIndex++ {
						startTime := time.Now()
						_, err := r.OCIClient.Networking(nil).CreatePrivateIp(ctx, additionalIpsByVnic[outerIndex].vnicId, nil)
						if err != nil {
							parallelLog.Error(err, "failed to create IPv4")
						}
						ipv4Allocations[innerIndex].err = err
						ipv4Allocations[innerIndex].timeTaken = float64(time.Since(startTime).Seconds())
					}
				}
				errIPv4 = validateVnicIpAllocation(ipv4Allocations)
				allocations.V4 = ipv4Allocations
			}
		}
		if contains(vnicIpFamilies, IPv6) {
			if additionalIpsByVnic[outerIndex].ips.V6 > 0 {
				parallelLog.Info("Need to allocate secondary IPv6 for VNIC")
				ipv6Allocations := make([]IPAllocation, additionalIpsByVnic[outerIndex].ips.V6)
				if isGvaNode(&npn) {
					//assign v6 ip cidr address (ex: 2603:4e3::/124) to accommodate ipCount addresses
					startTime := time.Now()
					ipsToAttach := additionalIpsByVnic[outerIndex].ips.V6
					if ipsToAttach > 0 {
						cidrPrefixLength, err := getCidrPrefixLengthForBlockSize(ipsToAttach, IPv6)
						if err != nil {
							parallelLog.Error(err, "failed to compute CIDR prefix length")
						} else {
							_, err = r.OCIClient.Networking(nil).CreateIpv6(ctx, additionalIpsByVnic[outerIndex].vnicId, cidrPrefixLength)
							if err != nil {
								parallelLog.Error(err, "failed to create IPv6")
							}
						}
						ipv6Allocations[0].err = err
						ipv6Allocations[0].timeTaken = float64(time.Since(startTime).Seconds())
					}

				} else {
					// assign individual IP v6 addresses for non-gva nodes.
					for innerIndex := 0; innerIndex < additionalIpsByVnic[outerIndex].ips.V6; innerIndex++ {
						startTime := time.Now()
						_, err := r.OCIClient.Networking(nil).CreateIpv6(ctx, additionalIpsByVnic[outerIndex].vnicId, nil)
						if err != nil {
							parallelLog.Error(err, "failed to create IPv6")
						}
						ipv6Allocations[innerIndex].err = err
						ipv6Allocations[innerIndex].timeTaken = float64(time.Since(startTime).Seconds())
					}
				}
				errIPv6 = validateVnicIpAllocation(ipv6Allocations)
				allocations.V6 = ipv6Allocations
			}
		}
		mutex.Lock()
		vnicAdditionalIpAllocations[outerIndex] = VnicIPAllocationResponse{additionalIpsByVnic[outerIndex].vnicId, errIPv4, errIPv6, vnicIpFamilies, allocations}
		mutex.Unlock()
	})
	for _, ips := range vnicAdditionalIpAllocations {
		if len(ips.ipFamilies) == 0 || contains(ips.ipFamilies, IPv4) {
			if ips.errIPv4 != nil {
				failReason, failMessage = "CreatePrivateIPFailed", ips.errIPv4.Error()
				r.handleError(ctx, req, ips.errIPv4, "CreatePrivateIP")
				r.PushMetric(IPAllocationSlice(ips.ipAllocations.V4).ErrorMetric(), IPv4)
				return ctrl.Result{}, ips.errIPv4
			}
			r.PushMetric(IPAllocationSlice(ips.ipAllocations.V4).ErrorMetric(), IPv4)
		}
		if contains(ips.ipFamilies, IPv6) {
			if ips.errIPv6 != nil {
				failReason, failMessage = "CreateIPv6Failed", ips.errIPv6.Error()
				r.handleError(ctx, req, ips.errIPv6, "CreateIPv6")
				r.PushMetric(IPAllocationSlice(ips.ipAllocations.V6).ErrorMetric(), IPv6)
				return ctrl.Result{}, ips.errIPv6
			}
			r.PushMetric(IPAllocationSlice(ips.ipAllocations.V6).ErrorMetric(), IPv6)
		}
	}

	log.Info(FetchingSecondaryVNICsAndIPsForInstance)
	_, existingSecondaryVNICs, err = r.getPrimaryAndSecondaryVNICs(ctx, *instance.CompartmentId, *instance.Id)
	if err != nil {
		failReason = "GetVNICsFailed"
		r.handleError(ctx, req, err, "GetVNIC")
		return ctrl.Result{}, err
	}
	log.WithValues("existingSecondaryVNICs", existingSecondaryVNICs).
		WithValues("countOfExistingSecondaryVNICs", len(existingSecondaryVNICs)).
		Info(FetchedExistingSecondaryVNICsForInstance)
	existingSecondaryIpsbyVNIC, err = r.getSecondaryIpsByVNICs(ctx, existingSecondaryVNICs, nodeIpFamilies, &npn)
	if err != nil {
		failReason = "ListPrivateIPsFailed"
		r.handleError(ctx, req, err, "ListPrivateIP")
		return ctrl.Result{}, err
	}

	totalAllocatedSecondaryIPs = totalAllocatedSecondaryIpsForInstance(existingSecondaryIpsbyVNIC)
	log.WithValues("secondaryIpsbyVNIC", existingSecondaryIpsbyVNIC).
		WithValues("countOfExistingSecondaryIps", totalAllocatedSecondaryIPs).
		Info("Fetched existingSecondaryIp for instance")

	// assign host IP address here per vnic
	existingSecondaryIpsbyVNIC = assignHostIpAddressForVnic(existingSecondaryIpsbyVNIC, nodeIpFamilies, &npn)

	// validate if maxPodCount = number of secondary IPs available on the vnics
	if !isGvaNode(&npn) {
		err = validateMaxPodCountWithSecondaryIPCount(existingSecondaryIpsbyVNIC, *npn.Spec.MaxPodCount, nodeIpFamilies)
		if err != nil {
			failReason = "IPsNotEqualToMaxPodCount"
			log.Error(err, "secondary IPs are not equal to MaxPodCount")
			r.handleError(ctx, req, err, "validateMaxPodCountWithSecondaryIPCount")
			return ctrl.Result{}, err
		}
	}

	// For GVA, we need to validate against vnic[].NicConfiguration.ipCount
	if isGvaNode(&npn) {
		err = validateVnicIPCountsAgainstConfiguredIpCount(existingSecondaryIpsbyVNIC, npn.Status.VNICs, nodeIpFamilies)
		if err != nil {
			failReason = "VnicIpsNotEqualToConfiguredCount"
			log.Error(err, "VNIC IPs are not equal to configured ipCount")
			r.handleError(ctx, req, err, "validateVnicIPCountsAgainstConfiguredIpCount")
			return ctrl.Result{}, err
		}
	}

	log.Info("Fetching NPN CR for owner ref & status update")
	updateNPN := npnv1beta1.NativePodNetwork{}
	err = r.Get(context.TODO(), req.NamespacedName, &updateNPN)
	if err != nil {
		failReason = "GetNPNFailed"
		log.Error(err, "failed to get NPN CR")
		r.handleError(ctx, req, err, "GetNPN_CR")
		return ctrl.Result{}, err
	}
	log.Info("Fetched NPN CR")

	log.Info("Getting v1 Node object to set ownerref on NPN CR")
	// Set OwnerRef on the CR and mark CR status as SUCCESS
	nodeObject, err := r.getNodeObjectInCluster(ctx, req.NamespacedName, nodeName)
	if err != nil {
		failReason = "GetV1NodeFailed"
		r.handleError(ctx, req, err, "GetV1Node")
		return ctrl.Result{}, err
	}

	if isGvaNode(&npn) && checkApplicationResourcesUsedOnAllVnics(gvaNics) {
		log.Info("Add taint to node based on vnic application resource advertisements")
		nodeObject, err = r.setApplicationResourceTaint(ctx, nodeObject)
		if err != nil {
			failReason = "SetTaintOnNodeFailed"
			r.handleError(ctx, req, err, "SetTaintOnNode")
			return ctrl.Result{}, err
		}
	}

	if err = controllerutil.SetOwnerReference(nodeObject, &updateNPN, r.Scheme); err != nil {
		failReason, failMessage = "UpdateOwnerRefrenceFailed", "failed to update owner ref on NPN CR"
		log.Error(err, failMessage)
		return ctrl.Result{}, err
	}
	log.Info("Updating ownerref and NPN CR status as COMPLETED")
	err = r.Client.Update(ctx, &updateNPN)
	if err != nil {
		failReason, failMessage = "SetOwnerRefrenceFailed", "failed to set ownerref on NPN CR"
		log.Error(err, failMessage)
		return ctrl.Result{}, err
	}

	updateNPN.Status.State = &STATE_SUCCESS
	updateNPN.Status.Reason = &COMPLETED
	updateNPN.Status.VNICs = convertCoreVNICtoNPNStatus(existingSecondaryVNICs, existingSecondaryIpsbyVNIC, nodeIpFamilies, gvaNics)
	r.Recorder.Event(&npn, v1.EventTypeNormal, "NPN_CR_Success", "NPN CR reconciled successfully")
	err = r.Status().Update(ctx, &updateNPN)
	if err != nil {
		failReason, failMessage = "FinalUpdateNPNStatusFailed", "failed to set status on NPN CR"
		log.Error(err, failMessage)
		return ctrl.Result{}, err
	}
	log.Info("NativePodNetwork CR reconciled successfully")

	r.PushMetric(endToEndLatencySlice{{time.Since(startTime.(time.Time)).Seconds()}}.ErrorMetric(), "")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NativePodNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := zap.L().Sugar()
	log.Info("Setting up NPN controller with manager")
	r.Recorder = mgr.GetEventRecorderFor("nativepodnetwork")
	return ctrl.NewControllerManagedBy(mgr).
		For(&npnv1beta1.NativePodNetwork{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 20, CacheSyncTimeout: time.Hour}).
		Complete(r)
}

// return the primary and secondary vnics for the given compute instance
func (r *NativePodNetworkReconciler) getPrimaryAndSecondaryVNICs(ctx context.Context, CompartmentId, InstanceId string) (primaryVnic *core.Vnic, existingSecondaryVNICAttachments []SubnetVnic, err error) {
	log := log.FromContext(ctx, "instanceId", InstanceId)
	vnicAttachments, err := r.OCIClient.Compute().ListVnicAttachments(ctx, CompartmentId, InstanceId)
	if err != nil {
		log.Error(err, "failed to get VNIC Attachments for OCI Instance")
		return nil, nil, err
	}
	existingSecondaryVNICAttachments = make([]SubnetVnic, 0)
	for _, vnicAttachment := range vnicAttachments {
		// ignore VNIC attachments in detached/detaching state
		if vnicAttachment.Id == nil ||
			vnicAttachment.VnicId == nil ||
			vnicAttachment.LifecycleState == core.VnicAttachmentLifecycleStateDetached ||
			vnicAttachment.LifecycleState == core.VnicAttachmentLifecycleStateDetaching {
			continue
		}
		vNIC, err := r.OCIClient.Networking(nil).GetVNIC(ctx, *vnicAttachment.VnicId)
		if err != nil {
			log.Error(err, "failed to get VNIC from VNIC attachment")
			return nil, nil, err
		}
		log = log.WithValues("vnicId", vNIC.Id)
		if *vNIC.IsPrimary {
			primaryVnic = vNIC
			continue
		}
		// ignore terminating/terminated VNICs
		if vNIC.LifecycleState == core.VnicLifecycleStateTerminating || vNIC.LifecycleState == core.VnicLifecycleStateTerminated {
			log.Info("Ignoring VNIC in terminating/terminated state")
			continue
		}
		subnet, err := r.OCIClient.Networking(nil).GetSubnet(ctx, *vNIC.SubnetId)
		if err != nil {
			log.Error(err, "failed to get subnet for VNIC")
			return nil, nil, err
		}
		existingSecondaryVNICAttachments = append(existingSecondaryVNICAttachments, SubnetVnic{vNIC, subnet, &vnicAttachment})
	}
	return
}

// get the list of secondary private ips allocated on the given VNIC
func (r *NativePodNetworkReconciler) getSecondaryIpsByVNICs(ctx context.Context, existingSecondaryVNICs []SubnetVnic, nodeIpFamilies []string, npn *npnv1beta1.NativePodNetwork) (map[string]*vnicSecondaryAddresses, error) {
	ipsByVNICs := make(map[string]*vnicSecondaryAddresses)
	log := log.FromContext(ctx)
	for _, secondary := range existingSecondaryVNICs {
		vnicSecondaryAddresses := &vnicSecondaryAddresses{}
		var ipFamiliesPerVnic []string
		log := log.WithValues("vnicId", *secondary.Vnic.Id)
		var err error
		if len(nodeIpFamilies) == 0 || contains(nodeIpFamilies, IPv4) {
			vnicSecondaryAddresses.V4, err = r.OCIClient.Networking(nil).ListPrivateIps(ctx, *secondary.Vnic.Id)
			if err != nil {
				log.Error(err, "failed to list secondary IPv4 IPs for VNIC")
				return nil, err
			}
			if len(vnicSecondaryAddresses.V4) > 0 {
				ipFamiliesPerVnic = append(ipFamiliesPerVnic, IPv4)
			}
		}
		if contains(nodeIpFamilies, IPv6) {
			vnicSecondaryAddresses.V6, err = r.OCIClient.Networking(nil).ListIpv6s(ctx, *secondary.Vnic.Id)
			if err != nil {
				log.Error(err, "failed to list secondary IPv6 IPs for VNIC")
				return nil, err
			}
			if len(vnicSecondaryAddresses.V6) > 0 {
				ipFamiliesPerVnic = append(ipFamiliesPerVnic, IPv6)
			}
		}
		vnicSecondaryAddresses.ipFamilies = ipFamiliesPerVnic

		// Dont filter PrimaryIP for GVA nodes.
		// PrimaryIP address is the GVA vnic's v4 HostAddress on the NPN
		if !isGvaNode(npn) {
			vnicSecondaryAddresses = filterPrimaryIp(vnicSecondaryAddresses)

		}
		ipsByVNICs[*secondary.Vnic.Id] = vnicSecondaryAddresses
	}
	return ipsByVNICs, nil
}

// assignHostIpAddressForVnic is a util method to get HostIP Address per vnic
// a non-IP CIDR Address is assigned as Host address
func assignHostIpAddressForVnic(existingSecondaryVNICs map[string]*vnicSecondaryAddresses, ipFamilies []string, npn *npnv1beta1.NativePodNetwork) map[string]*vnicSecondaryAddresses {
	IPsbyVNICs := make(map[string]*vnicSecondaryAddresses)

	for k, vnic := range existingSecondaryVNICs {

		if len(ipFamilies) == 0 || contains(ipFamilies, IPv4) {
			if len(vnic.V4) >= 1 {
				if isGvaNode(npn) {

					primaryIPIdx := -1
					ipAddressIdx := -1
					for i := 0; i < len(vnic.V4); i++ {
						// for IPv4, the PrimaryIP (non-IP CIDR Address) address will be assigned as HostAddress
						// a non-IP CIDR Address will prefix = 32.

						// skip all the IP CIDR Addresses (example: 10.0.3.0/30)
						if vnic.V4[i].CidrPrefixLength != nil && *vnic.V4[i].CidrPrefixLength != 32 {
							continue
						}

						// IP address found
						if ipAddressIdx == -1 {
							ipAddressIdx = i
						}

						// store primaryIP index
						if vnic.V4[i].IsPrimary != nil && *vnic.V4[i].IsPrimary {
							primaryIPIdx = i
							break
						}
					}
					hostIpIdx := primaryIPIdx

					// no primary IP found
					if hostIpIdx == -1 {
						hostIpIdx = ipAddressIdx
					}
					if hostIpIdx != -1 {
						vnic.hostIpv4 = vnic.V4[hostIpIdx].IpAddress
						vnic.V4 = append(vnic.V4[:hostIpIdx], vnic.V4[hostIpIdx+1:]...)
					}

				} else {
					vnic.hostIpv4, vnic.V4 = vnic.V4[0].IpAddress, vnic.V4[1:]
				}

			}
		}
		if contains(ipFamilies, IPv6) {
			if len(vnic.V6) >= 1 {
				if isGvaNode(npn) {
					for i := 0; i < len(vnic.V6); i++ {
						// non-ipCidrAddress for ipv4 has prefix == nil
						if vnic.V6[i].CidrPrefixLength == nil {
							vnic.hostIpv6 = vnic.V6[i].IpAddress
							vnic.V6 = append(vnic.V6[:i], vnic.V6[i+1:]...)
							break
						}
					}

				} else {
					vnic.hostIpv6, vnic.V6 = vnic.V6[0].IpAddress, vnic.V6[1:]
				}
			}
		}
		IPsbyVNICs[k] = vnic
	}
	return IPsbyVNICs
}

func validateMaxPodCountWithSecondaryIPCount(existingSecondaryVNICs map[string]*vnicSecondaryAddresses, maxPodCount int, ipFamilies []string) error {
	V4IPs, V6IPs := 0, 0
	for _, vnic := range existingSecondaryVNICs {
		V4IPs = V4IPs + len(vnic.V4)
		V6IPs = V6IPs + len(vnic.V6)
	}
	if (len(ipFamilies) == 0 || contains(ipFamilies, IPv4)) && V4IPs != maxPodCount {
		return errors2.Errorf("Allocated IPv4 count != maxPodCount (%d != %d)", maxPodCount, V4IPs)
	}
	if contains(ipFamilies, IPv6) && V6IPs != maxPodCount {
		return errors2.Errorf("Allocated IPv6 count != maxPodCount (%d != %d)", maxPodCount, V6IPs)
	}
	return nil
}

// util method to handle logging when thre is an error and updating the NPN status appropriately
func (r *NativePodNetworkReconciler) handleError(ctx context.Context, req ctrl.Request, err error, operation string) {
	log := log.FromContext(ctx).WithValues("name", req.Name)

	log.Error(err, "received error for operation "+operation, "parsedError", util.GetError(err))
	updateNPN := npnv1beta1.NativePodNetwork{}
	err = r.Get(context.TODO(), req.NamespacedName, &updateNPN)
	if err != nil {
		log.Error(err, "failed to get CR")
		return
	}
	reason := "FailedTo" + operation
	updateNPN.Status.State = &STATE_BACKOFF
	updateNPN.Status.Reason = &reason
	err = r.Status().Update(context.Background(), &updateNPN)
	if err != nil {
		log.Error(err, "failed to set status on CR")
	}
}

// contains is a utility method to check if a string is part of a slice
func contains(slice []string, searchString string) bool {
	for _, element := range slice {
		if element == searchString {
			return true
		}
	}
	return false
}

// exclude the primary IPs in the list of private IPs on VNIC
func filterPrimaryIp(ips *vnicSecondaryAddresses) *vnicSecondaryAddresses {
	Ips := &vnicSecondaryAddresses{
		V4:         []core.PrivateIp{},
		V6:         []core.Ipv6{},
		ipFamilies: ips.ipFamilies,
	}
	for _, ip := range ips.V4 {
		// ignore primary IP
		if *ip.IsPrimary {
			continue
		}
		Ips.V4 = append(Ips.V4, ip)
	}
	for _, ip := range ips.V6 {
		Ips.V6 = append(Ips.V6, ip)
	}

	return Ips
}

// compute the total number of allocated secondary ips on secondary vnics for this compute instance
func totalAllocatedSecondaryIpsForInstance(vnicToIpMap map[string]*vnicSecondaryAddresses) IpAddressCountByVersion {
	totalSecondaryIPv4, totalSecondaryIPv6 := 0, 0

	for _, Ips := range vnicToIpMap {
		totalSecondaryIPv4 += len(Ips.V4)
		totalSecondaryIPv6 += len(Ips.V6)
	}
	totalSecondaryIps := IpAddressCountByVersion{
		V4: totalSecondaryIPv4,
		V6: totalSecondaryIPv6,
	}
	return totalSecondaryIps
}

// check if there were any errors during attaching vnics
func validateAdditionalVnicAttachments(vnics []VnicAttachmentResponse) error {
	for _, vnic := range vnics {
		if vnic.err != nil {
			return vnic.err
		}
	}
	return nil
}

// compute the number of (additional) IPs needed to be allocated per VNIC
func getAdditionalSecondaryIPsNeededPerVNIC(existingIpsByVnic map[string]*vnicSecondaryAddresses, npn *npnv1beta1.NativePodNetwork, totalAllocated IpAddressCountByVersion, nodeIpFamilies []string, gvaNics []GvaNics, log logr.Logger) ([]VnicIPAllocations, error) {
	if *npn.Spec.MaxPodCount == 0 && len(npn.Status.VNICs) == 0 {
		return nil, nil
	}

	var podCount int
	if *npn.Spec.MaxPodCount > 0 {
		podCount = *npn.Spec.MaxPodCount
	}

	vnicCount := len(existingIpsByVnic)
	requireIPv4 := len(nodeIpFamilies) == 0 || contains(nodeIpFamilies, IPv4)
	requireIPv6 := contains(nodeIpFamilies, IPv6)

	requiredIPv4ByVnic := map[string]int{}
	requiredIPv6ByVnic := map[string]int{}
	var totalRequiredIPv4, totalRequiredIPv6 int

	perVnicIpFamilies := make(map[string][]string)
	if isGvaNode(npn) {
		for _, vnic := range gvaNics {
			var existingIPsOnVnic IpAddressCountByVersion

			if vnic.vnicId == nil {
				continue
			}
			vnicID := *vnic.vnicId
			ipCount := vnic.SecondaryVnicSpec.CreateVnicDetails.IpCount
			vnicIpFamilies := vnic.SecondaryVnicSpec.CreateVnicDetails.IpFamilies
			perVnicIpFamilies[vnicID] = vnicIpFamilies
			// vnicIpFamilies is internally derived based on GVA vnic subnet and
			// AssignIpv6Ip property in spec.secondaryVnics.createVnicDetails
			// can never be empty for GVA
			if contains(vnicIpFamilies, IPv4) {
				for _, v4Ips := range existingIpsByVnic[vnicID].V4 {
					if v4Ips.CidrPrefixLength != nil && *v4Ips.CidrPrefixLength < 32 {
						log.Info(fmt.Sprintf("%s is an ip cidr address with prefix %d. expanding the cidr range", *v4Ips.IpAddress, *v4Ips.CidrPrefixLength))
						existingIPsOnVnic.V4 += getBlockSizeForCidrPrefix(*v4Ips.CidrPrefixLength, IPv4)
					} else {
						existingIPsOnVnic.V4++
					}
				}

				// For a GVA node, the IPv4 host IP is the primary IP address.
				// When a seondary VNIC is created on an IPv4-compatible subnet with an IPv4 stack,
				// the IPv4 primary address is assigned.
				// Required IPs per VNIC = ipCount - (existing IPs on the VNIC - 1 primary IP address)
				requiredIPv4ByVnic[vnicID] = max(0, ipCount+1-existingIPsOnVnic.V4)
			}
			if contains(vnicIpFamilies, IPv6) {
				for _, v6Ips := range existingIpsByVnic[vnicID].V6 {
					if v6Ips.CidrPrefixLength != nil {
						log.Info(fmt.Sprintf("%s is an ip cidr address with prefix %d. expanding the cidr range", *v6Ips.IpAddress, *v6Ips.CidrPrefixLength))

						// For IPv6, the CIDR prefix must be a multiple of 4 (e.g., 128, 124, 120, ...).
						// Therefore, the block size can be greater than or equal to ipCount.
						// In such cases, use ipCount as the block size.
						blockSize := getBlockSizeForCidrPrefix(*v6Ips.CidrPrefixLength, IPv6)
						if blockSize > ipCount {
							blockSize = ipCount
						}
						existingIPsOnVnic.V6 += blockSize
					} else {
						existingIPsOnVnic.V6++
					}

				}

				// For a GVA node, the IPv6 stack is determined by the 'assignIpv6Ip' parameter.
				// This parameter attaches the VNIC with an IPv6 address, which will be used as the IPv6 host IP address.
				// Required IPs per VNIC = ipCount - (existing IPs on the VNIC - 1 IPv6 address assigned during VNIC attachment)
				requiredIPv6ByVnic[vnicID] = max(0, ipCount+1-existingIPsOnVnic.V6)
			}
		}
	} else {
		if requireIPv4 {
			totalRequiredIPv4 = podCount - totalAllocated.V4 + vnicCount
		}
		if requireIPv6 {
			totalRequiredIPv6 = podCount - totalAllocated.V6 + vnicCount
		}
	}

	additional := []VnicIPAllocations{}
	for vnicID, existing := range existingIpsByVnic {
		ipAlloc := IpAddressCountByVersion{}

		var ipFamilies []string
		var needV4, needV6 int

		if isGvaNode(npn) {
			ipFamilies = perVnicIpFamilies[vnicID]
			needV4 = requiredIPv4ByVnic[vnicID]
			needV6 = requiredIPv6ByVnic[vnicID]
		} else {
			ipFamilies = nodeIpFamilies
			needV4 = totalRequiredIPv4
			needV6 = totalRequiredIPv6
		}

		var available int
		isGva := isGvaNode(npn)
		if contains(ipFamilies, IPv4) {
			if isGva {
				// Calculate available IPv4 addresses for allocation:
				// (Maximum allocatable IP CIDR Address) + 1 (primary IP) - (number of existing IPv4 addresses)
				available = GVA_MAX_IP_COUNT + 1 - len(existing.V4)
				ipAlloc.V4 = min(needV4, available)
				// GVA requires a power-of-2 CIDR-compatible IP count
				if !isPowerOf2(ipAlloc.V4) {
					return nil, errors.New(fmt.Sprintf("the required number of IPs to allocate on VNIC is not v4 IP CIDR address compatible. %d is not a power of 2", ipAlloc.V4))
				}
			} else {
				available = maxSecondaryIpsPerVNIC - len(existing.V4)
				ipAlloc.V4 = min(needV4, available)
				totalRequiredIPv4 -= ipAlloc.V4
			}
		}

		if contains(ipFamilies, IPv6) {
			if isGva {
				// Calculate available IPv6 addresses for allocation.
				// (Maximum allocatable IP CIDR Address) + 1 v6 IP Address for v6 HostIP - (number of existing IPv6 addresses)
				available = GVA_MAX_IP_COUNT + 1 - len(existing.V6)
				ipAlloc.V6 = min(needV6, available)
				if !isPowerOf2(ipAlloc.V6) {
					return nil, errors.New(fmt.Sprintf("the required number of IPs to allocate on VNIC is not v6 IP CIDR address compatible. %d is not a power of 2", ipAlloc.V6))
				}
			} else {
				available = maxSecondaryIpsPerVNIC - len(existing.V6)
				ipAlloc.V6 = min(needV6, available)
				totalRequiredIPv6 -= ipAlloc.V6
			}
		}

		additional = append(additional, VnicIPAllocations{
			vnicId:     vnicID,
			ips:        ipAlloc,
			ipFamilies: ipFamilies,
		})
	}

	if !isGvaNode(npn) && (totalRequiredIPv4 > 0 || totalRequiredIPv6 > 0) {
		return nil, errors.New("failed to allocate required number of IPs with existing VNICs")
	}
	return additional, nil
}

// check if there were any errors during secondary ip allocation
func validateVnicIpAllocation(ipAllocations []IPAllocation) error {
	for _, ip := range ipAllocations {
		if ip.err != nil {
			return ip.err
		}
	}
	return nil
}

// util method to translate OCI objects to NPN status fields
func convertCoreVNICtoNPNStatus(existingSecondaryVNICs []SubnetVnic, existingSecondaryIpsByVNIC map[string]*vnicSecondaryAddresses, nodeIpFamilies []string, nics []GvaNics) []npnv1beta1.VNICAddress {
	var ipCount int
	npnVNICAddresses := make([]npnv1beta1.VNICAddress, 0, len(existingSecondaryIpsByVNIC))
	for _, vnic := range existingSecondaryVNICs {
		nicConfiguration := getNicConfigurationForNodes(vnic.Vnic, vnic.Subnet, nics)

		ipFamilies := nodeIpFamilies
		if len(nicConfiguration.IpFamilies) > 0 {
			ipFamilies = nicConfiguration.IpFamilies
		}
		requireIPv6s, requireIPv4s := contains(ipFamilies, IPv6), contains(ipFamilies, IPv4)
		npnVNICAddress := npnv1beta1.VNICAddress{
			VNICID:     vnic.Vnic.Id,
			MACAddress: vnic.Vnic.MacAddress,
		}
		if len(nics) > 0 && *nicConfiguration.IpCount > 0 {
			npnVNICAddress.NicConfiguration = nicConfiguration
			ipCount = *nicConfiguration.IpCount
		}

		vnicSecondaryAddresses := existingSecondaryIpsByVNIC[*vnic.Vnic.Id]
		var hostIPv4, hostIPv6, subnetCidrV4, subnetCidrV6, routerIPv4, routerIPv6 *string
		if requireIPv4s {
			subnetCidrV4 = vnic.Subnet.CidrBlock
			routerIPv4 = vnic.Subnet.VirtualRouterIp
		}
		if requireIPv6s {
			if vnic.Subnet.Ipv6CidrBlock != nil {
				// this value will be nil in case of Ipv6 of type ULA
				subnetCidrV6 = vnic.Subnet.Ipv6CidrBlock
			} else if len(vnic.Subnet.Ipv6CidrBlocks) > 0 {
				// default to first IPv6 ULA prefix. eventually we want this CIDR block to be passed via the NPN CRD as a parameter as Ipv6AddressIpv6SubnetCidrPairDetails
				subnetCidrV6 = common.String(vnic.Subnet.Ipv6CidrBlocks[0])
			}
			routerIPv6 = vnic.Subnet.Ipv6VirtualRouterIp
		}
		if len(nodeIpFamilies) > 0 {
			hostIPv4 = vnicSecondaryAddresses.hostIpv4
			hostIPv6 = vnicSecondaryAddresses.hostIpv6
			// Populate new fields only in case of IPFamilies being present in CRD
			npnVNICAddress.HostAddresses = []npnv1beta1.HostAddress{
				{
					V4: hostIPv4,
					V6: hostIPv6,
				},
			}
			npnVNICAddress.RouterIPs = []npnv1beta1.RouterIP{
				{
					V4: routerIPv4,
					V6: routerIPv6,
				},
			}
			npnVNICAddress.SubnetCidrs = []npnv1beta1.SubnetCidr{
				{
					V4: subnetCidrV4,
					V6: subnetCidrV6,
				},
			}
		}
		var secondaryIpCount int
		if len(nodeIpFamilies) == 0 || requireIPv4s {
			secondaryIpCount = len(existingSecondaryIpsByVNIC[*vnic.Vnic.Id].V4)
		}
		if requireIPv6s {
			secondaryIpCount = len(existingSecondaryIpsByVNIC[*vnic.Vnic.Id].V6)
		}
		vnicAddresses := make([]*string, 0, secondaryIpCount)
		vnicPodAddresses := make([]npnv1beta1.PodAddress, 0, secondaryIpCount)
		//var ipv4IP, ipv6IP *string
		var ipv4IPs, ipv6IPs []*string
		for i := 0; i < secondaryIpCount; i++ {
			if len(nodeIpFamilies) == 0 || requireIPv4s {
				if vnicSecondaryAddresses.V4[i].CidrPrefixLength != nil && *vnicSecondaryAddresses.V4[i].CidrPrefixLength < 32 {
					ipsInRange, err := expandIpAddressesInRange(*vnicSecondaryAddresses.V4[i].IpAddress,
						*vnicSecondaryAddresses.V4[i].CidrPrefixLength)
					if err != nil {
						return nil
					}
					for _, address := range ipsInRange {
						vnicAddresses = append(vnicAddresses, &address)
					}
				} else {
					vnicAddresses = append(vnicAddresses, vnicSecondaryAddresses.V4[i].IpAddress)
				}
				// Populate the old fields in case of IPv4 or nodeIpFamilies length == 0

			}
			if requireIPv4s {
				if vnicSecondaryAddresses.V4[i].CidrPrefixLength != nil && *vnicSecondaryAddresses.V4[i].CidrPrefixLength < 32 {
					ipsInRange, err := expandIpAddressesInRange(*vnicSecondaryAddresses.V4[i].IpAddress,
						*vnicSecondaryAddresses.V4[i].CidrPrefixLength)
					if err != nil {
						return nil
					}
					for _, address := range ipsInRange {
						fmt.Println(&address)
						ipv4IPs = append(ipv4IPs, &address)
					}
				} else {
					ipv4IPs = append(ipv4IPs, vnicSecondaryAddresses.V4[i].IpAddress)
				}
			}
			if requireIPv6s {
				if vnicSecondaryAddresses.V6[i].CidrPrefixLength != nil {
					ipsInRange, err := expandIpAddressesInRange(*vnicSecondaryAddresses.V6[i].IpAddress,
						*vnicSecondaryAddresses.V6[i].CidrPrefixLength)
					if err != nil {
						return nil
					}

					// account ips of length ipCount
					for j := 0; j < ipCount; j++ {
						ipv6IPs = append(ipv6IPs, &ipsInRange[j])
					}
				} else {
					ipv6IPs = append(ipv6IPs, vnicSecondaryAddresses.V6[i].IpAddress)
				}
			}
		}
		if len(nodeIpFamilies) > 0 {
			// Populate new fields only in case of IPFamilies being present in CRD
			expandedIpv4Count := len(ipv4IPs)
			expandedIpv6Count := len(ipv6IPs)

			if expandedIpv4Count > 0 && expandedIpv6Count > 0 {
				// Dual-stack
				// not update any spilled over ips (should not happen)
				minCount := expandedIpv4Count
				if expandedIpv6Count < minCount {
					minCount = expandedIpv6Count
				}
				for j := 0; j < minCount; j++ {
					addr := npnv1beta1.PodAddress{}
					if j < expandedIpv4Count {
						addr.V4 = ipv4IPs[j]
					}
					if j < expandedIpv6Count {
						addr.V6 = ipv6IPs[j]
					}
					vnicPodAddresses = append(vnicPodAddresses, addr)
				}
			} else if expandedIpv4Count > 0 {
				// Single-stack ipv4
				for _, v4 := range ipv4IPs {
					vnicPodAddresses = append(vnicPodAddresses, npnv1beta1.PodAddress{
						V4: v4,
					})
				}
			} else if expandedIpv6Count > 0 && requireIPv6s {
				// Single-stack ipv6
				for _, v6 := range ipv6IPs {
					vnicPodAddresses = append(vnicPodAddresses, npnv1beta1.PodAddress{
						V6: v6,
					})
				}
			}
		}

		if len(nodeIpFamilies) == 0 || requireIPv4s {
			npnVNICAddress.HostAddress = vnicSecondaryAddresses.hostIpv4
			npnVNICAddress.RouterIP = vnic.Subnet.VirtualRouterIp
			npnVNICAddress.SubnetCidr = vnic.Subnet.CidrBlock
			npnVNICAddress.Addresses = vnicAddresses
		}
		if requireIPv6s || requireIPv4s {
			npnVNICAddress.PodAddresses = vnicPodAddresses
		}
		npnVNICAddresses = append(npnVNICAddresses, npnVNICAddress)
	}
	return npnVNICAddresses
}

// wait for the Kubernetes object to be created in the cluster so that the owner reference of the NPN CR
// can be set to the Node object
func (r *NativePodNetworkReconciler) getNodeObjectInCluster(ctx context.Context, cr types.NamespacedName, nodeName string) (*v1.Node, error) {
	log := log.FromContext(ctx, "namespacedName", cr).WithValues("nodeName", nodeName)
	nodeObject := v1.Node{}
	nodePresentInCluster := func() (bool, error) {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		err := r.Client.Get(ctx, types.NamespacedName{
			Name: nodeName,
		}, &nodeObject)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Error(err, "node object does not exist in cluster")
				return false, nil
			}
			log.Error(err, "failed to get node object")
			return false, err
		}
		return true, nil
	}

	err := wait.PollImmediate(time.Second*5, GetNodeTimeout, func() (bool, error) {
		present, err := nodePresentInCluster()
		if err != nil {
			log.Error(err, "failed to get node from cluster")
			return false, err
		}
		return present, nil
	})
	if err != nil {
		log.Error(err, "timed out waiting for node object to be present in the cluster")
	}
	return &nodeObject, err
}

// getIpFamilies is a method to get ip families (IPv4/IPv6) from the NPN CRD
func getIpFamilies(ctx context.Context, npn npnv1beta1.NativePodNetwork) ([]string, error) {
	log := log.FromContext(ctx, "name", npn.Name)

	ipFamilies := []string{}
	if npn.Spec.IPFamilies != nil {
		for _, ipFamily := range npn.Spec.IPFamilies {
			if ipFamily != nil && len(*ipFamily) != 0 {
				ipFamilies = append(ipFamilies, *ipFamily)
			}
		}
	}
	log.WithValues("ipFamilies", fmt.Sprint(ipFamilies)).Info("IpFamily for NPN CR")

	return ipFamilies, nil
}

// wait for the compute instance to move to running state
func (r *NativePodNetworkReconciler) waitForInstanceToReachRunningState(ctx context.Context, npn npnv1beta1.NativePodNetwork) error {
	log := log.FromContext(ctx, "name", npn.Name)
	log = log.WithValues("instanceId", *npn.Spec.Id)

	instanceIsInRunningState := func() (bool, error) {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		instance, err := r.OCIClient.Compute().GetInstance(ctx, *npn.Spec.Id)
		if err != nil || instance.Id == nil {
			return false, err

		}
		if instance.LifecycleState != core.InstanceLifecycleStateRunning {
			log.WithValues("instanceLifecycle", instance.LifecycleState).Info("Instance is still not in running state")
			return false, nil
		}
		return true, nil
	}

	err := wait.PollImmediate(time.Second*10, GetNodeTimeout, func() (bool, error) {
		running, err := instanceIsInRunningState()
		if err != nil {
			log.Error(err, "failed to get OCI instance")
			return false, err
		}
		return running, nil
	})
	if err != nil {
		log.Error(err, "timed out waiting for instance to reach running state")
	}
	return err
}

// ensureVnicAttachedAndAvailable polls until vnic attachment is attached and vnic is available.
// We might keep waiting for 2 minutes when VNIC attach fails i.e. VNIC Attachment goes to Detaching/Detached
// and Vnic to Terminated/Terminating states so we error out in those situations and stop retrying
func (r *NativePodNetworkReconciler) ensureVnicAttachedAndAvailable(ctx context.Context, vnicAttachment *core.VnicAttachment) (vnicId *string, ensured bool, err error) {
	err = wait.PollImmediate(time.Second*5, ensureVnicAttachedAndAvailablePollDuration, func() (bool, error) {
		log := log.FromContext(ctx)
		if vnicAttachment.Id == nil {
			return false, errors.New("vnic attachment Id is nil")
		}
		vnicAttachment, err = r.OCIClient.Compute().GetVnicAttachment(ctx, vnicAttachment.Id)
		if err != nil {
			return false, err
		}
		if vnicAttachment.LifecycleState == core.VnicAttachmentLifecycleStateDetached ||
			vnicAttachment.LifecycleState == core.VnicAttachmentLifecycleStateDetaching {
			log.Error(err, "vnic attachment is detaching/detached", "vnicAttachment", vnicAttachment.Id)
			return false, errors.New("vnic attachment is in detaching/detached state")
		}
		if vnicAttachment.VnicId == nil {
			return false, nil
		}
		if vnicAttachment.LifecycleState != core.VnicAttachmentLifecycleStateAttached {
			log.WithValues("vnicAttachment", vnicAttachment.Id, "LifecycleState", vnicAttachment.LifecycleState).Info("vnic attachment is not in attached state, will retry")
			return false, nil
		}

		vNIC, err := r.OCIClient.Networking(nil).GetVNIC(ctx, *vnicAttachment.VnicId)
		if err != nil {
			log.Error(err, "failed to ensure vnic attached and available")
			return false, errors2.Wrap(err, "failed to get VNIC from VNIC attachment")
		}
		log = log.WithValues("vnic", vNIC.Id)
		if vNIC.LifecycleState == core.VnicLifecycleStateTerminating || vNIC.LifecycleState == core.VnicLifecycleStateTerminated {
			log.Error(err, "vnic is terminating/terminated")
			return false, errors.New("vnic is in terminating/terminated state")
		}
		if vNIC.LifecycleState != core.VnicLifecycleStateAvailable {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return nil, false, err
	}
	return vnicAttachment.VnicId, true, nil
}

// validateVnicAttachmentsAreInAttachedState will validate if the vnics have been attached
func (r *NativePodNetworkReconciler) validateVnicAttachmentsAreInAttachedState(ctx context.Context, InstanceId string, requiredSecondaryVNICs int, attachedSecondaryVnics []SubnetVnic) (attached bool, err error) {
	log := log.FromContext(ctx, "instanceId", InstanceId)

	if requiredSecondaryVNICs != len(attachedSecondaryVnics) {
		return false, errNotEnoughVnicsAttached
	}

	for _, vnicAttachment := range attachedSecondaryVnics {
		if _, ensured, err := r.ensureVnicAttachedAndAvailable(ctx, vnicAttachment.Attachment); !ensured {
			log.Error(err, "Failed to ensure Vnic is attached & available")
			return false, err
		}
	}
	return true, nil
}

func getNodeNameFromPrimaryVnic(ip *core.Vnic, ipFamilies []string) string {
	if contains(ipFamilies, IPv6) {
		if ip.PrivateIp != nil && *ip.PrivateIp != "" {
			return *ip.PrivateIp
		}
		if len(ip.Ipv6Addresses) > 0 {
			return strings.ReplaceAll(ip.Ipv6Addresses[0], ":", "-")
		}
	}
	if ip.PrivateIp != nil {
		return *ip.PrivateIp
	}
	return ""
}

func toOCIIpv6PairDetails(source []npnv1beta1.Ipv6AddressIpv6SubnetCidrPairDetail) []core.Ipv6AddressIpv6SubnetCidrPairDetails {
	if source == nil {
		return nil
	}
	out := make([]core.Ipv6AddressIpv6SubnetCidrPairDetails, 0, len(source))
	for _, s := range source {
		out = append(out, core.Ipv6AddressIpv6SubnetCidrPairDetails{
			Ipv6Address:    common.String(s.Ipv6Address),
			Ipv6SubnetCidr: common.String(s.Ipv6SubnetCidr),
		})
	}
	return out
}

func isGvaNode(npn *npnv1beta1.NativePodNetwork) bool {
	return npn.Spec.SecondaryVnics != nil && len(npn.Spec.SecondaryVnics) > 0
}

func validateVnicIPCountsAgainstConfiguredIpCount(existingIpsByVnic map[string]*vnicSecondaryAddresses, vnics []npnv1beta1.VNICAddress, ipFamilies []string) error {
	for _, vnic := range vnics {
		if vnic.NicConfiguration == nil {
			continue
		}
		if vnic.VNICID == nil || vnic.NicConfiguration.IpCount == nil {
			continue // skip if ID or IpCount is missing
		}
		vnicID := *vnic.VNICID
		configuredCount := *vnic.NicConfiguration.IpCount

		actualV4 := len(existingIpsByVnic[vnicID].V4)
		actualV6 := len(existingIpsByVnic[vnicID].V6)

		if (len(ipFamilies) == 0 || contains(ipFamilies, IPv4)) && actualV4 != configuredCount {
			return errors2.Errorf("VNIC %s: IPv4 count (%d) != configured ipCount (%d)", vnicID, actualV4, configuredCount)
		}
		if contains(ipFamilies, IPv6) && actualV6 != configuredCount {
			return errors2.Errorf("VNIC %s: IPv6 count (%d) != configured ipCount (%d)", vnicID, actualV6, configuredCount)
		}
	}
	return nil
}

func (r *NativePodNetworkReconciler) attachUserSpecifiedSecondaryVnics(ctx context.Context, npn npnv1beta1.NativePodNetwork) ([]GvaNics, error) {
	log := log.FromContext(ctx)
	attachmentResponses := make([]VnicAttachmentResponse, len(npn.Spec.SecondaryVnics))

	gvaNics := make([]GvaNics, 0, len(npn.Spec.SecondaryVnics))
	for idx, sv := range npn.Spec.SecondaryVnics {
		c := sv.CreateVnicDetails
		subnet, err := r.OCIClient.Networking(nil).GetSubnet(ctx, c.SubnetId)
		if err != nil {
			log.Error(err, "failed to get subnet for VNIC")
			return gvaNics, err
		}
		ipFamilies := getVnicIpFamily(subnet, sv.CreateVnicDetails)

		opts := ociclient.AttachVnicOptions{
			InstanceID:                           npn.Spec.Id,
			SubnetID:                             &c.SubnetId,
			NsgIds:                               c.NsgIds,
			SkipSourceDestCheck:                  &c.SkipSourceDestCheck,
			DisplayName:                          &c.DisplayName,
			AssignPublicIp:                       &c.AssignPublicIp,
			AssignIpv6Ip:                         &c.AssignIpv6Ip,
			DefinedTags:                          c.DefinedTags,
			FreeformTags:                         c.FreeformTags,
			Ipv6AddressIpv6SubnetCidrPairDetails: toOCIIpv6PairDetails(c.Ipv6AddressIpv6SubnetCidrPairDetails),
			SecurityAttributes:                   c.SecurityAttributes,
		}
		sv.CreateVnicDetails.IpFamilies = ipFamilies

		startTime := time.Now()
		vnicAttachment, err := r.OCIClient.Compute().AttachVnic(ctx, opts)
		if err != nil {
			log.Error(err, "failed to attach user-specified SecondaryVNIC", "secondaryVnicIndex", idx, "displayName", sv.DisplayName)
			r.PushMetric(VnicAttachmentResponseSlice(attachmentResponses).ErrorMetric(), "")
			return gvaNics, err
		}

		attachmentResponses[idx].VnicAttachment = vnicAttachment
		attachmentResponses[idx].err = err
		attachmentResponses[idx].timeTaken = float64(time.Since(startTime).Seconds())

		log.Info("User-specified SecondaryVNIC attached", "secondaryVnicIndex", idx, "vnicAttachmentId", vnicAttachment.Id)
		vnicID, ensured, err := r.ensureVnicAttachedAndAvailable(ctx, &vnicAttachment)
		if !ensured || err != nil {
			log.Error(err, "failed to ensure SecondaryVNIC is available", "secondaryVnicIndex", idx)
			r.PushMetric(VnicAttachmentResponseSlice(attachmentResponses).ErrorMetric(), "")
			return gvaNics, err
		}
		log.WithValues("vnicId", vnicID).Info("User-specified SecondaryVNIC attached")
		gvaNics = append(gvaNics, GvaNics{
			vnicId:            vnicID,
			SecondaryVnicSpec: &sv,
		})

	}

	// Push metrics for all attaches in this batch
	r.PushMetric(VnicAttachmentResponseSlice(attachmentResponses).ErrorMetric(), "")
	log.Info("All user-specified SecondaryVnics successfully attached")
	return gvaNics, nil
}

func extractIpCountAndApplicationResource(vnicId string, nics []GvaNics) (ipCount int, applicationResource []string) {
	for _, nic := range nics {
		if *nic.vnicId == vnicId {
			return nic.SecondaryVnicSpec.CreateVnicDetails.IpCount, nic.SecondaryVnicSpec.CreateVnicDetails.ApplicationResources
		}
	}
	return GVA_DEFAULT_IP_COUNT, []string{}
}

// getSecurityTag : method unused, to be supported whenever NPWF passes SecurityAttributes via NPN.Spec
func getSecurityTag(attributes map[string]map[string]interface{}, attributeName, tagName, field string) (value string, ok bool) {
	if attr, exists := attributes[attributeName]; exists {
		if tagAny, exists := attr[tagName]; exists {
			if tag, typeOk := tagAny.(map[string]interface{}); typeOk {
				if v, exists := tag[field]; exists {
					if strVal, strOk := v.(string); strOk {
						return strVal, true
					}
				}
			}
		}
	}
	return "", false
}

func getNicConfigurationForNodes(vnic *core.Vnic, subnet *core.Subnet, nics []GvaNics) *npnv1beta1.NicConfiguration {
	ipCount, applicationResources := extractIpCountAndApplicationResource(*vnic.Id, nics)
	var ipFamilies []string
	for _, nic := range nics {
		if nic.vnicId != nil && vnic.Id != nil && *nic.vnicId == *vnic.Id {
			ipFamilies = getVnicIpFamily(subnet, nic.SecondaryVnicSpec.CreateVnicDetails)
			break
		}
	}

	return &npnv1beta1.NicConfiguration{
		IpCount:                 common.Int(ipCount),
		IpFamilies:              ipFamilies,
		SubnetId:                vnic.SubnetId,
		NetworkSecurityGroupIDs: vnic.NsgIds,
		ApplicationResources:    applicationResources,
	}
}

func getVnicIpFamily(subnet *core.Subnet, details npnv1beta1.CreateVnicDetails) []string {
	if utils.IsIpv6SingleStackSubnet(subnet) {
		return []string{IPv6}
	}
	if utils.IsIpv4SingleStackSubnet(subnet) {
		return []string{IPv4}
	}
	// At this point, must be dual-stack
	if details.AssignIpv6Ip {
		return []string{IPv4, IPv6}
	}
	return []string{IPv4}
}
func stringPointerToStringSlice(original []*string) []string {
	stringArray := make([]string, 0, len(original))
	for _, value := range original {
		stringArray = append(stringArray, *value)
	}
	return stringArray
}

func (r *NativePodNetworkReconciler) getGvaNic(ctx context.Context, npn *npnv1beta1.NativePodNetwork, existingSecondaryVNICs []SubnetVnic) ([]GvaNics, error) {
	gvaNics := make([]GvaNics, 0, len(existingSecondaryVNICs))

	for _, vnic := range existingSecondaryVNICs {
		var matchedSpec *npnv1beta1.SecondaryVnic
		vnicDisplayName := ""
		if vnic.Vnic.DisplayName != nil {
			vnicDisplayName = *vnic.Vnic.DisplayName
		}
		subnet, err := r.OCIClient.Networking(nil).GetSubnet(ctx, *vnic.Vnic.SubnetId)
		if err != nil {
			return gvaNics, err
		}
		for idx, sv := range npn.Spec.SecondaryVnics {
			if sv.CreateVnicDetails.DisplayName == vnicDisplayName {
				ipFamilies := getVnicIpFamily(subnet, sv.CreateVnicDetails)
				matchedSpec = &npn.Spec.SecondaryVnics[idx]
				matchedSpec.CreateVnicDetails.IpFamilies = ipFamilies
				break
			}
		}
		if matchedSpec == nil {
			continue
		}

		gvaNics = append(gvaNics, GvaNics{
			vnicId:            vnic.Vnic.Id,
			SecondaryVnicSpec: matchedSpec,
		})
	}
	return gvaNics, nil
}

func checkApplicationResourcesUsedOnAllVnics(gvaNics []GvaNics) bool {
	for _, vnic := range gvaNics {
		if vnic.SecondaryVnicSpec != nil {
			applicationResources := vnic.SecondaryVnicSpec.CreateVnicDetails.ApplicationResources
			if len(applicationResources) == 0 {
				return false
			}
		}
	}
	return true
}

func (r *NativePodNetworkReconciler) setApplicationResourceTaint(ctx context.Context, nodeObject *v1.Node) (*v1.Node, error) {
	taint := v1.Taint{
		Key:    TaintKeyApplicationResourceOnly,
		Value:  "",
		Effect: v1.TaintEffectNoSchedule,
	}

	taintFound := false
	for _, t := range nodeObject.Spec.Taints {
		if t.Key == taint.Key && t.Effect == taint.Effect {
			taintFound = true
			break
		}
	}
	if !taintFound {
		nodeObject.Spec.Taints = append(nodeObject.Spec.Taints, taint)
		if err := r.Client.Update(ctx, nodeObject); err != nil {
			return nil, fmt.Errorf("failed to update node with taint: %w", err)
		}
	}
	return nodeObject, nil
}

func expandIpAddressesInRange(ipAddress string, cidrPrefix int) ([]string, error) {
	prefix, err := netip.ParsePrefix(fmt.Sprintf("%s/%d", ipAddress, cidrPrefix))
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %s/%d to extract the addresses in the range", ipAddress, cidrPrefix)
	}

	// normalized to the network address—that is, all host bits set to zero.
	// needed for the loop start from the correct starting network address
	// ex:192.168.1.10/24 should start from 192.168.1.0
	prefix = prefix.Masked()

	var ips []string
	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {
		ips = append(ips, addr.String())
	}
	return ips, nil
}

func getBlockSizeForCidrPrefix(cidrPrefix int, ipFamily string) int {
	var addressBits int
	addressBits = 32
	if ipFamily == IPv6 {
		addressBits = 128
	}
	return int(math.Exp2(float64(addressBits - cidrPrefix)))
}

// getCidrPrefixLengthForBlockSize returns a CIDR prefix length for a block that can hold at least `size` IPs.
func getCidrPrefixLengthForBlockSize(size int, ipFamily string) (*int, error) {
	if size <= 0 {
		return nil, nil
	}
	if size == 1 && strings.EqualFold(ipFamily, IPv6) {
		return nil, nil
	}

	if !isPowerOf2(size) {
		return nil, fmt.Errorf("invalid cidr block size: %d. must be power of 2", size)
	}

	k := int(math.Log2(float64(size)))

	if strings.EqualFold(ipFamily, IPv6) {
		prefix := 128 - k
		// Round down to a multiple of 4 (larger block or equal capacity), still >= size.
		prefix -= prefix % 4
		return &prefix, nil
	}

	// Default IPv4
	prefix := 32 - k
	return &prefix, nil
}

func isPowerOf2(i int) bool {
	l := math.Log2(float64(i))

	// valdiate if the i is power of 2
	// return false if 3.23 != 3
	return math.Trunc(l) == l
}
