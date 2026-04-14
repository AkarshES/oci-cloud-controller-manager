package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNodeDrainExecutorReturnsCompleteWhenNoDrainablePods(t *testing.T) {
	executor := newNodeDrainExecutor(k8sfake.NewSimpleClientset())

	result, err := executor.DrainNode(context.Background(), "node-a", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Outcome != drainOutcomeComplete {
		t.Fatalf("expected complete outcome, got %s", result.Outcome)
	}
	if result.Status != drainTrackerStatusComplete {
		t.Fatalf("expected complete status, got %s", result.Status)
	}
	if result.RemainingPods != 0 {
		t.Fatalf("expected 0 remaining pods, got %d", result.RemainingPods)
	}
}

func TestNodeDrainExecutorTreatsDeletingPodAsInProgress(t *testing.T) {
	now := metav1.NewTime(time.Now().UTC())
	pod := newDrainTestPod("node-a", "pod-a")
	pod.DeletionTimestamp = &now

	stubEvictionSupport(t)
	clientset := k8sfake.NewSimpleClientset(pod)
	executor := newNodeDrainExecutor(clientset)

	result, err := executor.DrainNode(context.Background(), "node-a", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Outcome != drainOutcomeInProgress {
		t.Fatalf("expected in progress outcome, got %s", result.Outcome)
	}
	if result.Status != drainTrackerStatusInProgress {
		t.Fatalf("expected in progress status, got %s", result.Status)
	}
}

func TestNodeDrainExecutorClassifiesEviction404AsInProgress(t *testing.T) {
	stubEvictionSupport(t)
	clientset := newPolicyEnabledClientset(newDrainTestPod("node-a", "pod-a"))
	clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		return true, nil, apierrors.NewNotFound(schema.GroupResource{Group: "policy", Resource: "evictions"}, "pod-a")
	})

	result, err := newNodeDrainExecutor(clientset).DrainNode(context.Background(), "node-a", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Outcome != drainOutcomeInProgress {
		t.Fatalf("expected in progress outcome, got %s", result.Outcome)
	}
}

func TestNodeDrainExecutorClassifiesEviction429AsPDBBackoff(t *testing.T) {
	stubEvictionSupport(t)
	clientset := newPolicyEnabledClientset(newDrainTestPod("node-a", "pod-a"))
	clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		return true, nil, apierrors.NewTooManyRequests("pdb blocked", 0)
	})

	result, err := newNodeDrainExecutor(clientset).DrainNode(context.Background(), "node-a", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Outcome != drainOutcomeBackoff {
		t.Fatalf("expected backoff outcome, got %s", result.Outcome)
	}
	if result.FailureType != drainFailureTypePDB {
		t.Fatalf("expected pdb failure type, got %s", result.FailureType)
	}
}

func TestNodeDrainExecutorClassifiesNamespaceTerminatingAsBackoff(t *testing.T) {
	stubEvictionSupport(t)
	clientset := newPolicyEnabledClientset(newDrainTestPod("node-a", "pod-a"))
	clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		return true, nil, newNamespaceTerminatingError("pod-a")
	})

	result, err := newNodeDrainExecutor(clientset).DrainNode(context.Background(), "node-a", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Outcome != drainOutcomeBackoff {
		t.Fatalf("expected backoff outcome, got %s", result.Outcome)
	}
	if result.FailureType != drainFailureTypeNamespaceTerminating {
		t.Fatalf("expected namespace terminating failure type, got %s", result.FailureType)
	}
}

func TestNodeDrainExecutorClassifiesUnexpectedEvictionErrorAsHardError(t *testing.T) {
	stubEvictionSupport(t)
	clientset := newPolicyEnabledClientset(newDrainTestPod("node-a", "pod-a"))
	clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		return true, nil, apierrors.NewUnauthorized("not allowed")
	})

	result, err := newNodeDrainExecutor(clientset).DrainNode(context.Background(), "node-a", false)
	if err == nil {
		t.Fatalf("expected an error")
	}
	if result.Outcome != drainOutcomeHardError {
		t.Fatalf("expected hard error outcome, got %s", result.Outcome)
	}
	if result.FailureType != drainFailureTypeHardError {
		t.Fatalf("expected hard error failure type, got %s", result.FailureType)
	}
}

func TestNodeDrainExecutorUsesDeleteWhenForced(t *testing.T) {
	clientset := newPolicyEnabledClientset(newDrainTestPod("node-a", "pod-a"))
	deleteCalls := 0
	evictCalls := 0
	clientset.PrependReactor("delete", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deleteCalls++
		return false, nil, nil
	})
	clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		evictCalls++
		return true, &policyv1.Eviction{}, nil
	})

	result, err := newNodeDrainExecutor(clientset).DrainNode(context.Background(), "node-a", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Outcome != drainOutcomeInProgress {
		t.Fatalf("expected in progress outcome, got %s", result.Outcome)
	}
	if deleteCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", deleteCalls)
	}
	if evictCalls != 0 {
		t.Fatalf("expected 0 eviction calls, got %d", evictCalls)
	}
}

