/*
Copyright 2015 The Kubernetes Authors.

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

package oci

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	errors2 "github.com/pkg/errors"
	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	discoveryinformers "k8s.io/client-go/informers/discovery/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	cloudprovider "k8s.io/cloud-provider"
	servicehelper "k8s.io/cloud-provider/service/helpers"
	"k8s.io/component-base/featuregate"
	"k8s.io/component-base/metrics"
	"k8s.io/klog/v2"
)

const (
	// Interval of synchronizing service status from apiserver
	serviceSyncPeriod = 30 * time.Second
	// Interval of synchronizing node status from apiserver
	nodeSyncPeriod = 100 * time.Second

	// How long to wait before retrying the processing of a service change.
	// If this changes, the sleep in hack/jenkins/e2e.sh before downing a cluster
	// should be changed appropriately.
	minRetryDelay = 5 * time.Second
	maxRetryDelay = 300 * time.Second
	// ToBeDeletedTaint is a taint used by the CLuster Autoscaler before marking a node for deletion. Defined in
	// https://github.com/kubernetes/autoscaler/blob/e80ab518340f88f364fe3ef063f8303755125971/cluster-autoscaler/utils/deletetaint/delete.go#L36
	ToBeDeletedTaint = "ToBeDeletedByClusterAutoscaler"

	// subSystemName is the name of this subsystem name used for prometheus metrics.
	subSystemName = "service_controller"
)

var register sync.Once

// registerMetrics registers service-controller metrics.
func registerMetrics() {
	register.Do(func() {
		// Removing the metric registry as it is registered from upstream
		// at vendor/k8s.io/cloud-provider/controllers/service/metrics.go:37
		// Created https://jira.oci.oraclecorp.com/browse/SKE-4577 to clean-up this code
		// legacyregistry.MustRegister(nodeSyncLatency)
		// legacyregistry.MustRegister(updateLoadBalancerHostLatency)
	})
}

var (
	nodeSyncLatency = metrics.NewHistogram(&metrics.HistogramOpts{
		Name:      "nodesync_latency_seconds",
		Subsystem: subSystemName,
		Help:      "A metric measuring the latency for nodesync which updates loadbalancer hosts on cluster node updates.",
		// Buckets from 1s to 16384s
		Buckets:        metrics.ExponentialBuckets(1, 2, 15),
		StabilityLevel: metrics.ALPHA,
	})

	updateLoadBalancerHostLatency = metrics.NewHistogram(&metrics.HistogramOpts{
		Name:      "update_loadbalancer_host_latency_seconds",
		Subsystem: subSystemName,
		Help:      "A metric measuring the latency for updating each load balancer hosts.",
		// Buckets from 1s to 16384s
		Buckets:        metrics.ExponentialBuckets(1, 2, 15),
		StabilityLevel: metrics.ALPHA,
	})
)

type cachedService struct {
	// The cached state of the service
	state *v1.Service
}

type serviceCache struct {
	mu         sync.RWMutex // protects serviceMap
	serviceMap map[string]*cachedService
}

// This controller is based on https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/cloud-provider/controllers/service/controller.go

// ServiceController keeps cloud provider service resources
// (like load balancers) in sync with the registry.
type ServiceController struct {
	cloud                           cloudprovider.Interface
	knownHosts                      []*v1.Node
	servicesToUpdate                sets.String
	kubeClient                      clientset.Interface
	balancer                        cloudprovider.LoadBalancer
	cache                           *serviceCache
	serviceLister                   corelisters.ServiceLister
	serviceListerSynced             cache.InformerSynced
	eventBroadcaster                record.EventBroadcaster
	eventRecorder                   record.EventRecorder
	nodeLister                      corelisters.NodeLister
	nodeListerSynced                cache.InformerSynced
	endpointSliceListerSynced       cache.InformerSynced
	endpointSliceUpdatesBatchPeriod time.Duration
	clusterName                     string
	// services that need to be synced
	queue workqueue.RateLimitingInterface

	// nodeSyncLock ensures there is only one instance of triggerNodeSync getting executed at one time
	// and protects internal states (needFullSync) of nodeSync
	nodeSyncLock sync.Mutex
	// nodeSyncCh triggers nodeSyncLoop to run
	nodeSyncCh chan interface{}
	// needFullSync indicates if the nodeSyncInternal will do a full node sync on all LB services.
	needFullSync bool

	// Feature gate for mixed cluster support in CCM
	mixedClustersEnabled bool
}

// getSpamKey builds unique event key based on source, involvedObject
func getNewSpamKey(event *v1.Event) string {
	return strings.Join([]string{
		event.Source.Component,
		event.Source.Host,
		event.InvolvedObject.Kind,
		event.InvolvedObject.Namespace,
		event.InvolvedObject.Name,
		string(event.InvolvedObject.UID),
		event.InvolvedObject.APIVersion,
		event.Type,
		event.Reason,
	},
		"")
}

// NewServiceController returns a new service controller to keep cloud provider service resources
// (like load balancers) in sync with the registry.
func NewServiceController(
	cloud cloudprovider.Interface,
	kubeClient clientset.Interface,
	serviceInformer coreinformers.ServiceInformer,
	nodeInformer coreinformers.NodeInformer,
	endpointSliceInformer discoveryinformers.EndpointSliceInformer,
	clusterName string,
	endpointSliceUpdatesBatchPeriod time.Duration,
	featureGate featuregate.FeatureGate,
) (*ServiceController, error) {
	opts := record.CorrelatorOptions{SpamKeyFunc: getNewSpamKey}
	broadcaster := record.NewBroadcasterWithCorrelatorOptions(opts)
	broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "service-controller"})

	registerMetrics()
	s := &ServiceController{
		cloud:                           cloud,
		knownHosts:                      []*v1.Node{},
		kubeClient:                      kubeClient,
		cache:                           &serviceCache{serviceMap: make(map[string]*cachedService)},
		eventBroadcaster:                broadcaster,
		eventRecorder:                   recorder,
		endpointSliceUpdatesBatchPeriod: endpointSliceUpdatesBatchPeriod,
		clusterName:                     clusterName,
		queue:                           workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(minRetryDelay, maxRetryDelay), "service"),
		// nodeSyncCh has a size 1 buffer. Only one pending sync signal would be cached.
		nodeSyncCh:           make(chan interface{}, 1),
		mixedClustersEnabled: GetIsFeatureEnabledFromEnv(zap.L().Sugar(), "ENABLE_MIXED_CLUSTERS_SUPPORT", false),
	}

	serviceInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(cur interface{}) {
				svc, ok := cur.(*v1.Service)
				// Check cleanup here can provide a remedy when controller failed to handle
				// changes before it exiting (e.g. crashing, restart, etc.).
				if ok && (wantsLoadBalancer(svc) || needsCleanup(svc)) {
					s.enqueueService(cur)
				}
			},
			UpdateFunc: func(old, cur interface{}) {
				oldSvc, ok1 := old.(*v1.Service)
				curSvc, ok2 := cur.(*v1.Service)
				if ok1 && ok2 && (s.needsUpdate(oldSvc, curSvc) || needsCleanup(curSvc)) {
					s.enqueueService(cur)
				}
			},
			// No need to handle deletion event because the deletion would be handled by
			// the update path when the deletion timestamp is added.
		},
		serviceSyncPeriod,
	)
	s.serviceLister = serviceInformer.Lister()
	s.serviceListerSynced = serviceInformer.Informer().HasSynced

	nodeInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(cur interface{}) {
				s.triggerNodeSync()
			},
			UpdateFunc: func(old, cur interface{}) {
				oldNode, ok := old.(*v1.Node)
				if !ok {
					return
				}

				curNode, ok := cur.(*v1.Node)
				if !ok {
					return
				}

				if !shouldSyncNode(oldNode, curNode) {
					return
				}

				s.triggerNodeSync()
			},
			DeleteFunc: func(old interface{}) {
				s.triggerNodeSync()
			},
		},
		time.Duration(0),
	)
	s.nodeLister = nodeInformer.Lister()
	s.nodeListerSynced = nodeInformer.Informer().HasSynced

	// Use shared informer to listen to endpointSlice events
	endpointSliceInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(cur interface{}) {
				endpointSlice, ok := cur.(*discovery.EndpointSlice)
				if !ok {
					return
				}
				s.enqueueServiceForEndpointSliceUpdate(endpointSlice, nil)
			},
			UpdateFunc: func(old, cur interface{}) {
				oldEndpointSlice, ok := old.(*discovery.EndpointSlice)
				if !ok {
					return
				}
				curEndpointSlice := cur.(*discovery.EndpointSlice)
				if !ok {
					return
				}
				// EndpointSlice generation does not change when labels change. Although the
				// controller will never change LabelServiceName, users might. This check
				// ensures that we handle changes to this label.
				// Refer: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/endpointslice/endpointslice_controller.go
				curSvcName := curEndpointSlice.Labels[discovery.LabelServiceName]
				oldSvcName := oldEndpointSlice.Labels[discovery.LabelServiceName]
				if curSvcName != oldSvcName {
					s.enqueueServiceForEndpointSliceUpdate(curEndpointSlice, nil)
					s.enqueueServiceForEndpointSliceUpdate(oldEndpointSlice, nil)
					return
				}
				if endpointsChanged(curEndpointSlice.Endpoints, oldEndpointSlice.Endpoints) {
					s.enqueueServiceForEndpointSliceUpdate(curEndpointSlice, oldEndpointSlice)
				}
			},
			DeleteFunc: func(old interface{}) {
				endpointSlice, ok := old.(*discovery.EndpointSlice)
				if !ok {
					return
				}
				s.enqueueServiceForEndpointSliceUpdate(endpointSlice, nil)
			},
		},
	)
	s.endpointSliceListerSynced = endpointSliceInformer.Informer().HasSynced

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

// needFullSyncAndUnmark returns the value and needFullSync and marks the field to false.
func (s *ServiceController) needFullSyncAndUnmark() bool {
	s.nodeSyncLock.Lock()
	defer s.nodeSyncLock.Unlock()
	ret := s.needFullSync
	s.needFullSync = false
	return ret
}

// obj could be an *v1.Service, or a DeletionFinalStateUnknown marker item.
func (s *ServiceController) enqueueService(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", obj, err))
		return
	}
	s.queue.Add(key)
}
func (s *ServiceController) enqueueServiceForEndpointSliceUpdate(targetEndpointSlice, oldEndpointSlice *discovery.EndpointSlice) {
	/* TODO: Keeping Support for Mixed Clusters behind feature gate since there some uncovered edge cases.
	 *  One rare edge case is possible where the VN is deleted before the endpoint is deleted,
	 *  in this case get node would fail for the corresponding endpoint and service enqueue won't happen
	 */
	if s.mixedClustersEnabled {
		nodesMap, err := GetNodesMap(s.nodeLister)
		if err != nil {
			runtime.HandleError(fmt.Errorf("couldn't determine nodes in the cluster: %v", err))
		}
		targetVirtualPodExists := s.virtualPodExistsInEndpointSlice(targetEndpointSlice, nodesMap)
		if oldEndpointSlice != nil {
			oldVirtualPodExists := s.virtualPodExistsInEndpointSlice(oldEndpointSlice, nodesMap)

			// Virtual pods neither exist before nor do they exist now
			if !targetVirtualPodExists && !oldVirtualPodExists {
				return
			}
		} else if !targetVirtualPodExists {
			return
		}
	} else {
		virtualNodeExists, err := VirtualNodeExists(s.nodeLister)
		if err != nil {
			runtime.HandleError(fmt.Errorf("couldn't determine if a virtual node exists in the cluster: %v", err))
			return
		}
		if !virtualNodeExists {
			// We don't want to queue a service that fronts non-virtual pods when there is an endpointslice event,
			// only virtual pods are added as backends
			return
		}
	}
	s.enqueueServiceForEndpointSlice(targetEndpointSlice)
}

