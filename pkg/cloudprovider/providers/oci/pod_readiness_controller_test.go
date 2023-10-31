package oci

import (
	"errors"
	"reflect"
	"testing"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/workqueue"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

func Test_getBackendHealthMap(t *testing.T) {
	conditionType := v1.PodConditionType("podReadinessCondition")
	testCases := []struct {
		desc             string
		backendSetHealth *client.GenericBackendSetHealth
		out              map[string]v1.PodCondition
	}{
		{
			desc: "warning backends",
			backendSetHealth: &client.GenericBackendSetHealth{
				WarningStateBackendNames:  []string{"10.0.10.1", "10.0.10.2"},
				CriticalStateBackendNames: []string{},
				UnknownStateBackendNames:  []string{},
			},
			out: map[string]v1.PodCondition{
				"10.0.10.1": {
					Type:   conditionType,
					Reason: "backend health is WARNING",
					Status: v1.ConditionFalse,
				},
				"10.0.10.2": {
					Type:   conditionType,
					Reason: "backend health is WARNING",
					Status: v1.ConditionFalse,
				},
			},
		},
		{
			desc: "critical backends",
			backendSetHealth: &client.GenericBackendSetHealth{
				WarningStateBackendNames:  []string{},
				CriticalStateBackendNames: []string{"10.0.10.2"},
				UnknownStateBackendNames:  []string{},
			},
			out: map[string]v1.PodCondition{
				"10.0.10.2": {
					Type:   conditionType,
					Reason: "backend health is CRITICAL",
					Status: v1.ConditionFalse,
				},
			},
		},
		{
			desc: "unknown backends",
			backendSetHealth: &client.GenericBackendSetHealth{
				WarningStateBackendNames:  []string{},
				CriticalStateBackendNames: []string{},
				UnknownStateBackendNames:  []string{"10.0.10.3", "10.0.10.1"},
			},
			out: map[string]v1.PodCondition{
				"10.0.10.1": {
					Type:   conditionType,
					Reason: "backend health is UNKNOWN",
					Status: v1.ConditionFalse,
				},
				"10.0.10.3": {
					Type:   conditionType,
					Reason: "backend health is UNKNOWN",
					Status: v1.ConditionFalse,
				},
			},
		},
		{
			desc: "mixed unhealthy backends",
			backendSetHealth: &client.GenericBackendSetHealth{
				WarningStateBackendNames:  []string{"10.0.10.1"},
				CriticalStateBackendNames: []string{"10.0.10.2"},
				UnknownStateBackendNames:  []string{"10.0.10.3"},
			},
			out: map[string]v1.PodCondition{
				"10.0.10.1": {
					Type:   conditionType,
					Reason: "backend health is WARNING",
					Status: v1.ConditionFalse,
				},
				"10.0.10.2": {
					Type:   conditionType,
					Reason: "backend health is CRITICAL",
					Status: v1.ConditionFalse,
				},
				"10.0.10.3": {
					Type:   conditionType,
					Reason: "backend health is UNKNOWN",
					Status: v1.ConditionFalse,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := getBackendHealthMap(tc.backendSetHealth, conditionType)
			if !reflect.DeepEqual(got, tc.out) {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.out, got)
			}
		})
	}
}

func Test_getUpdatedPodCondition(t *testing.T) {
	conditionType := v1.PodConditionType("podReadinessCondition")
	testCases := []struct {
		desc             string
		backendHealthMap map[string]v1.PodCondition
		backendName      string
		out              v1.PodCondition
	}{
		{
			desc: "unhealthy backend",
			backendHealthMap: map[string]v1.PodCondition{
				"10.0.10.1": {
					Type:   conditionType,
					Status: v1.ConditionFalse,
					Reason: "backend health is CRITICAL",
				},
				"10.0.10.2": {
					Type:   conditionType,
					Status: v1.ConditionFalse,
					Reason: "backend health is WARNING",
				},
			},
			backendName: "10.0.10.1",
			out: v1.PodCondition{
				Type:   conditionType,
				Status: v1.ConditionFalse,
				Reason: "backend health is CRITICAL",
			},
		},
		{
			desc: "healthy backend",
			backendHealthMap: map[string]v1.PodCondition{
				"10.0.10.1": {
					Type:   conditionType,
					Status: v1.ConditionFalse,
					Reason: "backend health is CRITICAL",
				},
				"10.0.10.2": {
					Type:   conditionType,
					Status: v1.ConditionFalse,
					Reason: "backend health is WARNING",
				},
			},
			backendName: "10.0.10.3",
			out: v1.PodCondition{
				Type:   conditionType,
				Status: v1.ConditionTrue,
				Reason: "backend is OK",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := getUpdatedPodCondition(tc.backendHealthMap, conditionType, tc.backendName)
			if !reflect.DeepEqual(got, tc.out) {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.out, got)
			}
		})
	}
}

