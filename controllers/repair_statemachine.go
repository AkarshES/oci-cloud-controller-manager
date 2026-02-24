package controllers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"bitbucket.oci.oraclecorp.com/oke/oke-common/ociclient"
	"github.com/go-logr/logr"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
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
	narStateAnnotationKey        = "oci.oraclecloud.com/nodeautorepair-state"
	narRepairIDAnnotationKey     = "oci.oraclecloud.com/nodeautorepair-repair-id"
	narLastTransitionAnnotation  = "oci.oraclecloud.com/nodeautorepair-last-transition"
	narAttemptsAnnotationKey     = "oci.oraclecloud.com/nodeautorepair-attempts"
	narRebootIssuedAnnotationKey = "oci.oraclecloud.com/nodeautorepair-reboot-issued"
	maxRepairAttempts            = 3
	defaultRetryBase             = 10 * time.Second
	defaultRetryCap              = 5 * time.Minute
	metricRepairTotal            = "nodeautorepair_repair_total"
	metricRepairFailures         = "nodeautorepair_repair_failures_total"
	eventRepairDetected          = "NodeRepairDetected"
	eventRepairCordoned          = "NodeRepairCordoned"
	eventRepairDraining          = "NodeRepairDraining"
	eventRepairRebooting         = "NodeRepairRebooting"
	eventRepairUncordoned        = "NodeRepairUncordoned"
	eventRepairFailed            = "NodeRepairFailed"
	instanceRunningPollInterval  = 15 * time.Second
	instanceRunningWait          = 2 * time.Minute
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

var repairStateConfigs = map[repairState]stateConfig{
	stateCordoning: {
		timeout:        30 * time.Second,
		successRequeue: 5 * time.Second,
		retryBase:      10 * time.Second,
	},
	stateDraining: {
		timeout:        10 * time.Minute,
		successRequeue: 10 * time.Second,
		retryBase:      20 * time.Second,
	},
	stateRebooting: {
		timeout:        5 * time.Minute,
		successRequeue: 30 * time.Second,
		retryBase:      30 * time.Second,
	},
	stateUncordon: {
		timeout:        30 * time.Second,
		successRequeue: 0,
		retryBase:      10 * time.Second,
	},
}

var repairAnnotationKeys = []string{
	narStateAnnotationKey,
	narRepairIDAnnotationKey,
	narLastTransitionAnnotation,
	narAttemptsAnnotationKey,
	narRebootIssuedAnnotationKey,
}

type nodeRepairStateMachine struct {
	reconciler *NodeAutoRepairReconciler
	node       *v1.Node
	nodeKey    client.ObjectKey
	logger     logr.Logger
}

func newNodeRepairStateMachine(r *NodeAutoRepairReconciler, node *v1.Node, logger logr.Logger) *nodeRepairStateMachine {
	return &nodeRepairStateMachine{
		reconciler: r,
		node:       node,
		nodeKey:    client.ObjectKey{Name: node.Name},
		logger:     logger,
	}
}

func (sm *nodeRepairStateMachine) Run(ctx context.Context) (ctrl.Result, error) {
	if sm.node.Annotations == nil {
		sm.node.Annotations = map[string]string{}
	}

	state := sm.currentState()
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
		sm.logger.Info("Node auto repair already succeeded", "node", sm.node.Name)
		return ctrl.Result{}, nil
	case stateFailed:
		sm.logger.Info("Node auto repair is in failed state. Awaiting manual remediation", "node", sm.node.Name)
		return ctrl.Result{}, nil
	default:
		sm.logger.Info("Unknown node auto repair state, resetting", "state", state, "node", sm.node.Name)
		if err := sm.setState(ctx, stateDetected); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRetryBase}, nil
	}
}

func (sm *nodeRepairStateMachine) handleDetected(ctx context.Context) (ctrl.Result, error) {
	if err := sm.ensureRepairID(ctx); err != nil {
		return ctrl.Result{}, err
	}
	sm.emitEvent(eventRepairDetected, "Detected unhealthy node; transitioning to Cordoning")
	sm.recordMetric(metricRepairTotal, 1)
	if err := sm.setState(ctx, stateCordoning); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: repairStateConfigs[stateCordoning].successRequeue}, nil
}

func (sm *nodeRepairStateMachine) handleCordoning(ctx context.Context) (ctrl.Result, error) {
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
	if err := sm.cordonNode(ctx); err != nil {
		sm.logger.Error(err, "Cordoning node failed", "attempt", attempt, "node", sm.node.Name)
		return ctrl.Result{RequeueAfter: sm.retryDelay(stateCordoning, attempt)}, nil
	}
	sm.emitEvent(eventRepairCordoned, "Node cordoned; moving to Draining")
	if err := sm.setState(ctx, stateDraining); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: repairStateConfigs[stateDraining].successRequeue}, nil
}

func (sm *nodeRepairStateMachine) handleDraining(ctx context.Context) (ctrl.Result, error) {
	if sm.stateTimedOut(stateDraining) {
		return sm.failState(ctx, stateDraining, fmt.Errorf("draining timed out"))
	}
	attempt, err := sm.recordAttempt(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if attempt > maxRepairAttempts {
		return sm.failState(ctx, stateDraining, fmt.Errorf("draining exceeded maximum attempts"))
	}
	if err := sm.drainNode(ctx); err != nil {
		sm.logger.Error(err, "Draining node failed", "attempt", attempt, "node", sm.node.Name)
		sm.emitEvent(eventRepairDraining, fmt.Sprintf("Draining failed (attempt %d): %v", attempt, err))
		return ctrl.Result{RequeueAfter: sm.retryDelay(stateDraining, attempt)}, nil
	}
	sm.emitEvent(eventRepairDraining, "Drain succeeded; moving to Rebooting")
	if err := sm.setState(ctx, stateRebooting); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: repairStateConfigs[stateRebooting].successRequeue}, nil
}