// enqueueServiceForEndpointSlice attempts to queue the corresponding Service for
// the provided EndpointSlice.
func (s *ServiceController) enqueueServiceForEndpointSlice(endpointSlice *discovery.EndpointSlice) {
	serviceName, ok := endpointSlice.Labels[discovery.LabelServiceName]
	if !ok || serviceName == "" {
		runtime.HandleError(fmt.Errorf("couldn't get service name for EndpointSlice %s", endpointSlice.Name))
		return
	}
	service, err := s.serviceLister.Services(endpointSlice.Namespace).Get(serviceName)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get service for EndpointSlice %s: %v", endpointSlice.Name, err))
		return
	}
	if !wantsLoadBalancer(service) {
		return
	}

	key := fmt.Sprintf("%s/%s", service.Namespace, serviceName)
	// queue after delay of endpointUpdatesBatchPeriod
	// this effectively batches endpointslice updates
	s.queue.AddAfter(key, s.endpointSliceUpdatesBatchPeriod)
}

// virtualPodExistsInEndpointSlice returns true if the given endpointSlice contains a pod
// that is scheduled on a virtual node
func (s *ServiceController) virtualPodExistsInEndpointSlice(endpointSlice *discovery.EndpointSlice, nodesMap map[string]*v1.Node) bool {
	limitLog := false
	for _, e := range endpointSlice.Endpoints {
		if e.TargetRef == nil {
			b, err := json.Marshal(e)
			if err != nil {
				runtime.HandleError(fmt.Errorf("target ref was nil for an Endpoint which could not be parsed"))
			}
			runtime.HandleError(fmt.Errorf("target ref was nil for Endpoint: %s", string(b)))
			continue
		}
		if e.TargetRef.Kind == "Pod" {
			if nodeName := e.NodeName; nodeName != nil {
				node, exist := nodesMap[*nodeName]
				if !exist {
					if !limitLog && e.Hostname != nil {
						runtime.HandleError(fmt.Errorf("node object does not exist: %s. Pod with hostname %s is probably unschedulable", *nodeName, *e.Hostname))
						limitLog = true
					}
					continue
				}

				if IsVirtualNode(node) {
					return true
				}
			}
		}
	}
	return false
}