func TestHandleDrainingTransitionsToRebootingWhenNoPodsRemain(t *testing.T) {
	node := newDrainingTestNode("node-complete", time.Now().UTC())
	sm, kubeClient, nodeClient := newDrainingStateMachine(t, node, k8sfake.NewSimpleClientset())
	_ = kubeClient

	res, err := sm.handleDraining(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateRebooting].successRequeue {
		t.Fatalf("expected rebooting requeue %v, got %v", repairStateConfigs[stateRebooting].successRequeue, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := nodeClient.Get(context.Background(), ctrlclient.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch node: %v", err)
	}
	if latest.Annotations[narStateAnnotationKey] != string(stateRebooting) {
		t.Fatalf("expected node state to be rebooting, got %q", latest.Annotations[narStateAnnotationKey])
	}

	tracker := loadRepairStateTracker(latest)
	entry := tracker.current(stateDraining)
	if entry == nil || entry.Drain == nil {
		t.Fatalf("expected draining tracker metadata to be persisted")
	}
	if entry.Drain.Status != drainTrackerStatusComplete {
		t.Fatalf("expected drain status complete, got %s", entry.Drain.Status)
	}
}

func TestHandleDrainingBackoffPDBKeepsStateAndPersistsMetadata(t *testing.T) {
	withDrainForceSettings(t, false, 0)
	stubEvictionSupport(t)
	node := newDrainingTestNode("node-pdb", time.Now().UTC())
	kubeClient := newPolicyEnabledClientset(newDrainTestPod(node.Name, "pod-a"))
	kubeClient.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		return true, nil, apierrors.NewTooManyRequests("pdb blocked", 0)
	})

	sm, _, nodeClient := newDrainingStateMachine(t, node, kubeClient)

	res, err := sm.handleDraining(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateDraining].successRequeue {
		t.Fatalf("expected draining requeue %v, got %v", repairStateConfigs[stateDraining].successRequeue, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := nodeClient.Get(context.Background(), ctrlclient.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch node: %v", err)
	}
	if latest.Annotations[narStateAnnotationKey] != string(stateDraining) {
		t.Fatalf("expected node to remain in draining, got %q", latest.Annotations[narStateAnnotationKey])
	}

	tracker := loadRepairStateTracker(latest)
	entry := tracker.current(stateDraining)
	if entry == nil || entry.Drain == nil {
		t.Fatalf("expected draining metadata")
	}
	if entry.Drain.Status != drainTrackerStatusBackoff {
		t.Fatalf("expected backoff status, got %s", entry.Drain.Status)
	}
	if entry.Drain.FailureType != drainFailureTypePDB {
		t.Fatalf("expected pdb failure type, got %s", entry.Drain.FailureType)
	}
	if entry.Drain.RemainingPods != 1 {
		t.Fatalf("expected remaining pods 1, got %d", entry.Drain.RemainingPods)
	}
}

func TestHandleDrainingHardErrorUsesRetryBackoff(t *testing.T) {
	withDrainForceSettings(t, false, 0)
	stubEvictionSupport(t)
	node := newDrainingTestNode("node-hard-error", time.Now().UTC())
	kubeClient := newPolicyEnabledClientset(newDrainTestPod(node.Name, "pod-a"))
	kubeClient.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		return true, nil, apierrors.NewUnauthorized("forbidden")
	})

	sm, _, nodeClient := newDrainingStateMachine(t, node, kubeClient)

	res, err := sm.handleDraining(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.RequeueAfter != sm.retryDelay(stateDraining, 1) {
		t.Fatalf("expected retry delay %v, got %v", sm.retryDelay(stateDraining, 1), res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := nodeClient.Get(context.Background(), ctrlclient.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch node: %v", err)
	}
	tracker := loadRepairStateTracker(latest)
	entry := tracker.current(stateDraining)
	if entry == nil || entry.Drain == nil {
		t.Fatalf("expected draining metadata")
	}
	if entry.Drain.FailureType != drainFailureTypeHardError {
		t.Fatalf("expected hard error failure type, got %s", entry.Drain.FailureType)
	}
}

func TestHandleDrainingUsesDeleteByDefaultWhenForceEnabled(t *testing.T) {
	withDrainForceSettings(t, true, 0)
	node := newDrainingTestNode("node-force-default", time.Now().UTC())
	kubeClient := newPolicyEnabledClientset(newDrainTestPod(node.Name, "pod-a"))
	deleteCalls := 0
	evictCalls := 0
	kubeClient.PrependReactor("delete", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deleteCalls++
		return false, nil, nil
	})
	kubeClient.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		evictCalls++
		return true, &policyv1.Eviction{}, nil
	})

	sm, _, nodeClient := newDrainingStateMachine(t, node, kubeClient)

	res, err := sm.handleDraining(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateDraining].successRequeue {
		t.Fatalf("expected draining requeue %v, got %v", repairStateConfigs[stateDraining].successRequeue, res.RequeueAfter)
	}
	if deleteCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", deleteCalls)
	}
	if evictCalls != 0 {
		t.Fatalf("expected no eviction calls, got %d", evictCalls)
	}

	latest := &v1.Node{}
	if err := nodeClient.Get(context.Background(), ctrlclient.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch node: %v", err)
	}
	tracker := loadRepairStateTracker(latest)
	entry := tracker.current(stateDraining)
	if entry == nil || entry.Drain == nil {
		t.Fatalf("expected draining metadata")
	}
	if !entry.Forced {
		t.Fatalf("expected forced draining to be persisted")
	}
}

