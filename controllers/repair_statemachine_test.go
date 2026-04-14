package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type failingUncordonClient struct {
	client.Client
	fail bool
}

func (c *failingUncordonClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if c.fail {
		if node, ok := obj.(*v1.Node); ok && !node.Spec.Unschedulable {
			return errors.New("simulated uncordon patch failure")
		}
	}
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func TestRetryDelayBackoffAndCap(t *testing.T) {
	// Base from env default is 10s; cap is 5m by default
	sm := &nodeRepairStateMachine{node: &v1.Node{}}
	// verify exponential growth and cap
	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{1, 10 * time.Second, 10 * time.Second},
		{2, 20 * time.Second, 20 * time.Second},
		{3, 40 * time.Second, 40 * time.Second},
		{4, 80 * time.Second, 80 * time.Second},
		{5, 160 * time.Second, 160 * time.Second},
		// shifted attempts beyond 5 should cap at base<<5 = 320s but then overall cap 300s applies
		{6, 300 * time.Second, 300 * time.Second},
		{10, 300 * time.Second, 300 * time.Second},
	}
	for _, tt := range tests {
		d := sm.retryDelay(stateCordoning, tt.attempt)
		if d < tt.min || d > tt.max {
			t.Fatalf("attempt %d expected %v..%v got %v", tt.attempt, tt.min, tt.max, d)
		}
	}
}

func TestStateTimedOut_TrueWhenExceedsTimeout(t *testing.T) {
	sm := &nodeRepairStateMachine{node: &v1.Node{}}
	sm.tracker = newRepairStateTracker("rid-1")
	entry := sm.tracker.begin(stateDraining)
	// backdate start time to exceed default draining timeout
	entry.StartTime = time.Now().UTC().Add(-31 * time.Minute).Format(time.RFC3339)
	if !sm.stateTimedOut(stateDraining) {
		t.Fatalf("expected draining state to be timed out")
	}
}

func TestStateTimedOut_FalseBeforeThirtyMinuteDrainingTimeout(t *testing.T) {
	sm := &nodeRepairStateMachine{node: &v1.Node{}}
	sm.tracker = newRepairStateTracker("rid-1")
	entry := sm.tracker.begin(stateDraining)
	entry.StartTime = time.Now().UTC().Add(-29 * time.Minute).Format(time.RFC3339)
	if sm.stateTimedOut(stateDraining) {
		t.Fatalf("expected draining state to remain within the 30 minute timeout")
	}
}

func TestBeginState_PreservesExistingStartTime(t *testing.T) {
	start := time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339)
	sm := &nodeRepairStateMachine{
		node:    &v1.Node{},
		tracker: newRepairStateTracker("rid-1"),
	}
	sm.tracker.States[stateCordoning] = &stateTrackerEntry{
		StartTime: start,
		Attempts:  2,
	}

	sm.beginState(context.Background(), stateCordoning)
	entry := sm.tracker.current(stateCordoning)
	if entry == nil {
		t.Fatalf("expected state tracker entry to exist")
	}
	if entry.StartTime != start {
		t.Fatalf("expected state start to remain %q, got %q", start, entry.StartTime)
	}
}

func TestHandleUncordoning_RequeuesWhileNodeUnhealthy(t *testing.T) {
	orig := postRebootObservationWindow
	postRebootObservationWindow = 30 * time.Second
	defer func() { postRebootObservationWindow = orig }()

	now := time.Now().UTC().Format(time.RFC3339)
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nar-uncordon-health-gate",
			Annotations: map[string]string{
				narStateAnnotationKey:       string(stateUncordon),
				narRepairIDAnnotationKey:    "rid-1",
				narLastTransitionAnnotation: now,
			},
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeConditionType("GPUCount"),
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	sm := &nodeRepairStateMachine{
		node:     node,
		repairID: "rid-1",
		tracker:  newRepairStateTracker("rid-1"),
	}
	sm.tracker.States[stateUncordon] = &stateTrackerEntry{
		StartTime: time.Now().UTC().Add(-5 * time.Second).Format(time.RFC3339),
	}

	res, err := sm.handleUncordoning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter <= 0 || res.RequeueAfter > postRebootObservationWindow {
		t.Fatalf("expected requeue during post-reboot observation window, got %v", res.RequeueAfter)
	}
}