// Run starts a background goroutine that watches for changes to services that
// have (or had) LoadBalancers=true and ensures that they have
// load balancers created and deleted appropriately.
// serviceSyncPeriod controls how often we check the cluster's services to
// ensure that the correct load balancers exist.
// nodeSyncPeriod controls how often we check the cluster's nodes to determine
// if load balancers need to be updated to point to a new set.
//
// It's an error to call Run() more than once for a given ServiceController
// object.
func (s *ServiceController) Run(ctx context.Context, workers int) {
	defer runtime.HandleCrash()
	defer s.queue.ShutDown()

	klog.Info("Starting OCI service controller")
	defer klog.Info("Shutting down OCI service controller")

	if !cache.WaitForNamedCacheSync("oci-service", ctx.Done(), s.serviceListerSynced, s.nodeListerSynced, s.endpointSliceListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, time.Second)
	}

	go s.nodeSyncLoop(ctx, workers)
	go wait.Until(s.triggerNodeSync, nodeSyncPeriod, ctx.Done())

	<-ctx.Done()
}

// triggerNodeSync triggers a nodeSync asynchronously
func (s *ServiceController) triggerNodeSync() {
	s.nodeSyncLock.Lock()
	defer s.nodeSyncLock.Unlock()
	newHosts, err := listWithPredicate(s.nodeLister, s.getNodeConditionPredicate())
	if err != nil {
		runtime.HandleError(fmt.Errorf("failed to retrieve current set of nodes from node lister: %v", err))
		// if node list cannot be retrieve, trigger full node sync to be safe.
		s.needFullSync = true
	} else if !nodeSlicesEqualForLB(newHosts, s.knownHosts) {
		// Here the last known state is recorded as knownHosts. For each
		// LB update, the latest node list is retrieved. This is to prevent
		// a stale set of nodes were used to be update loadbalancers when
		// there are many loadbalancers in the clusters. nodeSyncInternal
		// would be triggered until all loadbalancers are updated to the new state.
		klog.V(2).Infof("Node changes detected, triggering a full node sync on all loadbalancer services")
		s.needFullSync = true
		s.knownHosts = newHosts
	}

	select {
	case s.nodeSyncCh <- struct{}{}:
		klog.V(4).Info("Triggering nodeSync")
		return
	default:
		klog.V(4).Info("A pending nodeSync is already in queue")
		return
	}
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (s *ServiceController) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

// nodeSyncLoop takes nodeSync signal and triggers nodeSync
func (s *ServiceController) nodeSyncLoop(ctx context.Context, workers int) {
	klog.V(4).Info("nodeSyncLoop Started")
	for range s.nodeSyncCh {
		klog.V(4).Info("nodeSync has been triggered")
		s.nodeSyncInternal(ctx, workers)
	}
	klog.V(2).Info("s.nodeSyncCh is closed. Exiting nodeSyncLoop")
}

func (s *ServiceController) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	err := s.syncService(ctx, key.(string))
	if err == nil {
		s.queue.Forget(key)
		return true
	}

	runtime.HandleError(fmt.Errorf("error processing service %v (will retry): %v", key, err))
	s.queue.AddRateLimited(key)
	return true
}