func TestHandleDrainingUsesEvictionWhenCustomerDisablesForce(t *testing.T) {
	withDrainForceSettings(t, false, 0)
	stubEvictionSupport(t)
	node := newDrainingTestNode("node-force-disabled", time.Now().UTC())
	kubeClient := newPolicyEnabledClientset(newDrainTestPod(node.Name, "pod-a"))
	deleteCalls := 0
	evictCalls := 0
	kubeClient.PrependReactor("delete", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deleteCalls++
		return false, nil, nil
	})
	kubeClient.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		evictCalls++
		return true, &policyv1.Eviction{}, nil
	})

	sm, _, nodeClient := newDrainingStateMachine(t, node, kubeClient)

	res, err := sm.handleDraining(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateDraining].successRequeue {
		t.Fatalf("expected draining requeue %v, got %v", repairStateConfigs[stateDraining].successRequeue, res.RequeueAfter)
	}
	if deleteCalls != 0 {
		t.Fatalf("expected no delete calls when force is disabled, got %d", deleteCalls)
	}
	if evictCalls != 1 {
		t.Fatalf("expected 1 eviction call when force is disabled, got %d", evictCalls)
	}

	latest := &v1.Node{}
	if err := nodeClient.Get(context.Background(), ctrlclient.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch node: %v", err)
	}
	tracker := loadRepairStateTracker(latest)
	entry := tracker.current(stateDraining)
	if entry == nil || entry.Drain == nil {
		t.Fatalf("expected draining metadata")
	}
	if entry.Forced {
		t.Fatalf("expected force drain to remain disabled")
	}
}

func TestHandleDrainingSwitchesToForcedModeAfterConfiguredThreshold(t *testing.T) {
	withDrainForceSettings(t, true, 5*time.Minute)
	start := time.Now().UTC().Add(-drainForceAfter).Add(-time.Minute)
	node := newDrainingTestNode("node-force", start)
	kubeClient := newPolicyEnabledClientset(newDrainTestPod(node.Name, "pod-a"))
	deleteCalls := 0
	evictCalls := 0
	kubeClient.PrependReactor("delete", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deleteCalls++
		return false, nil, nil
	})
	kubeClient.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		evictCalls++
		return true, &policyv1.Eviction{}, nil
	})

	sm, _, nodeClient := newDrainingStateMachine(t, node, kubeClient)

	res, err := sm.handleDraining(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateDraining].successRequeue {
		t.Fatalf("expected draining requeue %v, got %v", repairStateConfigs[stateDraining].successRequeue, res.RequeueAfter)
	}
	if deleteCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", deleteCalls)
	}
	if evictCalls != 0 {
		t.Fatalf("expected no eviction calls, got %d", evictCalls)
	}

	latest := &v1.Node{}
	if err := nodeClient.Get(context.Background(), ctrlclient.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch node: %v", err)
	}
	tracker := loadRepairStateTracker(latest)
	entry := tracker.current(stateDraining)
	if entry == nil || entry.Drain == nil {
		t.Fatalf("expected draining metadata")
	}
	if !entry.Forced {
		t.Fatalf("expected forced draining to be persisted")
	}
	if entry.Drain.Status != drainTrackerStatusInProgress {
		t.Fatalf("expected in progress drain status, got %s", entry.Drain.Status)
	}
}

