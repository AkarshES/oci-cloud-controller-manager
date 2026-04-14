package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	kubectldrain "k8s.io/kubectl/pkg/drain"
)

const (
	legacyPDBMultipleBudgetError = "This pod has more than one PodDisruptionBudget, which the eviction subresource does not support."
)

type drainOutcome string

const (
	drainOutcomeInProgress drainOutcome = "in_progress"
	drainOutcomeBackoff    drainOutcome = "backoff"
	drainOutcomeComplete   drainOutcome = "complete"
	drainOutcomeHardError  drainOutcome = "hard_error"
)

type drainTrackerStatus string

const (
	drainTrackerStatusInProgress drainTrackerStatus = "in_progress"
	drainTrackerStatusBackoff    drainTrackerStatus = "backoff"
	drainTrackerStatusComplete   drainTrackerStatus = "complete"
	drainTrackerStatusSkipped    drainTrackerStatus = "skipped"
)

type drainFailureType string

const (
	drainFailureTypePDB                  drainFailureType = "pdb"
	drainFailureTypeNamespaceTerminating drainFailureType = "namespace_terminating"
	drainFailureTypeHardError            drainFailureType = "hard_error"
)

type drainTrackerState struct {
	Status         drainTrackerStatus `json:"status,omitempty"`
	FailureType    drainFailureType   `json:"failureType,omitempty"`
	RemainingPods  int                `json:"remainingPods,omitempty"`
	BlockingPod    string             `json:"blockingPod,omitempty"`
	LastMessage    string             `json:"lastMessage,omitempty"`
	LastUpdateTime string             `json:"lastUpdateTime,omitempty"`
}

type drainResult struct {
	Outcome       drainOutcome
	Status        drainTrackerStatus
	FailureType   drainFailureType
	RemainingPods int
	BlockingPod   string
	Message       string
	Forced        bool
}

type nodeDrainExecutor struct {
	client               kubernetes.Interface
	checkEvictionSupport func(kubernetes.Interface) (schema.GroupVersion, error)
}

var drainCheckEvictionSupport = kubectldrain.CheckEvictionSupport

func newNodeDrainExecutor(client kubernetes.Interface) *nodeDrainExecutor {
	return &nodeDrainExecutor{
		client:               client,
		checkEvictionSupport: drainCheckEvictionSupport,
	}
}

func (e *nodeDrainExecutor) DrainNode(ctx context.Context, nodeName string, force bool) (drainResult, error) {
	helper := &kubectldrain.Helper{
		Ctx:                             ctx,
		Client:                          e.client,
		Force:                           force,
		GracePeriodSeconds:              -1,
		IgnoreAllDaemonSets:             drainIgnoreDaemonSets,
		DeleteEmptyDirData:              drainDeleteEmptyDirData,
		Timeout:                         drainAttemptTimeout,
		DisableEviction:                 force,
		Out:                             nopWriter{},
		ErrOut:                          nopWriter{},
		AdditionalFilters:               nil,
		SkipWaitForDeleteTimeoutSeconds: 0,
	}

	list, errs := helper.GetPodsForDeletion(nodeName)
	if len(errs) > 0 {
		err := utilerrors.NewAggregate(errs)
		return hardDrainResult(force, 0, "", fmt.Sprintf("drain candidate filtering failed: %v", err)), err
	}

	pods := list.Pods()
	if len(pods) == 0 {
		msg := "all drainable pods removed"
		if warnings := trimDrainMessage(list.Warnings()); warnings != "" {
			msg = warnings
		}
		return drainResult{
			Outcome:       drainOutcomeComplete,
			Status:        drainTrackerStatusComplete,
			RemainingPods: 0,
			Message:       msg,
			Forced:        force,
		}, nil
	}

	evictionGV := schema.GroupVersion{}
	if !force {
		gv, err := e.checkEvictionSupport(e.client)
		if err != nil {
			wrapped := fmt.Errorf("checking eviction support: %w", err)
			return hardDrainResult(force, len(pods), "", wrapped.Error()), wrapped
		}
		evictionGV = gv
	}

	acc := newDrainAccumulator(len(pods))
	for _, pod := range pods {
		podRef := podRef(pod)
		if !pod.ObjectMeta.DeletionTimestamp.IsZero() {
			acc.observeInProgress(podRef, fmt.Sprintf("pod %s is already terminating", podRef))
			continue
		}

		if force || evictionGV.Empty() {
			if err := helper.DeletePod(pod); err != nil {
				if apierrors.IsNotFound(err) {
					acc.observeInProgress(podRef, fmt.Sprintf("pod %s no longer exists", podRef))
					continue
				}
				wrapped := fmt.Errorf("deleting pod %s: %w", podRef, err)
				return acc.hardResult(force, podRef, wrapped.Error()), wrapped
			}
			acc.observeInProgress(podRef, fmt.Sprintf("pod %s deletion started", podRef))
			continue
		}

		if err := helper.EvictPod(pod, evictionGV); err != nil {
			switch {
			case apierrors.IsNotFound(err):
				acc.observeInProgress(podRef, fmt.Sprintf("pod %s no longer exists", podRef))
			case apierrors.IsTooManyRequests(err):
				acc.observeBackoff(drainFailureTypePDB, podRef, fmt.Sprintf("pod %s is blocked by a PodDisruptionBudget", podRef))
			case apierrors.IsForbidden(err) && apierrors.HasStatusCause(err, corev1.NamespaceTerminatingCause):
				if !pod.ObjectMeta.DeletionTimestamp.IsZero() {
					acc.observeInProgress(podRef, fmt.Sprintf("pod %s is already terminating", podRef))
				} else {
					acc.observeBackoff(drainFailureTypeNamespaceTerminating, podRef, fmt.Sprintf("pod %s is in a terminating namespace", podRef))
				}
			case apierrors.IsInternalError(err) && isPDBConflictError(err):
				acc.observeBackoff(drainFailureTypePDB, podRef, fmt.Sprintf("pod %s is blocked by a PodDisruptionBudget", podRef))
			default:
				wrapped := fmt.Errorf("evicting pod %s: %w", podRef, err)
				return acc.hardResult(force, podRef, wrapped.Error()), wrapped
			}
			continue
		}

		acc.observeInProgress(podRef, fmt.Sprintf("pod %s eviction started", podRef))
	}

	return acc.result(force), nil
}

