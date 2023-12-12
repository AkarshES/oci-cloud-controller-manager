package oci

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
)

const (
	// How long to wait before retrying the pod readiness check for a service
	prcMinRetryDelay = 10 * time.Second
	prcMaxRetryDelay = 300 * time.Second
	// Pod readiness check sync interval
	prcSyncPeriod = 10 * time.Second
	// Pod readiness gate string
	podReadinessConditionPrefix = "podreadiness.loadbalancer.oraclecloud.com"
	// Context timeout
	clientTimeout = 10 * time.Second
)

type PodReadinessController struct {
	kubeClient    clientset.Interface
	recorder      record.EventRecorder
	ociClient     client.Interface
	logger        *zap.SugaredLogger
	metricPusher  *metrics.MetricPusher
	config        *providercfg.Config
	podLister     corelisters.PodLister
	serviceLister corelisters.ServiceLister
	nodeLister    corelisters.NodeLister
	queue         workqueue.RateLimitingInterface
}

func NewPodReadinessController(
	kubeClient clientset.Interface,
	ociClient client.Interface,
	logger *zap.SugaredLogger,
	metricPusher *metrics.MetricPusher,
	config *providercfg.Config,
	serviceLister corelisters.ServiceLister,
	podLister corelisters.PodLister,
	nodeLister corelisters.NodeLister,
) *PodReadinessController {
	controllerName := "pod-readiness-controller"
	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: controllerName})
	eventBroadcaster.StartLogging(klog.Infof)
	if kubeClient != nil {
		logger.Info("Sending events to api server.")
		eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	} else {
		logger.Info("No api server defined - no events will be sent to API server.")
	}

	return &PodReadinessController{
		kubeClient:    kubeClient,
		recorder:      recorder,
		ociClient:     ociClient,
		logger:        logger.With("controller", controllerName),
		podLister:     podLister,
		metricPusher:  metricPusher,
		config:        config,
		serviceLister: serviceLister,
		nodeLister:    nodeLister,
		queue:         workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(prcMinRetryDelay, prcMaxRetryDelay)),
	}
}

func (p *PodReadinessController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer p.queue.ShutDown()

	p.logger.Info("Starting pod readiness controller")

	for i := 0; i < workers; i++ {
		go wait.Until(p.worker, time.Second, stopCh)
	}

	go wait.Until(p.pusher, prcSyncPeriod, stopCh)

	<-stopCh
	p.logger.Info("Stopping pod readiness controller")
}

func (p *PodReadinessController) pusher() {
	services, err := p.serviceLister.List(labels.Everything())
	if err != nil {
		p.logger.With(zap.Error(err)).Error("unable to list services")
		return
	}

	for _, service := range services {
		if !wantsLoadBalancer(service) {
			continue
		}

		key, err := cache.MetaNamespaceKeyFunc(service)
		if err != nil {
			p.logger.With(zap.Error(err)).Errorf("couldn't get key for service %+v", service)
			continue
		}

		p.queue.Add(key)
	}
}

func (p *PodReadinessController) worker() {
	for p.processNextWorkItem() {
	}
}

func (p *PodReadinessController) processNextWorkItem() bool {
	key, quit := p.queue.Get()
	if quit {
		return false
	}
	defer p.queue.Done(key)

	err := p.sync(key.(string))
	if err == nil {
		p.queue.Forget(key)
		return true
	}

	p.logger.With(zap.Error(err)).Errorf("error syncing pod readiness for service %s (will retry)", key.(string))
	p.queue.AddRateLimited(key)
	return true
}