func Test_hasReadinessGate(t *testing.T) {
	testCases := []struct {
		desc          string
		pod           *v1.Pod
		readinessGate v1.PodConditionType
		out           bool
	}{
		{
			desc: "readiness gate exists",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					ReadinessGates: []v1.PodReadinessGate{
						{
							v1.PodConditionType("podReadinessCondition1"),
						},
						{
							v1.PodConditionType("podReadinessCondition2"),
						},
					},
				},
			},
			readinessGate: v1.PodConditionType("podReadinessCondition1"),
			out:           true,
		},
		{
			desc: "readiness gate does not exist",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					ReadinessGates: []v1.PodReadinessGate{
						{
							v1.PodConditionType("podReadinessCondition1"),
						},
						{
							v1.PodConditionType("podReadinessCondition2"),
						},
					},
				},
			},
			readinessGate: v1.PodConditionType("podReadinessCondition3"),
			out:           false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := hasReadinessGate(tc.pod, tc.readinessGate)
			if got != tc.out {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.out, got)
			}
		})
	}
}

func Test_getPodCondition(t *testing.T) {
	testCases := []struct {
		desc               string
		pod                *v1.Pod
		conditionType      v1.PodConditionType
		outCondition       v1.PodCondition
		outConditionExists bool
	}{
		{
			desc: "condition exists",
			pod: &v1.Pod{
				Status: v1.PodStatus{
					Conditions: []v1.PodCondition{
						{
							Type: v1.PodConditionType("podReadinessCondition1"),
						},
						{
							Type: v1.PodConditionType("podReadinessCondition2"),
						},
					},
				},
			},
			conditionType: v1.PodConditionType("podReadinessCondition1"),
			outCondition: v1.PodCondition{
				Type: v1.PodConditionType("podReadinessCondition1"),
			},
			outConditionExists: true,
		},
		{
			desc: "condition does not exist",
			pod: &v1.Pod{
				Status: v1.PodStatus{
					Conditions: []v1.PodCondition{
						{
							Type: v1.PodConditionType("podReadinessCondition1"),
						},
						{
							Type: v1.PodConditionType("podReadinessCondition2"),
						},
					},
				},
			},
			conditionType:      v1.PodConditionType("podReadinessCondition3"),
			outCondition:       v1.PodCondition{},
			outConditionExists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cond, exists := getPodCondition(tc.pod, tc.conditionType)
			if exists != tc.outConditionExists {
				t.Errorf("Expected exists: \n%+v\nbut got\n%+v", tc.outConditionExists, exists)
			}
			if cond != tc.outCondition {
				t.Errorf("Expected condition: \n%+v\nbut got\n%+v", tc.outCondition, cond)
			}
		})
	}
}

