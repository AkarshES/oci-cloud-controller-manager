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
	"sync"
	"time"

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
	vnicId string
	ips    IpAddressCountByVersion
}

type VnicIPAllocationResponse struct {
	vnicId        string
	errIPv4       error
	errIPv6       error
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
	var npn npnv1beta1.NativePodNetwork
	defer func() {
		if failMessage != "" && err != nil {
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
	requiredSecondaryVNICs := int(math.Ceil(float64(*npn.Spec.MaxPodCount) / maxPodIpsPerVNIC))
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
	nodeName := primaryVnic.PrivateIp
	log.WithValues("existingSecondaryVNICs", existingSecondaryVNICs).
		WithValues("countOfExistingSecondaryVNICs", len(existingSecondaryVNICs)).
		Info(FetchedExistingSecondaryVNICsForInstance)

	requiredAdditionalSecondaryVNICs := requiredSecondaryVNICs - len(existingSecondaryVNICs)

	if requiredAdditionalSecondaryVNICs > 0 {
		log.WithValues("requiredAdditionalSecondaryVNICs", requiredAdditionalSecondaryVNICs).Info(AllocateAdditionalVNICsToInstance)
		additionalVNICAttachments := make([]VnicAttachmentResponse, requiredAdditionalSecondaryVNICs)
		for index := 0; index < requiredAdditionalSecondaryVNICs; index++ {
			startTime := time.Now()
			vnicAttachment, err := r.OCIClient.Compute().AttachVnic(ctx, npn.Spec.Id, npn.Spec.PodSubnetIds[0], npn.Spec.NetworkSecurityGroupIds, &SKIP_SOURCE_DEST_CHECK)
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

			if ensured, err := r.ensureVnicAttachedAndAvailable(ctx, &vnicAttachment); !ensured {
				failReason, failMessage = "AttachAdditionalVNICsFailed", "failed to ensure required additional VNICs"
				log.WithValues("requiredAdditionalSecondaryVNICs", requiredAdditionalSecondaryVNICs).
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
		log.WithValues("requiredAdditionalSecondaryVNICs", requiredAdditionalSecondaryVNICs).Info(AllocatedAdditionalVNICsToInstance)
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

	vnicAttached, err := r.validateVnicAttachmentsAreInAttachedState(ctx, *instance.Id, requiredSecondaryVNICs, existingSecondaryVNICs)
	if vnicAttached == false || err != nil {
		failReason, failMessage = "AttachAdditionalVNICsFailed", "failed to validate required VNICs"
		log.Error(err, failMessage)
		r.handleError(ctx, req, err, "AttachVNIC")
		return ctrl.Result{}, err
	}

	ipFamilies, err := getIpFamilies(ctx, npn)
	if err != nil {
		log.Error(err, "failed to get IpFamilies from NPN CR")
		r.handleError(ctx, req, err, "GetNPN_IPFamilies")
		return ctrl.Result{}, err
	}

	log.Info(FetchingPrivateIPsForSecondaryVNICs)
	existingSecondaryIpsbyVNIC, err := r.getSecondaryIpsByVNICs(ctx, existingSecondaryVNICs, ipFamilies)
	if err != nil {
		failReason = "ListPrivateIPsFailed"
		r.handleError(ctx, req, err, "ListPrivateIP")
		return ctrl.Result{}, err
	}
	totalAllocatedSecondaryIPs := totalAllocatedSecondaryIpsForInstance(existingSecondaryIpsbyVNIC)
	log.WithValues("countOfExistingSecondaryIps", totalAllocatedSecondaryIPs).Info(FetchedPrivateIPsForSecondaryVNICs)

	log.Info(ComputingAdditionalIpsByVnic)
	additionalIpsByVnic, err := getAdditionalSecondaryIPsNeededPerVNIC(existingSecondaryIpsbyVNIC, *npn.Spec.MaxPodCount, totalAllocatedSecondaryIPs, ipFamilies)
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
		parallelLog := log.WithValues("vnicId", additionalIpsByVnic[outerIndex].vnicId).WithValues("requiredIPs", additionalIpsByVnic[outerIndex].ips)
		var errIPv4 error = nil
		var errIPv6 error = nil
		allocations := IpAllocations{}
		if len(ipFamilies) == 0 || contains(ipFamilies, IPv4) {
			if additionalIpsByVnic[outerIndex].ips.V4 > 0 {
				parallelLog.Info("Need to allocate secondary IPv4 for VNIC")
				ipv4Allocations := make([]IPAllocation, additionalIpsByVnic[outerIndex].ips.V4)
				for innerIndex := 0; innerIndex < additionalIpsByVnic[outerIndex].ips.V4; innerIndex++ {
					startTime := time.Now()
					_, err := r.OCIClient.Networking(nil).CreatePrivateIp(ctx, additionalIpsByVnic[outerIndex].vnicId)
					if err != nil {
						parallelLog.Error(err, "failed to create IPv4")
					}
					ipv4Allocations[innerIndex].err = err
					ipv4Allocations[innerIndex].timeTaken = float64(time.Since(startTime).Seconds())
				}
				errIPv4 = validateVnicIpAllocation(ipv4Allocations)
				allocations.V4 = ipv4Allocations
			}
		}
		if contains(ipFamilies, IPv6) {
			if additionalIpsByVnic[outerIndex].ips.V6 > 0 {
				parallelLog.Info("Need to allocate secondary IPv6 for VNIC")
				ipv6Allocations := make([]IPAllocation, additionalIpsByVnic[outerIndex].ips.V6)
				for innerIndex := 0; innerIndex < additionalIpsByVnic[outerIndex].ips.V6; innerIndex++ {
					startTime := time.Now()
					_, err := r.OCIClient.Networking(nil).CreateIpv6(ctx, additionalIpsByVnic[outerIndex].vnicId)
					if err != nil {
						parallelLog.Error(err, "failed to create IPv6")
					}
					ipv6Allocations[innerIndex].err = err
					ipv6Allocations[innerIndex].timeTaken = float64(time.Since(startTime).Seconds())
				}
				errIPv6 = validateVnicIpAllocation(ipv6Allocations)
				allocations.V6 = ipv6Allocations
			}
		}
		mutex.Lock()
		vnicAdditionalIpAllocations[outerIndex] = VnicIPAllocationResponse{additionalIpsByVnic[outerIndex].vnicId, errIPv4, errIPv6, allocations}
		mutex.Unlock()
	})
	for _, ips := range vnicAdditionalIpAllocations {
		if len(ipFamilies) == 0 || contains(ipFamilies, IPv4) {
			if ips.errIPv4 != nil {
				failReason, failMessage = "CreatePrivateIPFailed", ips.errIPv4.Error()
				r.handleError(ctx, req, ips.errIPv4, "CreatePrivateIP")
				r.PushMetric(IPAllocationSlice(ips.ipAllocations.V4).ErrorMetric(), IPv4)
				return ctrl.Result{}, ips.errIPv4
			}
			r.PushMetric(IPAllocationSlice(ips.ipAllocations.V4).ErrorMetric(), IPv4)
		}
		if contains(ipFamilies, IPv6) {
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
	existingSecondaryIpsbyVNIC, err = r.getSecondaryIpsByVNICs(ctx, existingSecondaryVNICs, ipFamilies)
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
	existingSecondaryIpsbyVNIC = assignHostIpAddressForVnic(existingSecondaryIpsbyVNIC, ipFamilies)

	// validate if maxPodCount = number of secondary IPs available on the vnics
	err = validateMaxPodCountWithSecondaryIPCount(existingSecondaryIpsbyVNIC, *npn.Spec.MaxPodCount, ipFamilies)
	if err != nil {
		failReason = "IPsNotEqualToMaxPodCount"
		log.Error(err, "secondary IPs are not equal to MaxPodCount")
		r.handleError(ctx, req, err, "validateMaxPodCountWithSecondaryIPCount")
		return ctrl.Result{}, err
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
	nodeObject, err := r.getNodeObjectInCluster(ctx, req.NamespacedName, *nodeName)
	if err != nil {
		failReason = "GetV1NodeFailed"
		r.handleError(ctx, req, err, "GetV1Node")
		return ctrl.Result{}, err
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
	updateNPN.Status.VNICs = convertCoreVNICtoNPNStatus(existingSecondaryVNICs, existingSecondaryIpsbyVNIC, ipFamilies)
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
func (r *NativePodNetworkReconciler) getSecondaryIpsByVNICs(ctx context.Context, existingSecondaryVNICs []SubnetVnic, ipFamilies []string) (map[string]*vnicSecondaryAddresses, error) {
	ipsByVNICs := make(map[string]*vnicSecondaryAddresses)
	log := log.FromContext(ctx)
	for _, secondary := range existingSecondaryVNICs {
		vnicSecondaryAddresses := &vnicSecondaryAddresses{}
		log := log.WithValues("vnicId", *secondary.Vnic.Id)
		var err error
		if len(ipFamilies) == 0 || contains(ipFamilies, IPv4) {
			vnicSecondaryAddresses.V4, err = r.OCIClient.Networking(nil).ListPrivateIps(ctx, *secondary.Vnic.Id)
			if err != nil {
				log.Error(err, "failed to list secondary IPv4 IPs for VNIC")
				return nil, err
			}
		}
		if contains(ipFamilies, IPv6) {
			vnicSecondaryAddresses.V6, err = r.OCIClient.Networking(nil).ListIpv6s(ctx, *secondary.Vnic.Id)
			if err != nil {
				log.Error(err, "failed to list secondary IPv6 IPs for VNIC")
				return nil, err
			}
		}
		vnicSecondaryAddresses = filterPrimaryIp(vnicSecondaryAddresses)
		ipsByVNICs[*secondary.Vnic.Id] = vnicSecondaryAddresses
	}
	return ipsByVNICs, nil
}

// assignHostIpAddressForVnic is a util method to get HostIP Address per vnic
func assignHostIpAddressForVnic(existingSecondaryVNICs map[string]*vnicSecondaryAddresses, ipFamilies []string) map[string]*vnicSecondaryAddresses {
	IPsbyVNICs := make(map[string]*vnicSecondaryAddresses)
	for k, vnic := range existingSecondaryVNICs {
		if len(ipFamilies) == 0 || contains(ipFamilies, IPv4) {
			if len(vnic.V4) >= 1 {
				vnic.hostIpv4, vnic.V4 = vnic.V4[0].IpAddress, vnic.V4[1:]
			}
		}
		if contains(ipFamilies, IPv6) {
			if len(vnic.V6) >= 1 {
				vnic.hostIpv6, vnic.V6 = vnic.V6[0].IpAddress, vnic.V6[1:]
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
		V6: []core.Ipv6{},
		V4: []core.PrivateIp{},
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
func getAdditionalSecondaryIPsNeededPerVNIC(existingIpsByVnic map[string]*vnicSecondaryAddresses, maxPodCount int, allocatedSecondaryIps IpAddressCountByVersion, ipFamilies []string) ([]VnicIPAllocations, error) {
	if maxPodCount == 0 {
		return []VnicIPAllocations{}, nil
	}

	// Required Host Addresses is supposed to be one host address per vnic
	requiredHostAddressesIPv4 := 0
	requiredHostAddressesIPv6 := 0
	for _, vnic := range existingIpsByVnic {
		if len(vnic.V4) == 0 {
			requiredHostAddressesIPv4++
		}
		if len(vnic.V6) == 0 {
			requiredHostAddressesIPv6++
		}
	}

	requiredSecondaryIPv4 := 0
	requiredSecondaryIPv6 := 0
	var requireIPv6, requireIPv4 bool
	if len(ipFamilies) == 0 || contains(ipFamilies, IPv4) {
		requiredSecondaryIPv4 = maxPodCount - allocatedSecondaryIps.V4 + requiredHostAddressesIPv4
		requireIPv4 = true
	}
	if contains(ipFamilies, IPv6) {
		requiredSecondaryIPv6 = maxPodCount - allocatedSecondaryIps.V6 + requiredHostAddressesIPv6
		requireIPv6 = true
	}
	additionalIpsByVnic := make([]VnicIPAllocations, 0)
	for vnic, existingIps := range existingIpsByVnic {
		ipAllocation := IpAddressCountByVersion{}
		if requireIPv4 {
			allocatableIPv4 := maxSecondaryIpsPerVNIC - len(existingIps.V4)
			if len(existingIps.V4) == maxSecondaryIpsPerVNIC {
				ipAllocation.V4 = 0
			} else if allocatableIPv4 > requiredSecondaryIPv4 {
				ipAllocation.V4 = requiredSecondaryIPv4
				requiredSecondaryIPv4 -= requiredSecondaryIPv4
			} else {
				ipAllocation.V4 = allocatableIPv4
				requiredSecondaryIPv4 -= allocatableIPv4
			}
		}
		if requireIPv6 {
			allocatableIPv6 := maxSecondaryIpsPerVNIC - len(existingIps.V6)
			if len(existingIps.V6) == maxSecondaryIpsPerVNIC {
				ipAllocation.V6 = 0
			} else if allocatableIPv6 > requiredSecondaryIPv6 {
				ipAllocation.V6 = requiredSecondaryIPv6
				requiredSecondaryIPv6 -= requiredSecondaryIPv6
			} else {
				ipAllocation.V6 = allocatableIPv6
				requiredSecondaryIPv6 -= allocatableIPv6
			}
		}
		additionalIpsByVnic = append(additionalIpsByVnic, VnicIPAllocations{vnic, ipAllocation})
	}
	if requiredSecondaryIPv4 > 0 || requiredSecondaryIPv6 > 0 {
		return nil, errors.New("failed to allocate the required number of IPs with existing VNICs")
	}
	return additionalIpsByVnic, nil
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
func convertCoreVNICtoNPNStatus(existingSecondaryVNICs []SubnetVnic, existingSecondaryIpsByVNIC map[string]*vnicSecondaryAddresses, ipFamilies []string) []npnv1beta1.VNICAddress {
	requireIPv6s, requireIPv4s := contains(ipFamilies, IPv6), contains(ipFamilies, IPv4)

	npnVNICAddresses := make([]npnv1beta1.VNICAddress, 0, len(existingSecondaryIpsByVNIC))
	for _, vnic := range existingSecondaryVNICs {
		npnVNICAddress := npnv1beta1.VNICAddress{
			VNICID:     vnic.Vnic.Id,
			MACAddress: vnic.Vnic.MacAddress,
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
		if len(ipFamilies) > 0 {
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
		if len(ipFamilies) == 0 || requireIPv4s {
			secondaryIpCount = len(existingSecondaryIpsByVNIC[*vnic.Vnic.Id].V4)
		}
		if requireIPv6s {
			secondaryIpCount = len(existingSecondaryIpsByVNIC[*vnic.Vnic.Id].V6)
		}
		vnicAddresses := make([]*string, 0, secondaryIpCount)
		vnicPodAddresses := make([]npnv1beta1.PodAddress, 0, secondaryIpCount)
		var ipv4IP, ipv6IP *string
		for i := 0; i < secondaryIpCount; i++ {
			if len(ipFamilies) == 0 || requireIPv4s {
				// Populate the old fields in case of IPv4 or ipFamilies length == 0
				vnicAddresses = append(vnicAddresses, vnicSecondaryAddresses.V4[i].IpAddress)
			}
			if requireIPv4s {
				ipv4IP = vnicSecondaryAddresses.V4[i].IpAddress
			}
			if requireIPv6s {
				ipv6IP = vnicSecondaryAddresses.V6[i].IpAddress
			}
			if len(ipFamilies) > 0 {
				// Populate new fields only in case of IPFamilies being present in CRD
				vnicPodAddresses = append(vnicPodAddresses, npnv1beta1.PodAddress{
					V4: ipv4IP,
					V6: ipv6IP,
				})
			}
		}

		if len(ipFamilies) == 0 || requireIPv4s {
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
func (r *NativePodNetworkReconciler) ensureVnicAttachedAndAvailable(ctx context.Context, vnicAttachment *core.VnicAttachment) (ensured bool, err error) {
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
		return false, err
	}
	return true, nil
}

// validateVnicAttachmentsAreInAttachedState will validate if the vnics have been attached
func (r *NativePodNetworkReconciler) validateVnicAttachmentsAreInAttachedState(ctx context.Context, InstanceId string, requiredSecondaryVNICs int, attachedSecondaryVnics []SubnetVnic) (attached bool, err error) {
	log := log.FromContext(ctx, "instanceId", InstanceId)

	if requiredSecondaryVNICs != len(attachedSecondaryVnics) {
		return false, errNotEnoughVnicsAttached
	}

	for _, vnicAttachment := range attachedSecondaryVnics {
		if ensured, err := r.ensureVnicAttachedAndAvailable(ctx, vnicAttachment.Attachment); !ensured {
			log.Error(err, "Failed to ensure Vnic is attached & available")
			return false, err
		}
	}
	return true, nil
}
