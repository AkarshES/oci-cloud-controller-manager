package controllers

import (
	"context"
	"reflect"
	"sync"
	"time"

	norv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/node-cycling/v1beta1"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	narName string = "narName"
)

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
	log.Info("Setting up NPN controller with manager")
	r.Recorder = mgr.GetEventRecorderFor("nodeautoRepair")

	return ctrl.NewControllerManagedBy(mgr).
		// Watch for changes to Node objects. This is crucial for checking persistent
		// conditions like DiskPressure, MemoryPressure, or custom NPD conditions.
		For(&v1.Node{}, builder.WithPredicates(ConditionChangedPredicate{log: log})).
		// Watch for new Event objects. This is used for temporary issues reported
		// by NPD, like a transient IMDS check failure.
		Watches(
			&v1.Event{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				event, ok := obj.(*v1.Event)
				if !ok {
					return nil
				}

				// Filter events to only those from Node Problem Detector.
				if event.Reason == "IMDSCheckFailed" {
					return []reconcile.Request{
						{
							NamespacedName: types.NamespacedName{
								Name: event.InvolvedObject.Name,
							},
						},
					}
				}

				return nil
			}),
		).
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
		log.Info("CCM NAR: found condition " + condition.String())
		// 3. Look for the "IMDSUnreachable" condition and check if its status is True.
		if condition.Type == "IMDSUnreachable" && condition.Status == v1.ConditionTrue {
			// Log a warning event to Kubernetes to make the issue visible.
			r.Recorder.Event(node, v1.EventTypeWarning, "IMDSUnreachableFromCCM", "Node condition IMDSUnreachable is now: True")

			// Trigger the auto-repair logic here. For example, you can call a separate function.
			log.Info("CCM: IMDS is unreachable, triggering repair action.", "node", req.NamespacedName.Name)

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
	}

	// 4. If the loop completes without finding a "True" IMDSUnreachable condition,
	// the node is considered healthy.
	log.Info("CCM: Node auto repair Node is healthy and has no IMDSUnreachable condition.", "node", req.NamespacedName.Name)

	// The reconciliation is complete. Return an empty result to indicate no need to re-process.
	return ctrl.Result{}, nil
}

func (r *NodeAutoRepairReconciler) rebootNode(ctx context.Context, nodeId string, clusterId string, nor norv1beta1.NodeOperationRule) (string, error) {
	logger := log.FromContext(ctx, norInstanceId, nodeId, norClusterId, clusterId)
	var workRequestId string
	var err error

	workRequestId, err = r.OCIClient.ContainerEngine().RebootClusterNode(ctx, nodeId, clusterId, nor)
	logger.Info("CCM: Trigger reboot action")
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

	if !reflect.DeepEqual(oldNode.Status.Conditions, newNode.Status.Conditions) {
		p.log.Infow("CCM: Node conditions have changed, triggering reconciliation.",
			"node", newNode.Name,
			"oldConditions", oldNode.Status.Conditions,
			"newConditions", newNode.Status.Conditions,
		)
		return true
	}
	p.log.Info("CCM: Node conditions haven't changed")

	return false
}

func (p ConditionChangedPredicate) Create(e event.CreateEvent) bool {
	return true
}
