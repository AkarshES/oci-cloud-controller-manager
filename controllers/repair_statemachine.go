package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	ociclientpkg "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubectl/pkg/drain"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	narStateAnnotationKey            = "oci.oraclecloud.com/nodeautorepair-state"
	narRepairIDAnnotationKey         = "oci.oraclecloud.com/nodeautorepair-repair-id"
	narRepairOriginAnnotationKey     = "oci.oraclecloud.com/nodeautorepair-repair-origin"
	narLastTransitionAnnotation      = "oci.oraclecloud.com/nodeautorepair-last-transition"
	narAttemptsAnnotationKey         = "oci.oraclecloud.com/nodeautorepair-attempts"
	narRebootIssuedAnnotationKey     = "oci.oraclecloud.com/nodeautorepair-reboot-issued"
	narStateMetadataAnnotationKey    = "oci.oraclecloud.com/nodeautorepair-state-meta"
	narRepairCycleAttemptsKey        = "oci.oraclecloud.com/nodeautorepair-cycle-attempts"
	narRepairCycleLockKey            = "oci.oraclecloud.com/nodeautorepair-cycle-lock"
	narUnhealthySinceAnnotationKey   = "oci.oraclecloud.com/nodeautorepair-unhealthy-since"
	narCordonedByRepairAnnotationKey = "oci.oraclecloud.com/nodeautorepair-cordoned-by-repair"
	narFailureCleanupPendingKey      = "oci.oraclecloud.com/nodeautorepair-cleanup-pending"
	narFailureCleanupStateKey        = "oci.oraclecloud.com/nodeautorepair-cleanup-failed-state"
	// Terminal repair summary annotations (preserved across cleanups)
	narLastRepairEndAnnotation    = "oci.oraclecloud.com/nodeautorepair-last-repair-end"
	narLastRepairResultAnnotation = "oci.oraclecloud.com/nodeautorepair-last-result"
	metricRepairTotal             = "nodeautorepair_repair_total"
	metricRepairFailures          = "nodeautorepair_repair_failures_total"
	metricRepairDuration          = "nodeautorepair_repair_duration_seconds"
	eventRepairDetected           = "NodeRepairDetected"
	eventRepairCordoned           = "NodeRepairCordoned"
	eventRepairDraining           = "NodeRepairDraining"
	eventRepairRebooting          = "NodeRepairRebooting"
	eventRepairUncordoned         = "NodeRepairUncordoned"
	eventRepairThrottled          = "NodeRepairThrottled"
	eventRepairDisabled           = "NodeRepairDisabled"
	eventRepairLeaseLost          = "NodeRepairLeaseLost"
	eventRepairSucceeded          = "NodeRepairSucceeded"
	eventRepairFailed             = "NodeRepairFailed"
	eventRepairHumanIntervention  = "NodeRepairHumanInterventionRequired"
)

type repairState string

const (
	stateDetected  repairState = "Detected"
	stateCordoning repairState = "Cordoning"
	stateDraining  repairState = "Draining"
	stateRebooting repairState = "Rebooting"
	stateUncordon  repairState = "Uncordoning"
	stateSucceeded repairState = "Succeeded"
	stateFailed    repairState = "Failed"
)

type stateConfig struct {
	timeout        time.Duration
	successRequeue time.Duration
	retryBase      time.Duration
}

var (
	maxRepairAttempts  = getEnvInt("NODE_AUTOREPAIR_MAX_ATTEMPTS", 3)
	maxRepairCycles    = getEnvInt("NODE_AUTOREPAIR_MAX_REPAIR_CYCLES", 3)
	defaultRetryBase   = getEnvDuration("NODE_AUTOREPAIR_RETRY_BASE", 10*time.Second)
	defaultRetryCap    = getEnvDuration("NODE_AUTOREPAIR_RETRY_CAP", 5*time.Minute)
	repairStateConfigs = map[repairState]stateConfig{
		stateCordoning: {
			timeout:        getEnvDuration("NODE_AUTOREPAIR_TIMEOUT_CORDONING", 60*time.Second),
			successRequeue: 10 * time.Second,
			retryBase:      defaultRetryBase,
		},
		stateDraining: {
			timeout:        getEnvDuration("NODE_AUTOREPAIR_TIMEOUT_DRAINING", 15*time.Minute),
			successRequeue: 10 * time.Second,
			retryBase:      defaultRetryBase,
		},
		stateRebooting: {
			timeout:        getEnvDuration("NODE_AUTOREPAIR_TIMEOUT_REBOOTING", 20*time.Minute),
			successRequeue: 30 * time.Second,
			retryBase:      defaultRetryBase,
		},
		stateUncordon: {
			timeout:        getEnvDuration("NODE_AUTOREPAIR_TIMEOUT_UNCORDONING", 5*time.Minute),
			successRequeue: 5 * time.Second,
			retryBase:      defaultRetryBase,
		},
	}
	instanceRunningPollInterval = getEnvDuration("NODE_AUTOREPAIR_REBOOT_POLL_INTERVAL", 20*time.Second)
)

var (
	repairAnnotationKeys = []string{
		narStateAnnotationKey,
		narRepairIDAnnotationKey,
		narRepairOriginAnnotationKey,
		narLastTransitionAnnotation,
		narAttemptsAnnotationKey,
		narRebootIssuedAnnotationKey,
		narStateMetadataAnnotationKey,
		narRepairCycleAttemptsKey,
		narRepairCycleLockKey,
		narUnhealthySinceAnnotationKey,
		narCordonedByRepairAnnotationKey,
		narFailureCleanupPendingKey,
		narFailureCleanupStateKey,
	}
	// Respect PDB up to 10 minutes by default, then force repair per design doc
	drainForceAfter         = getEnvDuration("NODE_AUTOREPAIR_DRAIN_FORCE_AFTER", 10*time.Minute)
	drainForceAlways        = getEnvBool("NODE_AUTOREPAIR_DRAIN_FORCE", false)
	drainIgnoreDaemonSets   = getEnvBool("NODE_AUTOREPAIR_DRAIN_IGNORE_DAEMONSETS", true)
	drainDeleteEmptyDirData = getEnvBool("NODE_AUTOREPAIR_DRAIN_DELETE_EMPTYDIR", false)
	drainAttemptTimeout     = getEnvDuration("NODE_AUTOREPAIR_DRAIN_ATTEMPT_TIMEOUT", 60*time.Second)
	healthyStabilization    = getEnvDuration("NODE_AUTOREPAIR_HEALTHY_STABILIZATION", 75*time.Second)
)