func (sm *nodeRepairStateMachine) handleRebooting(ctx context.Context) (ctrl.Result, error) {
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
			sm.logger.Error(err, "Reboot request failed", "attempt", attempt, "node", sm.node.Name)
			sm.emitEvent(eventRepairRebooting, fmt.Sprintf("Reboot failed (attempt %d): %v", attempt, err))
			return ctrl.Result{RequeueAfter: sm.retryDelay(stateRebooting, attempt)}, nil
		}
		sm.emitEvent(eventRepairRebooting, fmt.Sprintf("Reboot work request %s submitted", workRequestID))
		if err := sm.setRebootIssued(ctx, true); err != nil {
			return ctrl.Result{}, err
		}
		requeue := repairStateConfigs[stateRebooting].successRequeue
		if requeue == 0 {
			requeue = 30 * time.Second
		}
		return ctrl.Result{RequeueAfter: requeue}, nil
	}
	running, err := sm.instanceIsRunning(ctx)
	if err != nil {
		sm.logger.Error(err, "Failed to query instance state", "node", sm.node.Name)
		return ctrl.Result{RequeueAfter: sm.retryDelay(stateRebooting, 1)}, nil
	}
	if !running {
		return ctrl.Result{RequeueAfter: sm.retryDelay(stateRebooting, 1)}, nil
	}
	if err := sm.setRebootIssued(ctx, false); err != nil {
		return ctrl.Result{}, err
	}
	if err := sm.setState(ctx, stateUncordon); err != nil {
		return ctrl.Result{}, err
	}
	requeue := repairStateConfigs[stateRebooting].successRequeue
	if requeue == 0 {
		requeue = 30 * time.Second
	}
	return ctrl.Result{RequeueAfter: requeue}, nil
}

func (sm *nodeRepairStateMachine) handleUncordoning(ctx context.Context) (ctrl.Result, error) {
	if sm.stateTimedOut(stateUncordon) {
		return sm.failState(ctx, stateUncordon, fmt.Errorf("uncordoning timed out"))
	}
	attempt, err := sm.recordAttempt(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if attempt > maxRepairAttempts {
		return sm.failState(ctx, stateUncordon, fmt.Errorf("uncordoning exceeded maximum attempts"))
	}
	if err := sm.uncordonNode(ctx); err != nil {
		sm.logger.Error(err, "Uncordoning node failed", "attempt", attempt, "node", sm.node.Name)
		return ctrl.Result{RequeueAfter: sm.retryDelay(stateUncordon, attempt)}, nil
	}
	sm.emitEvent(eventRepairUncordoned, "Node uncordoned; marking repair succeeded")
	if err := sm.setState(ctx, stateSucceeded); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (sm *nodeRepairStateMachine) stateTimedOut(state repairState) bool {
	cfg, ok := repairStateConfigs[state]
	if !ok || cfg.timeout == 0 {
		return false
	}
	last := sm.lastTransitionTime()
	if last.IsZero() {
		return false
	}
	return time.Since(last) > cfg.timeout
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
	return sm.updateAnnotations(ctx, func(ann map[string]string) {
		ann[narStateAnnotationKey] = string(next)
		ann[narLastTransitionAnnotation] = time.Now().UTC().Format(time.RFC3339)
		ann[narAttemptsAnnotationKey] = "0"
	})
}

func (sm *nodeRepairStateMachine) recordAttempt(ctx context.Context) (int, error) {
	current := sm.currentAttempts()
	newVal := current + 1
	return newVal, sm.updateAnnotations(ctx, func(ann map[string]string) {
		ann[narAttemptsAnnotationKey] = strconv.Itoa(newVal)
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

func (sm *nodeRepairStateMachine) failState(ctx context.Context, state repairState, reason error) (ctrl.Result, error) {
	sm.logger.Error(reason, "Repair state failed", "state", state, "node", sm.node.Name)
	sm.emitEvent(eventRepairFailed, fmt.Sprintf("State %s failed: %v", state, reason))
	sm.recordMetric(metricRepairFailures, 1)
	if err := sm.setState(ctx, stateFailed); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
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
	helper := &drain.Helper{
		Ctx:                 ctx,
		Client:              sm.reconciler.KubeClient,
		Force:               false,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  false,
		Timeout:             repairStateConfigs[stateDraining].timeout,
		Out:                 nopWriter{},
		ErrOut:              nopWriter{},
	}
	return drain.RunNodeDrain(helper, sm.node.Name)
}

func (sm *nodeRepairStateMachine) triggerReboot(ctx context.Context) (string, error) {
	providerID := sm.node.Spec.ProviderID
	if providerID == "" {
		return "", fmt.Errorf("node providerID is empty")
	}
	req := core.InstanceActionRequest{
		InstanceId: &providerID,
		Action:     common.String("RESET"),
		RequestMetadata: common.RequestMetadata{
			RetryPolicy: ociclient.NewRetryPolicyWithMaxAttempts(3),
		},
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
	instance, err := sm.reconciler.OCIClient.Compute().GetInstance(ctx, providerID)
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
	if sm.reconciler.Recorder == nil {
		return
	}
	sm.reconciler.Recorder.Event(sm.node, v1.EventTypeNormal, reason, message)
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

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) {
	return len(p), nil
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
