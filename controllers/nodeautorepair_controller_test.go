package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	coordinationv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestHandleUnhealthyNode_ThrottleEventSuppressedWhenRepairInProgress(t *testing.T) {
	now := time.Now().UTC()
	node := &v1.Node{}
	node.Name = "nar-throttle-in-progress"
	node.Annotations = map[string]string{
		narStateAnnotationKey:      string(stateDraining),
		narLastRepairEndAnnotation: now.Add(-30 * time.Minute).Format(time.RFC3339),
	}
	cond := &v1.NodeCondition{Type: v1.NodeReady, Status: v1.ConditionFalse}

	rec := record.NewFakeRecorder(1)
	r := &NodeAutoRepairReconciler{Recorder: rec}
	logger := logr.Discard()

	_, err := r.handleUnhealthyNode(context.Background(), logger, node, []*v1.NodeCondition{cond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case evt := <-rec.Events:
		t.Fatalf("expected no throttle event when repair in progress, got %q", evt)
	default:
	}
}

func TestGetNodeCooldownDuration_Default(t *testing.T) {
	orig := repairCoolDown
	repairCoolDown = 42 * time.Minute
	defer func() { repairCoolDown = orig }()

	node := &v1.Node{}
	if got := getNodeCooldownDuration(node); got != repairCoolDown {
		t.Fatalf("expected default cooldown %v, got %v", repairCoolDown, got)
	}
}

func TestGetNodeCooldownDuration_DurationString(t *testing.T) {
	node := &v1.Node{}
	node.Labels = map[string]string{
		repairCooldownLabel: "5m30s",
	}

	if got := getNodeCooldownDuration(node); got != 5*time.Minute+30*time.Second {
		t.Fatalf("expected custom duration 5m30s, got %v", got)
	}
}

func TestGetNodeCooldownDuration_MinutesValue(t *testing.T) {
	node := &v1.Node{}
	node.Labels = map[string]string{
		repairCooldownLabel: "15",
	}

	if got := getNodeCooldownDuration(node); got != 15*time.Minute {
		t.Fatalf("expected custom duration 15m, got %v", got)
	}
}

func TestHandleUnhealthyNode_DisabledLabelSkipsRepair(t *testing.T) {
	node := &v1.Node{}
	node.Name = "nar-disabled-node"
	node.Labels = map[string]string{
		repairDisabledLabel: "true",
	}
	cond := &v1.NodeCondition{Type: v1.NodeReady, Status: v1.ConditionFalse}

	rec := record.NewFakeRecorder(1)
	r := &NodeAutoRepairReconciler{Recorder: rec}
	logger := logr.Discard()

	res, err := r.handleUnhealthyNode(context.Background(), logger, node, []*v1.NodeCondition{cond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Requeue || res.RequeueAfter != 0 {
		t.Fatalf("expected no requeue when repair disabled, got %#v", res)
	}

	select {
	case evt := <-rec.Events:
		t.Fatalf("expected no event when repair is disabled, got %q", evt)
	default:
	}
}

// Test that cleanup logic will not propose pruning last-repair-end annotation
func TestNodeAnnotationsToPrune_DoesNotIncludeLastRepairEnd(t *testing.T) {
	node := &v1.Node{}
	node.Name = "nar-cleanup-node"
	node.Annotations = map[string]string{
		narStateAnnotationKey:         "Draining",
		narRepairIDAnnotationKey:      "rid-123",
		narRepairOriginAnnotationKey:  "controller-a",
		narLastTransitionAnnotation:   time.Now().UTC().Format(time.RFC3339),
		narAttemptsAnnotationKey:      "2",
		narRebootIssuedAnnotationKey:  "true",
		narStateMetadataAnnotationKey: "{}",
		narLastRepairEndAnnotation:    time.Now().UTC().Format(time.RFC3339),
	}

	keys := nodeAnnotationsToPrune(node)
	// Ensure last-repair-end is not among pruned keys
	for _, k := range keys {
		if k == narLastRepairEndAnnotation {
			t.Fatalf("narLastRepairEndAnnotation should not be pruned, but found in prune set")
		}
	}
	// And ensure working-state annotations are included
	want := map[string]bool{
		narStateAnnotationKey:         true,
		narRepairIDAnnotationKey:      true,
		narRepairOriginAnnotationKey:  true,
		narLastTransitionAnnotation:   true,
		narAttemptsAnnotationKey:      true,
		narRebootIssuedAnnotationKey:  true,
		narStateMetadataAnnotationKey: true,
	}
	for _, k := range keys {
		delete(want, k)
	}
	if len(want) != 0 {
		t.Fatalf("expected working-state annotations to be pruned, missing: %v", want)
	}
}

func TestConditionChangedPredicate_Update_TriggersOnRepairDisabledLabelAdded(t *testing.T) {
	p := ConditionChangedPredicate{log: zap.NewNop().Sugar()}
	oldNode := &v1.Node{}
	oldNode.Name = "nar-predicate-label-added"
	newNode := oldNode.DeepCopy()
	newNode.Labels = map[string]string{
		repairDisabledLabel: "true",
	}

	changed := p.Update(event.UpdateEvent{
		ObjectOld: oldNode,
		ObjectNew: newNode,
	})
	if !changed {
		t.Fatalf("expected update predicate to trigger when %s label is added", repairDisabledLabel)
	}
}

func TestConditionChangedPredicate_Update_TriggersOnRepairDisabledLabelRemoved(t *testing.T) {
	p := ConditionChangedPredicate{log: zap.NewNop().Sugar()}
	oldNode := &v1.Node{}
	oldNode.Name = "nar-predicate-label-removed"
	oldNode.Labels = map[string]string{
		repairDisabledLabel: "true",
	}
	newNode := oldNode.DeepCopy()
	delete(newNode.Labels, repairDisabledLabel)

	changed := p.Update(event.UpdateEvent{
		ObjectOld: oldNode,
		ObjectNew: newNode,
	})
	if !changed {
		t.Fatalf("expected update predicate to trigger when %s label is removed", repairDisabledLabel)
	}
}

func TestConditionChangedPredicate_Update_TriggersOnRepairDisabledLabelValueChange(t *testing.T) {
	p := ConditionChangedPredicate{log: zap.NewNop().Sugar()}
	oldNode := &v1.Node{}
	oldNode.Name = "nar-predicate-label-toggled"
	oldNode.Labels = map[string]string{
		repairDisabledLabel: "false",
	}
	newNode := oldNode.DeepCopy()
	newNode.Labels[repairDisabledLabel] = "true"

	changed := p.Update(event.UpdateEvent{
		ObjectOld: oldNode,
		ObjectNew: newNode,
	})
	if !changed {
		t.Fatalf("expected update predicate to trigger when %s label value changes", repairDisabledLabel)
	}
}

func TestConditionChangedPredicate_Update_DoesNotTriggerWhenRepairDisabledLabelUnchanged(t *testing.T) {
	p := ConditionChangedPredicate{log: zap.NewNop().Sugar()}
	oldNode := &v1.Node{}
	oldNode.Name = "nar-predicate-label-unchanged"
	oldNode.Labels = map[string]string{
		repairDisabledLabel: "true",
	}
	newNode := oldNode.DeepCopy()

	changed := p.Update(event.UpdateEvent{
		ObjectOld: oldNode,
		ObjectNew: newNode,
	})
	if changed {
		t.Fatalf("expected update predicate to remain false when %s label is unchanged", repairDisabledLabel)
	}
}

func TestHandleUnhealthyNode_DisabledLabelDoesNotBlockInProgressRepair(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	if err := coordinationv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add coordinationv1 scheme: %v", err)
	}

	node := &v1.Node{}
	node.Name = "nar-disabled-in-progress"
	node.Labels = map[string]string{
		repairDisabledLabel: "true",
	}
	node.Annotations = map[string]string{
		narStateAnnotationKey: string(stateDetected),
	}
	node.Status.Conditions = []v1.NodeCondition{
		{
			Type:   v1.NodeConditionType("GPUCount"),
			Status: v1.ConditionTrue,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	rec := record.NewFakeRecorder(5)
	r := &NodeAutoRepairReconciler{
		Client:       fakeClient,
		Recorder:     rec,
		ControllerID: "nar-test-controller",
	}
	logger := logr.Discard()

	res, err := r.handleUnhealthyNode(context.Background(), logger, node, findUnhealthyConditions(node))
	defer func() {
		_ = r.stopLeaseHeartbeat(context.Background(), node.Name)
	}()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateCordoning].successRequeue {
		t.Fatalf("expected in-progress repair to continue with requeueAfter %v, got %v", repairStateConfigs[stateCordoning].successRequeue, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if got := latest.Annotations[narStateAnnotationKey]; got != string(stateCordoning) {
		t.Fatalf("expected node state to transition to %q, got %q", stateCordoning, got)
	}
	if latest.Annotations[narRepairIDAnnotationKey] == "" {
		t.Fatalf("expected repair-id annotation to be set for in-progress repair")
	}
}