func (s *ServiceController) init() error {
	if s.cloud == nil {
		return fmt.Errorf("WARNING: no cloud provider provided, services of type LoadBalancer will fail")
	}

	balancer, ok := s.cloud.LoadBalancer()
	if !ok {
		return fmt.Errorf("the cloud provider does not support external load balancers")
	}
	s.balancer = balancer

	return nil
}

// processServiceCreateOrUpdate operates loadbalancers for the incoming service accordingly.
// Returns an error if processing the service update failed.
func (s *ServiceController) processServiceCreateOrUpdate(ctx context.Context, service *v1.Service, key string) error {
	cachedService := s.cache.getOrCreate(key)
	if cachedService.state != nil && cachedService.state.UID != service.UID {
		// This happens only when a service is deleted and re-created
		// in a short period, which is only possible when it doesn't
		// contain finalizer.
		if err := s.processLoadBalancerDelete(ctx, cachedService.state, key); err != nil {
			return err
		}
	}
	// Always cache the service, we need the info for service deletion in case
	// when load balancer cleanup is not handled via finalizer.
	cachedService.state = service
	op, err := s.syncLoadBalancerIfNeeded(ctx, service, key)
	if err != nil {
		if !errors2.Is(err, LbOperationAlreadyExists) {
			s.eventRecorder.Eventf(service, v1.EventTypeWarning, "SyncLoadBalancerFailed", "Error syncing load balancer: %v", err)
		}
		return err
	}
	if op == deleteLoadBalancer {
		// Only delete the cache upon successful load balancer deletion.
		s.cache.delete(key)
	}

	return nil
}

type loadBalancerOperation int

const (
	deleteLoadBalancer loadBalancerOperation = iota
	ensureLoadBalancer
)

// syncLoadBalancerIfNeeded ensures that service's status is synced up with loadbalancer
// i.e. creates loadbalancer for service if requested and deletes loadbalancer if the service
// doesn't want a loadbalancer no more. Returns whatever error occurred.
func (s *ServiceController) syncLoadBalancerIfNeeded(ctx context.Context, service *v1.Service, key string) (loadBalancerOperation, error) {
	// Note: It is safe to just call EnsureLoadBalancer.  But, on some clouds that requires a delete & create,
	// which may involve service interruption.  Also, we would like user-friendly events.

	// Save the state so we can avoid a write if it doesn't change
	previousStatus := service.Status.LoadBalancer.DeepCopy()
	var newStatus *v1.LoadBalancerStatus
	var op loadBalancerOperation
	var err error

	if !wantsLoadBalancer(service) || needsCleanup(service) {
		// Delete the load balancer if service no longer wants one, or if service needs cleanup.
		op = deleteLoadBalancer
		newStatus = &v1.LoadBalancerStatus{}
		_, exists, err := s.balancer.GetLoadBalancer(ctx, s.clusterName, service)
		if err != nil {
			return op, fmt.Errorf("failed to check if load balancer exists before cleanup: %v", err)
		}
		if exists {
			klog.V(2).Infof("Deleting existing load balancer for service %s", key)
			s.eventRecorder.Event(service, v1.EventTypeNormal, "DeletingLoadBalancer", "Deleting load balancer")
			if err := s.balancer.EnsureLoadBalancerDeleted(ctx, s.clusterName, service); err != nil {
				if err == cloudprovider.ImplementedElsewhere {
					klog.V(4).Infof("LoadBalancer for service %s implemented by a different controller %s, Ignoring error on deletion", key, s.cloud.ProviderName())
				} else {
					return op, fmt.Errorf("failed to delete load balancer: %v", err)
				}
			}
		}
		// Always remove finalizer when load balancer is deleted, this ensures Services
		// can be deleted after all corresponding load balancer resources are deleted.
		if err := s.removeFinalizer(service); err != nil {
			return op, fmt.Errorf("failed to remove load balancer cleanup finalizer: %v", err)
		}
		s.eventRecorder.Event(service, v1.EventTypeNormal, "DeletedLoadBalancer", "Deleted load balancer")
	} else {
		// Create or update the load balancer if service wants one.
		op = ensureLoadBalancer
		klog.V(2).Infof("Ensuring load balancer for service %s", key)
		s.eventRecorder.Event(service, v1.EventTypeNormal, "EnsuringLoadBalancer", "Ensuring load balancer")
		// Always add a finalizer prior to creating load balancers, this ensures Services
		// can't be deleted until all corresponding load balancer resources are also deleted.
		if err := s.addFinalizer(service); err != nil {
			return op, fmt.Errorf("failed to add load balancer cleanup finalizer: %v", err)
		}
		newStatus, err = s.ensureLoadBalancer(ctx, service)
		if err != nil {
			if err == cloudprovider.ImplementedElsewhere {
				// ImplementedElsewhere indicates that the ensureLoadBalancer is a nop and the
				// functionality is implemented by a different controller.  In this case, we
				// return immediately without doing anything.
				klog.V(4).Infof("LoadBalancer for service %s implemented by a different controller %s, Ignoring error", key, s.cloud.ProviderName())
				return op, nil
			}
			return op, fmt.Errorf("failed to ensure load balancer: %v", err)
		}
		if newStatus == nil {
			return op, fmt.Errorf("service status returned by EnsureLoadBalancer is nil")
		}

		s.eventRecorder.Event(service, v1.EventTypeNormal, "EnsuredLoadBalancer", "Ensured load balancer")
	}

	if err := s.patchStatus(service, previousStatus, newStatus); err != nil {
		// Only retry error that isn't not found:
		// - Not found error mostly happens when service disappears right after
		//   we remove the finalizer.
		// - We can't patch status on non-exist service anyway.
		if !errors.IsNotFound(err) {
			return op, fmt.Errorf("failed to update load balancer status: %v", err)
		}
	}

	return op, nil
}

