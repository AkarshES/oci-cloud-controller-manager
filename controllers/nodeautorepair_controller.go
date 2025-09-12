package controllers

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	norv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/node-cycling/v1beta1"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	narName                    string         = "narName"
	repairProblemDetectedLabel string         = "oci.oraclecloud.com/node-problem-detected"
	repairEnabledLabel         string         = "oci.oraclecloud.com/node-auto-repair-enabled"
	repairFrequencyLabel       string         = "oci.oraclecloud.com/node-auto-repair-freq"
	REPAIR_TAINT_KEY           string         = "oci.oraclecloud.com/node-auto-repair-scheduled"
	REPAIR_TAINT_EFFECT        v1.TaintEffect = v1.TaintEffectNoSchedule
)

var CONDITIONS map[string]string = map[string]string{
	"IMDSUnreachable": "True",
	"GPUCount":        "True",
	"GPUCdfpCable":    "True",
	"GPURowRemap":     "True",
}

type NodeAutoRepairReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	MetricPusher     *metrics.MetricPusher
	OCIClient        ociclient.Interface
	KubeClient       clientset.Interface
	TimeTakenTracker sync.Map
	Recorder         record.EventRecorder
	Config           *providercfg.Config
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeAutoRepairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := zap.L().Sugar()
	log.Info("Setting up NAR controller with manager")
	r.Recorder = mgr.GetEventRecorderFor("nodeAutoRepair")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Node{}, builder.WithPredicates(ConditionChangedPredicate{log: log})).
		WithOptions(controller.Options{MaxConcurrentReconciles: 20, CacheSyncTimeout: time.Hour}).
		Complete(r)
}

// Reconcile is the main controller loop, now acting as an orchestrator.
func (r *NodeAutoRepairReconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	// Use the logger from the context, as requested.
	logger := log.FromContext(ctx)
	logger = logger.WithValues(narName, "NAR")

	// 1. Fetch the Node object.
	node := &v1.Node{}
	if err := r.Client.Get(ctx, req.NamespacedName, node); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Check if the node has an unhealthy condition.
	if unhealthyCondition := findUnhealthyCondition(node); unhealthyCondition != nil {
		// 2a. If unhealthy, execute the repair logic.
		logger.Info("CCM: Node is unhealthy, starting repair process", "condition", unhealthyCondition.Type)
		return r.handleUnhealthyNode(ctx, logger, node, unhealthyCondition)
	}

	// 3. If healthy, execute the cleanup logic for any previous repairs.
	return r.cleanupRepairArtifacts(ctx, logger, node)
}

// findUnhealthyCondition checks if a node has a condition that warrants repair.
// It returns the first matching unhealthy condition, or nil if the node is healthy.
func findUnhealthyCondition(node *v1.Node) *v1.NodeCondition {
	for _, condition := range node.Status.Conditions {
		if conditionStatus, ok := CONDITIONS[string(condition.Type)]; ok && (conditionStatus == string(condition.Status)) {
			return &condition
		}
	}
	return nil
}

