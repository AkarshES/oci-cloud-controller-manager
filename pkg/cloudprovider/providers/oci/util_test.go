// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oci

import (
	"testing"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	v1 "k8s.io/api/core/v1"
)

func TestMapProviderIDToInstanceID(t *testing.T) {
	testCases := map[string]struct {
		providerID string
		instanceID string
		error      bool
	}{
		"no cloud prefix": {
			providerID: "ocid1.testid",
			instanceID: "ocid1.testid",
			error:      false,
		},
		"cloud prefix": {
			providerID: providerPrefix + "ocid1.testid",
			instanceID: "ocid1.testid",
			error:      false,
		},
		"empty string": {
			providerID: "",
			instanceID: "",
			error:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := MapProviderIDToResourceID(tc.providerID)
			if result != tc.instanceID {
				t.Errorf("Expected instance id %q, but got %q", tc.instanceID, result)
			}
			if (err == nil && tc.error) || (!tc.error && err != nil) {
				t.Errorf("Expected an error condition for input %q, but did no receive one; or received one, when not expecting", tc.providerID)
			}
		})
	}
}

func TestDeepEqualLists(t *testing.T) {
	testCases := map[string]struct {
		listA   []string
		listB   []string
		isEqual bool
	}{
		"lists are empty": {
			listA:   []string{},
			listB:   []string{},
			isEqual: true,
		},
		"lists are equal": {
			listA:   []string{"ocid1", "ocid2"},
			listB:   []string{"ocid1", "ocid2"},
			isEqual: true,
		},
		"compare lists with different ordering": {
			listA:   []string{"ocid2", "ocid1"},
			listB:   []string{"ocid1", "ocid2"},
			isEqual: true,
		},
		"lists are duplicates in one of the list": {
			listA:   []string{"ocid1", "ocid2", "ocid1"},
			listB:   []string{"ocid1", "ocid2"},
			isEqual: true,
		},
		"lists are equal with duplicates and irregular order": {
			listA:   []string{"a", "c", "b", "d"},
			listB:   []string{"a", "b", "c", "d"},
			isEqual: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := DeepEqualLists(tc.listA, tc.listB)
			if result != tc.isEqual {
				t.Errorf("Expected Lists comparison to be %v, but got %v", tc.isEqual, result)
			}
		})
	}
}

func TestRemoveDuplicatesFromList(t *testing.T) {
	testCases := map[string]struct {
		list   []string
		result []string
	}{
		"List with Duplicates": {
			list:   []string{"ocid1", "ocid2", "ocid1"},
			result: []string{"ocid1", "ocid2"},
		},
		"List with irregular order": {
			list:   []string{"a", "c", "b"},
			result: []string{"a", "b", "c"},
		},
		"List with duplicates and irregular order": {
			list:   []string{"a", "a", "c", "b", "d", "c", "d"},
			result: []string{"a", "b", "c", "d"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := RemoveDuplicatesFromList(tc.list)
			if !DeepEqualLists(tc.result, result) {
				t.Errorf("Expected Lists comparison to be %v, but got %v", tc.result, result)
			}
		})
	}
}

func TestIsVirtualNode(t *testing.T) {
	tests := map[string]struct {
		node          *v1.Node
		isVirtualNode bool
	}{
		"Virtual node": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: "ocid1.virtualnode.xyz",
				},
			},
			isVirtualNode: true,
		},
		"Virtual node dev": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: "ocid1.virtualnodedev.xyz",
				},
			},
			isVirtualNode: true,
		},
		"Virtual node integ": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: "ocid1.virtualnodeinteg.xyz",
				},
			},
			isVirtualNode: true,
		},
		"Virtual node rbaas": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: "ocid1.virtualnoderbaas.xyz",
				},
			},
			isVirtualNode: true,
		},
		"Unknown virtual node resource": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: "ocid1.virtualnodelimit.xyz",
				},
			},
			isVirtualNode: false,
		},
		"Provisioned node": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: "ocid1.instance.xyz",
				},
			},
			isVirtualNode: false,
		},
		"Unknown provider Id": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: "ocid1.resource.virtualnode",
				},
			},
			isVirtualNode: false,
		},
		"Empty provider Id": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: "",
				},
			},
			isVirtualNode: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsVirtualNode(tc.node)
			if got != tc.isVirtualNode {
				t.Errorf("expected: %+v got %+v", tc.isVirtualNode, got)
			}
		})
	}
}

func TestVirtualNodeExists(t *testing.T) {
	tests := map[string]struct {
		nodeList []*v1.Node
		exists   bool
	}{
		"No nodes exist": {
			nodeList: []*v1.Node{},
			exists:   false,
		},
		"No virtual nodes exist": {
			nodeList: []*v1.Node{
				{
					Spec: v1.NodeSpec{
						ProviderID: "ocid1.instance.xyz1",
					},
				},
				{
					Spec: v1.NodeSpec{
						ProviderID: "ocid1.instance.xyz2",
					},
				},
			},
			exists: false,
		},
		"Virtual nodes and provisioned nodes exist": {
			nodeList: []*v1.Node{
				{
					Spec: v1.NodeSpec{
						ProviderID: "ocid1.instance.xyz1",
					},
				},
				{
					Spec: v1.NodeSpec{
						ProviderID: "ocid1.instance.xyz2",
					},
				},
				{
					Spec: v1.NodeSpec{
						ProviderID: "ocid1.instance.xyz3",
					},
				},
				{
					Spec: v1.NodeSpec{
						ProviderID: virtualNodeOcidPrefix + ".xyz",
					},
				},
			},
			exists: true,
		},
		"Virtual node exists": {
			nodeList: []*v1.Node{
				{
					Spec: v1.NodeSpec{
						ProviderID: virtualNodeOcidPrefix + ".xyz",
					},
				},
			},
			exists: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fakeInformerFactory := informers.NewSharedInformerFactory(&fake.Clientset{}, 0*time.Second)
			for _, node := range tc.nodeList {
				fakeInformerFactory.Core().V1().Nodes().Informer().GetStore().Add(node)
			}
			got, err := VirtualNodeExists(fakeInformerFactory.Core().V1().Nodes().Lister())
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tc.exists {
				t.Errorf("expected: %+v got %+v", tc.exists, got)
			}
		})
	}
}