func TestHandleUncordoning_EntersCleanupWhenConditionsRemainUnhealthyAfterObservationWindow(t *testing.T) {
	orig := postRebootObservationWindow
	postRebootObservationWindow = 30 * time.Second
	defer func() { postRebootObservationWindow = orig }()

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nar-uncordon-observation-fail",
			Annotations: map[string]string{
				narStateAnnotationKey:            string(stateUncordon),
				narRepairIDAnnotationKey:         "rid-observe",
				narLastTransitionAnnotation:      time.Now().UTC().Add(-31 * time.Second).Format(time.RFC3339),
				narCordonedByRepairAnnotationKey: "true",
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: true,
			Taints:        []v1.Taint{CreateRepairTaint()},
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeConditionType("GPUCount"), Status: v1.ConditionTrue},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: fakeClient}
	sm := &nodeRepairStateMachine{
		reconciler: r,
		node:       node.DeepCopy(),
		nodeKey:    client.ObjectKey{Name: node.Name},
		repairID:   "rid-observe",
		tracker:    newRepairStateTracker("rid-observe"),
	}
	sm.tracker.States[stateUncordon] = &stateTrackerEntry{
		StartTime: time.Now().UTC().Add(-31 * time.Second).Format(time.RFC3339),
	}

	res, err := sm.handleUncordoning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateUncordon].successRequeue {
		t.Fatalf("expected cleanup requeueAfter %v, got %v", repairStateConfigs[stateUncordon].successRequeue, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if !latest.Spec.Unschedulable {
		t.Fatalf("expected node to remain cordoned until cleanup uncordon runs")
	}
	if got := latest.Annotations[narFailureCleanupPendingKey]; got != "true" {
		t.Fatalf("expected cleanup-pending annotation after unhealthy observation window, got %q", got)
	}
}

func TestHandleUncordoning_CleanupPathUncordonsEvenWhenStillUnhealthy(t *testing.T) {
	orig := unhealthyThreshold
	unhealthyThreshold = 10 * time.Minute
	defer func() { unhealthyThreshold = orig }()

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nar-cleanup-uncordon",
			Annotations: map[string]string{
				narStateAnnotationKey:            string(stateUncordon),
				narRepairIDAnnotationKey:         "rid-cleanup",
				narLastTransitionAnnotation:      now,
				narCordonedByRepairAnnotationKey: "true",
				narFailureCleanupPendingKey:      "true",
				narFailureCleanupStateKey:        string(stateRebooting),
				narRepairCycleAttemptsKey:        "0",
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: true,
			Taints:        []v1.Taint{CreateRepairTaint()},
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeConditionType("GPUCount"), Status: v1.ConditionTrue},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: fakeClient}
	sm := &nodeRepairStateMachine{
		reconciler: r,
		node:       node.DeepCopy(),
		nodeKey:    client.ObjectKey{Name: node.Name},
		repairID:   "rid-cleanup",
		tracker:    newRepairStateTracker("rid-cleanup"),
		logger:     logr.Discard(),
	}

	res, err := sm.handleUncordoning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != unhealthyThreshold {
		t.Fatalf("expected failed cycle to requeue after %v, got %v", unhealthyThreshold, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if latest.Spec.Unschedulable {
		t.Fatalf("expected cleanup path to uncordon node before finalizing failure")
	}
	if got := latest.Annotations[narLastRepairResultAnnotation]; got != "failed" {
		t.Fatalf("expected last repair result failed, got %q", got)
	}
	if got := latest.Annotations[narRepairCycleAttemptsKey]; got != "1" {
		t.Fatalf("expected cycle attempts to be incremented to 1, got %q", got)
	}
	if _, ok := latest.Annotations[narFailureCleanupPendingKey]; ok {
		t.Fatalf("expected cleanup-pending annotation to be pruned after failed cycle completion")
	}
	for _, taint := range latest.Spec.Taints {
		if taint.Key == REPAIR_TAINT_KEY && taint.Effect == REPAIR_TAINT_EFFECT {
			t.Fatalf("expected repair taint to be removed after failed cycle cleanup")
		}
	}
}

func TestHandleUncordoning_CleanupPathRetriesUntilNodeIsSchedulable(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nar-cleanup-retry",
			Annotations: map[string]string{
				narStateAnnotationKey:            string(stateUncordon),
				narRepairIDAnnotationKey:         "rid-cleanup-retry",
				narLastTransitionAnnotation:      now,
				narCordonedByRepairAnnotationKey: "true",
				narFailureCleanupPendingKey:      "true",
				narFailureCleanupStateKey:        string(stateDraining),
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: true,
			Taints:        []v1.Taint{CreateRepairTaint()},
		},
	}

	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: &failingUncordonClient{Client: baseClient, fail: true}}
	sm := &nodeRepairStateMachine{
		reconciler: r,
		node:       node.DeepCopy(),
		nodeKey:    client.ObjectKey{Name: node.Name},
		repairID:   "rid-cleanup-retry",
		tracker:    newRepairStateTracker("rid-cleanup-retry"),
		logger:     logr.Discard(),
	}

	res, err := sm.handleUncordoning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter <= 0 {
		t.Fatalf("expected retry requeue after cleanup uncordon failure, got %v", res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := baseClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if !latest.Spec.Unschedulable {
		t.Fatalf("expected node to remain cordoned when cleanup uncordon patch fails")
	}
	if _, ok := latest.Annotations[narLastRepairResultAnnotation]; ok {
		t.Fatalf("expected failed cycle not to finalize while cleanup uncordon is still failing")
	}
	if got := latest.Annotations[narFailureCleanupPendingKey]; got != "true" {
		t.Fatalf("expected cleanup-pending annotation to remain set, got %q", got)
	}
}

func TestFailState_DoesNotEnterCleanupWhenRepairNeverCordonedNode(t *testing.T) {
	orig := unhealthyThreshold
	unhealthyThreshold = 10 * time.Minute
	defer func() { unhealthyThreshold = orig }()

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nar-fail-before-cordon",
			Annotations: map[string]string{
				narStateAnnotationKey:       string(stateCordoning),
				narRepairIDAnnotationKey:    "rid-no-cordon",
				narLastTransitionAnnotation: now,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	sm := &nodeRepairStateMachine{
		reconciler: &NodeAutoRepairReconciler{Client: fakeClient},
		node:       node.DeepCopy(),
		nodeKey:    client.ObjectKey{Name: node.Name},
		repairID:   "rid-no-cordon",
		tracker:    newRepairStateTracker("rid-no-cordon"),
		logger:     logr.Discard(),
	}

	res, err := sm.failState(context.Background(), stateCordoning, errors.New("cordon failed"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != unhealthyThreshold {
		t.Fatalf("expected failed cycle to requeue after %v, got %v", unhealthyThreshold, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if _, ok := latest.Annotations[narFailureCleanupPendingKey]; ok {
		t.Fatalf("expected cleanup not to be scheduled when repair never cordoned node")
	}
	if got := latest.Annotations[narRepairCycleAttemptsKey]; got != "1" {
		t.Fatalf("expected cycle attempts to increment to 1, got %q", got)
	}
}

func TestFailState_AddsHumanInterventionLabelOnThirdFailure(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nar-third-failure",
			Annotations: map[string]string{
				narStateAnnotationKey:         string(stateCordoning),
				narRepairIDAnnotationKey:      "rid-third",
				narLastTransitionAnnotation:   now,
				narRepairCycleAttemptsKey:     "2",
				narLastRepairResultAnnotation: "failed",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	sm := &nodeRepairStateMachine{
		reconciler: &NodeAutoRepairReconciler{Client: fakeClient},
		node:       node.DeepCopy(),
		nodeKey:    client.ObjectKey{Name: node.Name},
		repairID:   "rid-third",
		tracker:    newRepairStateTracker("rid-third"),
		logger:     logr.Discard(),
	}

	res, err := sm.failState(context.Background(), stateCordoning, errors.New("third failure"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != 0 {
		t.Fatalf("expected no requeue after human intervention is required, got %v", res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if got := latest.Labels[repairHumanInterventionLabel]; got != "true" {
		t.Fatalf("expected human intervention label to be set, got %q", got)
	}
	if _, ok := latest.Annotations[narRepairCycleAttemptsKey]; ok {
		t.Fatalf("expected cycle attempts to be reset once human intervention label is applied")
	}
}
