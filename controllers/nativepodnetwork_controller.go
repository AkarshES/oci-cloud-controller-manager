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
	"math"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	npnv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/v1beta1"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/core"
)

const (
	CREATE_PRIVATE_IP             = "CREATE_PRIVATE_IP"
	ATTACH_VNIC                   = "ATTACH_VNIC"
	INITIALIZE_NPN_NODE           = "INITIALIZE_NPN_NODE"
	maxSecondaryPrivateIPsPerVNIC = 31
	// GetNodeTimeout is the timeout for the node object to be created in Kubernetes
	GetNodeTimeout = 20 * time.Minute
	// RunningInstanceTimeout is the timeout for the instance to reach running state
	// before we try to attach VNIC(s) to them
	RunningInstanceTimeout                   = 5 * time.Minute
	FetchedExistingSecondaryVNICsForInstance = "Fetched existingSecondaryVNICs for instance"
)

var (
	STATE_SUCCESS     = "SUCCESS"
	STATE_IN_PROGRESS = "IN_PROGRESS"
	STATE_BACKOFF     = "BACKOFF"
	COMPLETED         = "COMPLETED"

	SKIP_SOURCE_DEST_CHECK = true
	errPrimaryVnicNotFound = errors.New("failed to get primary vnic for instance")
	errInstanceNotRunning  = errors.New("instance is not in running state")
)

// NativePodNetworkReconciler reconciles a NativePodNetwork object
type NativePodNetworkReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	MetricPusher     *metrics.MetricPusher
	OCIClient        ociclient.Interface
	TimeTakenTracker map[string]time.Time
}

// VnicAttachmentResponse is used to store the response for attach VNIC
type VnicAttachmentResponse struct {
	VnicAttachment core.VnicAttachment
	err            error
	timeTaken      float64
}

type VnicIPAllocations struct {
	vnicId string
	ips    int
}

type VnicIPAllocationResponse struct {
	vnicId        string
	err           error
	ipAllocations []IPAllocation
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
	Vnic   *core.Vnic
	Subnet *core.Subnet
}

type ErrorMetric interface {
	GetMetricName() string
	GetTimeTaken() float64
	GetError() error
}
type ConvertToErrorMetric interface {
	ErrorMetric() []ErrorMetric
}

func (r NativePodNetworkReconciler) PushMetric(errorArray []ErrorMetric) {
	averageByReturnCode := computeAveragesByReturnCode(errorArray)
	if len(errorArray) == 0 {
		return
	}
	metricName := errorArray[0].GetMetricName()
	for k, v := range averageByReturnCode {
		dimensions := map[string]string{"component": k}
		metrics.SendMetricData(r.MetricPusher, metricName, v, dimensions)
	}
}

func (v IPAllocation) GetTimeTaken() float64 {
	return v.timeTaken
}
func (v IPAllocation) GetMetricName() string {
	return CREATE_PRIVATE_IP
}
func (v IPAllocation) GetError() error {
	return v.err
}

func (v VnicAttachmentResponse) GetTimeTaken() float64 {
	return v.timeTaken
}
func (v VnicAttachmentResponse) GetMetricName() string {
	return ATTACH_VNIC
}
func (v VnicAttachmentResponse) GetError() error {
	return v.err
}

