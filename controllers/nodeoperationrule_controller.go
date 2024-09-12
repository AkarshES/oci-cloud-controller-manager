/*
Copyright 2024.

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
	"github.com/go-logr/logr"
	norv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/node-cycling/v1beta1"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"maps"
	"math"
	"runtime/debug"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// NOR Validation error and event message
	// error msg + event msg
	errMatchOkeLabelFormat = errors.New("MatchOKELabel in NodeSelector in the Spec of Node Operation Rule is invalid")
	// TODO: do we want to enable some help link in the event
	// event msg
	errMatchCustomLabelsFormat = errors.New("MatchCustomLabels in NodeSelector in the Spec of NodeOperationRule is invalid")
	// error msg + event msg
	errMatchOkeLabelKey = errors.New("MatchOKELabel in NodeSelector in the Spec of NodeOperationRule does not contain required key")
	// event msg
	errNodeSelectorFormat  = errors.New("NodeSelector in the Spec of NodeOperationRule is invalid")
	errActionsFormat       = errors.New("actions in the Spec of NodeOperationRule is invalid")
	errBvrRateLimited      = errors.New("rate limited for operation replace boot volume on cluster")
	errRebootRateLimited   = errors.New("rate limited for operation reboot on cluster")
	errNonRetryableError   = errors.New("nonRetryable Error shows up")
	errUnexpectedNodeScope = errors.New("there are no nodes with labels retrieved from Kube API server but there are still some nodes need operation")
	errFailedToCancelWR    = errors.New("fail to cancel work request")
	errUnexpected          = errors.New("unexpected situation happens")
	errStaleNor            = errors.New(staleNorMsg)
)

// OKE reserved resource constants related
const (
	okeReservedLabelKey string = "oke.oraclecloud.com/node_operation"
	finalizer           string = "nodeoperationrule.oci.oraclecloud.com/finalizers"
)

// Garbage Collection related
const (
	maxSuccededNodesBeforeGarbageCollection = 1000
)

// event record related
const (
	eventReasonValidatedNOR           string = "ValidatedNOR"
	eventReasonRemovedLabel           string = "RemovedOKEReservedLabel"
	eventMsgRemoveLabel               string = "Removed OKE Reserved Label on node: "
	eventReasonStartedNodeOperation   string = "StartedNodeOperation"
	eventMsgStartedNodeOperation      string = "Started node operation on node with work request ID: "
	eventReasonCompletedNodeOperation string = "CompletedNodeOperation"
	eventMsgCompletedNodeOperation    string = "Completed operation on node: "
	eventReasonFailedNodeOperation    string = "FailedNodeOperation"
	eventMsgFailedNodeOperation       string = "Failed with operation on node with work request ID: "
	eventReasonCancelledNodeOperation string = "CancelledNodeOperation"
	eventMsgCancelledNodeOperation    string = "Cancelled operation on node: "
	eventMsgGarbageCollection         string = " nodes garbage collected in the succeededNodes section"
	evenReasonGarbageCollection       string = "GarbageCollectNOR"
	eventReasonCheckPolicy            string = "ValidatedPolicy"
	eventMsgCheckPolicy               string = "Please validate the policy configuration of cluster and instance"
)

// logging related
const (
	norName            string = "norName"
	norInstanceId      string = "norInstanceId"
	norMergedLabel     string = "norMergedLabels"
	norWorkRequestId   string = "norWorkRequestId"
	norWorkRequest     string = "norWorkRequest"
	norOpcId           string = "norOpcId"
	norNodeName        string = "norNodeName"
	norClusterId       string = "norClusterId"
	norInProgressNodes string = "norInProgressNodes"
	norPendingNodes    string = "norPendingNodes"
	norBackOffNodes    string = "norBackOffNodes"
	norHttpStatusCode  string = "norHttpStatusCode"
	norErrorMessage    string = "norErrorMessage"
	norErrorCode       string = "norErrorCode"
)

// API related
const (
	rebootAction            string = "reboot"
	replaceBootVolumeAction string = "replaceBootVolume"
	// the error message is retrieved from CP
	errOKEAPIDisabledMsg string = "Node Action is not enabled"
)

// error related
const (
	unknownErrorMsg     string = "unknownError"
	rateLimitedErrorMsg string = "rateLimitedError"
	staleNorMsg         string = "stale version of nor"
	blankNodeId         string = "Invalid nodeId"
	invalidNodeIdFormat string = "Invalid node Id"
	blankClusterId      string = "Invalid clusterId"
	wrongNodeMember            = "The node does not belong to the specified cluster"
)

const (
	RequeueDurationForUpdateWithoutError = 60 * time.Second
)

var globalResourceVersionMap = make(map[types.NamespacedName]string)

// https://bitbucket.oci.oraclecorp.com/projects/OKE/repos/clusters-api-spec/pull-requests/454/diff#shepherd%2Fgenerated-src%2Fpreview%2Fclusters-api-spec.yaml
const (
	HTTP400BadRequestCode              = "BadRequest"
	HTTP400InvalidParameterCode        = "InvalidParameter"
	HTTP401NotAuthenticatedCode        = "NotAuthenticated"
	HTTP404NotAuthorizedOrNotFoundCode = "NotAuthorizedOrNotFound"
	HTTP409IncorrectStateCode          = "IncorrectState"
	HTTP409ConflictCode                = "Conflict"
	HTTP412NoEtagMatchCode             = "NoEtagMatch"
	HTTP429TooManyRequestsCode         = "TooManyRequests"
	HTTP500TooInternalServerErrorCode  = "InternalServerError"
)

type NodeOperationFailureType string

const (
	NodeOperationFailureTypeRetryable    NodeOperationFailureType = "retryableError"
	NodeOperationFailureTypeNonRetryable NodeOperationFailureType = "nonRetryableError"
	NodeOperationFailureTypeRateLimited  NodeOperationFailureType = "rateLimitedError"
)

type NodeOperationFailure struct {
	nodeName   string
	errorType  NodeOperationFailureType
	errorCode  string //ServiceError.getCode
	errorMsg   string
	trackingId string // opc request Id / work request id / ""
}

type NodeOperationResultUpdate struct {
	NodeName      string
	WorkRequestId string
	update        UpdateType
	timeFinished  metav1.Time
}

type UpdateType string

const (
	UpdateTypeInProgress   UpdateType = "inProgress"
	UpdateTypeRetryable    UpdateType = "retryable"
	UpdateTypeNonRetryable UpdateType = "nonRetryable"
	UpdateTypeSucceeded    UpdateType = "succeeded"
	UpdateTypeCanceled     UpdateType = "canceled"
)

// metrics related
const (
	// metrics name
	baseMetricName = "NOR"
	// customer usage metrics related
	operation      = "OPERATION"
	useCustomLabel = "USE_CUSTOM_LABEL"
	// performance related
	schedulingLatency  = "SCHEDULE_LATENCY_IN_SECONDS"
	operationLatency   = "OPERATION_LATENCY_IN_SECONDS"
	operationFailure   = "OPERATION_FAILURE"
	removeLabelFailure = "LABEL_REMOVAL_FAILURE"
	gcTriggered        = "GC_TRIGGERED"
	// health
	panic     = "PANIC"
	invalidId = "INVALIDID"

	// dimension
	tenancy         = "tenancyId"
	action          = "action"
	cluster         = "clusterId"
	failureCategory = "category"
	finalStatus     = "status"
	failure         = "failed"
	success         = "succeeded"
)

var schedulingTimer sync.Map
var operationTimer sync.Map

// NodeOperationRuleReconciler reconciles a NodeOperationRule object
type NodeOperationRuleReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	KubeClient        clientset.Interface
	OCIClient         ociclient.Interface
	Config            *providercfg.Config
	MetricPusher      *metrics.MetricPusher
	Recorder          record.EventRecorder
	nodeList          v1.NodeList
	BvrRateLimiter    *rate.Limiter
	RebootRateLimiter *rate.Limiter
}

// +kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nodeoperationrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nodeoperationrules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nodeoperationrules/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=nodes/status,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the NodesOperationRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
// TODO: need to think backoff retry for all the cases
func (r *NodeOperationRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	log.Info("nor reconciliation in progress")

	defer func() {
		if e := recover(); e != nil {
			log.Error(fmt.Errorf("error is %v. stack is %s", e, string(debug.Stack())), "Panic occurred in reconciling nor")
			r.sendMetricsWithClusterId(panic, 1, make(map[string]string))
		}
	}()

	var nor norv1beta1.NodeOperationRule
	if err := r.Get(ctx, req.NamespacedName, &nor); err != nil {
		log.Error(err, "unable to fetch nor")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues(norName, nor.Name)
	log.Info(fmt.Sprint("Generation: ", nor.Generation, " Resource Version: ", nor.ResourceVersion,
		" original status: in progress: ", nor.Status.InProgressNodes, " pending: ", nor.Status.PendingNodes, " backoff: ", nor.Status.BackOffNodes, " canceled: ", nor.Status.CanceledNodes, " succeeded: ", nor.Status.SucceededNodes))

	latestVersion := getLatestResourceVersion(req.NamespacedName, nor.ResourceVersion)
	log.Info(fmt.Sprint("latest retrieved version: ", latestVersion))
	if isNorStale(nor.ResourceVersion, latestVersion) {
		return ctrl.Result{}, errStaleNor
	}

	// 1. handle finalizer
	// 1.1 add finalizer for the newly created NOR CR
	if nor.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(&nor, finalizer) {
		// if the deletionTime is not set, and finalizer is not existing, add it
		// edge case: keep failing add finalizer and customer delete the nor, it does not matter as no node operation is reconciled at all
		err := r.addFinalizer(ctx, nor)
		if err != nil {
			return ctrl.Result{}, err
		} else {
			return ctrl.Result{RequeueAfter: RequeueDurationForUpdateWithoutError}, nil
		}
	}

	// nodes that have in-flight node operation
	updatedInProgressNodes := initializeListInStatus(nor.Status.InProgressNodes)
	// nodes are pending triggered with node operation due to maxParallelism or limit rate throttling
	updatedPendingNodes := initializePendingList(nor.Status.PendingNodes)
	// nodes have finished operation but failed and need to retry
	updatedRetryableNodes := initializeListInStatus(nor.Status.BackOffNodes)
	// nodes that operations are succeeded
	updatedSucceededNodes := initializeSuccessListInStatus(nor.Status.SucceededNodes)
	// nodes that operations are canceled
	updatedCanceledNodes := initializeListInStatus(nor.Status.CanceledNodes)

	log.Info(fmt.Sprint("Generation: ", nor.Generation, " Resource Version: ", nor.ResourceVersion,
		" status after initialization: in progress: ", updatedInProgressNodes, " pending: ", updatedPendingNodes, " backoff: ", updatedRetryableNodes, " canceled: ", updatedCanceledNodes, " succeeded: ", updatedSucceededNodes))

	// 1.2 if nor is triggered with deletion, trigger work request cancellation and remove finalizer
	if !nor.DeletionTimestamp.IsZero() {
		updatedNorStatus := norv1beta1.NodeOperationRuleStatus{
			InProgressNodes: updatedInProgressNodes,
			PendingNodes:    updatedPendingNodes,
			BackOffNodes:    updatedRetryableNodes,
			SucceededNodes:  updatedSucceededNodes,
			CanceledNodes:   updatedCanceledNodes,
		}

		log.Info("detected DeletionTimestamp populated on nor")
		// if the deletionTime is set, start clean-up process: trigger cancellation and remove finalizer
		// 1.2.1 cancellation process
		updatedStatus, err := r.cancelOperations(ctx, nor, updatedNorStatus)
		if err != nil {
			log.Error(err, "fail to cancel in-progress work request")
			return ctrl.Result{}, err
		}
		// 1.2.2 remove finalizer
		nor.Status = updatedStatus
		controllerutil.RemoveFinalizer(&nor, finalizer)
		if err := r.Update(ctx, &nor); err != nil {
			log.Error(err, "fail to remove finalizer and update status of nor "+nor.Name)
			return ctrl.Result{}, err
		}
		log.Info("nor finalizer cleanup finishes")
		deleteEntryFromGlobalResourceVersion(req.NamespacedName)
		return ctrl.Result{RequeueAfter: RequeueDurationForUpdateWithoutError}, nil
	}

	// 2. validate NOR
	overallLabelsSelector, err := r.validateNOR(ctx, nor)
	if err != nil {
		return ctrl.Result{}, err
	}

	// list all the qualified nodes that are matched with the labels defined in the nor
	nodesWithLabel := r.filterMatchedNodes(ctx, overallLabelsSelector)

	log.Info(fmt.Sprint("retrieve nodes with qualified labels and its size is ", len(nodesWithLabel)))

	// 3. check whether nodes need to be canceled, which is invoked by label being removed from node
	// trigger cancelWR and in progress -> canceled
	// pending nodes -> canceled nodes
	// retryable nodes -> canceled nodes
	nodesWithLabelMap := convertV1NodesToMap(nodesWithLabel)
	inProgressNodesMap := convertNodeOperationResultsToMap(updatedInProgressNodes)
	isInProgressNodesSubset, toBeCancelled := isSubset(inProgressNodesMap, nodesWithLabelMap)
	if !isInProgressNodesSubset {
		// cancel WR //TODO: whether need to parallel cancel
		// remove from in progress nodes list
		// add it to cancelled nodes list
		for _, nodeName := range toBeCancelled {
			err = r.cancelWorkRequest(ctx, inProgressNodesMap[nodeName].WorkRequestId)
			removeLatency(operationTimer, nor.Name, nodeName)
			if err != nil {
				log.Error(err, "fail to cancel work request", norWorkRequestId, inProgressNodesMap[nodeName].WorkRequestId, norNodeName, nodeName)
			} else {
				updatedInProgressNodes = removeResultFromList(inProgressNodesMap[nodeName].NodeName, updatedInProgressNodes)
				updatedCanceledNodes = updateResultToList(inProgressNodesMap[nodeName], updatedCanceledNodes)
				r.Recorder.Event(&nor, v1.EventTypeNormal, eventReasonCancelledNodeOperation, fmt.Sprint(eventMsgCancelledNodeOperation, inProgressNodesMap[nodeName].NodeName))
			}
		}
		log.Info("canceled work request, will update nor status")
		return r.updateStatus(ctx, req.NamespacedName, updatedInProgressNodes, updatedPendingNodes, updatedRetryableNodes, updatedSucceededNodes, updatedCanceledNodes)
	}
	isPendingSubset, toBeCancelled := isNodeNameSubset(updatedPendingNodes, nodesWithLabelMap)
	if !isPendingSubset {
		// remove from pending nodes list
		// add it to cancelled nodes list
		for _, nodeName := range toBeCancelled {
			removeLatency(schedulingTimer, nor.Name, nodeName)
			updatedPendingNodes = removeNodeFromList(nodeName, updatedPendingNodes)
			updatedCanceledNodes = updateResultToList(norv1beta1.NodeOperationResult{
				NodeName:      nodeName,
				WorkRequestId: "",
			}, updatedCanceledNodes)
		}
		log.Info("moved pending nodes to canceled nodes, will update nor status")
		return r.updateStatus(ctx, req.NamespacedName, updatedInProgressNodes, updatedPendingNodes, updatedRetryableNodes, updatedSucceededNodes, updatedCanceledNodes)
	}

	retryableNodesMap := convertNodeOperationResultsToMap(updatedRetryableNodes)
	isRetryableNodesSubset, toBeCancelled := isSubset(retryableNodesMap, nodesWithLabelMap)
	if !isRetryableNodesSubset {
		// remove from retryable nodes list
		// add it to cancelled nodes list
		for _, nodeName := range toBeCancelled {
			removeLatency(schedulingTimer, nor.Name, nodeName)
			removeLatency(operationTimer, nor.Name, nodeName)
			updatedRetryableNodes = removeResultFromList(retryableNodesMap[nodeName].NodeName, updatedRetryableNodes)
			updatedCanceledNodes = updateResultToList(norv1beta1.NodeOperationResult{
				NodeName:      nodeName,
				WorkRequestId: "",
			}, updatedCanceledNodes)
		}
		log.Info("moved retryable nodes to canceled nodes, will update nor status")
		return r.updateStatus(ctx, req.NamespacedName, updatedInProgressNodes, updatedPendingNodes, updatedRetryableNodes, updatedSucceededNodes, updatedCanceledNodes)
	}

	// 4. poll the latest status of in-progress operation
	// if the in-progress WRs finish successfully or finish cancellation, will update status immediately
	// in order to avoid re-triggering on successful nodes or cancelled nodes
	hasInProgressNodesSucceededOrCanceled := false
	if updatedInProgressNodes != nil && len(updatedInProgressNodes) != 0 {
		updatedInProgressNodesCopy := make([]norv1beta1.NodeOperationResult, len(updatedInProgressNodes))
		copy(updatedInProgressNodesCopy, updatedInProgressNodes)
		for _, inProgressNode := range updatedInProgressNodesCopy {
			nodeOperationUpdate := r.getWorkRequest(ctx, nor, inProgressNode.WorkRequestId, inProgressNode.NodeName)
			switch nodeOperationUpdate.update {
			case UpdateTypeInProgress:
				updatedInProgressNodes = updateResultToList(norv1beta1.NodeOperationResult{
					NodeName:      nodeOperationUpdate.NodeName,
					WorkRequestId: nodeOperationUpdate.WorkRequestId,
				}, updatedInProgressNodes)
				log.Info("work request is in progress", norNodeName, nodeOperationUpdate.NodeName, norWorkRequestId, nodeOperationUpdate.WorkRequestId)
				break
			case UpdateTypeNonRetryable: // TODO: how to deal with non-retryable case
				updatedRetryableNodes = updateResultToList(norv1beta1.NodeOperationResult{
					NodeName:      nodeOperationUpdate.NodeName,
					WorkRequestId: nodeOperationUpdate.WorkRequestId,
				}, updatedRetryableNodes)
				updatedInProgressNodes = removeResultFromList(nodeOperationUpdate.NodeName, updatedInProgressNodes)
				log.Error(errNonRetryableError, fmt.Sprint(errNonRetryableError.Error(), " will add to backOff nodes"), norNodeName, nodeOperationUpdate.NodeName, norWorkRequestId, nodeOperationUpdate.WorkRequestId)
				r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonFailedNodeOperation, fmt.Sprint(eventMsgFailedNodeOperation, inProgressNode.NodeName, ": ", nodeOperationUpdate.WorkRequestId))
				latency := calculateLatency(operationTimer, nor.Name, inProgressNode.NodeName)
				if latency != 0 {
					nodeDimension := map[string]string{norNodeName: inProgressNode.NodeName, finalStatus: failure}
					r.sendMetricsWithClusterId(operationLatency, latency, nodeDimension)
				}
				break
			case UpdateTypeRetryable:
				updatedRetryableNodes = updateResultToList(norv1beta1.NodeOperationResult{
					NodeName:      nodeOperationUpdate.NodeName,
					WorkRequestId: nodeOperationUpdate.WorkRequestId,
				}, updatedRetryableNodes)
				updatedInProgressNodes = removeResultFromList(nodeOperationUpdate.NodeName, updatedInProgressNodes)
				log.Info("work request is failed. will add to backoff nodes", norNodeName, nodeOperationUpdate.NodeName, norWorkRequestId, nodeOperationUpdate.WorkRequestId)
				r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonFailedNodeOperation, fmt.Sprint(eventMsgFailedNodeOperation, inProgressNode.NodeName, ": ", nodeOperationUpdate.WorkRequestId))
				latency := calculateLatency(operationTimer, nor.Name, inProgressNode.NodeName)
				if latency != 0 {
					nodeDimension := map[string]string{norNodeName: inProgressNode.NodeName, finalStatus: failure}
					r.sendMetricsWithClusterId(operationLatency, latency, nodeDimension)
				}
				break
			case UpdateTypeSucceeded:
				err := r.removeOkeReservedLabel(log, inProgressNode.NodeName)
				if err != nil {
					// only print out error, and will update status to trigger a new round of reconcile so that retry remove label
					log.Error(err, "fail to remove OKE Reserved Label", norNodeName, inProgressNode.NodeName, norWorkRequestId, nodeOperationUpdate.WorkRequestId)
				} else {
					updatedSucceededNodes = append(updatedSucceededNodes,
						norv1beta1.NodeOperationSuccess{
							NodeName:         nodeOperationUpdate.NodeName,
							SuccessTimestamp: nodeOperationUpdate.timeFinished,
						})
					updatedInProgressNodes = removeResultFromList(nodeOperationUpdate.NodeName, updatedInProgressNodes)
					log.Info("work request is succeeded", norWorkRequestId, nodeOperationUpdate.WorkRequestId)
					r.Recorder.Event(&nor, v1.EventTypeNormal, eventReasonCompletedNodeOperation, fmt.Sprint(eventMsgCompletedNodeOperation, inProgressNode.NodeName))
					latency := calculateLatency(operationTimer, nor.Name, inProgressNode.NodeName)
					if latency != 0 {
						nodeDimension := map[string]string{norNodeName: inProgressNode.NodeName, finalStatus: success}
						r.sendMetricsWithClusterId(operationLatency, latency, nodeDimension)
					}
				}
				hasInProgressNodesSucceededOrCanceled = true
				break
			case UpdateTypeCanceled:
				updatedCanceledNodes = updateResultToList(norv1beta1.NodeOperationResult{
					NodeName:      nodeOperationUpdate.NodeName,
					WorkRequestId: nodeOperationUpdate.WorkRequestId,
				}, updatedCanceledNodes)
				updatedInProgressNodes = removeResultFromList(nodeOperationUpdate.NodeName, updatedInProgressNodes)
				hasInProgressNodesSucceededOrCanceled = true
				log.Info("work request is canceling or canceled", norNodeName, nodeOperationUpdate.NodeName, norWorkRequestId, nodeOperationUpdate.WorkRequestId)
				r.Recorder.Event(&nor, v1.EventTypeNormal, eventReasonCancelledNodeOperation, fmt.Sprint(eventMsgCancelledNodeOperation, inProgressNode.NodeName))
				removeLatency(operationTimer, nor.Name, nodeOperationUpdate.NodeName)
				break
			}
		}
		log.Info(fmt.Sprint("current status: in progress: ", updatedInProgressNodes, " pending: ", updatedPendingNodes, " backoff: ", updatedRetryableNodes, " canceled: ", updatedCanceledNodes, " succeeded: ", updatedSucceededNodes))
		if hasInProgressNodesSucceededOrCanceled {
			log.Info("there are succeeded or canceled work requests, will update nor status")

			//Perform Garbage collection if the number of SucceededNodes > 1000
			// Rate limits is 10 per minute. So no way we would have removed succeeded nodes that are just introduced as part of this cycle
			//Sort SucceededNodes and remove the ones that are the oldest in the queue by SuccessTimestamp
			updatedSucceededNodes, numberOfNodesGarbageCollected := sortSucceededNodesAndGarbageCollect(updatedSucceededNodes)

			if numberOfNodesGarbageCollected > 0 {
				log.Info(fmt.Sprint("Garbage collection kick off. ", nor.Name, " is calculated as ", strconv.Itoa(numberOfNodesGarbageCollected)))
				r.sendMetricsWithClusterId(gcTriggered, 1, make(map[string]string))
				r.Recorder.Event(&nor, v1.EventTypeNormal, evenReasonGarbageCollection, strconv.Itoa(numberOfNodesGarbageCollected)+eventMsgGarbageCollection)
			}

			return r.updateStatus(ctx, req.NamespacedName, updatedInProgressNodes, updatedPendingNodes, updatedRetryableNodes, updatedSucceededNodes, updatedCanceledNodes)
		}
	}
	// trigger objects: newly found per evaluation, pending nodes, retryable nodes
	// order: newly found -> pending -> retryable
	candidates, err := getAndSortCandidates(ctx, nor.Name, nodesWithLabel, updatedInProgressNodes, updatedPendingNodes, updatedRetryableNodes)
	if err != nil {
		return ctrl.Result{}, err
	}

	// calculate how many requests could be sent in parallel in this round
	parallelism := calculateParallelism(nor.Spec.MaxParallelism-len(updatedInProgressNodes), candidates)
	log.Info(fmt.Sprint("the parallelism of nor ", nor.Name, " is calculated as ", strconv.Itoa(parallelism)))

	// not all the node candidates could be triggered with operation in the current round due to parallelism
	if parallelism < len(candidates) {
		leftOverCandidates := candidates[len(candidates)-parallelism:]
		updatedPendingNodes = addToListIfNotExist(getNodeNames(leftOverCandidates), updatedPendingNodes)
	}

	for index := 0; index < parallelism; index++ {
		node := candidates[index]
		// remove the node from pending nodes or backoff nodes if the node is contained in these 2 lists
		updatedPendingNodes = removeNodeFromList(node.Name, updatedPendingNodes)
		updatedRetryableNodes = removeResultFromList(node.Name, updatedRetryableNodes)

		workRequestId, nodeOperationFailure := r.triggerNodeAction(ctx, node, nor)
		r.sendMetricsWithTenancy(operation, 1, map[string]string{action: string(nor.Spec.Actions[0])})
		var wrId string
		if len(workRequestId) != 0 {
			// node action API request has been accepted, work request ID is generated
			wrId = workRequestId
			updatedInProgressNodes = updateResultToList(norv1beta1.NodeOperationResult{
				NodeName:      node.Name,
				WorkRequestId: wrId,
			}, updatedInProgressNodes)
			log.Info("node action is triggered", norWorkRequestId, wrId, norNodeName, node.Name)
			r.Recorder.Event(&nor, v1.EventTypeNormal, eventReasonStartedNodeOperation, fmt.Sprint(eventMsgStartedNodeOperation, node.Name, ": ", wrId))
			latency := calculateLatency(schedulingTimer, nor.Name, node.Name)
			if latency != 0 {
				nodeDimension := map[string]string{norNodeName: node.Name}
				r.sendMetricsWithClusterId(schedulingLatency, latency, nodeDimension)
			}
			operationTimer.LoadOrStore(getNorNameNodeName(nor.Name, node.Name), time.Now())
		} else {
			// node action API request is not accepted
			if len(nodeOperationFailure.trackingId) != 0 {
				// opc ID could be in-place
				wrId = nodeOperationFailure.trackingId
			} else {
				wrId = ""
			}
			switch nodeOperationFailure.errorType {
			case NodeOperationFailureTypeRetryable:
				updatedRetryableNodes = updateResultToList(norv1beta1.NodeOperationResult{
					NodeName:      node.Name,
					WorkRequestId: wrId,
				}, updatedRetryableNodes)
				log.Info("node action is failed", norOpcId, wrId, norNodeName, node.Name)
				break
			case NodeOperationFailureTypeNonRetryable:
				log.Error(errNonRetryableError, errNonRetryableError.Error(), norOpcId, wrId, norNodeName, node.Name)
				break
			case NodeOperationFailureTypeRateLimited:
				updatedPendingNodes = addToListIfNotExist([]string{node.Name}, updatedPendingNodes)
				log.Info("node operation is pending due to rate limit")
				break
			default:
				updatedRetryableNodes = updateResultToList(norv1beta1.NodeOperationResult{
					NodeName:      node.Name,
					WorkRequestId: wrId,
				}, updatedRetryableNodes)
				log.Error(errUnexpected, errUnexpected.Error(), norOpcId, wrId, norNodeName, node.Name)
				break
			}
		}
		log.Info("triggered node operation, will update nor status")
		result, err := r.updateStatus(ctx, req.NamespacedName, updatedInProgressNodes, updatedPendingNodes, updatedRetryableNodes, updatedSucceededNodes, updatedCanceledNodes)
		// if err pops during update status, need to return the result and restart reconcile
		// if no error, continue with triggering next node action
		if err != nil {
			return result, err
		}
		if index == parallelism-1 {
			// after the last one is triggered with API, return reconcile result
			return result, err
		}
	}
	log.Info("finished the reconciliation logic, will update nor status")
	return r.updateStatus(ctx, req.NamespacedName, updatedInProgressNodes, updatedPendingNodes, updatedRetryableNodes, updatedSucceededNodes, updatedCanceledNodes)
}

// bvrNode initiates a cycling operation for a specified node within a cluster.
// It takes the node ID, cluster ID, and a NodeOperationRule object as input.
// The function returns the work request ID associated with the cycling operation.
//
// Parameters:
// - nodeId: A string representing the unique identifier of the node to be cycled.
// - clusterId: A string representing the unique identifier of the cluster where the node resides.
// - nor: An instance of norv1beta1.NodeOperationRule containing additional details for the cycling operation.
//
// Returns:
// - A string representing the work request ID associated with the cycling operation.
// - An error indicating any issues encountered during the cycling operation; otherwise, returns nil.
func (r *NodeOperationRuleReconciler) bvrNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) (string, error) {
	logger := log.FromContext(ctx, norInstanceId, nodeId, norClusterId, clusterId)
	var workRequestId string
	var err error
	if r.BvrRateLimiter.Allow() {
		workRequestId, err = r.OCIClient.ContainerEngine().ReplaceBootVolumeClusterNode(ctx, nodeId, clusterId, nor)
	} else {
		logger.Error(errBvrRateLimited, errBvrRateLimited.Error(), norClusterId, clusterId, norInstanceId, nodeId)
		return "", errBvrRateLimited
	}

	return workRequestId, err
}

// rebootNode initiates a reboot operation for a specified node within a cluster.
// It takes the node ID, cluster ID, and a NodeOperationRule object as input.
// The function returns the work request ID associated with the reboot operation.
//
// Parameters:
// - nodeId: A string representing the unique identifier of the node to be rebooted.
// - clusterId: A string representing the unique identifier of the cluster where the node resides.
// - nor: An instance of norv1beta1.NodeOperationRule containing additional details for the reboot operation.
//
// Returns:
// - A string representing the work request ID associated with the reboot operation.
// - An error indicating any issues encountered during the reboot operation; otherwise, returns nil.
func (r *NodeOperationRuleReconciler) rebootNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) (string, error) {
	logger := log.FromContext(ctx, norInstanceId, nodeId, norClusterId, clusterId)
	var workRequestId string
	var err error
	if r.RebootRateLimiter.Allow() {
		workRequestId, err = r.OCIClient.ContainerEngine().RebootClusterNode(ctx, nodeId, clusterId, nor)
	} else {
		logger.Error(errRebootRateLimited, errRebootRateLimited.Error(), norClusterId, clusterId, norInstanceId, nodeId)
		return "", errRebootRateLimited
	}
	return workRequestId, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeOperationRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := zap.L().Sugar()
	log.Info("Setting up NOR controller with manager")
	r.Recorder = mgr.GetEventRecorderFor("NodeOperationRule")
	return ctrl.NewControllerManagedBy(mgr).
		For(&norv1beta1.NodeOperationRule{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 20, CacheSyncTimeout: time.Hour}).
		WatchesRawSource(source.Kind(
			mgr.GetCache(),
			&v1.Node{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, node *v1.Node) []reconcile.Request {

				log.With(norNodeName, node.Name).Info("detect node has label changed")

				// List all nor CRs
				var norList norv1beta1.NodeOperationRuleList
				if err := mgr.GetCache().List(ctx, &norList); err != nil {
					// Handle error appropriately
					log.With(zap.Error(err)).Error("fail to list nor list")
					return nil
				}

				// List all nodes
				var allNodes v1.NodeList
				if err := mgr.GetCache().List(ctx, &allNodes); err != nil {
					log.With(zap.Error(err)).Error("fail to list node list")
					return nil
				}
				r.nodeList = allNodes

				// Create a reconcile.Request for each nor CR
				requests := make([]reconcile.Request, len(norList.Items))
				for i, nor := range norList.Items {
					requests[i] = reconcile.Request{
						NamespacedName: client.ObjectKey{
							Namespace: nor.Namespace,
							Name:      nor.Name,
						},
					}
				}

				return requests
			}),
			predicate.Or[*v1.Node](predicate.TypedLabelChangedPredicate[*v1.Node]{}))).
		Complete(r)
}

// validateNorFormat is to validate NodeOperationRule CR. Even though it is supposed to be validated during CR creation via CRD, just in case validate again.
// if any error happens, will reconcile with error and add error log and record event
// TODO: Post GA: if the Actions support multiple entries in the future, both CRD and this function need modification
func (r *NodeOperationRuleReconciler) validateNorFormat(ctx context.Context, nor norv1beta1.NodeOperationRule) (ctrl.Result, error) {
	log := log.FromContext(ctx, norName, nor.Name)
	// validate MatchOKELabel
	okeRequiredLabel := nor.Spec.NodeSelector.MatchTriggerLabel
	if okeRequiredLabel == nil || len(okeRequiredLabel) != 1 {
		log.Error(errMatchOkeLabelFormat, errMatchOkeLabelFormat.Error())
		r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonValidatedNOR, errMatchOkeLabelFormat.Error())
		return ctrl.Result{}, errMatchOkeLabelFormat
	}
	if _, ok := okeRequiredLabel[okeReservedLabelKey]; !ok {
		log.Error(errMatchOkeLabelKey, errMatchOkeLabelKey.Error())
		r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonValidatedNOR, errMatchOkeLabelKey.Error())
		return ctrl.Result{}, errMatchOkeLabelKey
	}

	// validate labels
	_, err1 := labels.ValidatedSelectorFromSet(okeRequiredLabel)
	if err1 != nil {
		log.Error(err1, "MatchTriggerLabel format is invalid")
		r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonValidatedNOR, errMatchOkeLabelKey.Error())
		// if the label cannot be converted to selector, it means the label format has error, will not requeue
		return ctrl.Result{}, err1
	}
	_, err2 := labels.ValidatedSelectorFromSet(nor.Spec.NodeSelector.MatchCustomLabels)
	if err2 != nil {
		log.Error(err1, "MatchCustomLabels format is invalid")
		r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonValidatedNOR, errMatchCustomLabelsFormat.Error())
		// if the label cannot be converted to selector, it means the label format has error, will not requeue
		return ctrl.Result{}, err2
	}
	if nor.Spec.NodeSelector.MatchCustomLabels != nil {
		r.sendMetricsWithTenancy(useCustomLabel, 1, make(map[string]string))
	}
	// validate Actions
	actions := nor.Spec.Actions
	if actions == nil || len(actions) != 1 || (string(actions[0]) != replaceBootVolumeAction && string(actions[0]) != rebootAction) {
		log.Error(errActionsFormat, errActionsFormat.Error())
		r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonValidatedNOR, errActionsFormat.Error())
		return ctrl.Result{}, errActionsFormat
	}

	return ctrl.Result{}, nil
}

// calculateParallelism calculates the parallelism of triggering node operation API requests
// The calculation is based on the maxUnavailable configured in NOR and the size of nodes are qualified with operation
func calculateParallelism(currentQuota int, nodeCandidates []*v1.Node) int {
	if nodeCandidates == nil || len(nodeCandidates) == 0 {
		return 0
	}
	return int(math.Min(float64(currentQuota), float64(len(nodeCandidates))))
}

// getAndSortCandidates get the nodes which have match labels, and sort the nodes by order
// nodeCandidates - all the nodes retrieved from kube-api server and have labels. Exclude in-progress nodes
// order: newly found node -> pending nodes -> nodes need retry
func getAndSortCandidates(ctx context.Context, norName string, nodesWithLabel []*v1.Node, inProgressNodes []norv1beta1.NodeOperationResult, pendingNodes []string, retryableNodes []norv1beta1.NodeOperationResult) ([]*v1.Node, error) {
	logger := log.FromContext(ctx)
	if nodesWithLabel == nil || len(nodesWithLabel) == 0 {
		if len(inProgressNodes) == 0 && len(pendingNodes) == 0 && len(retryableNodes) == 0 {
			return make([]*v1.Node, 0), nil
		} else {
			logger.Error(errUnexpectedNodeScope, errUnexpectedNodeScope.Error(), norInProgressNodes, inProgressNodes, norPendingNodes, pendingNodes, norBackOffNodes, retryableNodes)
			return make([]*v1.Node, 0), errUnexpectedNodeScope
		}
	}
	inProgressNodesMap := convertNodeOperationResultsToMap(inProgressNodes)
	retryableNodesMap := convertNodeOperationResultsToMap(retryableNodes)

	sortedCandidates := make([]*v1.Node, 0)
	pendingV1Nodes := make([]*v1.Node, 0)
	retryableV1Nodes := make([]*v1.Node, 0)

	for _, nc := range nodesWithLabel {
		_, isProgressNode := inProgressNodesMap[nc.Name]
		_, isRetryableNode := retryableNodesMap[nc.Name]
		isPendingNode := slices.Contains(pendingNodes, nc.Name)
		if !isProgressNode && !isRetryableNode && !isPendingNode {
			logger.Info(fmt.Sprint("newly found node in the scope: ", nc.Name), norNodeName, nc.Name)
			// add new node
			sortedCandidates = append(sortedCandidates, nc)
			// store the starting timestamp when the first time found the node
			schedulingTimer.LoadOrStore(getNorNameNodeName(norName, nc.Name), time.Now())
		}
		if isPendingNode {
			pendingV1Nodes = append(pendingV1Nodes, nc)
		}
		if isRetryableNode {
			retryableV1Nodes = append(retryableV1Nodes, nc)
		}
	}
	sortedCandidates = slices.Concat(sortedCandidates, pendingV1Nodes)
	sortedCandidates = slices.Concat(sortedCandidates, retryableV1Nodes)

	logger.Info(fmt.Sprint("the size of nodes need to be triggered with action is ", len(sortedCandidates)))

	return sortedCandidates, nil
}

func convertNodeOperationResultsToMap(results []norv1beta1.NodeOperationResult) map[string]norv1beta1.NodeOperationResult {
	resultMap := make(map[string]norv1beta1.NodeOperationResult)
	for _, result := range results {
		resultMap[result.NodeName] = result
	}
	return resultMap
}

// convert v1.Nodes list to map: key: nodeName, value: *v1.Node
func convertV1NodesToMap(nodes []*v1.Node) map[string]*v1.Node {
	nodesMap := make(map[string]*v1.Node)
	for _, node := range nodes {
		nodesMap[node.Name] = node
	}
	return nodesMap
}

func (r *NodeOperationRuleReconciler) triggerNodeAction(ctx context.Context, node *v1.Node, nor norv1beta1.NodeOperationRule) (string, NodeOperationFailure) {
	logger := log.FromContext(ctx, norName, nor.Name, norNodeName, node.Name, norInstanceId, node.Spec.ProviderID)
	var workRequestId string
	var err error
	if &nor.Spec != nil && &nor.Spec.Actions != nil && len(nor.Spec.Actions) == 1 {
		if string(nor.Spec.Actions[0]) == replaceBootVolumeAction {
			workRequestId, err = r.bvrNode(ctx, node.Spec.ProviderID, r.Config.ClusterID, nor)
		} else if string(nor.Spec.Actions[0]) == rebootAction {
			workRequestId, err = r.rebootNode(ctx, node.Spec.ProviderID, r.Config.ClusterID, nor)
		} else {
			logger.Error(errActionsFormat, errActionsFormat.Error())
			return "", NodeOperationFailure{
				nodeName:   node.Name,
				errorType:  NodeOperationFailureTypeRetryable,
				errorCode:  "",
				errorMsg:   errActionsFormat.Error(),
				trackingId: "",
			}
		}
		if err != nil {
			nodeOperationFailure := r.convertApiErrorToNodeOperationFailureType(ctx, nor, node.Name, err, "")
			logger.Info(string(nor.Spec.Actions[0])+" work request "+workRequestId+"fails", norOpcId, nodeOperationFailure.trackingId)
			return "", nodeOperationFailure
		}
		return workRequestId, NodeOperationFailure{}
	}
	logger.Error(errActionsFormat, errActionsFormat.Error())
	return "", NodeOperationFailure{
		nodeName:   node.Name,
		errorType:  NodeOperationFailureTypeRetryable,
		errorCode:  "",
		errorMsg:   errActionsFormat.Error(),
		trackingId: "",
	}

}

func (r *NodeOperationRuleReconciler) convertApiErrorToNodeOperationFailureType(ctx context.Context, nor norv1beta1.NodeOperationRule, nodeName string, err error, workRequestId string) NodeOperationFailure {
	logger := log.FromContext(ctx, norName, nor.Name, norNodeName, nodeName, norWorkRequestId, workRequestId)
	logger.Error(err, "error retrieved from API response")

	dimensions := make(map[string]string)
	if err != nil {
		if errBvrRateLimited.Error() == err.Error() || errRebootRateLimited.Error() == err.Error() {
			dimensions[failureCategory] = "RateLimited"
			r.sendMetricsWithClusterId(operationFailure, 1, dimensions)
			return NodeOperationFailure{
				nodeName:   nodeName,
				errorType:  NodeOperationFailureTypeRateLimited,
				errorCode:  "",
				errorMsg:   rateLimitedErrorMsg,
				trackingId: "",
			}
		}

		var ociServiceError common.ServiceError

		if errors.As(err, &ociServiceError) {
			logger.Info("API response ociServiceError received", norHttpStatusCode, ociServiceError.GetHTTPStatusCode(),
				norErrorMessage, ociServiceError.GetMessage(),
				norErrorCode, ociServiceError.GetCode(),
				norOpcId, ociServiceError.GetOpcRequestID())
			dimensions[failureCategory] = ociServiceError.GetCode()
			r.sendMetricsWithClusterId(operationFailure, 1, dimensions)

			retryableFailure := NodeOperationFailure{
				nodeName:   nodeName,
				errorType:  NodeOperationFailureTypeRetryable,
				errorCode:  ociServiceError.GetCode(),
				errorMsg:   ociServiceError.GetMessage(),
				trackingId: ociServiceError.GetOpcRequestID(),
			}
			switch errorCode := ociServiceError.GetCode(); errorCode {
			case HTTP400BadRequestCode:
				if strings.Contains(ociServiceError.GetMessage(), errOKEAPIDisabledMsg) {
					// if feature flag is disabled, non-retryable error
					logger.Error(err, "will not retry on node "+nodeName, norNodeName, nodeName)
					return NodeOperationFailure{
						nodeName:   nodeName,
						errorType:  NodeOperationFailureTypeNonRetryable,
						errorCode:  ociServiceError.GetCode(),
						errorMsg:   ociServiceError.GetMessage(),
						trackingId: ociServiceError.GetOpcRequestID(),
					}
				} else {
					return NodeOperationFailure{
						nodeName:   nodeName,
						errorType:  NodeOperationFailureTypeRetryable,
						errorCode:  ociServiceError.GetCode(),
						errorMsg:   ociServiceError.GetMessage(),
						trackingId: ociServiceError.GetOpcRequestID(),
					}
				}
			case HTTP400InvalidParameterCode:
				// invalid cluster id / node id -> need to send metrics and alarm because there might be a bug in operator
				// https://dyn.slack.com/archives/C073EMW8E30/p1730769071221349?thread_ts=1730742868.854419&cid=C073EMW8E30
				if invalidParameterErrorNeedsAlarm(ociServiceError.GetMessage()) {
					dimensions[norNodeName] = nodeName
					dimensions[cluster] = r.Config.ClusterID
					r.sendMetricsWithTenancy(invalidId, 1, dimensions)
				}
				return retryableFailure
			case HTTP401NotAuthenticatedCode:
				return retryableFailure
			case HTTP404NotAuthorizedOrNotFoundCode:
				// if it is missing system policy from customer to non-prebake image SMN, need to populate clear k8s events
				// https://dyn.slack.com/archives/C073EMW8E30/p1730841700068339?thread_ts=1730835883.140629&cid=C073EMW8E30
				r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonCheckPolicy, eventMsgCheckPolicy)
				return retryableFailure
			case HTTP409ConflictCode:
				return retryableFailure
			case HTTP409IncorrectStateCode:
				return retryableFailure
			case HTTP412NoEtagMatchCode:
				return retryableFailure
			case HTTP429TooManyRequestsCode:
				return retryableFailure
			case HTTP500TooInternalServerErrorCode:
				trackingId := ociServiceError.GetOpcRequestID()
				if len(workRequestId) != 0 {
					trackingId = workRequestId
				}
				return NodeOperationFailure{
					nodeName:   nodeName,
					errorType:  NodeOperationFailureTypeRetryable,
					errorCode:  ociServiceError.GetCode(),
					errorMsg:   ociServiceError.GetMessage(),
					trackingId: trackingId,
				}

			default:
				errMsg := "uncategorizedError: " + ociServiceError.GetMessage()
				trackingId := ""
				if len(ociServiceError.GetOpcRequestID()) != 0 {
					trackingId = ociServiceError.GetOpcRequestID()
				}
				return NodeOperationFailure{
					nodeName:   nodeName,
					errorType:  NodeOperationFailureTypeRetryable,
					errorCode:  ociServiceError.GetCode(),
					errorMsg:   errMsg,
					trackingId: trackingId,
				}
			}
		}
	}

	dimensions[failureCategory] = "unknownError"
	r.sendMetricsWithTenancy(operationFailure, 1, dimensions)
	return NodeOperationFailure{
		nodeName:   nodeName,
		errorType:  NodeOperationFailureTypeRetryable,
		errorCode:  "",
		errorMsg:   unknownErrorMsg,
		trackingId: "",
	}
}

func getNodeNames(nodes []*v1.Node) []string {
	nodeNames := make([]string, len(nodes))
	for i, node := range nodes {
		nodeNames[i] = node.Name
	}
	return nodeNames
}

func (r *NodeOperationRuleReconciler) getWorkRequest(ctx context.Context, nor norv1beta1.NodeOperationRule, workRequestId string, nodeName string) NodeOperationResultUpdate {
	logger := log.FromContext(ctx, norNodeName, nodeName, norWorkRequestId, workRequestId, norName, nor.Name)
	var update UpdateType
	var timeFinished metav1.Time
	workRequest, err := r.OCIClient.ContainerEngine().GetWorkRequest(ctx, workRequestId)
	if err != nil {
		logger.Error(err, "failed to pull work request")
		update = UpdateTypeInProgress
		return NodeOperationResultUpdate{
			NodeName:      nodeName,
			WorkRequestId: workRequestId,
			update:        update,
		}
	}
	logger.Info(fmt.Sprint("get work request ID is ", workRequestId), norWorkRequest, workRequest)
	switch workRequest.Status {
	case containerengine.WorkRequestStatusAccepted:
		update = UpdateTypeInProgress
		break
	case containerengine.WorkRequestStatusInProgress:
		update = UpdateTypeInProgress
		break
	case containerengine.WorkRequestStatusFailed:
		nodeOperationFailure := r.convertApiErrorToNodeOperationFailureType(ctx, nor, nodeName, err, workRequestId)
		if nodeOperationFailure.errorType == NodeOperationFailureTypeRetryable {
			update = UpdateTypeRetryable
		} else {
			update = UpdateTypeNonRetryable
		}
		break
	case containerengine.WorkRequestStatusSucceeded:
		update = UpdateTypeSucceeded
		if workRequest.TimeFinished == nil {
			timeFinished = metav1.Now()
		} else {
			timeFinished = metav1.NewTime(workRequest.TimeFinished.Time)
		}
		break
	case containerengine.WorkRequestStatusCanceling:
		update = UpdateTypeCanceled
		break
	case containerengine.WorkRequestStatusCanceled:
		update = UpdateTypeCanceled
		break
	default:
		update = UpdateTypeRetryable

	}
	return NodeOperationResultUpdate{
		NodeName:      nodeName,
		WorkRequestId: workRequestId,
		update:        update,
		timeFinished:  timeFinished,
	}
}

// cancel work request at best efforts
func (r *NodeOperationRuleReconciler) cancelWorkRequest(ctx context.Context, workRequestId string) error {
	logger := log.FromContext(ctx, norWorkRequestId, workRequestId)
	opcRequestId, err := r.OCIClient.ContainerEngine().DeleteWorkRequest(ctx, workRequestId)
	if err != nil {
		logger.Error(err, "fail to cancel work request")
		var ociServiceError common.ServiceError

		if errors.As(err, &ociServiceError) {
			logger.Info("API response ociServiceError received", norHttpStatusCode, ociServiceError.GetHTTPStatusCode(),
				norErrorMessage, ociServiceError.GetMessage(),
				norErrorCode, ociServiceError.GetCode(),
				norOpcId, ociServiceError.GetOpcRequestID())
			// if work request is already in terminal status, it is expected to be failed to cancel work request. And will not retry cancellation
			// https://bitbucket.oci.oraclecorp.com/projects/OKE/repos/oke-control-plane/browse/api/src/main/java/com/oracle/oci/oke/resources/WorkRequestsHandler.java#220-222
			if ociServiceError.GetCode() == HTTP409ConflictCode {
				return nil
			}
		}
		return err
	} else {
		logger.Info(fmt.Sprint("canceling work request ", workRequestId, " opc request Id ", opcRequestId), norOpcId, opcRequestId)
		return nil
	}
}

func (r *NodeOperationRuleReconciler) removeOkeReservedLabel(log logr.Logger, nodeName string) error {
	node, err := r.KubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		log.Error(err, fmt.Sprint("fail to retrieve node ", nodeName), norNodeName, nodeName)
		return err
	}

	if _, ok := node.Labels[okeReservedLabelKey]; ok {
		log.Info(fmt.Sprint("removing oke reserved label: ", okeReservedLabelKey, "=", node.Labels[okeReservedLabelKey], " from node "+nodeName), norNodeName, nodeName)
		delete(node.Labels, okeReservedLabelKey)
	} else {
		log.Info(fmt.Sprint("oke reserved label: ", okeReservedLabelKey, "=", node.Labels[okeReservedLabelKey], "does not exist"), norNodeName, nodeName)
		return nil
	}

	_, err = r.KubeClient.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err, fmt.Sprint("fail to remove oke reserved label from node ", nodeName), norNodeName, nodeName)
		r.sendMetricsWithClusterId(removeLabelFailure, 1, make(map[string]string))
		return err
	} else {
		r.Recorder.Event(node, v1.EventTypeNormal, eventReasonRemovedLabel, eventMsgRemoveLabel+nodeName)
		// TODO: add nor related event
	}
	return nil
}

func updateResultToList(nodeOperationResult norv1beta1.NodeOperationResult, list []norv1beta1.NodeOperationResult) []norv1beta1.NodeOperationResult {
	if nodeOperationResult == (norv1beta1.NodeOperationResult{}) {
		if list == nil {
			return make([]norv1beta1.NodeOperationResult, 0)
		}
		return list
	}

	var updateList = list
	index := slices.IndexFunc(list, func(item norv1beta1.NodeOperationResult) bool {
		return item.NodeName == nodeOperationResult.NodeName
	})

	if index == -1 {
		updateList = append(list, nodeOperationResult)
	} else {
		updateList[index] = nodeOperationResult
	}
	return updateList
}

func addToListIfNotExist(newNodeNames []string, nodeNames []string) []string {
	if newNodeNames == nil || len(newNodeNames) == 0 {
		if nodeNames == nil {
			return make([]string, 0)
		}
		return nodeNames
	}
	if nodeNames == nil || len(nodeNames) == 0 {
		if newNodeNames == nil {
			return make([]string, 0)
		}
		return newNodeNames
	}

	var updatedList = nodeNames
	for _, n := range newNodeNames {
		if !slices.Contains(nodeNames, n) {
			updatedList = append(updatedList, n)
		}
	}

	return updatedList
}

func removeResultFromList(nodeName string, list []norv1beta1.NodeOperationResult) []norv1beta1.NodeOperationResult {
	if list == nil {
		return make([]norv1beta1.NodeOperationResult, 0)
	}
	if !slices.ContainsFunc(list, func(item norv1beta1.NodeOperationResult) bool {
		return item.NodeName == nodeName
	}) {
		return list
	}
	slices.DeleteFunc(list, func(item norv1beta1.NodeOperationResult) bool {
		return item.NodeName == nodeName
	})
	list = list[0 : len(list)-1]
	return list
}

func removeNodeFromList(nodeName string, nodes []string) []string {
	if nodes == nil {
		return make([]string, 0)
	}
	if !slices.Contains(nodes, nodeName) {
		return nodes
	}
	slices.DeleteFunc(nodes, func(n string) bool {
		return n == nodeName
	})

	nodes = nodes[0 : len(nodes)-1]
	return nodes
}

func mergeFilterLabels(triggerLabel map[string]string, customLabels map[string]string) (labels.Selector, error) {
	overallLabels := make(map[string]string, 0)
	maps.Copy(overallLabels, triggerLabel)
	if customLabels != nil {
		maps.Copy(overallLabels, customLabels)
	}
	return labels.ValidatedSelectorFromSet(overallLabels)
}

// isSubset is to whether the node list in the nor.Status is contained in the nodes with labels retrieved from kube-api server
// if not, return the nodes which are not in the nodes with labels
func isSubset(nodesInStatus map[string]norv1beta1.NodeOperationResult, nodesWithLabel map[string]*v1.Node) (bool, []string) {
	var isSubset = true
	results := make([]string, 0)
	for nodeName, _ := range nodesInStatus {
		if _, ok := nodesWithLabel[nodeName]; !ok {
			isSubset = false
			results = append(results, nodeName)
		}
	}
	return isSubset, results
}

// isNodeNameSubset is to whether the nodeNames in the nor.Status is contained in the nodes with labels retrieved from kube-api server
// if any of nodeNames is not contained in nodesWithLabel, return such nodes
func isNodeNameSubset(nodeNames []string, nodesWithLabel map[string]*v1.Node) (bool, []string) {
	var isSubset = true
	results := make([]string, 0)
	for _, nodeName := range nodeNames {
		if _, ok := nodesWithLabel[nodeName]; !ok {
			isSubset = false
			results = append(results, nodeName)
		}
	}
	return isSubset, results
}

func initializeListInStatus(resultList []norv1beta1.NodeOperationResult) []norv1beta1.NodeOperationResult {
	if resultList == nil || len(resultList) == 0 {
		return make([]norv1beta1.NodeOperationResult, 0)
	}
	return resultList
}

func initializeSuccessListInStatus(successList []norv1beta1.NodeOperationSuccess) []norv1beta1.NodeOperationSuccess {
	if successList == nil || len(successList) == 0 {
		return make([]norv1beta1.NodeOperationSuccess, 0)
	}
	return successList
}

func initializePendingList(pendingList []string) []string {
	if pendingList == nil {
		return make([]string, 0)
	}
	return pendingList
}

func (r *NodeOperationRuleReconciler) updateStatus(ctx context.Context,
	namespacedName types.NamespacedName,
	updatedInProgressNodes []norv1beta1.NodeOperationResult,
	updatedPendingNodes []string,
	updatedRetryableNodes []norv1beta1.NodeOperationResult,
	updatedSucceededNodes []norv1beta1.NodeOperationSuccess,
	updatedCanceledNodes []norv1beta1.NodeOperationResult) (ctrl.Result, error) {
	logger := log.FromContext(ctx, norName, namespacedName.Name)

	updatedNor := norv1beta1.NodeOperationRule{}
	err := r.Get(ctx, namespacedName, &updatedNor)
	if err != nil {
		logger.Error(err, "failed ot get updated NOR CR")
		return ctrl.Result{}, err
	}

	updatedNor.Status.InProgressNodes = updatedInProgressNodes
	updatedNor.Status.BackOffNodes = updatedRetryableNodes
	updatedNor.Status.PendingNodes = updatedPendingNodes
	updatedNor.Status.SucceededNodes = updatedSucceededNodes
	updatedNor.Status.CanceledNodes = updatedCanceledNodes

	logger.Info(fmt.Sprint("will update nor status: in progress: ", updatedInProgressNodes, " pending: ", updatedPendingNodes, " backoff: ", updatedRetryableNodes, " canceled: ", updatedCanceledNodes, " succeeded: ", updatedSucceededNodes))

	err = r.Status().Update(ctx, &updatedNor)
	if err != nil {
		logger.Error(err, "fail to update NOR status")
		return ctrl.Result{}, err
	} else {
		updatedResourceVersion := updatedNor.ResourceVersion
		logger.Info(fmt.Sprint("successfully update nor status. Generation: ", updatedNor.Generation, " Resource Version: ", updatedResourceVersion))
		updateGlobalResourceVersion(namespacedName, updatedResourceVersion)
	}

	return ctrl.Result{RequeueAfter: RequeueDurationForUpdateWithoutError}, nil

}

func (r *NodeOperationRuleReconciler) addFinalizer(ctx context.Context, nor norv1beta1.NodeOperationRule) error {
	log := log.FromContext(ctx, norName, nor.Name)

	controllerutil.AddFinalizer(&nor, finalizer)
	if err := r.Update(ctx, &nor); err != nil {
		log.Error(err, fmt.Sprint("fail to add finalizer to nor ", nor.Name))
		return err
	} else {
		log.Info("Add finalizer to nor")
		return nil
	}
}

// cancelOperations is to cancel node operations
// pending nodes: move all pending nodes to the canceled nodes
// backoff nodes : move all backoff nodes to the canceled nodes
// in progress nodes : cancel in-progress WRs and move them to canceled nodes //TODO: whether need to parallel cancel
func (r *NodeOperationRuleReconciler) cancelOperations(ctx context.Context,
	nor norv1beta1.NodeOperationRule,
	updateNorStatus norv1beta1.NodeOperationRuleStatus) (norv1beta1.NodeOperationRuleStatus, error) {
	log := log.FromContext(ctx, norName, nor.Name)

	// move pending nodes to canceled nodes
	for _, pendingNode := range updateNorStatus.PendingNodes {
		updateNorStatus.CanceledNodes = updateResultToList(norv1beta1.NodeOperationResult{
			NodeName:      pendingNode,
			WorkRequestId: "",
		}, updateNorStatus.CanceledNodes)
	}
	updateNorStatus.PendingNodes = make([]string, 0)

	// move backOff nodes to canceled nodes
	for _, retryableNode := range updateNorStatus.BackOffNodes {
		updateNorStatus.CanceledNodes = updateResultToList(norv1beta1.NodeOperationResult{
			NodeName:      retryableNode.NodeName,
			WorkRequestId: "",
		}, updateNorStatus.CanceledNodes)
	}
	updateNorStatus.BackOffNodes = make([]norv1beta1.NodeOperationResult, 0)

	copiedInProgressNode := make([]norv1beta1.NodeOperationResult, len(updateNorStatus.InProgressNodes))
	copy(copiedInProgressNode, updateNorStatus.InProgressNodes)

	// cancel the in-progress node operations
	for _, inProgressNode := range copiedInProgressNode {
		err := r.cancelWorkRequest(ctx, inProgressNode.WorkRequestId)
		if err != nil {
			log.Error(err, errFailedToCancelWR.Error(), norWorkRequestId, inProgressNode.WorkRequestId)
			return updateNorStatus, err
		} else {
			updateNorStatus.CanceledNodes = updateResultToList(inProgressNode, updateNorStatus.CanceledNodes)
			updateNorStatus.InProgressNodes = removeResultFromList(inProgressNode.NodeName, updateNorStatus.InProgressNodes)
		}
	}
	return updateNorStatus, nil
}

func (r *NodeOperationRuleReconciler) validateNOR(ctx context.Context, nor norv1beta1.NodeOperationRule) (labels.Selector, error) {
	log := log.FromContext(ctx, norName, nor.Name)

	// if NOR format is invalid, will not requeue
	_, err := r.validateNorFormat(ctx, nor)
	if err != nil {
		return nil, err
	}

	// construct label selector consisting of MatchTriggerLabel and customer's labels
	overallLabelsSelector, err := mergeFilterLabels(nor.Spec.NodeSelector.MatchTriggerLabel, nor.Spec.NodeSelector.MatchCustomLabels)
	if err != nil {
		log.Error(err, "NodeSelector fails to convert to label selector", norMergedLabel, overallLabelsSelector)
		r.Recorder.Event(&nor, v1.EventTypeWarning, eventReasonValidatedNOR, errNodeSelectorFormat.Error())
		// if the label cannot be converted to selector, it means the label format has error, will not requeue
		return nil, err
	}
	log.Info("merged label selector declared in the NOR spec", norMergedLabel, overallLabelsSelector)
	return overallLabelsSelector, nil
}

// filterMatchedNodes is to filter all the nodes that are matched with the labels defined in the nor
func (r *NodeOperationRuleReconciler) filterMatchedNodes(ctx context.Context, overallLabelsSelector labels.Selector) []*v1.Node {
	log := log.FromContext(ctx)

	var nodesWithLabel []*v1.Node
	allNodes := r.nodeList.Items

	for _, node := range allNodes {
		if overallLabelsSelector.Matches(labels.Set(node.Labels)) {
			nodesWithLabel = append(nodesWithLabel, &node)
			log.Info(fmt.Sprint("retrieved node with qualified label is ", node.Name), norNodeName, node.Name)
		}
	}
	return nodesWithLabel
}

// getLatestResourceVersion is to retrieve the latest recorded resource version
func getLatestResourceVersion(key types.NamespacedName, version string) string {
	_, exists := globalResourceVersionMap[key]
	if !exists {
		updateGlobalResourceVersion(key, version)
	}
	return globalResourceVersionMap[key]
}

// updateGlobalResourceVersion is to update the latest resource version in the global map
func updateGlobalResourceVersion(key types.NamespacedName, version string) {
	if globalResourceVersionMap == nil {
		globalResourceVersionMap = make(map[types.NamespacedName]string)
	}
	globalResourceVersionMap[key] = version
}

// deleteEntryFromGlobalResourceVersion is to delete the key entry from the global map
// it is mainly used when the nor cr is deleted
func deleteEntryFromGlobalResourceVersion(key types.NamespacedName) {
	delete(globalResourceVersionMap, key)
}

func isNorStale(currentVersion string, latestVersion string) bool {
	if len(latestVersion) == 0 {
		return false
	}
	currentVersionInt, err1 := strconv.Atoi(currentVersion)
	latestVersionInt, err2 := strconv.Atoi(latestVersion)
	if err1 != nil || err2 != nil {
		return true
	}
	return currentVersionInt < latestVersionInt
}

// sendMetricsWithTenancy is to send metrics with Tenancy ID which is more used to tracking customer habit
func (r *NodeOperationRuleReconciler) sendMetricsWithTenancy(metricName string, value float64, dimensions map[string]string) {
	_, exists := dimensions[tenancy]
	if !exists {
		dimensions[tenancy] = r.Config.Auth.TenancyID
	}
	metricName = baseMetricName + "." + metricName
	metrics.SendMetricData(r.MetricPusher, metricName, value, dimensions)
}

// sendMetricsWithClusterId is to send metrics with Cluster ID which is more used to tracking nor operator performance and health
func (r *NodeOperationRuleReconciler) sendMetricsWithClusterId(metricName string, value float64, dimensions map[string]string) {
	_, exists := dimensions[cluster]
	if !exists {
		dimensions[cluster] = r.Config.ClusterID
	}
	metricName = baseMetricName + "." + metricName
	metrics.SendMetricData(r.MetricPusher, metricName, value, dimensions)
}

func getNorNameNodeName(norName string, nodeName string) string {
	return norName + nodeName
}

// calculateLatency is to calculate the time difference between starting time of the item (nor name + node name) and now
// it is used to calculate scheduling latency and operation latency
func calculateLatency(timer sync.Map, norName string, nodeName string) float64 {
	key := getNorNameNodeName(norName, nodeName)
	if val, ok := timer.Load(key); ok {
		nodeFoundTimestamp := val.(time.Time)
		latency := time.Since(nodeFoundTimestamp).Seconds()
		timer.Delete(key)
		return latency
	}
	return 0
}

// removeLatency is to remove the timestamp of the item (nor name + node name) if the item value exists
func removeLatency(timer sync.Map, norName string, nodeName string) {
	key := getNorNameNodeName(norName, nodeName)
	if _, ok := timer.Load(key); ok {
		timer.Delete(key)
	}
}

func sortSucceededNodesAndGarbageCollect(SucceededNodes []norv1beta1.NodeOperationSuccess) ([]norv1beta1.NodeOperationSuccess, int) {
	numberOFNodesTrimmed := 0
	if SucceededNodes == nil || len(SucceededNodes) < maxSuccededNodesBeforeGarbageCollection {
		return SucceededNodes, numberOFNodesTrimmed
	}
	//Given that we append to the list of SucceededNodes, ideally it should already be sorted by SuccessTimestamp
	//we sort it anyway just in case there is any discrepancy.Not a costly operation since the number of nodes sorted should always be ~1000
	sort.Slice(SucceededNodes, func(i, j int) bool {
		return SucceededNodes[i].SuccessTimestamp.Time.Before(SucceededNodes[j].SuccessTimestamp.Time)
	})
	numberOfSucceededNodes := len(SucceededNodes)
	numberOFNodesTrimmed = numberOfSucceededNodes - maxSuccededNodesBeforeGarbageCollection
	//Return the last maxSuccededNodesBeforeGarbageCollection number of elements
	return SucceededNodes[numberOFNodesTrimmed:], numberOFNodesTrimmed
}

func invalidParameterErrorNeedsAlarm(errorMsg string) bool {
	return strings.Contains(errorMsg, blankNodeId) || strings.Contains(errorMsg, invalidNodeIdFormat) || strings.Contains(errorMsg, blankClusterId) || strings.Contains(errorMsg, wrongNodeMember)
}