func (s *ServiceController) ensureLoadBalancer(ctx context.Context, service *v1.Service) (*v1.LoadBalancerStatus, error) {
	nodes, err := listWithPredicate(s.nodeLister, s.getNodeConditionPredicate())
	if err != nil {
		return nil, err
	}

	// If there are no available nodes for LoadBalancer service, make a EventTypeWarning event for it.
	if len(nodes) == 0 {
		s.eventRecorder.Event(service, v1.EventTypeWarning, "UnAvailableLoadBalancer", "There are no available provisioned nodes for LoadBalancer")
	}

	// - Only one protocol supported per service
	// - Not all cloud providers support all protocols and the next step is expected to return
	//   an error for unsupported protocols
	return s.balancer.EnsureLoadBalancer(ctx, s.clusterName, service, nodes)
}

// ListKeys implements the interface required by DeltaFIFO to list the keys we
// already know about.
func (s *serviceCache) ListKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.serviceMap))
	for k := range s.serviceMap {
		keys = append(keys, k)
	}
	return keys
}

// GetByKey returns the value stored in the serviceMap under the given key
func (s *serviceCache) GetByKey(key string) (interface{}, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.serviceMap[key]; ok {
		return v, true, nil
	}
	return nil, false, nil
}

// ListKeys implements the interface required by DeltaFIFO to list the keys we
// already know about.
func (s *serviceCache) allServices() []*v1.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	services := make([]*v1.Service, 0, len(s.serviceMap))
	for _, v := range s.serviceMap {
		services = append(services, v.state)
	}
	return services
}

func (s *serviceCache) get(serviceName string) (*cachedService, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	service, ok := s.serviceMap[serviceName]
	return service, ok
}

func (s *serviceCache) getOrCreate(serviceName string) *cachedService {
	s.mu.Lock()
	defer s.mu.Unlock()
	service, ok := s.serviceMap[serviceName]
	if !ok {
		service = &cachedService{}
		s.serviceMap[serviceName] = service
	}
	return service
}

func (s *serviceCache) set(serviceName string, service *cachedService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceMap[serviceName] = service
}

func (s *serviceCache) delete(serviceName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceMap, serviceName)
}

// needsCleanup checks if load balancer needs to be cleaned up as indicated by finalizer.
func needsCleanup(service *v1.Service) bool {
	if !servicehelper.HasLBFinalizer(service) {
		return false
	}

	if service.ObjectMeta.DeletionTimestamp != nil {
		return true
	}

	// Service doesn't want loadBalancer but owns loadBalancer finalizer also need to be cleaned up.
	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		return true
	}

	return false
}

