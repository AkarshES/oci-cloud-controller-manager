package controllers

import (
	"context"
	"sync"
	"time"

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
			return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
		}
	}

	// 4. If the loop completes without finding a "True" IMDSUnreachable condition,
	// the node is considered healthy.
	log.Info("CCM: Node auto repair Node is healthy and has no IMDSUnreachable condition.", "node", req.NamespacedName.Name)

	// The reconciliation is complete. Return an empty result to indicate no need to re-process.
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// It configures the controller to watch for changes to both Node and Event objects.
func (r *NodeAutoRepairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := zap.L().Sugar()
	log.Info("Setting up NPN controller with manager")
	r.Recorder = mgr.GetEventRecorderFor("nativepodnetwork")

	return ctrl.NewControllerManagedBy(mgr).
		// Watch for changes to Node objects. This is crucial for checking persistent
		// conditions like DiskPressure, MemoryPressure, or custom NPD conditions.
		For(&v1.Node{}).
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
				if event.Source.Component == "npd" && event.InvolvedObject.Kind == "Node" {
					// Only create a reconcile request for the "IMDSCheckFailed" event.
					if event.Reason == "IMDSCheckFailed" {
						return []reconcile.Request{
							{
								NamespacedName: types.NamespacedName{
									Name: event.InvolvedObject.Name,
								},
							},
						}
					}
				}
				return nil
			}),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 20, CacheSyncTimeout: time.Hour}).
		Complete(r)
}