type drainAccumulator struct {
	remainingPods int
	outcome       drainOutcome
	failureType   drainFailureType
	blockingPod   string
	message       string
}

func newDrainAccumulator(remainingPods int) *drainAccumulator {
	return &drainAccumulator{remainingPods: remainingPods}
}

func (a *drainAccumulator) observeInProgress(podRef, message string) {
	a.maybeUpdate(drainOutcomeInProgress, "", podRef, message)
}

func (a *drainAccumulator) observeBackoff(failureType drainFailureType, podRef, message string) {
	a.maybeUpdate(drainOutcomeBackoff, failureType, podRef, message)
}

func (a *drainAccumulator) hardResult(force bool, podRef, message string) drainResult {
	return hardDrainResult(force, a.remainingPods, podRef, message)
}

func (a *drainAccumulator) result(force bool) drainResult {
	outcome := a.outcome
	status := drainTrackerStatusInProgress
	if outcome == "" {
		outcome = drainOutcomeInProgress
	}
	if outcome == drainOutcomeBackoff || outcome == drainOutcomeHardError {
		status = drainTrackerStatusBackoff
	}

	return drainResult{
		Outcome:       outcome,
		Status:        status,
		FailureType:   a.failureType,
		RemainingPods: a.remainingPods,
		BlockingPod:   a.blockingPod,
		Message:       trimDrainMessage(a.message),
		Forced:        force,
	}
}

func (a *drainAccumulator) maybeUpdate(outcome drainOutcome, failureType drainFailureType, podRef, message string) {
	nextPriority := drainPriority(outcome, failureType)
	currentPriority := drainPriority(a.outcome, a.failureType)
	if nextPriority < currentPriority {
		return
	}
	if nextPriority > currentPriority || a.message == "" {
		a.outcome = outcome
		a.failureType = failureType
		a.blockingPod = podRef
		a.message = trimDrainMessage(message)
	}
}

func drainPriority(outcome drainOutcome, failureType drainFailureType) int {
	switch outcome {
	case drainOutcomeHardError:
		return 40
	case drainOutcomeBackoff:
		switch failureType {
		case drainFailureTypeNamespaceTerminating:
			return 30
		case drainFailureTypePDB:
			return 20
		default:
			return 15
		}
	case drainOutcomeInProgress:
		return 10
	default:
		return 0
	}
}

func hardDrainResult(force bool, remainingPods int, podRef, message string) drainResult {
	return drainResult{
		Outcome:       drainOutcomeHardError,
		Status:        drainTrackerStatusBackoff,
		FailureType:   drainFailureTypeHardError,
		RemainingPods: remainingPods,
		BlockingPod:   podRef,
		Message:       trimDrainMessage(message),
		Forced:        force,
	}
}

func shouldForceDrain(start time.Time) bool {
	if !drainForceEnabled {
		return false
	}
	if drainForceAfter <= 0 {
		return true
	}
	if start.IsZero() {
		return false
	}
	return time.Since(start) >= drainForceAfter
}

func isPDBConflictError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "disruption budget") ||
		strings.Contains(msg, strings.ToLower(legacyPDBMultipleBudgetError)) ||
		strings.Contains(msg, "poddisruptionbudget")
}

func podRef(pod corev1.Pod) string {
	return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
}

func trimDrainMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	if len(msg) <= 240 {
		return msg
	}
	return msg[:237] + "..."
}