// needsUpdate checks if load balancer needs to be updated due to change in attributes.
func (s *ServiceController) needsUpdate(oldService *v1.Service, newService *v1.Service) bool {
	if !wantsLoadBalancer(oldService) && !wantsLoadBalancer(newService) {
		return false
	}
	if wantsLoadBalancer(oldService) != wantsLoadBalancer(newService) {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "Type", "%v -> %v",
			oldService.Spec.Type, newService.Spec.Type)
		return true
	}

	if wantsLoadBalancer(newService) && !reflect.DeepEqual(oldService.Spec.LoadBalancerSourceRanges, newService.Spec.LoadBalancerSourceRanges) {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "LoadBalancerSourceRanges", "%v -> %v",
			oldService.Spec.LoadBalancerSourceRanges, newService.Spec.LoadBalancerSourceRanges)
		return true
	}

	if !portsEqualForLB(oldService, newService) || oldService.Spec.SessionAffinity != newService.Spec.SessionAffinity {
		return true
	}

	if !reflect.DeepEqual(oldService.Spec.SessionAffinityConfig, newService.Spec.SessionAffinityConfig) {
		return true
	}
	if !loadBalancerIPsAreEqual(oldService, newService) {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "LoadbalancerIP", "%v -> %v",
			oldService.Spec.LoadBalancerIP, newService.Spec.LoadBalancerIP)
		return true
	}
	if len(oldService.Spec.ExternalIPs) != len(newService.Spec.ExternalIPs) {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "ExternalIP", "Count: %v -> %v",
			len(oldService.Spec.ExternalIPs), len(newService.Spec.ExternalIPs))
		return true
	}
	for i := range oldService.Spec.ExternalIPs {
		if oldService.Spec.ExternalIPs[i] != newService.Spec.ExternalIPs[i] {
			s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "ExternalIP", "Added: %v",
				newService.Spec.ExternalIPs[i])
			return true
		}
	}
	if !reflect.DeepEqual(oldService.Annotations, newService.Annotations) {
		return true
	}
	if oldService.UID != newService.UID {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "UID", "%v -> %v",
			oldService.UID, newService.UID)
		return true
	}
	if oldService.Spec.ExternalTrafficPolicy != newService.Spec.ExternalTrafficPolicy {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "ExternalTrafficPolicy", "%v -> %v",
			oldService.Spec.ExternalTrafficPolicy, newService.Spec.ExternalTrafficPolicy)
		return true
	}
	if oldService.Spec.HealthCheckNodePort != newService.Spec.HealthCheckNodePort {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "HealthCheckNodePort", "%v -> %v",
			oldService.Spec.HealthCheckNodePort, newService.Spec.HealthCheckNodePort)
		return true
	}

	return false
}

func getPortsForLB(service *v1.Service) []*v1.ServicePort {
	ports := []*v1.ServicePort{}
	for i := range service.Spec.Ports {
		sp := &service.Spec.Ports[i]
		ports = append(ports, sp)
	}
	return ports
}

func portsEqualForLB(x, y *v1.Service) bool {
	xPorts := getPortsForLB(x)
	yPorts := getPortsForLB(y)
	return portSlicesEqualForLB(xPorts, yPorts)
}

func portSlicesEqualForLB(x, y []*v1.ServicePort) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if !portEqualForLB(x[i], y[i]) {
			return false
		}
	}
	return true
}

func portEqualForLB(x, y *v1.ServicePort) bool {
	if x.Name != y.Name {
		return false
	}

	if x.Protocol != y.Protocol {
		return false
	}

	if x.Port != y.Port {
		return false
	}

	if x.NodePort != y.NodePort {
		return false
	}

	if x.TargetPort != y.TargetPort {
		return false
	}

	return true
}

func nodeNames(nodes []*v1.Node) sets.String {
	ret := sets.NewString()
	for _, node := range nodes {
		ret.Insert(node.Name)
	}
	return ret
}

func nodeSlicesEqualForLB(x, y []*v1.Node) bool {
	if len(x) != len(y) {
		return false
	}
	return nodeNames(x).Equal(nodeNames(y))
}

func (s *ServiceController) getNodeConditionPredicate() NodeConditionPredicate {
	return func(node *v1.Node) bool {
		if IsVirtualNode(node) {
			// We don't want to trigger a reconciliation for Virtual Node add/update/delete
			return false
		}

		if _, hasExcludeBalancerLabel := node.Labels[v1.LabelNodeExcludeBalancers]; hasExcludeBalancerLabel {
			return false
		}

		// Remove nodes that are about to be deleted by the cluster autoscaler.
		for _, taint := range node.Spec.Taints {
			if taint.Key == ToBeDeletedTaint {
				klog.V(4).Infof("Ignoring node %v with autoscaler taint %+v", node.Name, taint)
				return false
			}
		}

		// If we have no info, don't accept
		if len(node.Status.Conditions) == 0 {
			return false
		}
		for _, cond := range node.Status.Conditions {
			// We consider the node for load balancing only when its NodeReady condition status
			// is ConditionTrue
			if cond.Type == v1.NodeReady && cond.Status != v1.ConditionTrue {
				klog.V(4).Infof("Ignoring node %v with %v condition status %v", node.Name, cond.Type, cond.Status)
				return false
			}
		}
		return true
	}
}

func shouldSyncNode(oldNode, newNode *v1.Node) bool {
	if oldNode.Spec.Unschedulable != newNode.Spec.Unschedulable {
		return true
	}

	if !reflect.DeepEqual(oldNode.Labels, newNode.Labels) {
		return true
	}

	return nodeReadyConditionStatus(oldNode) != nodeReadyConditionStatus(newNode)
}

func nodeReadyConditionStatus(node *v1.Node) v1.ConditionStatus {
	for _, condition := range node.Status.Conditions {
		if condition.Type != v1.NodeReady {
			continue
		}

		return condition.Status
	}

	return ""
}

// nodeSyncInternal handles updating the hosts pointed to by all load
// balancers whenever the set of nodes in the cluster changes.
func (s *ServiceController) nodeSyncInternal(ctx context.Context, workers int) {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Seconds()
		klog.V(4).Infof("It took %v seconds to finish nodeSyncInternal", latency)
		nodeSyncLatency.Observe(latency)
	}()

	if !s.needFullSyncAndUnmark() {
		// The set of nodes in the cluster hasn't changed, but we can retry
		// updating any services that we failed to update last time around.
		// It is required to call `s.cache.get()` on each Service in case there was
		// an update event that occurred between retries.
		var servicesToUpdate []*v1.Service
		for key := range s.servicesToUpdate {
			cachedService, exist := s.cache.get(key)
			if !exist {
				klog.Errorf("Service %q should be in the cache but not", key)
				continue
			}
			servicesToUpdate = append(servicesToUpdate, cachedService.state)
		}

		s.servicesToUpdate = s.updateLoadBalancerHosts(ctx, servicesToUpdate, workers)
		return
	}
	klog.V(2).Infof("Syncing backends for all LB services.")

	// Try updating all services, and save the failed ones to try again next
	// round.
	servicesToUpdate := s.cache.allServices()
	numServices := len(servicesToUpdate)
	s.servicesToUpdate = s.updateLoadBalancerHosts(ctx, servicesToUpdate, workers)
	klog.V(2).Infof("Successfully updated %d out of %d load balancers to direct traffic to the updated set of nodes",
		numServices-len(s.servicesToUpdate), numServices)
}