type repairStateTracker struct {
	RepairID string                             `json:"repairId,omitempty"`
	States   map[repairState]*stateTrackerEntry `json:"states,omitempty"`
}

type stateTrackerEntry struct {
	StartTime    string `json:"startTime,omitempty"`
	Attempts     int    `json:"attempts,omitempty"`
	Forced       bool   `json:"forced,omitempty"`
	HealthySince string `json:"healthySince,omitempty"`
}

func (sm *nodeRepairStateMachine) refreshStateTracker() {
	sm.tracker = loadRepairStateTracker(sm.node)
	if sm.tracker != nil {
		sm.tracker.RepairID = sm.node.Annotations[narRepairIDAnnotationKey]
		sm.tracker.ensureStart(sm.currentState(), sm.lastTransitionTime())
	}
}

func loadRepairStateTracker(node *v1.Node) *repairStateTracker {
	if node == nil || node.Annotations == nil {
		return newRepairStateTracker("")
	}
	meta := node.Annotations[narStateMetadataAnnotationKey]
	if meta == "" {
		return newRepairStateTracker(node.Annotations[narRepairIDAnnotationKey])
	}
	tracker := &repairStateTracker{}
	if err := json.Unmarshal([]byte(meta), tracker); err != nil {
		return newRepairStateTracker(node.Annotations[narRepairIDAnnotationKey])
	}
	if tracker.States == nil {
		tracker.States = map[repairState]*stateTrackerEntry{}
	}
	if tracker.RepairID == "" {
		tracker.RepairID = node.Annotations[narRepairIDAnnotationKey]
	}
	return tracker
}

func newRepairStateTracker(repairID string) *repairStateTracker {
	return &repairStateTracker{
		RepairID: repairID,
		States:   map[repairState]*stateTrackerEntry{},
	}
}

func (t *repairStateTracker) current(state repairState) *stateTrackerEntry {
	if t == nil {
		return nil
	}
	return t.States[state]
}

func (t *repairStateTracker) begin(state repairState) *stateTrackerEntry {
	if t == nil {
		return nil
	}
	entry := t.States[state]
	if entry == nil {
		entry = &stateTrackerEntry{}
		t.States[state] = entry
	}
	entry.StartTime = time.Now().UTC().Format(time.RFC3339)
	entry.Attempts = 0
	return entry
}

func (t *repairStateTracker) reset(state repairState) {
	if t == nil {
		return
	}
	delete(t.States, state)
}

func (t *repairStateTracker) ensureStart(state repairState, fallback time.Time) *stateTrackerEntry {
	if t == nil {
		return nil
	}
	entry := t.States[state]
	if entry == nil {
		entry = &stateTrackerEntry{}
		t.States[state] = entry
	}
	if entry.StartTime == "" && !fallback.IsZero() {
		entry.StartTime = fallback.UTC().Format(time.RFC3339)
	}
	return entry
}

func (t *repairStateTracker) toJSON() string {
	if t == nil {
		return ""
	}
	bytes, err := json.Marshal(t)
	if err != nil {
		return ""
	}
	return string(bytes)
}

type nodeRepairStateMachine struct {
	reconciler *NodeAutoRepairReconciler
	node       *v1.Node
	nodeKey    client.ObjectKey
	logger     logr.Logger
	repairID   string
	originID   string
	tracker    *repairStateTracker
}

// leaseRenewer is a tiny interface to allow unit testing of lease ownership checks
// by substituting a fake that implements Renew.
type leaseRenewer interface {
	Renew(ctx context.Context, nodeName string) error
}

// l returns a logger with consistent contextual fields for easier troubleshooting
func (sm *nodeRepairStateMachine) l() logr.Logger {
	return sm.logger.WithValues(
		"node", sm.node.Name,
		"repairID", sm.repairID,
		"state", sm.currentState(),
		"attempts", sm.currentAttempts(),
	)
}

func newNodeRepairStateMachine(r *NodeAutoRepairReconciler, node *v1.Node, logger logr.Logger) *nodeRepairStateMachine {
	sm := &nodeRepairStateMachine{
		reconciler: r,
		node:       node,
		nodeKey:    client.ObjectKey{Name: node.Name},
		logger:     logger,
		repairID:   node.Annotations[narRepairIDAnnotationKey],
		originID:   node.Annotations[narRepairOriginAnnotationKey],
	}
	sm.refreshStateTracker()
	sm.ensureStateStart(sm.currentState())
	return sm
}

func (sm *nodeRepairStateMachine) Run(ctx context.Context) (ctrl.Result, error) {
	if sm.node.Annotations == nil {
		sm.node.Annotations = map[string]string{}
	}

	state := sm.currentState()
	sm.l().Info("Current repair state " + string(state))
	if state == "" {
		if err := sm.setState(ctx, stateDetected); err != nil {
			return ctrl.Result{}, err
		}
		state = stateDetected
	}

	switch state {
	case stateDetected:
		return sm.handleDetected(ctx)
	case stateCordoning:
		return sm.handleCordoning(ctx)
	case stateDraining:
		return sm.handleDraining(ctx)
	case stateRebooting:
		return sm.handleRebooting(ctx)
	case stateUncordon:
		return sm.handleUncordoning(ctx)
	case stateSucceeded:
		// Ensure any global repair lease is released in terminal state
		if err := sm.reconciler.stopLeaseHeartbeat(ctx, sm.node.Name); err != nil {
			sm.l().Error(err, "Failed to stop lease heartbeat in Succeeded state")
		}
		sm.l().Info("Node auto repair already succeeded")
		return ctrl.Result{}, nil
	case stateFailed:
		// Ensure any global repair lease is released in terminal state
		if err := sm.reconciler.stopLeaseHeartbeat(ctx, sm.node.Name); err != nil {
			sm.l().Error(err, "Failed to stop lease heartbeat in Failed state")
		}
		sm.l().Info("Node auto repair is in failed state. Awaiting manual remediation")
		return ctrl.Result{}, nil
	default:
		sm.l().Info(fmt.Sprintf("Unknown node auto repair state, resetting (observedState=%s)", state))
		if err := sm.setState(ctx, stateDetected); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRetryBase}, nil
	}
}