// handleUnhealthyNode performs all actions when a node is found to be unhealthy.
func (r *NodeAutoRepairReconciler) handleUnhealthyNode(ctx context.Context, logger logr.Logger, node *v1.Node, condition *v1.NodeCondition) (ctrl.Result, error) {
	patch := client.MergeFrom(node.DeepCopy())
	var needsPatch bool

	// Get dry run label
	var repairEnabled bool
	if val, ok := node.Labels[repairEnabledLabel]; ok && val == "true" {
		repairEnabled = true
	}

	// Action 1: Add repair label if it doesn't exist.
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	if _, ok := node.Labels[repairProblemDetectedLabel]; !ok {
		node.Labels[repairProblemDetectedLabel] = string(condition.Type)
		// logger.Info("CCM: Adding label to unhealthy node", "node", node.Name, "label", labelKey)
		needsPatch = true
	}

	// Action 2: Add repair taint if it doesn't exist.
	taintFound := false
	for _, taint := range node.Spec.Taints {
		if taint.Key == REPAIR_TAINT_KEY && taint.Effect == REPAIR_TAINT_EFFECT {
			taintFound = true
			break
		}
	}
	if !taintFound {
		repairTaint := CreateRepairTaint()
		node.Spec.Taints = append(node.Spec.Taints, repairTaint)
		// logger.Info("CCM: Adding taint to unhealthy node", "node", node.Name, "taint", REPAIR_TAINT.Key)
		needsPatch = true
	}

	// Action 3: Apply a single patch for both label and taint to be efficient.
	if needsPatch {
		if err := r.Client.Patch(ctx, node, patch); err != nil {
			logger.Error(err, "CCM: Failed to patch node with repair label/taint")
			return ctrl.Result{}, err
		}
	}

	// Action 4: Trigger the repair action (terminate).
	// if string(condition.Type) == "GPUCountMismatch" && taintFound {
	// 	logger.Info("GPUCountMismatch condition detected on already tainted node. Triggering terminate action.", "node", node.Name)
	// 	workrequestId, _ := r.OCIClient.Compute().TerminateInstance(ctx, node.Spec.ProviderID)
	// 	logger.Info("CCM: Terminate instance workrequest id: " + workrequestId)
	// 	return ctrl.Result{}, nil
	// }

	var eventMessage string = "[Node Auto Repair]: Node condition " + string(condition.Type) + " is now: " + string(condition.Status) + ", OKE is triggering - REBOOT - repair action"
	if !repairEnabled {
		eventMessage += "(in dry run mode)"
	}
	r.Recorder.Event(node, v1.EventTypeWarning, string(condition.Type), eventMessage)
	logger.Info("CCM: Condition triggered repair action", "condition", string(condition.Type), "node", node.Name)

	if repairEnabled {
		workRequestId, _ := r.rebootNode(ctx, node.Spec.ProviderID, r.Config.ClusterID, norv1beta1.NodeOperationRule{
			Spec: norv1beta1.NodeOperationRuleSpec{
				NodeEvictionSettings: norv1beta1.NodeEvictionSettings{
					EvictionGracePeriod:             1,
					IsForceActionAfterGraceDuration: true,
				},
			},
		})
		r.Recorder.Event(node, v1.EventTypeWarning, "NodeAutoRepairController", "[Node Auto Repair]: OKE Reboot work request id: "+workRequestId)
		logger.Info("CCM: Auto Repair work request id for reboot: " + workRequestId)
	}

	requeDuration := getRequeueDuration(logger, node)
	logger.Info("CCM: Requeuing request for periodic check after " + requeDuration.String())
	return ctrl.Result{RequeueAfter: requeDuration}, nil
}

// cleanupRepairArtifacts checks a healthy node for leftover repair items and removes them.
func (r *NodeAutoRepairReconciler) cleanupRepairArtifacts(ctx context.Context, logger logr.Logger, node *v1.Node) (ctrl.Result, error) {
	patch := client.MergeFrom(node.DeepCopy())
	var needsPatch bool

	// 1. Clean up repair labels.
	for key := range node.Labels {
		if strings.HasPrefix(key, repairProblemDetectedLabel) {
			logger.Info("Node is healthy, removing repair label", "node", node.Name, "label", key)
			delete(node.Labels, key)
			needsPatch = true
		}
	}

	// 2. Clean up repair taints.
	var taintsToKeep []v1.Taint
	var taintToRemove bool
	for _, taint := range node.Spec.Taints {
		if taint.Key == REPAIR_TAINT_KEY && taint.Effect == REPAIR_TAINT_EFFECT {
			taintToRemove = true
		} else {
			taintsToKeep = append(taintsToKeep, taint)
		}
	}

	if taintToRemove {
		logger.Info("Node is healthy, removing repair taint", "node", node.Name, "taint", REPAIR_TAINT_KEY)
		node.Spec.Taints = taintsToKeep
		needsPatch = true
	}

	// 3. Apply a single patch only if changes were made.
	if needsPatch {
		logger.Info("Applying cleanup patch to node", "node", node.Name)
		if err := r.Client.Patch(ctx, node, patch); err != nil {
			logger.Error(err, "Failed to apply cleanup patch to node")
			return ctrl.Result{}, err
		}
	} else {
		logger.Info("Node is healthy and has no NPD conditions to clean up.", "node", node.Name)
	}

	return ctrl.Result{}, nil
}

