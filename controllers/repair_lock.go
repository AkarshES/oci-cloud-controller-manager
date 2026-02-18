package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultRepairLeaseName      = "oci-nodeautorepair-lock"
	defaultRepairLeaseNamespace = "kube-system"
	defaultRepairLeaseDuration  = 2 * time.Minute
	repairLeaseNodeAnnotation   = "oci.oraclecloud.com/nodeautorepair-lock-node"
)

type repairLeaseManager struct {
	client        client.Client
	leaseName     string
	leaseNS       string
	holderID      string
	leaseDuration time.Duration
}

func newRepairLeaseManager(c client.Client, holderID string) *repairLeaseManager {
	leaseName := getEnvString("NODE_AUTOREPAIR_LEASE_NAME", defaultRepairLeaseName)
	leaseNS := getEnvString("NODE_AUTOREPAIR_LEASE_NAMESPACE", defaultRepairLeaseNamespace)
	leaseDuration := getEnvDuration("NODE_AUTOREPAIR_LEASE_DURATION", defaultRepairLeaseDuration)
	if leaseDuration <= 0 {
		leaseDuration = defaultRepairLeaseDuration
	}
	return &repairLeaseManager{
		client:        c,
		leaseName:     leaseName,
		leaseNS:       leaseNS,
		holderID:      holderID,
		leaseDuration: leaseDuration,
	}
}

func (m *repairLeaseManager) leaseKey() types.NamespacedName {
	return types.NamespacedName{Name: m.leaseName, Namespace: m.leaseNS}
}

// TryAcquire attempts to acquire the repair lease for the given node. It returns
// whether the caller now owns the lease and the name of the node currently
// recorded in the lease (if any).
func (m *repairLeaseManager) TryAcquire(ctx context.Context, nodeName string) (bool, string, error) {
	if m == nil {
		return true, "", nil
	}
	var activeNode string
	acquired := false
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		lease := &coordinationv1.Lease{}
		if err := m.client.Get(ctx, m.leaseKey(), lease); err != nil {
			if apierrors.IsNotFound(err) {
				if err := m.createLease(ctx, nodeName); err != nil {
					return err
				}
				acquired = true
				activeNode = nodeName
				return nil
			}
			return err
		}
		activeNode = lease.Annotations[repairLeaseNodeAnnotation]
		if m.leaseOwned(lease, nodeName) {
			if err := m.refreshLease(ctx, lease); err != nil {
				return err
			}
			acquired = true
			activeNode = nodeName
			return nil
		}
		if m.leaseExpired(lease) {
			if err := m.updateLease(ctx, lease, nodeName); err != nil {
				return err
			}
			acquired = true
			activeNode = nodeName
			return nil
		}
		acquired = false
		return nil
	})
	return acquired, activeNode, err
}

func (m *repairLeaseManager) Renew(ctx context.Context, nodeName string) error {
	if m == nil {
		return nil
	}
	lease := &coordinationv1.Lease{}
	if err := m.client.Get(ctx, m.leaseKey(), lease); err != nil {
		return err
	}
	if !m.leaseOwned(lease, nodeName) {
		return fmt.Errorf("repair lease held by another entity")
	}
	return m.refreshLease(ctx, lease)
}

func (m *repairLeaseManager) createLease(ctx context.Context, nodeName string) error {
	leasing := &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.leaseName,
			Namespace: m.leaseNS,
			Annotations: map[string]string{
				repairLeaseNodeAnnotation: nodeName,
			},
		},
		Spec: m.buildLeaseSpec(),
	}
	now := metav1.NewMicroTime(time.Now())
	leasing.Spec.HolderIdentity = pointerString(m.holderID)
	leasing.Spec.AcquireTime = &now
	leasing.Spec.RenewTime = &now
	if err := m.client.Create(ctx, leasing); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("lease already exists")
		}
		return err
	}
	return nil
}

func (m *repairLeaseManager) updateLease(ctx context.Context, lease *coordinationv1.Lease, nodeName string) error {
	orig := lease.DeepCopy()
	now := metav1.NewMicroTime(time.Now())
	lease.Spec.HolderIdentity = pointerString(m.holderID)
	lease.Spec.RenewTime = &now
	if lease.Spec.AcquireTime == nil {
		lease.Spec.AcquireTime = &now
	}
	lease.Spec.LeaseDurationSeconds = pointerInt32(int32(m.leaseDuration.Seconds()))
	if lease.Annotations == nil {
		lease.Annotations = map[string]string{}
	}
	lease.Annotations[repairLeaseNodeAnnotation] = nodeName
	lease.Spec.LeaseTransitions = pointerInt32(m.incrementTransitions(orig.Spec.LeaseTransitions))
	if err := m.client.Patch(ctx, lease, client.MergeFrom(orig)); err != nil {
		return err
	}
	return nil
}

func (m *repairLeaseManager) refreshLease(ctx context.Context, lease *coordinationv1.Lease) error {
	orig := lease.DeepCopy()
	now := metav1.NewMicroTime(time.Now())
	lease.Spec.RenewTime = &now
	if lease.Spec.LeaseDurationSeconds == nil {
		lease.Spec.LeaseDurationSeconds = pointerInt32(int32(m.leaseDuration.Seconds()))
	}
	return m.client.Patch(ctx, lease, client.MergeFrom(orig))
}

func (m *repairLeaseManager) leaseOwned(lease *coordinationv1.Lease, nodeName string) bool {
	if lease == nil {
		return false
	}
	if lease.Spec.HolderIdentity == nil || *lease.Spec.HolderIdentity != m.holderID {
		return false
	}
	if nodeName == "" {
		return true
	}
	return lease.Annotations[repairLeaseNodeAnnotation] == nodeName
}

func (m *repairLeaseManager) leaseExpired(lease *coordinationv1.Lease) bool {
	if lease == nil {
		return true
	}
	if lease.Spec.LeaseDurationSeconds == nil || lease.Spec.RenewTime == nil {
		return true
	}
	dur := time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second
	if dur <= 0 {
		return true
	}
	return lease.Spec.RenewTime.Add(dur).Before(time.Now())
}

func (m *repairLeaseManager) incrementTransitions(val *int32) int32 {
	if val == nil {
		return 1
	}
	return *val + 1
}

func (m *repairLeaseManager) buildLeaseSpec() coordinationv1.LeaseSpec {
	dur := int32(m.leaseDuration.Seconds())
	return coordinationv1.LeaseSpec{
		LeaseDurationSeconds: &dur,
	}
}

func (m *repairLeaseManager) ActiveNode(ctx context.Context) (string, error) {
	if m == nil {
		return "", nil
	}
	lease := &coordinationv1.Lease{}
	if err := m.client.Get(ctx, m.leaseKey(), lease); err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return lease.Annotations[repairLeaseNodeAnnotation], nil
}

func (m *repairLeaseManager) Release(ctx context.Context, nodeName string) error {
	if m == nil {
		return nil
	}
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		lease := &coordinationv1.Lease{}
		if err := m.client.Get(ctx, m.leaseKey(), lease); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if !m.leaseOwned(lease, nodeName) {
			return nil
		}
		return m.client.Delete(ctx, lease)
	})
}

func (m *repairLeaseManager) renewInterval() time.Duration {
	interval := m.leaseDuration / 3
	if interval < 30*time.Second {
		interval = 30 * time.Second
	}
	return interval
}


func getEnvString(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func pointerString(val string) *string {
	return &val
}

func pointerInt32(val int32) *int32 {
	return &val
}
