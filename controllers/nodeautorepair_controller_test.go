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

// Test that handleUnhealthyNode throttles when last-repair-end is within cooldown window
func TestHandleUnhealthyNode_Throttled(t *testing.T) {
	now := time.Now().UTC()
	node := &v1.Node{}
	node.Name = "nar-throttle-node"
	node.Annotations = map[string]string{
		narLastRepairEndAnnotation: now.Add(-30 * time.Minute).Format(time.RFC3339),
	}
	// a sample unhealthy condition (type doesn't matter for this direct call)
	cond := &v1.NodeCondition{Type: v1.NodeReady, Status: v1.ConditionFalse}

	r := &NodeAutoRepairReconciler{}
	logger := logr.Discard()

	res, err := r.handleUnhealthyNode(context.Background(), logger, node, []*v1.NodeCondition{cond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter <= 0 {
		t.Fatalf("expected positive RequeueAfter due to throttling, got %v", res.RequeueAfter)
	}
	// remaining should be less than cooldown (default 60m)
	if res.RequeueAfter >= repairCoolDown {
		t.Fatalf("expected RequeueAfter < cooldown %v, got %v", repairCoolDown, res.RequeueAfter)
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

func TestGetNodeCooldownDuration(t *testing.T) {
	defaultCooldown := repairCoolDown

	t.Run("missing annotation falls back to default", func(t *testing.T) {
		node := &v1.Node{}
		if got := getNodeCooldownDuration(node); got != defaultCooldown {
			t.Fatalf("expected default cooldown %v, got %v", defaultCooldown, got)
		}
	})

	t.Run("integer minutes annotation", func(t *testing.T) {
		node := &v1.Node{
			Annotations: map[string]string{
				narCooldownAnnotationKey: "30",
			},
		}
		want := 30 * time.Minute
		if got := getNodeCooldownDuration(node); got != want {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("duration string annotation", func(t *testing.T) {
		node := &v1.Node{
			Annotations: map[string]string{
				narCooldownAnnotationKey: "15m",
			},
		}
		want := 15 * time.Minute
		if got := getNodeCooldownDuration(node); got != want {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("invalid annotation reverts to default", func(t *testing.T) {
		node := &v1.Node{
			Annotations: map[string]string{
				narCooldownAnnotationKey: "bad-value",
			},
		}
		if got := getNodeCooldownDuration(node); got != defaultCooldown {
			t.Fatalf("expected default cooldown %v for invalid annotation, got %v", defaultCooldown, got)
		}
	})
}