// nodeSyncService syncs the nodes for one load balancer type service
func (s *ServiceController) nodeSyncService(svc *v1.Service) bool {
	if svc == nil || !wantsLoadBalancer(svc) {
		return false
	}
	klog.V(4).Infof("nodeSyncService started for service %s/%s", svc.Namespace, svc.Name)
	hosts, err := listWithPredicate(s.nodeLister, s.getNodeConditionPredicate())
	if err != nil {
		runtime.HandleError(fmt.Errorf("failed to retrieve node list: %v", err))
		return true
	}

	if err := s.lockedUpdateLoadBalancerHosts(svc, hosts); err != nil {
		runtime.HandleError(fmt.Errorf("failed to update load balancer hosts for service %s/%s: %v", svc.Namespace, svc.Name, err))
		return true
	}
	klog.V(4).Infof("nodeSyncService finished successfully for service %s/%s", svc.Namespace, svc.Name)
	return false
}

// updateLoadBalancerHosts updates all existing load balancers so that
// they will match the latest list of nodes with input number of workers.
// Returns the list of services that couldn't be updated.
func (s *ServiceController) updateLoadBalancerHosts(ctx context.Context, services []*v1.Service, workers int) (servicesToRetry sets.String) {
	klog.V(4).Infof("Running updateLoadBalancerHosts(len(services)==%d, workers==%d)", len(services), workers)

	// lock for servicesToRetry
	servicesToRetry = sets.NewString()
	lock := sync.Mutex{}
	doWork := func(piece int) {
		if shouldRetry := s.nodeSyncService(services[piece]); !shouldRetry {
			return
		}
		lock.Lock()
		defer lock.Unlock()
		key := fmt.Sprintf("%s/%s", services[piece].Namespace, services[piece].Name)
		servicesToRetry.Insert(key)
	}

	workqueue.ParallelizeUntil(ctx, workers, len(services), doWork)
	klog.V(4).Infof("Finished updateLoadBalancerHosts")
	return servicesToRetry
}

// Updates the load balancer of a service, assuming we hold the mutex
// associated with the service.
func (s *ServiceController) lockedUpdateLoadBalancerHosts(service *v1.Service, hosts []*v1.Node) error {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Seconds()
		klog.V(4).Infof("It took %v seconds to update load balancer hosts for service %s/%s", latency, service.Namespace, service.Name)
		updateLoadBalancerHostLatency.Observe(latency)
	}()

	klog.V(2).Infof("Updating backends for load balancer %s/%s with node set: %v", service.Namespace, service.Name, nodeNames(hosts))
	// This operation doesn't normally take very long (and happens pretty often), so we only record the final event
	err := s.balancer.UpdateLoadBalancer(context.TODO(), s.clusterName, service, hosts)
	if err == nil {
		// If there are no available nodes for LoadBalancer service, make a EventTypeWarning event for it.
		if len(hosts) == 0 {
			s.eventRecorder.Event(service, v1.EventTypeWarning, "UnAvailableLoadBalancer", "There are no available provisioned nodes for LoadBalancer")
		} else {
			s.eventRecorder.Event(service, v1.EventTypeNormal, "UpdatedLoadBalancer", "Updated load balancer with new hosts")
		}
		return nil
	}
	if err == cloudprovider.ImplementedElsewhere {
		// ImplementedElsewhere indicates that the UpdateLoadBalancer is a nop and the
		// functionality is implemented by a different controller.  In this case, we
		// return immediately without doing anything.
		return nil
	}
	// It's only an actual error if the load balancer still exists.
	if _, exists, err := s.balancer.GetLoadBalancer(context.TODO(), s.clusterName, service); err != nil {
		runtime.HandleError(fmt.Errorf("failed to check if load balancer exists for service %s/%s: %v", service.Namespace, service.Name, err))
	} else if !exists {
		return nil
	}

	if !errors2.Is(err, LbOperationAlreadyExists) {
		s.eventRecorder.Eventf(service, v1.EventTypeWarning, "UpdateLoadBalancerFailed", "Error updating load balancer with new hosts %v: %v", nodeNames(hosts), err)
	}
	return err
}

func wantsLoadBalancer(service *v1.Service) bool {
	// if LoadBalancerClass is set, the user does not want the default cloud-provider Load Balancer
	return service.Spec.Type == v1.ServiceTypeLoadBalancer && service.Spec.LoadBalancerClass == nil
}

func loadBalancerIPsAreEqual(oldService, newService *v1.Service) bool {
	return oldService.Spec.LoadBalancerIP == newService.Spec.LoadBalancerIP
}

