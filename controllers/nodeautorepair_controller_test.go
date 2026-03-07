package controllers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
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
	node := &v1.Node{
		Labels: map[string]string{
			repairCooldownLabel: "5m30s",
		},
	}
	if got := getNodeCooldownDuration(node); got != 5*time.Minute+30*time.Second {
		t.Fatalf("expected custom duration 5m30s, got %v", got)
	}
}

func TestGetNodeCooldownDuration_MinutesValue(t *testing.T) {
	node := &v1.Node{
		Labels: map[string]string{
			repairCooldownLabel: "15",
		},
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
		if !strings.Contains(evt, eventRepairDisabled) {
			t.Fatalf("expected disabled event %q, got %q", eventRepairDisabled, evt)
		}
	default:
		t.Fatalf("expected event to be emitted when repair is disabled")
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
