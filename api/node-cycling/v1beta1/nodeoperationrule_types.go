/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,singular=nodeoperationrule,shortName=nor;nors
// +kubebuilder:printcolumn:name="NodeSelector",type=string,JSONPath=.spec.nodeSelector,priority=0,description="The labels are used to filter nodes that could be operated"
// +kubebuilder:printcolumn:name="Actions",type=string,JSONPath=.spec.actions,priority=0,description="The action of node operation"
// +kubebuilder:printcolumn:name="CreationTimestamp",type=date,JSONPath=.metadata.creationTimestamp,priority=0
// +genclient

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NodeOperationRule is the Schema for the NodeOperationRule API
type NodeOperationRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeOperationRuleSpec   `json:"spec,omitempty"`
	Status NodeOperationRuleStatus `json:"status,omitempty"`
}

// NodeOperationRuleSpec defines the desired state of NodeOperationRule
type NodeOperationRuleSpec struct {

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="NodeSelector field is immutable"
	// +kubebuilder:validation:Required
	NodeSelector NodeSelector `json:"nodeSelector,omitempty"`
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Actions field is immutable"
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	Actions              []Action             `json:"actions,omitempty"`
	NodeEvictionSettings NodeEvictionSettings `json:"nodeEvictionSettings,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=20
	MaxParallelism int `json:"maxParallelism,omitempty"`
}

type NodeSelector struct {
	// +kubebuilder:validation:MinProperties=1
	// +kubebuilder:validation:MaxProperties=1
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self.all(key, key == 'oke.oraclecloud.com/node_operation') && self.all(key, self[key].size() >= 0 && self[key].size() <= 63) && self.all(key, self[key].matches('^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$'))",message="matchTriggerLabel key is 'oke.oraclecloud.com/node_operation' and the syntax should be complying with node label"
	MatchTriggerLabel map[string]string `json:"matchTriggerLabel,omitempty"`
	// +kubebuilder:validation:Optional
	MatchCustomLabels map[string]string `json:"matchCustomLabels,omitempty"`
}

// +kubebuilder:validation:Enum=reboot;replaceBootVolume
type Action string

type NodeEvictionSettings struct {
	// +kubebuilder:validation:Minimum=0
	EvictionGracePeriod             int  `json:"evictionGracePeriod,omitempty"`
	IsForceActionAfterGraceDuration bool `json:"isForceActionAfterGraceDuration,omitempty"`
}

// NodeOperationRuleStatus defines the observed state of NodeOperationRule
type NodeOperationRuleStatus struct {
	InProgressNodes []NodeOperationResult  `json:"inProgressNodes"`
	PendingNodes    []string               `json:"pendingNodes"`
	BackOffNodes    []NodeOperationResult  `json:"backOffNodes"`
	SucceededNodes  []NodeOperationSuccess `json:"succeededNodes"`
	CanceledNodes   []NodeOperationResult  `json:"canceledNodes"`
}

type NodeOperationResult struct {
	NodeName      string `json:"nodeName"`
	WorkRequestId string `json:"workRequestId"`
}

type NodeOperationSuccess struct {
	NodeName         string      `json:"nodeName"`
	SuccessTimestamp metav1.Time `json:"successTimestamp"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// +kubebuilder:object:root=true

// NodeOperationRuleList contains a list of NodeOperationRule
type NodeOperationRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeOperationRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeOperationRule{}, &NodeOperationRuleList{})
}
