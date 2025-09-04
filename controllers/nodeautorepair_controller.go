package controllers

import (
	"context"
	"sync"
	"time"

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
	narName string = "narName"
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
// It configures the controller to watch for changes to both Node and Event objects.
func (r *NodeAutoRepairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := zap.L().Sugar()
	log.Info("Setting up NAR controller with manager")
	r.Recorder = mgr.GetEventRecorderFor("nodeAutoRepair")

	return ctrl.NewControllerManagedBy(mgr).
		// Watch for changes to Node objects. This is crucial for checking persistent
		// conditions like DiskPressure, MemoryPressure, or custom NPD conditions.
		For(&v1.Node{}, builder.WithPredicates(ConditionChangedPredicate{log: log})).
		WithOptions(controller.Options{MaxConcurrentReconciles: 20, CacheSyncTimeout: time.Hour}).
		Complete(r)
}

// Reconcile implements reconcile.TypedReconciler.
func (r *NodeAutoRepairReconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)
	log = log.WithValues(norName, "NAR")
	// 1. Fetch the Node object that triggered the reconciliation.
	node := &v1.Node{}
	if err := r.Client.Get(ctx, req.NamespacedName, node); err != nil {
		// If the node is not found, it might have been deleted. Ignore the request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Iterate through the node's status conditions to check for problems.
	for _, condition := range node.Status.Conditions {
		// 3. Look for the "IMDSUnreachable" condition and check if its status is True.
		if conditionStatus, ok := CONDITIONS[string(condition.Type)]; ok {
			if conditionStatus != string(v1.ConditionTrue) {
				continue
			}
		}

		if string(condition.Type) == "GPUCOUNT" {
			log.Info("CCM: condition "+string(condition.Type)+" triggered. Triggering terminate action.", "node", req.NamespacedName.Name)
			workrequestId, _ := r.OCIClient.Compute().TerminateInstance(ctx, node.Spec.ProviderID)
			log.Info("CCM: Terminate instance workrequest id: " + workrequestId)
			return ctrl.Result{}, nil
		}

		// Log a warning event to Kubernetes to make the issue visible.
		r.Recorder.Event(node, v1.EventTypeWarning, string(condition.Type), "Node condition"+string(condition.Type)+" is now: True")

		// Trigger the auto-repair logic here. For example, you can call a separate function.
		log.Info("CCM: condition "+string(condition.Type)+" triggered. Triggering repair action.", "node", req.NamespacedName.Name)

		// Requeue the request after a delay to periodically check if the issue is resolved.
		log.Info("CCM: Requeue after 10 mins", "node", req.NamespacedName.Name)
		workRequestId, _ := r.rebootNode(ctx, node.Spec.ProviderID, r.Config.ClusterID, norv1beta1.NodeOperationRule{
			Spec: norv1beta1.NodeOperationRuleSpec{
				NodeEvictionSettings: norv1beta1.NodeEvictionSettings{
					EvictionGracePeriod:             1,
					IsForceActionAfterGraceDuration: true,
				},
			},
		})
		log.Info("CCM: Auto Repair work request id for reboot: " + workRequestId)
		return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil

	}

	// 4. If the loop completes without finding a "True" IMDSUnreachable condition,
	// the node is considered healthy.
	log.Info("CCM: Node auto repair Node is healthy and has no NPD conditions.", "node", req.NamespacedName.Name)

	// The reconciliation is complete. Return an empty result to indicate no need to re-process.
	return ctrl.Result{}, nil
}

func (r *NodeAutoRepairReconciler) rebootNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) (string, error) {
	// logger := log.FromContext(ctx, norInstanceId, nodeId, norClusterId, clusterId)
	var workRequestId string
	var err error

	workRequestId, err = r.OCIClient.ContainerEngine().RebootClusterNode(ctx, nodeId, clusterId, nor)
	// logger.Info("CCM: Trigger reboot action")
	return workRequestId, err
}

type ConditionChangedPredicate struct {
	predicate.Funcs
	log *zap.SugaredLogger
}

func (p ConditionChangedPredicate) Update(e event.UpdateEvent) bool {
	// Assert that both the old and new objects are of type *v1.Node.
	oldNode, ok := e.ObjectOld.(*v1.Node)
	if !ok {
		// If the old object is not a Node, we can't perform a comparison.
		return false
	}

	newNode, ok := e.ObjectNew.(*v1.Node)
	if !ok {
		// If the new object is not a Node, we can't perform a comparison.
		return false
	}

	// Convert the conditions slices to maps for efficient lookup by condition type.
	oldConditions := getConditionMap(oldNode.Status.Conditions)
	newConditions := getConditionMap(newNode.Status.Conditions)

	// Iterate through the new conditions to find added or changed conditions.
	for _, newCondition := range newConditions {
		oldCondition, exists := oldConditions[newCondition.Type]

		// Case 1: A new condition was added.
		if !exists {
			p.log.Infow("CCM: New condition added to node.",
				"node", newNode.Name,
				"conditionType", newCondition.Type,
				"status", newCondition.Status,
				"reason", newCondition.Reason,
			)
			return true
		}

		// Case 2: An existing condition's status, reason, or message changed.
		if oldCondition.Status != newCondition.Status ||
			oldCondition.Reason != newCondition.Reason ||
			oldCondition.Message != newCondition.Message {

			p.log.Infow("CCM: Node condition: "+string(newCondition.Type)+" changed",
				"node", newNode.Name,
				"conditionType", newCondition.Type,
				"oldStatus", oldCondition.Status,
				"newStatus", newCondition.Status,
				"oldReason", oldCondition.Reason,
				"newReason", newCondition.Reason,
				"oldHeartBeatTime", oldCondition.LastHeartbeatTime,
				"newHeartBeatTime", newCondition.LastHeartbeatTime,
				"oldTransitTime", oldCondition.LastTransitionTime,
				"newTransitTime", newCondition.LastHeartbeatTime,
			)
			if oldCondition.Type == "IMDSUnreachable" {
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
	// If we reach here, no significant changes were found in the conditions.
	return false
}

// getConditionMap is a helper function to convert a slice of NodeCondition
// objects into a map for quick lookup.
func getConditionMap(conditions []v1.NodeCondition) map[v1.NodeConditionType]v1.NodeCondition {
	conditionMap := make(map[v1.NodeConditionType]v1.NodeCondition, len(conditions))
	for _, condition := range conditions {
		conditionMap[condition.Type] = condition
	}
	return conditionMap
}

func (p ConditionChangedPredicate) Create(e event.CreateEvent) bool {
	return true
}
