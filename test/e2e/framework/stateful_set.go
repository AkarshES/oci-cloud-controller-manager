package framework

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	port                          int32 = 80
	terminationGracePeriodSeconds int64 = 30
	GracePeriodSeconds            int64 = 15
)

func (j *PVCTestJig) CreateService(namespace string) string {

	req := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: j.Name,
			Namespace:    namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: port,
					Name: "serviceport",
				},
			},
			Selector: j.Labels,
		},
	}

	resp, err := j.KubeClient.CoreV1().Services(namespace).Create(context.Background(), &req, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())

	return resp.Name
}

func createContainerTemplate(name string) []corev1.Container {
	return []corev1.Container{
		{
			Args: []string{
				"-c",
				"while true; do echo $(date -u) >> /data/out.txt; sleep 5; done",
			},
			Command: []string{
				"/bin/sh",
			},
			Image:                    centos,
			ImagePullPolicy:          "Always",
			Name:                     name,
			TerminationMessagePath:   "/dev/termination-log",
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      name,
					MountPath: "/data",
				},
			},
		},
	}
}

// CreateAndAwaitStatefulSet Creates StatefulSet for FVP
func (j *PVCTestJig) CreateAndAwaitStatefulSet(appLabel, serviceName, scName, size, adLabel string, replicas int32) *v1.StatefulSet {
	pvc := j.NewPVCTemplate(namespace, size, scName, adLabel)
	pvc.ObjectMeta = metav1.ObjectMeta{
		Name: j.Name + "-0",
	}
	return j.createAndAwaitStatefulSet(appLabel, serviceName, pvc, replicas)
}

// CreateAndAwaitStatefulSetDynamicFss Creates StatefulSet for FSS-Dynamic
func (j *PVCTestJig) CreateAndAwaitStatefulSetDynamicFss(appLabel, serviceName, scName, size string, replicas int32) *v1.StatefulSet {
	pvc := j.NewPVCTemplateDynamicFSS(namespace, size, scName)
	pvc.ObjectMeta = metav1.ObjectMeta{
		Name: j.Name + "-0",
	}
	Logf("ocid: %s", mntTargetOCID)
	return j.createAndAwaitStatefulSet(appLabel, serviceName, pvc, replicas)
}

// CreateAndAwaitStatefulSetCSI Creates StatefulSet for CSI
func (j *PVCTestJig) CreateAndAwaitStatefulSetCSI(appLabel, serviceName, scName, size string, replicas int32, volumeMode corev1.PersistentVolumeMode, accessMode corev1.PersistentVolumeAccessMode) *v1.StatefulSet {
	pvc := j.NewPVCTemplateCSI(namespace, size, scName, volumeMode, accessMode)
	pvc.ObjectMeta = metav1.ObjectMeta{
		Name: j.Name + "-0",
	}
	return j.createAndAwaitStatefulSet(appLabel, serviceName, pvc, replicas)
}

func (j *PVCTestJig) createAndAwaitStatefulSet(appLabel string, serviceName string, pvc *corev1.PersistentVolumeClaim, replicas int32) *v1.StatefulSet {
	containers := createContainerTemplate(j.Name + "-0")

	req := v1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: j.Name,
			Namespace:    namespace,
		},
		Spec: v1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": appLabel,
				},
			},
			ServiceName: serviceName,
			UpdateStrategy: v1.StatefulSetUpdateStrategy{
				Type: v1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": appLabel,
					},
				},
				Spec: corev1.PodSpec{
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 "Always",
					SchedulerName:                 "default-scheduler",
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Containers:                    containers,
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{*pvc},
		},
	}

	resp, err := j.KubeClient.AppsV1().StatefulSets(namespace).Create(context.Background(), &req, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = j.waitTimeoutForStsReadyInNamespace(resp.Name, namespace, 15*time.Minute)
	if err != nil {
		Failf("Statefulset %q is not Ready: %v", resp.Name, err)
	}
	zap.S().With(resp.Namespace).With(resp.Name).Info("Statefulset is created.")

	return resp
}

// Waits for the Statefulset to have all the replicas in running state
func (j *PVCTestJig) waitTimeoutForStsReadyInNamespace(stsName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, j.stsRunning(stsName, namespace))
}

func (j *PVCTestJig) stsRunning(stsName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		sts, err := j.KubeClient.AppsV1().StatefulSets(namespace).Get(context.Background(), stsName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if sts.Status.Size() == 0 {
			return false, fmt.Errorf("statefulset %s Status field is empty, retrying in %v", stsName, Poll)
		}
		if sts.Status.ReadyReplicas == *sts.Spec.Replicas {
			return true, nil
		}
		return false, nil
	}
}

// ValidateExistingResources Checks for statefulsets in the UpgradeTestingNamespace and verifies if all are in ready sate.
func (j *PVCTestJig) ValidateExistingResources() {

	Logf("Checking Pre-Upgrade Resources...")
	statefulSets, err := j.KubeClient.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	Logf("Found %d statefulsets...", statefulSets.Size())

	for _, sts := range statefulSets.Items {
		Logf("Statefulset: %s", sts.Name)
		Logf("Spec.Replicas = %d, ReadyReplicas = %d", *sts.Spec.Replicas, sts.Status.Replicas)
		if *sts.Spec.Replicas != sts.Status.ReadyReplicas {
			Failf("Number of ready replicas doesn't match the number of replicas defined in the Spec for statefulSet %s", sts.Name)
		}
	}

	Logf("Validated Existing resources.")
}

// RestartExistingResources For post-upgrade testing. Restarts one pod from each statefulset found in the namespace.
func (j *PVCTestJig) RestartExistingResources() {

	statefulsets, err := j.KubeClient.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	for _, sts := range statefulsets.Items {
		Logf("Checking statefulset: %s...", sts.Name)
		podName := sts.Name + "-0"

		pod, err := j.KubeClient.CoreV1().Pods(sts.Namespace).Get(context.Background(), podName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		nodeName := pod.Spec.NodeName

		Logf("Cordoning Node: %s", nodeName)
		j.CordonOrUncordonNode(nodeName, true)

		Logf("Deleting sts pod %s to restart", podName)
		err = j.KubeClient.CoreV1().Pods(sts.Namespace).Delete(context.Background(), podName, metav1.DeleteOptions{
			GracePeriodSeconds: &GracePeriodSeconds,
		})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Minute)

		err = j.waitTimeoutForStsReadyInNamespace(sts.Name, sts.Namespace, slowPodStartTimeout)
		if err != nil {
			Failf("Statefulset %q is not Ready: %v", sts.Name, err)
		}
		zap.S().With(namespace).With(sts).Info("Statefulset is ready.")

		Logf("Uncordoning Node: %s", nodeName)
		j.CordonOrUncordonNode(nodeName, false)
	}
}

func (j *PVCTestJig) CordonOrUncordonNode(nodeName string, cordon bool) {
	if cordon {
		KubectlCmd("cordon", nodeName)
	} else {
		KubectlCmd("uncordon", nodeName)
	}
}