func (r *NodeAutoRepairReconciler) rebootNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) (string, error) {
	var workRequestId string
	var err error
	workRequestId, err = r.OCIClient.ContainerEngine().RebootClusterNode(ctx, nodeId, clusterId, nor)
	return workRequestId, err
}

type ConditionChangedPredicate struct {
	predicate.Funcs
	log *zap.SugaredLogger
}

func (p ConditionChangedPredicate) Update(e event.UpdateEvent) bool {
	oldNode, ok := e.ObjectOld.(*v1.Node)
	if !ok {
		return false
	}
	newNode, ok := e.ObjectNew.(*v1.Node)
	if !ok {
		return false
	}
	oldConditions := getConditionMap(oldNode.Status.Conditions)
	newConditions := getConditionMap(newNode.Status.Conditions)
	for _, newCondition := range newConditions {
		oldCondition, exists := oldConditions[newCondition.Type]
		if !exists {
			p.log.Infow("CCM: New condition added to node.", "node", newNode.Name, "conditionType", newCondition.Type, "status", newCondition.Status, "reason", newCondition.Reason)
			return true
		}
		if oldCondition.Status != newCondition.Status || oldCondition.Reason != newCondition.Reason || oldCondition.Message != newCondition.Message {
			p.log.Infow("CCM: Node condition: "+string(newCondition.Type)+" changed", "node", newNode.Name, "conditionType", newCondition.Type, "oldStatus", oldCondition.Status, "newStatus", newCondition.Status, "oldReason", oldCondition.Reason, "newReason", newCondition.Reason, "oldHeartBeatTime", oldCondition.LastHeartbeatTime, "newHeartBeatTime", newCondition.LastHeartbeatTime, "oldTransitTime", oldCondition.LastTransitionTime, "newTransitTime", newCondition.LastHeartbeatTime)
			if _, ok := CONDITIONS[string(oldCondition.Type)]; ok {
				return true
			}
		}
	}
	var lastManager string
	var lastUpdateTime time.Time
	var fieldName string
	for _, field := range newNode.ManagedFields {
		if field.Time.After(lastUpdateTime) {
			lastUpdateTime = field.Time.Time
			lastManager = field.Manager
			fieldName = field.String()
		}
	}
	if lastManager != "" {
		p.log.Info("CCM: NPD Condition hasn't changed. Detected a change in Node object.", " manager: ", lastManager, " updateTime: ", lastUpdateTime, " fieldName:", fieldName)
	} else {
		p.log.Info("CCM: NPD Condition hasn't changed. No manager found for this update event.")
	}
	return false
}

func getConditionMap(conditions []v1.NodeCondition) map[v1.NodeConditionType]v1.NodeCondition {
	conditionMap := make(map[v1.NodeConditionType]v1.NodeCondition, len(conditions))
	for _, condition := range conditions {
		conditionMap[condition.Type] = condition
	}
	return conditionMap
}

func getRequeueDuration(logger logr.Logger, node *v1.Node) time.Duration {
	// Define the default requeue duration
	requeueDuration := 10 * time.Minute

	if val, ok := node.Labels[repairFrequencyLabel]; ok {
		// Parse the string value to an integer (base 10, 64-bit)
		if parsedValue, err := strconv.ParseInt(val, 10, 64); err == nil && parsedValue > 0 {
			// If parsing is successful and the value is positive, use it
			requeueDuration = time.Duration(parsedValue) * time.Minute
		} else {
			// Log a warning if the label value is invalid, but continue with the default
			logger.Info("CCM: Invalid value for repair frequency label on node: " + node.Name + " Frequency Lable has a value: " + val)
		}
	}

	return requeueDuration
}

func (p ConditionChangedPredicate) Create(e event.CreateEvent) bool {
	return true
}

// createRepairTaint creates a new Taint object with a timestamp as its value.
func CreateRepairTaint() v1.Taint {
	// Get the current time and format it as a string
	now := time.Now().UTC()
	const k8sTaintTimeFormat = "2006-01-02-15-04-05"

	// Format using a reference string
	timestamp := now.Format(k8sTaintTimeFormat)

	return v1.Taint{
		Key:    REPAIR_TAINT_KEY,
		Value:  timestamp, // Value is now a dynamic timestamp
		Effect: REPAIR_TAINT_EFFECT,
	}
}
