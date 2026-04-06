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
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func newNodeAutoRepairTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	if err := coordinationv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add coordinationv1 scheme: %v", err)
	}
	return scheme
}

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

func TestGetNodeUnhealthyThresholdDuration_Default(t *testing.T) {
	orig := unhealthyThreshold
	unhealthyThreshold = 11 * time.Minute
	defer func() { unhealthyThreshold = orig }()

	node := &v1.Node{}
	if got := getNodeUnhealthyThresholdDuration(node); got != unhealthyThreshold {
		t.Fatalf("expected default unhealthy threshold %v, got %v", unhealthyThreshold, got)
	}
}

func TestGetNodeUnhealthyThresholdDuration_DurationString(t *testing.T) {
	node := &v1.Node{}
	node.Labels = map[string]string{
		repairUnhealthyThresholdLabel: "7m30s",
	}

	if got := getNodeUnhealthyThresholdDuration(node); got != 7*time.Minute+30*time.Second {
		t.Fatalf("expected custom unhealthy threshold 7m30s, got %v", got)
	}
}

func TestGetNodeUnhealthyThresholdDuration_MinutesValue(t *testing.T) {
	node := &v1.Node{}
	node.Labels = map[string]string{
		repairUnhealthyThresholdLabel: "12",
	}

	if got := getNodeUnhealthyThresholdDuration(node); got != 12*time.Minute {
		t.Fatalf("expected custom unhealthy threshold 12m, got %v", got)
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

func TestHandleUnhealthyNode_WaitsForContinuousUnhealthyWindow(t *testing.T) {
	orig := unhealthyThreshold
	unhealthyThreshold = 10 * time.Minute
	defer func() { unhealthyThreshold = orig }()

	scheme := newNodeAutoRepairTestScheme(t)
	node := &v1.Node{}
	node.Name = "nar-unhealthy-dwell-wait"
	node.Status.Conditions = []v1.NodeCondition{
		{
			Type:   v1.NodeConditionType("GPUCount"),
			Status: v1.ConditionTrue,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: fakeClient}

	res, err := r.handleUnhealthyNode(context.Background(), logr.Discard(), node, findUnhealthyConditions(node))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != unhealthyThreshold {
		t.Fatalf("expected requeueAfter %v, got %v", unhealthyThreshold, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if latest.Annotations[narUnhealthySinceAnnotationKey] == "" {
		t.Fatalf("expected unhealthy-since annotation to be set")
	}
	if latest.Annotations[narStateAnnotationKey] != "" {
		t.Fatalf("expected repair state to remain unset before threshold, got %q", latest.Annotations[narStateAnnotationKey])
	}
	if latest.Spec.Unschedulable {
		t.Fatalf("expected node not to be cordoned before threshold")
	}
}

func TestHandleUnhealthyNode_StartsRepairAfterContinuousUnhealthyWindow(t *testing.T) {
	orig := unhealthyThreshold
	unhealthyThreshold = 10 * time.Minute
	defer func() { unhealthyThreshold = orig }()

	scheme := newNodeAutoRepairTestScheme(t)
	node := &v1.Node{}
	node.Name = "nar-unhealthy-dwell-ready"
	node.Annotations = map[string]string{
		narUnhealthySinceAnnotationKey: time.Now().UTC().Add(-unhealthyThreshold - time.Minute).Format(time.RFC3339),
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

	res, err := r.handleUnhealthyNode(context.Background(), logr.Discard(), node, findUnhealthyConditions(node))
	defer func() {
		_ = r.stopLeaseHeartbeat(context.Background(), node.Name)
	}()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateCordoning].successRequeue {
		t.Fatalf("expected repair to start with requeueAfter %v, got %v", repairStateConfigs[stateCordoning].successRequeue, res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if got := latest.Annotations[narStateAnnotationKey]; got != string(stateCordoning) {
		t.Fatalf("expected node state %q, got %q", stateCordoning, got)
	}
	foundRepairTaint := false
	for _, taint := range latest.Spec.Taints {
		if taint.Key == REPAIR_TAINT_KEY && taint.Effect == REPAIR_TAINT_EFFECT {
			foundRepairTaint = true
			break
		}
	}
	if !foundRepairTaint {
		t.Fatalf("expected repair taint to be added once repair starts")
	}
}

func TestHandleUnhealthyNode_PreservesUnhealthySinceAcrossConditionChanges(t *testing.T) {
	orig := unhealthyThreshold
	unhealthyThreshold = 10 * time.Minute
	defer func() { unhealthyThreshold = orig }()

	scheme := newNodeAutoRepairTestScheme(t)
	since := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)
	node := &v1.Node{}
	node.Name = "nar-unhealthy-dwell-preserve"
	node.Annotations = map[string]string{
		narUnhealthySinceAnnotationKey: since,
	}
	node.Status.Conditions = []v1.NodeCondition{
		{
			Type:   v1.NodeConditionType("RDMALink"),
			Status: v1.ConditionTrue,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: fakeClient}

	res, err := r.handleUnhealthyNode(context.Background(), logr.Discard(), node, findUnhealthyConditions(node))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter <= 0 || res.RequeueAfter >= unhealthyThreshold {
		t.Fatalf("expected partial remaining dwell time, got %v", res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if got := latest.Annotations[narUnhealthySinceAnnotationKey]; got != since {
		t.Fatalf("expected unhealthy-since to remain %q, got %q", since, got)
	}
}

func TestHandleUnhealthyNode_UsesNodeSpecificUnhealthyThresholdLabel(t *testing.T) {
	orig := unhealthyThreshold
	unhealthyThreshold = 10 * time.Minute
	defer func() { unhealthyThreshold = orig }()

	scheme := newNodeAutoRepairTestScheme(t)
	node := &v1.Node{}
	node.Name = "nar-unhealthy-dwell-label"
	node.Labels = map[string]string{
		repairUnhealthyThresholdLabel: "2m",
	}
	node.Annotations = map[string]string{
		narUnhealthySinceAnnotationKey: time.Now().UTC().Add(-3 * time.Minute).Format(time.RFC3339),
	}
	node.Status.Conditions = []v1.NodeCondition{
		{
			Type:   v1.NodeConditionType("GPUCount"),
			Status: v1.ConditionTrue,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{
		Client:       fakeClient,
		ControllerID: "nar-test-controller",
	}

	res, err := r.handleUnhealthyNode(context.Background(), logr.Discard(), node, findUnhealthyConditions(node))
	defer func() {
		_ = r.stopLeaseHeartbeat(context.Background(), node.Name)
	}()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateCordoning].successRequeue {
		t.Fatalf("expected repair to start using node-specific threshold, got %v", res.RequeueAfter)
	}
}

func TestCleanupRepairArtifacts_ClearsUnhealthySinceWhenNodeHealthy(t *testing.T) {
	scheme := newNodeAutoRepairTestScheme(t)
	node := &v1.Node{}
	node.Name = "nar-unhealthy-dwell-clear"
	node.Annotations = map[string]string{
		narUnhealthySinceAnnotationKey: time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339),
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: fakeClient}

	if _, err := r.cleanupRepairArtifacts(context.Background(), logr.Discard(), node.DeepCopy()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if _, ok := latest.Annotations[narUnhealthySinceAnnotationKey]; ok {
		t.Fatalf("expected unhealthy-since annotation to be removed when node is healthy")
	}
}

// Test that cleanup logic will not propose pruning last-repair-end annotation
func TestSanitizeRepairProblemLabelValue_CommaDelimitedProblems(t *testing.T) {
	got := sanitizeRepairProblemLabelValue("PCIeLanes,RDMALinkSpeed")
	if got != "PCIeLanes-RDMALinkSpeed" {
		t.Fatalf("expected sanitized label value %q, got %q", "PCIeLanes-RDMALinkSpeed", got)
	}
}

func TestEnsureRepairMarkers_StoresSanitizedLabelAndFullAnnotation(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	node := &v1.Node{}
	node.Name = "nar-sanitize-markers"
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: fakeClient}

	if err := r.ensureRepairMarkers(context.Background(), node, "PCIeLanes,RDMALinkSpeed"); err != nil {
		t.Fatalf("unexpected error ensuring repair markers: %v", err)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if got := latest.Labels[repairProblemDetectedLabel]; got != "PCIeLanes-RDMALinkSpeed" {
		t.Fatalf("expected sanitized label value %q, got %q", "PCIeLanes-RDMALinkSpeed", got)
	}
	if got := latest.Annotations["oci.oraclecloud.com/nodeautorepair-problems"]; got != "PCIeLanes,RDMALinkSpeed" {
		t.Fatalf("expected full problem list annotation to remain unchanged, got %q", got)
	}
}

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

func TestConditionChangedPredicate_Update_TriggersOnHumanInterventionLabelRemoved(t *testing.T) {
	p := ConditionChangedPredicate{log: zap.NewNop().Sugar()}
	oldNode := &v1.Node{}
	oldNode.Name = "nar-predicate-human-removed"
	oldNode.Labels = map[string]string{
		repairHumanInterventionLabel: "true",
	}
	newNode := oldNode.DeepCopy()
	delete(newNode.Labels, repairHumanInterventionLabel)

	changed := p.Update(event.UpdateEvent{
		ObjectOld: oldNode,
		ObjectNew: newNode,
	})
	if !changed {
		t.Fatalf("expected update predicate to trigger when %s label is removed", repairHumanInterventionLabel)
	}
}

func TestConditionChangedPredicate_Update_TriggersOnUnhealthyThresholdLabelChange(t *testing.T) {
	p := ConditionChangedPredicate{log: zap.NewNop().Sugar()}
	oldNode := &v1.Node{}
	oldNode.Name = "nar-predicate-threshold-changed"
	oldNode.Labels = map[string]string{
		repairUnhealthyThresholdLabel: "10m",
	}
	newNode := oldNode.DeepCopy()
	newNode.Labels[repairUnhealthyThresholdLabel] = "5m"

	changed := p.Update(event.UpdateEvent{
		ObjectOld: oldNode,
		ObjectNew: newNode,
	})
	if !changed {
		t.Fatalf("expected update predicate to trigger when %s changes", repairUnhealthyThresholdLabel)
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
	scheme := newNodeAutoRepairTestScheme(t)

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

func TestHandleUnhealthyNode_HumanInterventionLabelBlocksNewRepair(t *testing.T) {
	node := &v1.Node{}
	node.Name = "nar-human-block"
	node.Labels = map[string]string{
		repairHumanInterventionLabel: "true",
	}
	node.Status.Conditions = []v1.NodeCondition{
		{
			Type:   v1.NodeConditionType("GPUCount"),
			Status: v1.ConditionTrue,
		},
	}

	res, err := (&NodeAutoRepairReconciler{}).handleUnhealthyNode(context.Background(), logr.Discard(), node, findUnhealthyConditions(node))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Requeue || res.RequeueAfter != 0 {
		t.Fatalf("expected no requeue when human intervention label is present, got %#v", res)
	}
}

func TestHandleUnhealthyNode_HumanInterventionRemovalWaitsForCooldown(t *testing.T) {
	origCooldown := repairCoolDown
	repairCoolDown = 30 * time.Minute
	defer func() { repairCoolDown = origCooldown }()

	scheme := newNodeAutoRepairTestScheme(t)
	node := &v1.Node{}
	node.Name = "nar-human-removed-cooldown"
	node.Annotations = map[string]string{
		narLastRepairResultAnnotation: "failed",
		narLastRepairEndAnnotation:    time.Now().UTC().Add(-10 * time.Minute).Format(time.RFC3339),
	}
	node.Status.Conditions = []v1.NodeCondition{
		{
			Type:   v1.NodeConditionType("GPUCount"),
			Status: v1.ConditionTrue,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{Client: fakeClient}

	res, err := r.handleUnhealthyNode(context.Background(), logr.Discard(), node, findUnhealthyConditions(node))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter <= 0 || res.RequeueAfter > 20*time.Minute {
		t.Fatalf("expected remaining cooldown requeue, got %v", res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if got := latest.Annotations[narLastRepairResultAnnotation]; got != "failed" {
		t.Fatalf("expected failed summary to remain until cooldown expires, got %q", got)
	}
}

func TestHandleUnhealthyNode_HumanInterventionRemovalResetsStateAndRestarts(t *testing.T) {
	origCooldown := repairCoolDown
	repairCoolDown = 20 * time.Minute
	defer func() { repairCoolDown = origCooldown }()

	scheme := newNodeAutoRepairTestScheme(t)
	node := &v1.Node{}
	node.Name = "nar-human-removed-restart"
	node.Labels = map[string]string{
		repairProblemDetectedLabel: "stale",
	}
	node.Annotations = map[string]string{
		narLastRepairResultAnnotation:                 "failed",
		narLastRepairEndAnnotation:                    time.Now().UTC().Add(-25 * time.Minute).Format(time.RFC3339),
		narUnhealthySinceAnnotationKey:                time.Now().UTC().Add(-40 * time.Minute).Format(time.RFC3339),
		"oci.oraclecloud.com/nodeautorepair-problems": "GPUCount",
	}
	node.Status.Conditions = []v1.NodeCondition{
		{
			Type:   v1.NodeConditionType("GPUCount"),
			Status: v1.ConditionTrue,
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
	r := &NodeAutoRepairReconciler{
		Client:       fakeClient,
		ControllerID: "nar-test-controller",
	}

	res, err := r.handleUnhealthyNode(context.Background(), logr.Discard(), node, findUnhealthyConditions(node))
	defer func() {
		_ = r.stopLeaseHeartbeat(context.Background(), node.Name)
	}()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequeueAfter != repairStateConfigs[stateCordoning].successRequeue {
		t.Fatalf("expected repair to restart immediately after cooldown expiry, got %v", res.RequeueAfter)
	}

	latest := &v1.Node{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: node.Name}, latest); err != nil {
		t.Fatalf("failed to fetch latest node: %v", err)
	}
	if got := latest.Annotations[narStateAnnotationKey]; got != string(stateCordoning) {
		t.Fatalf("expected node state %q after restart, got %q", stateCordoning, got)
	}
	if _, ok := latest.Annotations[narLastRepairEndAnnotation]; ok {
		t.Fatalf("expected last repair end annotation to be cleared before restart")
	}
	if _, ok := latest.Annotations[narLastRepairResultAnnotation]; ok {
		t.Fatalf("expected last repair result annotation to be cleared before restart")
	}
	if got := latest.Labels[repairProblemDetectedLabel]; got == "stale" {
		t.Fatalf("expected stale repair problem label to be reset before restart")
	}
}