func Test_getBackendSetsNeedSync(t *testing.T) {
	testCases := []struct {
		desc    string
		service *v1.Service
		pods    []*v1.Pod
		out     map[string]v1.ServicePort
	}{
		{
			desc: "pod has readiness gate but not readiness condition",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
			},
		},
		{
			desc: "pod has readiness gate and readiness condition set to False",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
						},
					},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   getPodReadinessCondition("default", "svc1", "TCP-80"),
								Status: v1.ConditionFalse,
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
			},
		},
		{
			desc: "pod has readiness gate and readiness condition set to True",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
						},
					},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   getPodReadinessCondition("default", "svc1", "TCP-80"),
								Status: v1.ConditionTrue,
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
			},
		},
		{
			desc: "one pod has readiness gate and readiness condition",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
						},
					},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   getPodReadinessCondition("default", "svc1", "TCP-80"),
								Status: v1.ConditionFalse,
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
			},
		},
		{
			desc: "one pod has readiness gate for one backend set",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     81,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-81"),
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
				"TCP-81": {
					Protocol: v1.ProtocolTCP,
					Port:     81,
				},
			},
		},
		{
			desc: "all pods have readiness gate for all backend sets",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     81,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-81"),
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-81"),
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
				"TCP-81": {
					Protocol: v1.ProtocolTCP,
					Port:     81,
				},
			},
		},
		{
			desc: "pods have readiness gate for only one backend set",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     81,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
			},
		},
		{
			desc: "pods have readiness gate for backend sets that don't belong to the service",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     81,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-82"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-83"),
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-82"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-83"),
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{},
		},
		{
			desc: "pods don't have readiness gates",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     81,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
					},
				},
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
					},
				},
			},
			out: map[string]v1.ServicePort{},
		},
		{
			desc: "pods are not on virtual nodes",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     81,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-81"),
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-81"),
							},
						},
					},
				},
			},
			out: nil,
		},
		{
			desc: "one pod is on virtual node and other is not",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "svc1",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     81,
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "virtualNodeDefault",
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-81"),
							},
						},
					},
				},
				{
					Spec: v1.PodSpec{
						ReadinessGates: []v1.PodReadinessGate{
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-80"),
							},
							{
								ConditionType: getPodReadinessCondition("default", "svc1", "TCP-81"),
							},
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
				"TCP-81": {
					Protocol: v1.ProtocolTCP,
					Port:     81,
				},
			},
		},
	}

	prc := &PodReadinessController{
		nodeLister: &mockNodeLister{},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			out, _ := prc.getBackendSetsNeedSync(tc.service, tc.pods)
			if !reflect.DeepEqual(out, tc.out) {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.out, out)
			}
		})
	}
}

func Test_pusher(t *testing.T) {
	testCases := []struct {
		desc       string
		nodes      []*v1.Node
		services   []*v1.Service
		itemExists bool
	}{
		{
			desc: "no virtual nodes",
			nodes: []*v1.Node{
				nodeList["default"],
				nodeList["instance1"],
			},
			services: []*v1.Service{
				serviceList["default"],
			},
			itemExists: false,
		},
		{
			desc: "virtual nodes exists",
			nodes: []*v1.Node{
				nodeList["virtualNodeDefault"],
			},
			services: []*v1.Service{
				serviceList["default"],
			},
			itemExists: true,
		},
		{
			desc: "virtual nodes exists but service not of type loadbalancer",
			nodes: []*v1.Node{
				nodeList["virtualNodeDefault"],
			},
			services: []*v1.Service{
				serviceList["non-loadbalancer"],
			},
			itemExists: false,
		},
	}

	pcr := &PodReadinessController{
		logger: zap.S(),
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			pcr.queue = workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(prcMinRetryDelay, prcMaxRetryDelay))
			pcr.nodeLister = &mockNodeLister{
				nodes: tc.nodes,
			}
			pcr.serviceLister = &mockServiceLister{
				services: tc.services,
			}
			pcr.pusher()
			if pcr.queue.Len() > 0 != tc.itemExists {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.itemExists, !tc.itemExists)
			}
		})
	}
}

type mockServiceLister struct {
	services []*v1.Service
}

func (s *mockServiceLister) List(selector labels.Selector) (ret []*v1.Service, err error) {
	var services, allServices []*v1.Service
	if len(s.services) > 0 {
		allServices = s.services
	} else {
		for _, n := range serviceList {
			allServices = append(allServices, n)
		}
	}

	for _, service := range allServices {
		if selector != nil {
			if selector.Matches(labels.Set(service.ObjectMeta.GetLabels())) {
				services = append(services, service)
			}
		} else {
			services = append(services, service)
		}
	}
	return services, nil
}

func (s *mockServiceLister) Services(namespace string) v1corelisters.ServiceNamespaceLister {
	return &mockServiceNamespaceLister{
		services: s.services,
	}
}

type mockServiceNamespaceLister struct {
	services []*v1.Service
}

func (s *mockServiceNamespaceLister) List(selector labels.Selector) (ret []*v1.Service, err error) {
	var services, allServices []*v1.Service
	if len(s.services) > 0 {
		allServices = s.services
	}

	for _, service := range allServices {
		if selector != nil {
			if selector.Matches(labels.Set(service.ObjectMeta.GetLabels())) {
				services = append(services, service)
			}
		} else {
			services = append(services, service)
		}
	}
	return services, nil
}

func (s *mockServiceNamespaceLister) Get(name string) (ret *v1.Service, err error) {
	if len(s.services) > 0 {
		for _, service := range s.services {
			if service.Name == name {
				return service, nil
			}
		}
	} else if service, ok := serviceList[name]; ok {
		return service, nil
	}
	return nil, errors.New("get service error")
}