func (p *PodReadinessController) sync(key string) error {
	startTime := time.Now()
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	logger := p.logger.With("serviceNamespace", namespace, "serviceName", name)

	service, err := p.serviceLister.Services(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("service has been deleted, pod readiness sync will stop")
			return nil
		}
		logger.With(zap.Error(err)).Error("failed to get service")
		return err
	}
	if !wantsLoadBalancer(service) {
		logger.Info("service is no longer of type LoadBalancer, pod readiness sync will stop")
		return nil
	}

	podLabelSelector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
	pods, err := p.podLister.Pods(service.Namespace).List(podLabelSelector)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to list pods of service")
		return err
	}

	backendSets, err := p.getBackendSetsNeedSync(service, pods)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get backend sets that need pod readiness sync")
		return err
	}

	if len(backendSets) == 0 {
		logger.Debugf("pods of service %s are ready, pod readiness sync is not needed", service.Name)
		return nil
	}

	lbType := getLoadBalancerType(service)
	lbProvider := p.getLoadBalancerProvider(lbType)
	lbName := GetLoadBalancerName(service)

	logger = logger.With("loadBalancerName", lbName)
	dimensionsMap := make(map[string]string)
	metricName := ""
	if lbType == NLB {
		metricName = metrics.NLBPodReadinessSync
	} else {
		metricName = metrics.LBPodReadinessSync
	}

	lb, err := p.getLoadBalancerByName(lbProvider, p.config.CompartmentID, lbName)
	if err != nil {
		if client.IsNotFound(err) {
			logger.Infof("LB %s for service has been deleted, pod readiness sync will stop", lbName)
			return nil
		}
		logger.With(zap.Error(err)).Error("failed to get loadbalancer by name")
		dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.LoadBalancerType)
		dimensionsMap[metrics.ResourceOCIDDimension] = lbName
		metrics.SendMetricData(p.metricPusher, metricName, time.Since(startTime).Seconds(), dimensionsMap)
		return err
	}

	if lb != nil && lb.Id != nil {
		logger = logger.With("loadBalancerID", *lb.Id)
		dimensionsMap[metrics.ResourceOCIDDimension] = *lb.Id
	}

	for backendSetName, servicePort := range backendSets {
		logger = logger.With("backendSetName", backendSetName)
		backendSet, ok := lb.BackendSets[backendSetName]
		if !ok {
			// Backend set has not been added to LB yet
			logger.Debug("backend set has not been added yet")
			continue
		}

		backendSetHealth, err := p.getBackendSetHealth(lbProvider, *lb.Id, backendSetName)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to get backend set health")
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.LoadBalancerType)
			metrics.SendMetricData(p.metricPusher, metricName, time.Since(startTime).Seconds(), dimensionsMap)
			return err
		}

		backendsMap := getBackendsMap(backendSet)
		podReadinessCondition := getPodReadinessCondition(service.Namespace, service.Name, backendSetName)
		unhealthyBackendMap := getUnhealthyBackendMap(backendSetHealth, podReadinessCondition)

		for _, pod := range pods {
			if !hasReadinessGate(pod, podReadinessCondition) {
				continue
			}

			if _, ok := backendsMap[pod.Status.PodIP]; !ok {
				node, err := p.nodeLister.Get(pod.Spec.NodeName)
				if err != nil {
					logger.Error("could not retrieve node %s", pod.Spec.NodeName)
					continue
				}

				// Check if virtual pod
				if IsVirtualNode(node) {
					// Pod has not been added to the backend set yet
					logger.Debugf("pod has not been added yet:%s", pod.Status.PodIP)
					continue
				}
			}

			backendName := fmt.Sprintf("%s:%d", pod.Status.PodIP, servicePort.NodePort)
			if err := p.ensurePodReadinessCondition(logger, unhealthyBackendMap, backendName, pod, podReadinessCondition); err != nil {
				logger.With(zap.Error(err)).Errorf("failed to ensure pod readiness condition for pod %s", pod.Name)
				dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.LoadBalancerType)
				metrics.SendMetricData(p.metricPusher, metricName, time.Since(startTime).Seconds(), dimensionsMap)
				return err
			}
		}
	}

	logger.Info("Successfully completed pod readiness sync")
	dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.Success, util.LoadBalancerType)
	dimensionsMap[metrics.BackendSetsCountDimension] = strconv.Itoa(len(backendSets))
	metrics.SendMetricData(p.metricPusher, metricName, time.Since(startTime).Seconds(), dimensionsMap)
	return nil
}

func (p *PodReadinessController) ensurePodReadinessCondition(logger *zap.SugaredLogger, backendHealthMap map[string]v1.PodCondition, backendName string, pod *v1.Pod, prConditionType v1.PodConditionType) error {
	logger = logger.With("backendName", backendName, "podNamespace", pod.Namespace, "podName", pod.Name)

	condition, exists := getPodCondition(pod, prConditionType)
	updatedCondition := getUpdatedPodCondition(backendHealthMap, prConditionType, backendName)

	if exists &&
		condition.Status == updatedCondition.Status &&
		condition.Reason == updatedCondition.Reason {
		// no change in condition
		return nil
	}

	if !exists || updatedCondition.Status != condition.Status {
		updatedCondition.LastTransitionTime = metav1.Now()
		if condition.Status != "" {
			updatedCondition.Message = fmt.Sprintf("condition status moved from %s to %s", condition.Status, updatedCondition.Status)
		}
	}

	logger.Debugf("updating pod readiness gate from %+v to %+v", condition, updatedCondition)

	patchBytes, err := buildPodConditionPatch(pod, updatedCondition)
	if err != nil {
		logger.Errorf("unable to build pod readiness condition %+v", updatedCondition)
		return err
	}

	_, err = p.patchPod(pod.Namespace, pod.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}, "status")
	if err != nil {
		logger.Errorf("unable to update pod readiness condition %+v", updatedCondition)
		return err
	}

	logger.Infof("successfully updated pod readiness condition from %+v to %+v", condition, updatedCondition)
	return nil
}

func (p *PodReadinessController) getLoadBalancerByName(lbProvider CloudLoadBalancerProvider, compartmentId, name string) (*client.GenericLoadBalancer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	return lbProvider.lbClient.GetLoadBalancerByName(ctx, compartmentId, name)
}

func (p *PodReadinessController) getBackendSetHealth(lbProvider CloudLoadBalancerProvider, lbId, backendSetName string) (*client.GenericBackendSetHealth, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	return lbProvider.getBackendSetHealth(ctx, lbId, backendSetName)
}