func (sm *nodeRepairStateMachine) handleDetected(ctx context.Context) (ctrl.Result, error) {
	// Verify we still own the global repair lease before proceeding
	sm.l().Info("CCM: Handle detected state")
	if err := sm.ensureLeaseOwned(ctx); err != nil {
		return ctrl.Result{RequeueAfter: defaultRetryBase}, nil
	}
	if err := sm.ensureRepairID(ctx); err != nil {
		return ctrl.Result{}, err
	}
	sm.repairID = sm.node.Annotations[narRepairIDAnnotationKey]
	sm.tracker = newRepairStateTracker(sm.repairID)
	sm.ensureStateStart(stateCordoning)
	sm.emitEvent(eventRepairDetected, "Detected unhealthy node; transitioning to Cordoning")
	sm.l().Info("CCM: Emit detected event and transition to cordoning")
	sm.recordMetric(metricRepairTotal, 1)
	if err := sm.setState(ctx, stateCordoning); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: repairStateConfigs[stateCordoning].successRequeue}, nil
}

func (sm *nodeRepairStateMachine) beginState(ctx context.Context, state repairState) {
	if sm.tracker == nil {
		sm.tracker = newRepairStateTracker(sm.repairID)
	}
	// Preserve the original state start time so timeout windows are measured
	// from state entry, not from every reconcile retry.
	entry := sm.tracker.current(state)
	if entry != nil && entry.StartTime != "" {
		return
	}
	entry = sm.tracker.ensureStart(state, sm.lastTransitionTime())
	if entry == nil || entry.StartTime == "" {
		entry = sm.tracker.begin(state)
	}
	_ = sm.persistTracker(ctx)
}

func (sm *nodeRepairStateMachine) ensureStateStart(state repairState) {
	if sm.tracker == nil {
		sm.tracker = newRepairStateTracker(sm.repairID)
	}
	sm.tracker.ensureStart(state, sm.lastTransitionTime())
}

func (sm *nodeRepairStateMachine) persistTracker(ctx context.Context) error {
	return sm.updateAnnotations(ctx, func(ann map[string]string) {
		ann[narStateMetadataAnnotationKey] = sm.tracker.toJSON()
	})
}

func (sm *nodeRepairStateMachine) validateRepairID(ctx context.Context) error {
	if sm.node.Annotations == nil {
		return fmt.Errorf("node %s missing annotations for repair tracking", sm.node.Name)
	}
	annotatedID := sm.node.Annotations[narRepairIDAnnotationKey]
	if annotatedID == "" {
		return fmt.Errorf("node %s missing repair id annotation", sm.node.Name)
	}
	if sm.repairID == "" {
		sm.repairID = annotatedID
	}
	if sm.repairID != annotatedID {
		return fmt.Errorf("node %s repair id changed (expected %s, found %s)", sm.node.Name, sm.repairID, annotatedID)
	}
	annotationOrigin := sm.node.Annotations[narRepairOriginAnnotationKey]
	if annotationOrigin != "" && sm.reconciler.ControllerID != "" && annotationOrigin != sm.reconciler.ControllerID {
		// Allow takeover if this controller is currently orchestrating the repair (global lease
		// was acquired in handleUnhealthyNode before entering the state machine). Update origin.
		if err := sm.updateAnnotations(ctx, func(ann map[string]string) {
			ann[narRepairOriginAnnotationKey] = sm.reconciler.ControllerID
		}); err != nil {
			return fmt.Errorf("node %s repair appears owned by %s; takeover by %s failed: %w", sm.node.Name, annotationOrigin, sm.reconciler.ControllerID, err)
		}
	}
	if sm.tracker == nil {
		sm.tracker = newRepairStateTracker(sm.repairID)
	} else if sm.tracker.RepairID == "" || sm.tracker.RepairID != sm.repairID {
		sm.tracker = newRepairStateTracker(sm.repairID)
	}
	return nil
}

func (sm *nodeRepairStateMachine) handleCordoning(ctx context.Context) (ctrl.Result, error) {
	sm.l().Info("CCM: Handle cordoning state")
	if err := sm.validateRepairID(ctx); err != nil {
		return ctrl.Result{}, err
	}
	if err := sm.ensureLeaseOwned(ctx); err != nil {
		return ctrl.Result{RequeueAfter: defaultRetryBase}, nil
	}
	sm.beginState(ctx, stateCordoning)
	if sm.stateTimedOut(stateCordoning) {
		return sm.failState(ctx, stateCordoning, fmt.Errorf("cordoning timed out"))
	}
	attempt, err := sm.recordAttempt(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if attempt > maxRepairAttempts {
		return sm.failState(ctx, stateCordoning, fmt.Errorf("cordoning exceeded maximum attempts"))
	}
	wasSchedulable := !sm.node.Spec.Unschedulable
	if err := sm.cordonNode(ctx); err != nil {
		sm.l().Error(err, fmt.Sprintf("Cordoning node failed (attempt=%d)", attempt))
		sm.emitWarningEvent(eventRepairCordoned, fmt.Sprintf("Cordoning failed (attempt %d): %v", attempt, err))
		return ctrl.Result{RequeueAfter: sm.retryDelay(stateCordoning, attempt)}, nil
	}
	if wasSchedulable {
		if err := sm.setCordonedByRepair(ctx, true); err != nil {
			return ctrl.Result{}, err
		}
	}
	sm.l().Info("CCM: Emit cordoning event and transition to draining")
	sm.emitEvent(eventRepairCordoned, "Node cordoned; moving to Draining")
	sm.recordStateDuration(stateCordoning)
	if err := sm.setState(ctx, stateDraining); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: repairStateConfigs[stateDraining].successRequeue}, nil
}

