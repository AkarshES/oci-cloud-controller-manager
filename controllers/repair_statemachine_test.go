package controllers

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

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
	entry.StartTime = time.Now().UTC().Add(-20 * time.Minute).Format(time.RFC3339)
	if !sm.stateTimedOut(stateDraining) {
		t.Fatalf("expected draining state to be timed out")
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
		StartTime: now,
	}

	res, err := sm.handleUncordoning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateUncordon].successRequeue {
		t.Fatalf("expected requeueAfter %v, got %v", repairStateConfigs[stateUncordon].successRequeue, res.RequeueAfter)
	}
}

func TestHandleUncordoning_WaitsForHealthyStabilization(t *testing.T) {
	orig := healthyStabilization
	healthyStabilization = 30 * time.Second
	defer func() { healthyStabilization = orig }()

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nar-uncordon-stabilization",
			Annotations: map[string]string{
				narStateAnnotationKey:       string(stateUncordon),
				narRepairIDAnnotationKey:    "rid-stable",
				narLastTransitionAnnotation: now,
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: fakeClient}
	sm := &nodeRepairStateMachine{
		reconciler: r,
		node:       node.DeepCopy(),
		nodeKey:    client.ObjectKey{Name: node.Name},
		repairID:   "rid-stable",
		tracker:    newRepairStateTracker("rid-stable"),
	}
	sm.tracker.States[stateUncordon] = &stateTrackerEntry{StartTime: now}

	res, err := sm.handleUncordoning(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateUncordon].successRequeue {
		t.Fatalf("expected requeueAfter %v, got %v", repairStateConfigs[stateUncordon].successRequeue, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	tracker := loadRepairStateTracker(latest)
	entry := tracker.current(stateUncordon)
	if entry == nil || entry.HealthySince == "" {
		t.Fatalf("expected healthy stabilization timestamp to be recorded, got %+v", entry)
	}
	if got := latest.Annotations[narLastRepairResultAnnotation]; got != "" {
		t.Fatalf("expected repair result to remain unset while stabilizing, got %q", got)
	}
}