// syncService will sync the Service with the given key if it has had its expectations fulfilled,
// meaning it did not expect to see any more of its pods created or deleted. This function is not meant to be
// invoked concurrently with the same key.
func (s *ServiceController) syncService(ctx context.Context, key string) error {
	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing service %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// service holds the latest service info from apiserver
	service, err := s.serviceLister.Services(namespace).Get(name)
	switch {
	case errors.IsNotFound(err):
		// service absence in store means watcher caught the deletion, ensure LB info is cleaned
		err = s.processServiceDeletion(ctx, key)
	case err != nil:
		runtime.HandleError(fmt.Errorf("Unable to retrieve service %v from store: %v", key, err))
	default:
		err = s.processServiceCreateOrUpdate(ctx, service, key)
	}

	return err
}

func (s *ServiceController) processServiceDeletion(ctx context.Context, key string) error {
	cachedService, ok := s.cache.get(key)
	if !ok {
		// Cache does not contains the key means:
		// - We didn't create a Load Balancer for the deleted service at all.
		// - We already deleted the Load Balancer that was created for the service.
		// In both cases we have nothing left to do.
		return nil
	}
	klog.V(2).Infof("Service %v has been deleted. Attempting to cleanup load balancer resources", key)
	if err := s.processLoadBalancerDelete(ctx, cachedService.state, key); err != nil {
		return err
	}
	s.cache.delete(key)
	return nil
}

func (s *ServiceController) processLoadBalancerDelete(ctx context.Context, service *v1.Service, key string) error {
	// delete load balancer info only if the service type is LoadBalancer
	if !wantsLoadBalancer(service) {
		return nil
	}
	s.eventRecorder.Event(service, v1.EventTypeNormal, "DeletingLoadBalancer", "Deleting load balancer")
	if err := s.balancer.EnsureLoadBalancerDeleted(ctx, s.clusterName, service); err != nil {
		if !errors2.Is(err, LbOperationAlreadyExists) {
			s.eventRecorder.Eventf(service, v1.EventTypeWarning, "DeleteLoadBalancerFailed", "Error deleting load balancer: %v", err)
		}
		return err
	}
	s.eventRecorder.Event(service, v1.EventTypeNormal, "DeletedLoadBalancer", "Deleted load balancer")
	return nil
}

// addFinalizer patches the service to add finalizer.
func (s *ServiceController) addFinalizer(service *v1.Service) error {
	if servicehelper.HasLBFinalizer(service) {
		return nil
	}

	// Make a copy so we don't mutate the shared informer cache.
	updated := service.DeepCopy()
	updated.ObjectMeta.Finalizers = append(updated.ObjectMeta.Finalizers, servicehelper.LoadBalancerCleanupFinalizer)

	klog.V(2).Infof("Adding finalizer to service %s/%s", updated.Namespace, updated.Name)
	_, err := servicehelper.PatchService(s.kubeClient.CoreV1(), service, updated)
	return err
}

// removeFinalizer patches the service to remove finalizer.
func (s *ServiceController) removeFinalizer(service *v1.Service) error {
	if !servicehelper.HasLBFinalizer(service) {
		return nil
	}

	// Make a copy so we don't mutate the shared informer cache.
	updated := service.DeepCopy()
	updated.ObjectMeta.Finalizers = removeString(updated.ObjectMeta.Finalizers, servicehelper.LoadBalancerCleanupFinalizer)

	klog.V(2).Infof("Removing finalizer from service %s/%s", updated.Namespace, updated.Name)
	_, err := servicehelper.PatchService(s.kubeClient.CoreV1(), service, updated)
	return err
}

// removeString returns a newly created []string that contains all items from slice that
// are not equal to s.
func removeString(slice []string, s string) []string {
	var newSlice []string
	for _, item := range slice {
		if item != s {
			newSlice = append(newSlice, item)
		}
	}
	return newSlice
}

// patchStatus patches the service with the given LoadBalancerStatus.
func (s *ServiceController) patchStatus(service *v1.Service, previousStatus, newStatus *v1.LoadBalancerStatus) error {
	if servicehelper.LoadBalancerStatusEqual(previousStatus, newStatus) {
		return nil
	}
	// Make a copy so we don't mutate the shared informer cache.
	updated := service.DeepCopy()
	updated.Status.LoadBalancer = *newStatus

	klog.V(2).Infof("Patching status for service %s/%s", updated.Namespace, updated.Name)
	_, err := servicehelper.PatchService(s.kubeClient.CoreV1(), service, updated)
	return err
}

// NodeConditionPredicate is a function that indicates whether the given node's conditions meet
// some set of criteria defined by the function.
type NodeConditionPredicate func(node *v1.Node) bool

// listWithPredicate gets nodes that matches predicate function.
func listWithPredicate(nodeLister corelisters.NodeLister, predicate NodeConditionPredicate) ([]*v1.Node, error) {
	nodes, err := nodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var filtered []*v1.Node
	for i := range nodes {
		if predicate(nodes[i]) {
			filtered = append(filtered, nodes[i])
		}
	}

	return filtered, nil
}

func endpointsChanged(curEndpoints, oldEndpoints []discovery.Endpoint) bool {
	if len(curEndpoints) != len(oldEndpoints) {
		return true
	}

	endpointsSet := sets.NewString()
	for _, endpoint := range curEndpoints {
		if len(endpoint.Addresses) > 0 {
			endpointsSet.Insert(endpoint.Addresses[0])
		}
	}
	for _, endpoint := range oldEndpoints {
		if len(endpoint.Addresses) > 0 && !endpointsSet.Has(endpoint.Addresses[0]) {
			return true
		}
	}
	return false
}