func (p *PodReadinessController) patchPod(namespace, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Pod, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	return p.kubeClient.CoreV1().Pods(namespace).Patch(ctx, name, pt, data, opts, subresources...)
}

func (clb *CloudLoadBalancerProvider) getBackendSetHealth(ctx context.Context, lbId, backendSetName string) (*client.GenericBackendSetHealth, error) {
	res, err := clb.lbClient.GetBackendSetHealth(ctx, lbId, backendSetName)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (p *PodReadinessController) getLoadBalancerProvider(lbType string) CloudLoadBalancerProvider {
	return CloudLoadBalancerProvider{
		client:       p.ociClient,
		lbClient:     p.ociClient.LoadBalancer(p.logger, lbType, p.config.Auth.TenancyID, nil),
		logger:       p.logger,
		metricPusher: p.metricPusher,
		config:       p.config,
	}
}

func getPodReadinessCondition(serviceNamespace, serviceName string, backendSetName string) v1.PodConditionType {
	// The character limit for the pod condition string is 63
	// Refer: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
	h := sha256.New()
	h.Write([]byte(serviceNamespace))
	h.Write([]byte(serviceName))
	h.Write([]byte(backendSetName))
	hash := fmt.Sprintf("%.20s", hex.EncodeToString(h.Sum(nil)))
	return v1.PodConditionType(fmt.Sprintf("%s/%s", podReadinessConditionPrefix, hash))
}

func hasReadinessGate(pod *v1.Pod, readinessGate v1.PodConditionType) bool {
	for _, r := range pod.Spec.ReadinessGates {
		if r.ConditionType == readinessGate {
			return true
		}
	}
	return false
}

func getPodCondition(pod *v1.Pod, conditionType v1.PodConditionType) (v1.PodCondition, bool) {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == conditionType {
			return cond, true
		}
	}

	return v1.PodCondition{}, false
}

func buildPodConditionPatch(pod *v1.Pod, condition v1.PodCondition) ([]byte, error) {
	oldData, err := json.Marshal(v1.Pod{
		Status: v1.PodStatus{
			Conditions: nil,
		},
	})
	if err != nil {
		return nil, err
	}

	newData, err := json.Marshal(v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: pod.UID,
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{condition},
		},
	})
	if err != nil {
		return nil, err
	}

	return strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Pod{})
}

func getBackendsMap(backendSet client.GenericBackendSetDetails) map[string]struct{} {
	m := make(map[string]struct{})
	for _, backend := range backendSet.Backends {
		if backend.IpAddress != nil {
			m[*backend.IpAddress] = struct{}{}
		}
	}
	return m
}

func getUnhealthyBackendMap(backendSetHealth *client.GenericBackendSetHealth, conditionType v1.PodConditionType) map[string]v1.PodCondition {
	unhealthyBackendMap := make(map[string]v1.PodCondition)

	for _, backend := range backendSetHealth.UnknownStateBackendNames {
		unhealthyBackendMap[backend] = v1.PodCondition{
			Type:   conditionType,
			Reason: "backend health is UNKNOWN",
			Status: v1.ConditionFalse,
		}
	}
	for _, backend := range backendSetHealth.WarningStateBackendNames {
		unhealthyBackendMap[backend] = v1.PodCondition{
			Type:   conditionType,
			Reason: "backend health is WARNING",
			Status: v1.ConditionFalse,
		}
	}
	for _, backend := range backendSetHealth.CriticalStateBackendNames {
		unhealthyBackendMap[backend] = v1.PodCondition{
			Type:   conditionType,
			Reason: "backend health is CRITICAL",
			Status: v1.ConditionFalse,
		}
	}

	return unhealthyBackendMap
}

func getUpdatedPodCondition(backendHealthMap map[string]v1.PodCondition, condType v1.PodConditionType, backendName string) v1.PodCondition {
	if cond, exists := backendHealthMap[backendName]; exists {
		return cond
	}

	// For virtual pods, if their corresponding backend is healthy they will not have an entry in the backendHealthMap,
	// we return a healthy condition
	// For regular pods, since they are not added as backends, they will not have an entry in the backendHealthMap,
	// we return a healthy condition
	return v1.PodCondition{
		Type:   condType,
		Status: v1.ConditionTrue,
		Reason: "backend is OK",
	}
}

func (p *PodReadinessController) getBackendSetsNeedSync(service *v1.Service, pods []*v1.Pod) (map[string]v1.ServicePort, error) {
	backendSetsNeedSync := make(map[string]v1.ServicePort)

	for backendSetName, servicePort := range getBackendSetNamePortMap(service) {
		podReadinessCondition := getPodReadinessCondition(service.Namespace, service.Name, backendSetName)

		for _, pod := range pods {
			// Check if pod has readiness gate
			if !hasReadinessGate(pod, podReadinessCondition) {
				continue
			}

			backendSetsNeedSync[backendSetName] = servicePort
			break
		}
	}

	return backendSetsNeedSync, nil
}