func (v endToEndLatency) GetTimeTaken() float64 {
	return v.timeTaken
}
func (v endToEndLatency) GetMetricName() string {
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
	log := log.FromContext(ctx)
	if _, ok := r.TimeTakenTracker[req.Name]; !ok {
		r.TimeTakenTracker[req.Name] = time.Now()
	}
	startTime := r.TimeTakenTracker[req.Name]
	mutex := sync.Mutex{}
	var npn npnv1beta1.NativePodNetwork
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
	err := r.Status().Update(context.Background(), &npn)
	if err != nil {
		log.Error(err, "failed to set status on CR")
		return ctrl.Result{}, err
	}

	requiredSecondaryVNICs := int(math.Ceil(float64(*npn.Spec.MaxPodCount) / maxSecondaryPrivateIPsPerVNIC))
	instance, err := r.OCIClient.Compute().GetInstance(ctx, *npn.Spec.Id)
	if err != nil || instance.Id == nil {
		log.WithValues("instanceId", *npn.Spec.Id).Error(err, "failed to get OCI compute instance")
		return ctrl.Result{}, err
	}
	log = log.WithValues("instanceId", *instance.Id)

	// remove the CR in case the node never joined the cluster and the instance is terminated
	if instance.LifecycleState == core.InstanceLifecycleStateTerminated ||
		instance.LifecycleState == core.InstanceLifecycleStateTerminating {
		err = r.Client.Delete(ctx, &npn)
		if err != nil {
			log.Error(err, "failed to delete NPN CR for terminated instance")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		log.Info("Deleted the CR for Terminated compute instance")
		return ctrl.Result{}, nil
	}

	if instance.LifecycleState != core.InstanceLifecycleStateRunning {
		err = r.waitForInstanceToReachRunningState(ctx, npn)
		if err != nil {
			r.handleError(ctx, req, errInstanceNotRunning, "GetRunningInstance")
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
	}

	primaryVnic, existingSecondaryVNICs, err := r.getPrimaryAndSecondaryVNICs(ctx, *instance.CompartmentId, *instance.Id)
	if err != nil {
		r.handleError(ctx, req, err, "GetVNIC")
		return ctrl.Result{}, err
	}
	if primaryVnic == nil {
		r.handleError(ctx, req, errPrimaryVnicNotFound, "GetPrimaryVNIC")
		return ctrl.Result{}, errPrimaryVnicNotFound
	}
	nodeName := primaryVnic.PrivateIp
	log.WithValues("existingSecondaryVNICs", existingSecondaryVNICs).
		WithValues("countOfExistingSecondaryVNICs", len(existingSecondaryVNICs)).
		Info(FetchedExistingSecondaryVNICsForInstance)

	existingSecondaryIpsbyVNIC, err := r.getSecondaryPrivateIpsByVNICs(ctx, existingSecondaryVNICs)
	if err != nil {
		r.handleError(ctx, req, err, "ListPrivateIP")
		return ctrl.Result{}, err
	}
	totalAllocatedSecondaryIPs := totalAllocatedSecondaryIpsForInstance(existingSecondaryIpsbyVNIC)
	log.WithValues("countOfExistingSecondaryIps", totalAllocatedSecondaryIPs).Info("Fetched existingSecondaryIp for instance")

	requiredAdditionalSecondaryVNICs := requiredSecondaryVNICs - len(existingSecondaryVNICs)

	if requiredAdditionalSecondaryVNICs > 0 {
		log.WithValues("requiredAdditionalSecondaryVNICs", requiredAdditionalSecondaryVNICs).Info("Need to allocate VNICs for instance")
		additionalVNICAttachments := make([]VnicAttachmentResponse, requiredAdditionalSecondaryVNICs)
		workqueue.ParallelizeUntil(ctx, requiredAdditionalSecondaryVNICs, requiredAdditionalSecondaryVNICs, func(index int) {
			startTime := time.Now()
			vnicAttachment, err := r.OCIClient.Compute().AttachVnic(ctx, npn.Spec.Id, npn.Spec.PodSubnetIds[0], npn.Spec.NetworkSecurityGroupIds, &SKIP_SOURCE_DEST_CHECK)
			mutex.Lock()
			additionalVNICAttachments[index].VnicAttachment, additionalVNICAttachments[index].err = vnicAttachment, err
			if additionalVNICAttachments[index].err != nil {
				log.Error(additionalVNICAttachments[index].err, "failed to attach VNIC to instance")
			}
			additionalVNICAttachments[index].timeTaken = float64(time.Since(startTime).Seconds())
			if additionalVNICAttachments[index].err == nil {
				log.WithValues("vnic", additionalVNICAttachments[index].VnicAttachment).Info("VNIC attached to instance")
			}
			mutex.Unlock()
		})
		err = validateAdditionalVnicAttachments(additionalVNICAttachments)
		if err != nil {
			log.Error(err, "failed to provision required additional VNICs")
			r.handleError(ctx, req, err, "AttachVNIC")
			r.PushMetric(VnicAttachmentResponseSlice(additionalVNICAttachments).ErrorMetric())
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
		r.PushMetric(VnicAttachmentResponseSlice(additionalVNICAttachments).ErrorMetric())
		log.WithValues("requiredAdditionalSecondaryVNICs", requiredAdditionalSecondaryVNICs).Info("Allocated the required VNICs for instance")
	}

	_, existingSecondaryVNICs, err = r.getPrimaryAndSecondaryVNICs(ctx, *instance.CompartmentId, *instance.Id)
	if err != nil {
		r.handleError(ctx, req, err, "GetVNIC")
		return ctrl.Result{}, err
	}
	log.WithValues("existingSecondaryVNICs", existingSecondaryVNICs).
		WithValues("countOfExistingSecondaryVNICs", len(existingSecondaryVNICs)).
		Info(FetchedExistingSecondaryVNICsForInstance)

	existingSecondaryIpsbyVNIC, err = r.getSecondaryPrivateIpsByVNICs(ctx, existingSecondaryVNICs)
	if err != nil {
		r.handleError(ctx, req, err, "ListPrivateIP")
		return ctrl.Result{}, err
	}

	additionalIpsByVnic, err := getAdditionalSecondaryIPsNeededPerVNIC(existingSecondaryIpsbyVNIC, *npn.Spec.MaxPodCount-totalAllocatedSecondaryIPs)
	if err != nil {
		log.WithValues("additionalIpsRequired", *npn.Spec.MaxPodCount-totalAllocatedSecondaryIPs).Error(err, "failed to allocate the required IP addresses")
		r.handleError(ctx, req, err, "AllocatePrivateIP")
		return ctrl.Result{}, err
	}
	log.WithValues("additionalIpsByVnic", additionalIpsByVnic).Info("Computed required additionalIpsByVnic")

	vnicAdditionalIpAllocations := make([]VnicIPAllocationResponse, requiredSecondaryVNICs)
	workqueue.ParallelizeUntil(ctx, requiredSecondaryVNICs, requiredSecondaryVNICs, func(outerIndex int) {
		parallelLog := log.WithValues("vnicId", additionalIpsByVnic[outerIndex].vnicId).WithValues("requiredIPs", additionalIpsByVnic[outerIndex].ips)
		if additionalIpsByVnic[outerIndex].ips <= 0 {
			mutex.Lock()
			vnicAdditionalIpAllocations[outerIndex] = VnicIPAllocationResponse{additionalIpsByVnic[outerIndex].vnicId, nil, []IPAllocation{}}
			mutex.Unlock()
			return
		}
		parallelLog.Info("Need to allocate secondary IPs for VNIC")
		ipAllocations := make([]IPAllocation, additionalIpsByVnic[outerIndex].ips)
		workqueue.ParallelizeUntil(ctx, 8, additionalIpsByVnic[outerIndex].ips, func(innerIndex int) {
			startTime := time.Now()
			_, err := r.OCIClient.Networking().CreatePrivateIp(ctx, additionalIpsByVnic[outerIndex].vnicId)
			if err != nil {
				parallelLog.Error(err, "failed to create private-ip")
			}
			mutex.Lock()
			ipAllocations[innerIndex].err = err
			ipAllocations[innerIndex].timeTaken = float64(time.Since(startTime).Seconds())
			mutex.Unlock()
		})
		err = validateVnicIpAllocation(ipAllocations)
		mutex.Lock()
		vnicAdditionalIpAllocations[outerIndex] = VnicIPAllocationResponse{additionalIpsByVnic[outerIndex].vnicId, err, ipAllocations}
		mutex.Unlock()
	})
	for _, ips := range vnicAdditionalIpAllocations {
		if ips.err != nil {
			r.handleError(ctx, req, ips.err, "CreatePrivateIP")
			r.PushMetric(IPAllocationSlice(ips.ipAllocations).ErrorMetric())
			return ctrl.Result{}, ips.err
		}
		r.PushMetric(IPAllocationSlice(ips.ipAllocations).ErrorMetric())
	}

	_, existingSecondaryVNICs, err = r.getPrimaryAndSecondaryVNICs(ctx, *instance.CompartmentId, *instance.Id)
	if err != nil {
		r.handleError(ctx, req, err, "GetVNIC")
		return ctrl.Result{}, err
	}
	log.WithValues("existingSecondaryVNICs", existingSecondaryVNICs).
		WithValues("countOfExistingSecondaryVNICs", len(existingSecondaryVNICs)).
		Info(FetchedExistingSecondaryVNICsForInstance)
	existingSecondaryIpsbyVNIC, err = r.getSecondaryPrivateIpsByVNICs(ctx, existingSecondaryVNICs)
	if err != nil {
		r.handleError(ctx, req, err, "ListPrivateIP")
		return ctrl.Result{}, err
	}

	totalAllocatedSecondaryIPs = totalAllocatedSecondaryIpsForInstance(existingSecondaryIpsbyVNIC)
	log.WithValues("secondaryIpsbyVNIC", existingSecondaryIpsbyVNIC).
		WithValues("countOfExistingSecondaryIps", totalAllocatedSecondaryIPs).
		Info("Fetched existingSecondaryIp for instance")

	updateNPN := npnv1beta1.NativePodNetwork{}
	err = r.Get(context.TODO(), req.NamespacedName, &updateNPN)
	if err != nil {
		log.Error(err, "failed to get CR")
		r.handleError(ctx, req, err, "GetCR")
		return ctrl.Result{}, err
	}

	log.Info("Getting v1 Node object to set ownerref on CR")
	// Set OwnerRef on the CR and mark CR status as SUCCESS
	nodeObject, err := r.getNodeObjectInCluster(ctx, req.NamespacedName, *nodeName)
	if err != nil {
		r.handleError(ctx, req, err, "GetV1Node")
		return ctrl.Result{}, err
	}

	if err = controllerutil.SetOwnerReference(nodeObject, &updateNPN, r.Scheme); err != nil {
		log.Error(err, "failed to update owner ref on CR")
		return ctrl.Result{}, err
	}
	log.Info("Updating ownerref and CR status as COMPLETED")
	err = r.Client.Update(ctx, &updateNPN)
	if err != nil {
		log.Error(err, "failed to set ownerref on CR")
		return ctrl.Result{}, err
	}

	updateNPN.Status.State = &STATE_SUCCESS
	updateNPN.Status.Reason = &COMPLETED
	updateNPN.Status.VNICs = convertCoreVNICtoNPNStatus(existingSecondaryVNICs, existingSecondaryIpsbyVNIC)
	err = r.Status().Update(ctx, &updateNPN)
	if err != nil {
		log.Error(err, "failed to set status on CR")
		return ctrl.Result{}, err
	}
	r.PushMetric(endToEndLatencySlice{{time.Since(startTime).Seconds()}}.ErrorMetric())
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NativePodNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&npnv1beta1.NativePodNetwork{}).
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
		vNIC, err := r.OCIClient.Networking().GetVNIC(ctx, *vnicAttachment.VnicId)
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
		subnet, err := r.OCIClient.Networking().GetSubnet(ctx, *vNIC.SubnetId)
		if err != nil {
			log.Error(err, "failed to get subnet for VNIC")
			return nil, nil, err
		}
		existingSecondaryVNICAttachments = append(existingSecondaryVNICAttachments, SubnetVnic{vNIC, subnet})
	}
	return
}

// get the list of secondary private ips allocated on the given VNIC
func (r *NativePodNetworkReconciler) getSecondaryPrivateIpsByVNICs(ctx context.Context, existingSecondaryVNICs []SubnetVnic) (map[string][]core.PrivateIp, error) {
	privateIPsbyVNICs := make(map[string][]core.PrivateIp)
	log := log.FromContext(ctx)
	for _, secondary := range existingSecondaryVNICs {
		log := log.WithValues("vnicId", *secondary.Vnic.Id)
		privateIps, err := r.OCIClient.Networking().ListPrivateIps(ctx, *secondary.Vnic.Id)
		if err != nil {
			log.Error(err, "failed to list secondary IPs for VNIC")
			return nil, err
		}
		privateIps = filterPrivateIp(privateIps)
		privateIPsbyVNICs[*secondary.Vnic.Id] = privateIps
	}
	return privateIPsbyVNICs, nil
}

// util method to handle logging when thre is an error and updating the NPN status appropriately
func (r *NativePodNetworkReconciler) handleError(ctx context.Context, req ctrl.Request, err error, operation string) {
	log := log.FromContext(ctx).WithValues("name", req.Name)

	log.Error(err, "received error for operation", "parsedError", util.GetError(err))
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

// exclude the primary IPs in the list of private IPs on VNIC
func filterPrivateIp(privateIps []core.PrivateIp) []core.PrivateIp {
	secondaryIps := []core.PrivateIp{}
	for _, ip := range privateIps {
		// ignore primary IP
		if *ip.IsPrimary {
			continue
		}
		secondaryIps = append(secondaryIps, ip)
	}
	return secondaryIps
}

// compute the total number of allocated secondary ips on secondary vnics for this compute instance
func totalAllocatedSecondaryIpsForInstance(vnicToIpMap map[string][]core.PrivateIp) int {
	totalSecondaryIps := 0
	for _, Ips := range vnicToIpMap {
		totalSecondaryIps += len(Ips)
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
func getAdditionalSecondaryIPsNeededPerVNIC(existingIpsByVnic map[string][]core.PrivateIp, additionalSecondaryIps int) ([]VnicIPAllocations, error) {
	requiredAdditionalSecondaryIps := additionalSecondaryIps
	additionalIpsByVnic := make([]VnicIPAllocations, 0)
	for vnic, existingIps := range existingIpsByVnic {
		// VNIC already has max secondary IPs
		if len(existingIps) == maxSecondaryPrivateIPsPerVNIC {
			additionalIpsByVnic = append(additionalIpsByVnic, VnicIPAllocations{vnic, 0})
			continue
		}
		allocatableIps := maxSecondaryPrivateIPsPerVNIC - len(existingIps)
		if allocatableIps > requiredAdditionalSecondaryIps {
			additionalIpsByVnic = append(additionalIpsByVnic, VnicIPAllocations{vnic, requiredAdditionalSecondaryIps})
			requiredAdditionalSecondaryIps -= requiredAdditionalSecondaryIps
			continue
		}
		additionalIpsByVnic = append(additionalIpsByVnic, VnicIPAllocations{vnic, allocatableIps})
		requiredAdditionalSecondaryIps -= allocatableIps
	}
	if requiredAdditionalSecondaryIps > 0 {
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
func convertCoreVNICtoNPNStatus(existingSecondaryVNICs []SubnetVnic, existingSecondaryIpsbyVNIC map[string][]core.PrivateIp) []npnv1beta1.VNICAddress {
	npnVNICAddress := make([]npnv1beta1.VNICAddress, 0, len(existingSecondaryIpsbyVNIC))
	for _, vnic := range existingSecondaryVNICs {
		vnicIps := make([]*string, 0, len(existingSecondaryIpsbyVNIC[*vnic.Vnic.Id]))
		for _, ip := range existingSecondaryIpsbyVNIC[*vnic.Vnic.Id] {
			vnicIps = append(vnicIps, ip.IpAddress)
		}
		npnVNICAddress = append(npnVNICAddress, npnv1beta1.VNICAddress{
			VNICID:     vnic.Vnic.Id,
			MACAddress: vnic.Vnic.MacAddress,
			RouterIP:   vnic.Subnet.VirtualRouterIp,
			Addresses:  vnicIps,
			SubnetCidr: vnic.Subnet.CidrBlock,
		})
	}
	return npnVNICAddress
}

// wait for the Kubernetes object to be created in the cluster so that the owner reference of the NPN CR
// can be set to the Node object
func (r NativePodNetworkReconciler) getNodeObjectInCluster(ctx context.Context, cr types.NamespacedName, nodeName string) (*v1.Node, error) {
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

// wait for the compute instance to move to running state
func (r NativePodNetworkReconciler) waitForInstanceToReachRunningState(ctx context.Context, npn npnv1beta1.NativePodNetwork) error {
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