func (sm *nodeRepairStateMachine) handleDraining(ctx context.Context) (ctrl.Result, error) {
	sm.l().Info("CCM: Handle draining state")
	if err := sm.validateRepairID(ctx); err != nil {
		return ctrl.Result{}, err
	}
	if err := sm.ensureLeaseOwned(ctx); err != nil {
		return ctrl.Result{RequeueAfter: defaultRetryBase}, nil
	}
	sm.beginState(ctx, stateDraining)
	if sm.stateTimedOut(stateDraining) {
		return sm.failState(ctx, stateDraining, fmt.Errorf("draining timed out"))
	}
	attempt, err := sm.recordAttempt(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := sm.drainNode(ctx); err != nil {
		sm.l().Error(err, fmt.Sprintf("Draining node failed (attempt=%d)", attempt))
		sm.emitWarningEvent(eventRepairDraining, fmt.Sprintf("Draining failed (attempt %d): %v", attempt, err))
		return ctrl.Result{RequeueAfter: sm.retryDelay(stateDraining, attempt)}, nil
	}
	sm.l().Info("CCM: Emit draining event and transition to rebooting")
	sm.emitEvent(eventRepairDraining, "Drain succeeded; moving to Rebooting")
	sm.l().Info("CCM: Record state duration for draining")
	sm.recordStateDuration(stateDraining)
	if err := sm.setState(ctx, stateRebooting); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: repairStateConfigs[stateRebooting].successRequeue}, nil
}