func TestNewNodeRepairStateMachineRestoresDrainMetadata(t *testing.T) {
	node := newDrainingTestNode("node-restore", time.Now().UTC())
	node.Annotations[narStateMetadataAnnotationKey] = `{"repairId":"rid-test","states":{"Draining":{"startTime":"2026-04-14T01:00:00Z","forced":true,"drain":{"status":"backoff","failureType":"pdb","remainingPods":2,"blockingPod":"default/pod-a","lastMessage":"blocked","lastUpdateTime":"2026-04-14T01:01:00Z"}}}}`

	sm := newNodeRepairStateMachine(&NodeAutoRepairReconciler{}, node.DeepCopy(), logr.Discard())
	entry := sm.tracker.current(stateDraining)
	if entry == nil || entry.Drain == nil {
		t.Fatalf("expected drain metadata to be restored")
	}
	if !entry.Forced {
		t.Fatalf("expected forced flag to be restored")
	}
	if entry.Drain.FailureType != drainFailureTypePDB {
		t.Fatalf("expected pdb failure type, got %s", entry.Drain.FailureType)
	}
}

func newPolicyEnabledClientset(objects ...runtime.Object) *k8sfake.Clientset {
	clientset := k8sfake.NewSimpleClientset(objects...)
	discovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
	if !ok {
		return clientset
	}
	discovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{
					Name:    "pods/eviction",
					Kind:    "Eviction",
					Group:   "policy",
					Version: "v1",
				},
			},
		},
	}
	return clientset
}

func newDrainTestPod(nodeName, name string) *v1.Pod {
	controller := true
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "ReplicaSet",
					Name:       "rs-a",
					Controller: &controller,
				},
			},
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
}

func newNamespaceTerminatingError(name string) error {
	return &apierrors.StatusError{
		ErrStatus: metav1.Status{
			Status:  metav1.StatusFailure,
			Code:    403,
			Reason:  metav1.StatusReasonForbidden,
			Message: "Forbidden: NamespaceTerminating",
			Details: &metav1.StatusDetails{
				Name: name,
				Causes: []metav1.StatusCause{
					{
						Type:    v1.NamespaceTerminatingCause,
						Message: "NamespaceTerminating",
					},
				},
			},
		},
	}
}

func newDrainingTestNode(name string, lastTransition time.Time) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				narStateAnnotationKey:       string(stateDraining),
				narRepairIDAnnotationKey:    "rid-test",
				narLastTransitionAnnotation: lastTransition.UTC().Format(time.RFC3339),
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: true,
		},
	}
}

func newDrainingStateMachine(t *testing.T, node *v1.Node, kubeClient *k8sfake.Clientset) (*nodeRepairStateMachine, *k8sfake.Clientset, ctrlclient.Client) {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add node scheme: %v", err)
	}

	nodeClient := ctrlfake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	reconciler := &NodeAutoRepairReconciler{
		Client:     nodeClient,
		KubeClient: kubeClient,
	}

	sm := newNodeRepairStateMachine(reconciler, node.DeepCopy(), logr.Discard())
	sm.repairID = node.Annotations[narRepairIDAnnotationKey]
	sm.tracker = loadRepairStateTracker(node)
	if sm.tracker == nil {
		sm.tracker = newRepairStateTracker(sm.repairID)
	}
	return sm, kubeClient, nodeClient
}

func TestShouldForceDrainDefaultsToImmediateForce(t *testing.T) {
	withDrainForceSettings(t, true, 0)
	if !shouldForceDrain(time.Now().UTC()) {
		t.Fatalf("expected draining to force immediately by default")
	}
}

func TestShouldForceDrainCanBeDisabledByCustomer(t *testing.T) {
	withDrainForceSettings(t, false, 5*time.Minute)
	start := time.Now().UTC().Add(-10 * time.Minute)
	if shouldForceDrain(start) {
		t.Fatalf("expected customer-disabled force drain to stay disabled")
	}
}

func TestShouldForceDrainHonorsConfiguredDelay(t *testing.T) {
	withDrainForceSettings(t, true, 5*time.Minute)
	if shouldForceDrain(time.Now().UTC().Add(-4 * time.Minute)) {
		t.Fatalf("expected force drain to wait for configured delay")
	}
	if !shouldForceDrain(time.Now().UTC().Add(-6 * time.Minute)) {
		t.Fatalf("expected force drain after configured delay")
	}
}

func TestIsPDBConflictError(t *testing.T) {
	if !isPDBConflictError(errors.New(legacyPDBMultipleBudgetError)) {
		t.Fatalf("expected legacy pdb error to be classified")
	}
}

func stubEvictionSupport(t *testing.T) {
	t.Helper()
	original := drainCheckEvictionSupport
	drainCheckEvictionSupport = func(_ kubernetes.Interface) (schema.GroupVersion, error) {
		return schema.GroupVersion{Group: "policy", Version: "v1"}, nil
	}
	t.Cleanup(func() {
		drainCheckEvictionSupport = original
	})
}

func withDrainForceSettings(t *testing.T, enabled bool, after time.Duration) {
	t.Helper()
	originalEnabled := drainForceEnabled
	originalAfter := drainForceAfter
	drainForceEnabled = enabled
	drainForceAfter = after
	t.Cleanup(func() {
		drainForceEnabled = originalEnabled
		drainForceAfter = originalAfter
	})
}