func (sm *nodeRepairStateMachine) handleRebooting(ctx context.Context) (ctrl.Result, error) {
	sm.l().Info("CCM: Handle rebooting state")
	if err := sm.validateRepairID(ctx); err != nil {
		return ctrl.Result{}, err
	}
	if err := sm.ensureLeaseOwned(ctx); err != nil {
		return ctrl.Result{RequeueAfter: defaultRetryBase}, nil
	}
	sm.beginState(ctx, stateRebooting)
	if sm.stateTimedOut(stateRebooting) {
		return sm.failState(ctx, stateRebooting, fmt.Errorf("rebooting timed out"))
	}
	if !sm.rebootIssued() {
		attempt, err := sm.recordAttempt(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		if attempt > maxRepairAttempts {
			return sm.failState(ctx, stateRebooting, fmt.Errorf("rebooting exceeded maximum attempts"))
		}
		workRequestID, err := sm.triggerReboot(ctx)
		if err != nil {
			sm.l().Error(err, fmt.Sprintf("Reboot request failed (attempt=%d)", attempt))
			sm.emitWarningEvent(eventRepairRebooting, fmt.Sprintf("Reboot failed (attempt %d): %v", attempt, err))
			return ctrl.Result{RequeueAfter: sm.retryDelay(stateRebooting, attempt)}, nil
		}
		sm.emitEvent(eventRepairRebooting, fmt.Sprintf("Reboot work request %s submitted", workRequestID))
		sm.l().Info(fmt.Sprintf("CCM: Reboot issued, recording annotation and polling for instance state (workRequestID=%s)", workRequestID))
		if err := sm.setRebootIssued(ctx, true); err != nil {
			return ctrl.Result{}, err
		}
		// After submitting reboot, poll the instance state periodically
		return ctrl.Result{RequeueAfter: instanceRunningPollInterval}, nil
	}
	running, err := sm.instanceIsRunning(ctx)
	if err != nil {
		sm.l().Error(err, "Failed to query instance state")
		return ctrl.Result{RequeueAfter: instanceRunningPollInterval}, nil
	}
	if !running {
		return ctrl.Result{RequeueAfter: instanceRunningPollInterval}, nil
	}
	if err := sm.setRebootIssued(ctx, false); err != nil {
		return ctrl.Result{}, err
	}
	if err := sm.setState(ctx, stateUncordon); err != nil {
		return ctrl.Result{}, err
	}
	sm.recordStateDuration(stateRebooting)
	return ctrl.Result{RequeueAfter: repairStateConfigs[stateUncordon].successRequeue}, nil
}

func (sm *nodeRepairStateMachine) handleUncordoning(ctx context.Context) (ctrl.Result, error) {
	if err := sm.validateRepairID(ctx); err != nil {
		return ctrl.Result{}, err
	}
	if err := sm.ensureLeaseOwned(ctx); err != nil {
		return ctrl.Result{RequeueAfter: defaultRetryBase}, nil
	}
	sm.beginState(ctx, stateUncordon)
	cleanupPending := sm.failureCleanupPending()
	if !cleanupPending && sm.stateTimedOut(stateUncordon) {
		return sm.failState(ctx, stateUncordon, fmt.Errorf("uncordoning timed out"))
	}
	if !cleanupPending {
		if unhealthyConditions := findUnhealthyConditions(sm.node); len(unhealthyConditions) > 0 {
			if sm.clearHealthySince(ctx, stateUncordon) {
				sm.l().Info("Cleared healthy stabilization window because unhealthy conditions returned")
			}
			sm.l().Info(fmt.Sprintf(
				"Node still has unhealthy conditions after reboot; delaying uncordon (conditions=%s)",
				summarizeConditionTypes(conditionTypeValues(unhealthyConditions)),
			))
			return ctrl.Result{RequeueAfter: repairStateConfigs[stateUncordon].successRequeue}, nil
		}
		if healthyStabilization > 0 {
			stable, remaining, err := sm.ensureHealthyFor(ctx, stateUncordon, healthyStabilization)
			if err != nil {
				return ctrl.Result{}, err
			}
			if !stable {
				sm.l().Info(fmt.Sprintf(
					"Node is healthy but waiting for stabilization window before uncordon (required=%s remaining=%s)",
					healthyStabilization,
					remaining.Truncate(time.Second),
				))
				return ctrl.Result{RequeueAfter: repairStateConfigs[stateUncordon].successRequeue}, nil
			}
		}
	}
	if cleanupPending && !sm.node.Spec.Unschedulable {
		sm.recordStateDuration(stateUncordon)
		return sm.completeFailedCycle(ctx, sm.failureCleanupState(), fmt.Errorf("repair cleanup completed after node was already schedulable"))
	}
	attempt, err := sm.recordAttempt(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !cleanupPending && attempt > maxRepairAttempts {
		return sm.failState(ctx, stateUncordon, fmt.Errorf("uncordoning exceeded maximum attempts"))
	}
	if err := sm.uncordonNode(ctx); err != nil {
		if cleanupPending {
			sm.l().Error(err, fmt.Sprintf("Cleanup uncordon failed (attempt=%d)", attempt))
			sm.emitWarningEvent(eventRepairUncordoned, fmt.Sprintf("Cleanup uncordon failed (attempt %d): %v", attempt, err))
			return ctrl.Result{RequeueAfter: sm.retryDelay(stateUncordon, attempt)}, nil
		}
		sm.l().Error(err, fmt.Sprintf("Uncordoning node failed (attempt=%d)", attempt))
		sm.emitWarningEvent(eventRepairUncordoned, fmt.Sprintf("Uncordoning failed (attempt %d): %v", attempt, err))
		return ctrl.Result{RequeueAfter: sm.retryDelay(stateUncordon, attempt)}, nil
	}
	sm.recordStateDuration(stateUncordon)
	if cleanupPending {
		sm.emitEvent(eventRepairUncordoned, "Node uncordoned during failed repair cleanup")
		return sm.completeFailedCycle(ctx, sm.failureCleanupState(), fmt.Errorf("repair cycle failed and node was uncordoned during cleanup"))
	}
	sm.emitEvent(eventRepairUncordoned, "Node uncordoned; marking repair succeeded")
	if err := sm.setState(ctx, stateSucceeded); err != nil {
		return ctrl.Result{}, err
	}
	// Remove repair taint so the node becomes schedulable again
	if err := sm.removeRepairTaint(ctx); err != nil {
		sm.l().Error(err, "Failed to remove repair taint after success")
	}
	// Emit success event to mark completion
	sm.emitEvent(eventRepairSucceeded, "Node auto repair succeeded; repair taint removed")
	// Release global repair lease immediately upon success
	if err := sm.reconciler.stopLeaseHeartbeat(ctx, sm.node.Name); err != nil {
		sm.l().Error(err, "Failed to stop lease heartbeat after Succeeded")
	}
	// Finalize: record last repair end/result and prune transient annotations
	if err := sm.finalizeRepair(ctx, "succeeded"); err != nil {
		sm.l().Error(err, "Failed to finalize repair annotations after success")
	}
	sm.resetCycleAttempts(ctx)
	return ctrl.Result{}, nil
}

func (sm *nodeRepairStateMachine) ensureHealthyFor(ctx context.Context, state repairState, window time.Duration) (bool, time.Duration, error) {
	if sm.tracker == nil {
		sm.tracker = newRepairStateTracker(sm.repairID)
	}
	entry := sm.tracker.ensureStart(state, sm.lastTransitionTime())
	if entry == nil {
		return false, 0, nil
	}
	now := time.Now().UTC()
	if entry.HealthySince == "" {
		entry.HealthySince = now.Format(time.RFC3339)
		if err := sm.persistTracker(ctx); err != nil {
			return false, 0, err
		}
		return false, window, nil
	}
	healthySince, err := time.Parse(time.RFC3339, entry.HealthySince)
	if err != nil {
		entry.HealthySince = now.Format(time.RFC3339)
		if persistErr := sm.persistTracker(ctx); persistErr != nil {
			return false, 0, persistErr
		}
		return false, window, nil
	}
	elapsed := now.Sub(healthySince)
	if elapsed >= window {
		return true, 0, nil
	}
	return false, window - elapsed, nil
}

func (sm *nodeRepairStateMachine) clearHealthySince(ctx context.Context, state repairState) bool {
	if sm.tracker == nil {
		return false
	}
	entry := sm.tracker.current(state)
	if entry == nil || entry.HealthySince == "" {
		return false
	}
	entry.HealthySince = ""
	if err := sm.persistTracker(ctx); err != nil {
		sm.l().Error(err, "Failed to clear healthy stabilization timestamp")
	}
	return true
}

func (sm *nodeRepairStateMachine) stateTimedOut(state repairState) bool {
	cfg, ok := repairStateConfigs[state]
	if !ok || cfg.timeout == 0 {
		return false
	}
	entry := sm.tracker.current(state)
	if entry == nil || entry.StartTime == "" {
		return false
	}
	start, err := time.Parse(time.RFC3339, entry.StartTime)
	if err != nil {
		return false
	}
	return time.Since(start) > cfg.timeout
}

func (sm *nodeRepairStateMachine) retryDelay(state repairState, attempt int) time.Duration {
	cfg, ok := repairStateConfigs[state]
	base := defaultRetryBase
	if ok && cfg.retryBase > 0 {
		base = cfg.retryBase
	}
	if attempt < 1 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 5 {
		shift = 5
	}
	delay := base * time.Duration(1<<shift)
	if delay > defaultRetryCap {
		delay = defaultRetryCap
	}
	return delay
}

func (sm *nodeRepairStateMachine) ensureRepairID(ctx context.Context) error {
	if _, ok := sm.node.Annotations[narRepairIDAnnotationKey]; ok {
		return nil
	}
	return sm.updateAnnotations(ctx, func(ann map[string]string) {
		ann[narRepairIDAnnotationKey] = string(uuid.NewUUID())
		ann[narRepairOriginAnnotationKey] = sm.reconciler.ControllerID
	})
}

func (sm *nodeRepairStateMachine) currentState() repairState {
	if sm.node.Annotations == nil {
		return stateDetected
	}
	if state, ok := sm.node.Annotations[narStateAnnotationKey]; ok && state != "" {
		return repairState(state)
	}
	return stateDetected
}

func (sm *nodeRepairStateMachine) lastTransitionTime() time.Time {
	if sm.node.Annotations == nil {
		return time.Time{}
	}
	if val, ok := sm.node.Annotations[narLastTransitionAnnotation]; ok && val != "" {
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (sm *nodeRepairStateMachine) setState(ctx context.Context, next repairState) error {
	prev := sm.currentState()
	sm.l().Info(fmt.Sprintf("Transitioning node auto repair state from %s to %s", prev, next))
	return sm.updateAnnotations(ctx, func(ann map[string]string) {
		ann[narStateAnnotationKey] = string(next)
		ann[narLastTransitionAnnotation] = time.Now().UTC().Format(time.RFC3339)
		ann[narAttemptsAnnotationKey] = "0"
		if next == stateCordoning {
			if _, ok := ann[narRepairCycleAttemptsKey]; !ok {
				ann[narRepairCycleAttemptsKey] = "0"
			}
		}
		if sm.tracker == nil {
			sm.tracker = newRepairStateTracker(sm.repairID)
		}
		sm.tracker.reset(next)
		sm.tracker.ensureStart(next, time.Now())
		ann[narStateMetadataAnnotationKey] = sm.tracker.toJSON()
	})
}

func (sm *nodeRepairStateMachine) recordAttempt(ctx context.Context) (int, error) {
	current := sm.currentAttempts()
	newVal := current + 1
	return newVal, sm.updateAnnotations(ctx, func(ann map[string]string) {
		ann[narAttemptsAnnotationKey] = strconv.Itoa(newVal)
		if sm.tracker != nil {
			entry := sm.tracker.current(sm.currentState())
			if entry != nil {
				entry.Attempts = newVal
				ann[narStateMetadataAnnotationKey] = sm.tracker.toJSON()
			}
		}
	})
}

// incrementCycleFailure increments the per-node failed repair cycle counter and
// returns the updated value.
func (sm *nodeRepairStateMachine) incrementCycleFailure(ctx context.Context) (int, error) {
	next := sm.currentCycleAttempts() + 1
	return next, sm.updateAnnotations(ctx, func(ann map[string]string) {
		ann[narRepairCycleAttemptsKey] = strconv.Itoa(next)
		delete(ann, narRepairCycleLockKey)
	})
}

func (sm *nodeRepairStateMachine) currentAttempts() int {
	if sm.node.Annotations == nil {
		return 0
	}
	val, ok := sm.node.Annotations[narAttemptsAnnotationKey]
	if !ok {
		return 0
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return parsed
}

func (sm *nodeRepairStateMachine) currentCycleAttempts() int {
	if sm.node.Annotations == nil {
		return 0
	}
	val, ok := sm.node.Annotations[narRepairCycleAttemptsKey]
	if !ok {
		return 0
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return parsed
}

func (sm *nodeRepairStateMachine) wasCordonedByRepair() bool {
	if sm.node.Annotations == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(sm.node.Annotations[narCordonedByRepairAnnotationKey]), "true")
}

func (sm *nodeRepairStateMachine) setCordonedByRepair(ctx context.Context, cordoned bool) error {
	return sm.updateAnnotations(ctx, func(ann map[string]string) {
		if cordoned {
			ann[narCordonedByRepairAnnotationKey] = "true"
			return
		}
		delete(ann, narCordonedByRepairAnnotationKey)
	})
}

func (sm *nodeRepairStateMachine) failureCleanupPending() bool {
	if sm.node.Annotations == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(sm.node.Annotations[narFailureCleanupPendingKey]), "true")
}

func (sm *nodeRepairStateMachine) failureCleanupState() repairState {
	if sm.node.Annotations == nil {
		return stateFailed
	}
	if state := strings.TrimSpace(sm.node.Annotations[narFailureCleanupStateKey]); state != "" {
		return repairState(state)
	}
	return stateFailed
}

func (sm *nodeRepairStateMachine) failState(ctx context.Context, state repairState, reason error) (ctrl.Result, error) {
	if sm.needsFailureCleanupUncordon() {
		return sm.enterFailureCleanup(ctx, state, reason)
	}
	sm.recordStateDuration(state)
	return sm.completeFailedCycle(ctx, state, reason)
}

func (sm *nodeRepairStateMachine) needsFailureCleanupUncordon() bool {
	return sm.wasCordonedByRepair() && sm.node.Spec.Unschedulable
}

func (sm *nodeRepairStateMachine) enterFailureCleanup(ctx context.Context, failedState repairState, reason error) (ctrl.Result, error) {
	sm.l().Error(reason, fmt.Sprintf("Repair state failed; entering cleanup uncordon (state=%s)", failedState))
	if failedState != stateUncordon {
		sm.recordStateDuration(failedState)
	}
	if err := sm.updateAnnotations(ctx, func(ann map[string]string) {
		ann[narFailureCleanupPendingKey] = "true"
		ann[narFailureCleanupStateKey] = string(failedState)
	}); err != nil {
		return ctrl.Result{}, err
	}
	if sm.currentState() != stateUncordon {
		if err := sm.setState(ctx, stateUncordon); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{RequeueAfter: repairStateConfigs[stateUncordon].successRequeue}, nil
}

func (sm *nodeRepairStateMachine) completeFailedCycle(ctx context.Context, failedState repairState, reason error) (ctrl.Result, error) {
	sm.l().Error(reason, fmt.Sprintf("Repair cycle failed (state=%s)", failedState))
	sm.emitWarningEvent(eventRepairFailed, fmt.Sprintf("State %s failed: %v", failedState, reason))
	sm.recordMetric(metricRepairFailures, 1)
	failures, err := sm.incrementCycleFailure(ctx)
	if err != nil {
		sm.l().Error(err, "Failed to record cycle failure")
		return ctrl.Result{}, err
	}
	if err := sm.setState(ctx, stateFailed); err != nil {
		return ctrl.Result{}, err
	}
	if err := sm.removeRepairTaint(ctx); err != nil {
		sm.l().Error(err, "Failed to remove repair taint after failed cycle")
	}
	humanIntervention := failures >= maxRepairCycles
	if humanIntervention {
		if err := sm.markHumanInterventionRequired(ctx); err != nil {
			return ctrl.Result{}, err
		}
		sm.emitWarningEvent(eventRepairHumanIntervention, fmt.Sprintf("Repair failed %d consecutive times; manual intervention required", failures))
	}
	// Release global repair lease immediately upon failure
	if err := sm.reconciler.stopLeaseHeartbeat(ctx, sm.node.Name); err != nil {
		sm.l().Error(err, "Failed to stop lease heartbeat after Failed")
	}
	// Finalize: record last repair end/result and prune transient annotations
	if err := sm.finalizeRepair(ctx, "failed"); err != nil {
		sm.l().Error(err, "Failed to finalize repair annotations after failure")
		return ctrl.Result{}, err
	}
	if humanIntervention {
		sm.resetCycleAttempts(ctx)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{RequeueAfter: unhealthyThreshold}, nil
}

func (sm *nodeRepairStateMachine) cordonNode(ctx context.Context) error {
	return sm.patchNode(ctx, func(node *v1.Node) {
		if !node.Spec.Unschedulable {
			node.Spec.Unschedulable = true
		}
	})
}

func (sm *nodeRepairStateMachine) uncordonNode(ctx context.Context) error {
	return sm.patchNode(ctx, func(node *v1.Node) {
		if node.Spec.Unschedulable {
			node.Spec.Unschedulable = false
		}
	})
}

func (sm *nodeRepairStateMachine) drainNode(ctx context.Context) error {
	childCtx, cancel := context.WithTimeout(ctx, drainAttemptTimeout)
	defer cancel()
	// Determine whether to force drain (skip PDB by disabling evictions) based on elapsed time
	// in current Draining state, or if globally configured to always force.
	forcedMode := drainForceAlways
	if !forcedMode {
		// Use the state's last transition time (entering Draining) for a stable start timestamp
		// regardless of reconcile retries.
		start := sm.lastTransitionTime()
		if !start.IsZero() && drainForceAfter > 0 && time.Since(start) >= drainForceAfter {
			forcedMode = true
		}
	}
	helper := &drain.Helper{
		Ctx:    childCtx,
		Client: sm.reconciler.KubeClient,
		// If forcedMode, we both set Force and DisableEviction to bypass PDB protection
		// and proceed with deletions when evictions cannot make progress.
		Force:               forcedMode || drainForceAlways,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: drainIgnoreDaemonSets,
		DeleteEmptyDirData:  drainDeleteEmptyDirData,
		DisableEviction:     forcedMode,
		Timeout:             drainAttemptTimeout,
		Out:                 nopWriter{},
		ErrOut:              nopWriter{},
	}
	if forcedMode {
		// Mark we have switched to forced draining in the tracker metadata and emit an event once.
		var emitted bool
		if sm.tracker != nil {
			// Ensure the entry exists without clobbering its start when available.
			sm.tracker.ensureStart(stateDraining, sm.lastTransitionTime())
			if entry := sm.tracker.current(stateDraining); entry != nil {
				if !entry.Forced {
					entry.Forced = true
					emitted = true
					// Best-effort persist; ignore error here, actual drain proceeds regardless.
					_ = sm.persistTracker(ctx)
				}
			}
		}
		if emitted {
			sm.emitEvent(eventRepairDraining, "PDB wait exceeded; forcing evictions and deletions")
		}
	}
	return drain.RunNodeDrain(helper, sm.node.Name)
}

func (sm *nodeRepairStateMachine) triggerReboot(ctx context.Context) (string, error) {
	providerID := sm.node.Spec.ProviderID
	if providerID == "" {
		return "", fmt.Errorf("node providerID is empty")
	}
	instanceID := ociclientpkg.MapProviderIDToInstanceID(providerID)
	req := core.InstanceActionRequest{
		InstanceId:      &instanceID,
		Action:          common.String("SOFTRESET"),
		RequestMetadata: common.RequestMetadata{},
	}
	response, err := sm.reconciler.OCIClient.Compute().InstanceAction(ctx, req)
	if err != nil {
		return "", err
	}
	if response.OpcRequestId == nil || *response.OpcRequestId == "" {
		return "", errors.New("instance reboot response missing work request id")
	}
	return *response.OpcRequestId, nil
}

func (sm *nodeRepairStateMachine) rebootIssued() bool {
	if sm.node.Annotations == nil {
		return false
	}
	return sm.node.Annotations[narRebootIssuedAnnotationKey] == "true"
}

func (sm *nodeRepairStateMachine) setRebootIssued(ctx context.Context, issued bool) error {
	return sm.updateAnnotations(ctx, func(ann map[string]string) {
		if issued {
			ann[narRebootIssuedAnnotationKey] = "true"
		} else {
			delete(ann, narRebootIssuedAnnotationKey)
		}
	})
}

func (sm *nodeRepairStateMachine) instanceIsRunning(ctx context.Context) (bool, error) {
	providerID := sm.node.Spec.ProviderID
	if providerID == "" {
		return false, fmt.Errorf("node providerID is empty")
	}
	id := ociclientpkg.MapProviderIDToInstanceID(providerID)
	instance, err := sm.reconciler.OCIClient.Compute().GetInstance(ctx, id)
	if err != nil {
		return false, err
	}
	return instance.LifecycleState == core.InstanceLifecycleStateRunning, nil
}

func (sm *nodeRepairStateMachine) updateAnnotations(ctx context.Context, mutate func(map[string]string)) error {
	return sm.patchNode(ctx, func(node *v1.Node) {
		if node.Annotations == nil {
			node.Annotations = map[string]string{}
		}
		mutate(node.Annotations)
	})
}

func (sm *nodeRepairStateMachine) patchNode(ctx context.Context, mutate func(*v1.Node)) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &v1.Node{}
		if err := sm.reconciler.Client.Get(ctx, sm.nodeKey, latest); err != nil {
			return err
		}
		original := latest.DeepCopy()
		mutate(latest)
		if err := sm.reconciler.Client.Patch(ctx, latest, client.MergeFrom(original)); err != nil {
			return err
		}
		*sm.node = *latest
		return nil
	})
}

func (sm *nodeRepairStateMachine) emitEvent(reason, message string) {
	decoratedMsg := sm.decorateMessage(message)
	decoratedReason := sm.decorateReason(reason)
	sm.l().Info(fmt.Sprintf("Emitting repair event type=%s reason=%s message=%s", v1.EventTypeNormal, decoratedReason, decoratedMsg))
	if sm.reconciler.Recorder == nil {
		sm.l().Info("Event recorder not available; skipping event emission")
		return
	}
	sm.reconciler.Recorder.Event(sm.node, v1.EventTypeNormal, decoratedReason, decoratedMsg)
	sm.l().Info(fmt.Sprintf("Emitted repair event type=%s reason=%s message=%s", v1.EventTypeNormal, decoratedReason, decoratedMsg))
}

func (sm *nodeRepairStateMachine) emitWarningEvent(reason, message string) {
	decoratedMsg := sm.decorateMessage(message)
	decoratedReason := sm.decorateReason(reason)
	sm.l().Info(fmt.Sprintf("Emitting repair event type=%s reason=%s message=%s", v1.EventTypeWarning, decoratedReason, decoratedMsg))
	if sm.reconciler.Recorder == nil {
		return
	}
	sm.reconciler.Recorder.Event(sm.node, v1.EventTypeWarning, decoratedReason, decoratedMsg)
}

func (sm *nodeRepairStateMachine) recordMetric(metric string, value float64) {
	if sm.reconciler.MetricPusher == nil {
		return
	}
	dimensions := map[string]string{
		metrics.ComponentDimension: "nodeautorepair",
	}
	if sm.reconciler.Config != nil {
		dimensions[metrics.ClusterOCID] = sm.reconciler.Config.ClusterID
	}
	if sm.node.Spec.ProviderID != "" {
		dimensions[metrics.InstanceIdDimension] = sm.node.Spec.ProviderID
	}
	metrics.SendMetricData(sm.reconciler.MetricPusher, metric, value, dimensions)
}

func (sm *nodeRepairStateMachine) decorateMessage(msg string) string {
	if sm.repairID == "" {
		return msg
	}
	return fmt.Sprintf("[repair-id:%s] %s", sm.repairID, msg)
}

// decorateReason appends a repair-id-based suffix so each repair emits a distinct
// event reason, preventing aggregation from collapsing state transitions.
func (sm *nodeRepairStateMachine) decorateReason(reason string) string {
	if sm.repairID == "" || reason == "" {
		return reason
	}
	suffix := strings.ToUpper(strings.ReplaceAll(sm.repairID, "-", ""))
	if len(suffix) > 12 {
		suffix = suffix[len(suffix)-12:]
	}
	return fmt.Sprintf("%s_%s", reason, suffix)
}

// ensureLeaseOwned verifies that the reconciler still holds the global repair lease
// for this node. If not, it stops the heartbeat and emits a warning event, then returns error
// so the caller can safely stop progressing the state machine.
func (sm *nodeRepairStateMachine) ensureLeaseOwned(ctx context.Context) error {
	if sm.reconciler == nil || sm.reconciler.leaseManager == nil {
		return nil
	}
	var r leaseRenewer = sm.reconciler.leaseManager
	if err := r.Renew(ctx, sm.node.Name); err != nil {
		// Lost ownership (or lease missing). Stop our heartbeat to avoid noise and back off.
		_ = sm.reconciler.stopLeaseHeartbeat(ctx, sm.node.Name)
		sm.emitWarningEvent(eventRepairLeaseLost, "Lost global repair lease; pausing state progression")
		sm.l().Info("Lost global repair lease; stopping current repair progression")
		return fmt.Errorf("repair lease not owned: %w", err)
	}
	return nil
}

func (sm *nodeRepairStateMachine) recordStateDuration(state repairState) {
	if sm.reconciler.MetricPusher == nil {
		return
	}
	start := sm.lastTransitionTime()
	if start.IsZero() {
		return
	}
	duration := time.Since(start).Seconds()
	dims := map[string]string{
		metrics.ComponentDimension: "nodeautorepair",
		"state":                    string(state),
		"repair_id":                sm.repairID,
	}
	if sm.reconciler.Config != nil {
		dims[metrics.ClusterOCID] = sm.reconciler.Config.ClusterID
	}
	if sm.node.Spec.ProviderID != "" {
		dims[metrics.InstanceIdDimension] = sm.node.Spec.ProviderID
	}
	metrics.SendMetricData(sm.reconciler.MetricPusher, metricRepairDuration, duration, dims)
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// removeRepairTaint removes the NoSchedule taint used to mark a node under repair
func (sm *nodeRepairStateMachine) removeRepairTaint(ctx context.Context) error {
	return sm.patchNode(ctx, func(node *v1.Node) {
		if len(node.Spec.Taints) == 0 {
			return
		}
		kept := make([]v1.Taint, 0, len(node.Spec.Taints))
		for _, t := range node.Spec.Taints {
			if t.Key == REPAIR_TAINT_KEY && t.Effect == REPAIR_TAINT_EFFECT {
				continue
			}
			kept = append(kept, t)
		}
		node.Spec.Taints = kept
	})
}

func (sm *nodeRepairStateMachine) markHumanInterventionRequired(ctx context.Context) error {
	return sm.patchNode(ctx, func(node *v1.Node) {
		if node.Labels == nil {
			node.Labels = map[string]string{}
		}
		node.Labels[repairHumanInterventionLabel] = "true"
	})
}

func nodeAnnotationsToPrune(node *v1.Node) []string {
	if node.Annotations == nil {
		return nil
	}
	var keys []string
	for _, key := range repairAnnotationKeys {
		if _, ok := node.Annotations[key]; ok {
			keys = append(keys, key)
		}
	}
	return keys
}

// finalizeRepair records the terminal timestamp/result and prunes transient repair annotations.
func (sm *nodeRepairStateMachine) finalizeRepair(ctx context.Context, result string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	return sm.updateAnnotations(ctx, func(ann map[string]string) {
		// Record terminal metadata
		ann[narLastRepairEndAnnotation] = now
		if result != "" {
			ann[narLastRepairResultAnnotation] = result
		}
		// Prune transient annotations
		for _, k := range repairAnnotationKeys {
			// Preserve cycle counters/lock on failure so we can accumulate cycles
			if result == "failed" && (k == narRepairCycleAttemptsKey || k == narRepairCycleLockKey) {
				continue
			}
			delete(ann, k)
		}
	})
}

func (sm *nodeRepairStateMachine) resetCycleAttempts(ctx context.Context) {
	_ = sm.updateAnnotations(ctx, func(ann map[string]string) {
		delete(ann, narRepairCycleAttemptsKey)
		delete(ann, narRepairCycleLockKey)
	})
}

func getEnvInt(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		switch strings.ToLower(val) {
		case "1", "true", "t", "yes", "y":
			return true
		case "0", "false", "f", "no", "n":
			return false
		}
	}
	return def
}
