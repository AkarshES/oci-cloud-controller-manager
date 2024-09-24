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
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,singular=nodeoperationrequest,shortName=nor;nors
// +kubebuilder:printcolumn:name="NodeNames",type=string,JSONPath=.spec.nodeNames,priority=0,description="node names requested to be operated"
// +kubebuilder:printcolumn:name="NodeLabelSelector",type=string,JSONPath=.spec.nodeLabelSelector,priority=0,description="node label selector requested to be operated"
// +kubebuilder:printcolumn:name="CyclingActionDetails",type=string,JSONPath=.spec.cyclingActionDetails,priority=0,description="details of cycling action"
// +kubebuilder:printcolumn:name="MaxUnavailable",type=integer,JSONPath=.spec.maxUnavailable,priority=0,description="maximum number of unavailable nodes in the scope of nodes operation"
// +kubebuilder:printcolumn:name="NodeEvictionSettings",type=string,JSONPath=.spec.nodeEvictionSettings,priority=0,description="node eviction settings"
// +kubebuilder:printcolumn:name="IsPaused",type=boolean,JSONPath=.spec.isPaused,priority=0,description="whether to pause node operation"
// +kubebuilder:printcolumn:name="NodeOperationRequestState",type=string,JSONPath=.status.nodeOperationRequestState,priority=0,description="overall status of nodes operation"
// +kubebuilder:printcolumn:name="NodeCandidates",type=integer,JSONPath=.status.nodeCandidates,priority=0,description="nodes in the scope of operation"
// +kubebuilder:printcolumn:name="NumberSucceededNodes",type=integer,JSONPath=.status.numberSucceededNodes,priority=0,description="number of nodes finish operation successfully"
// +kubebuilder:printcolumn:name="NumberFailedNodes",type=integer,JSONPath=.status.numberFailedNodes,priority=0,description="number of nodes failed to finish operation"
// +kubebuilder:printcolumn:name="numberInProgressNodes",type=integer,JSONPath=.status.numberInProgressNodes,priority=0,description="number of nodes are in progress with operation"
// +kubebuilder:printcolumn:name="numberPendingNodes",type=integer,JSONPath=.status.numberPendingNodes,priority=0,description="number of nodes are pending with operation"
// +kubebuilder:printcolumn:name="PendingNodes",type=string,JSONPath=.status.pendingNodes,priority=0,description="nodes are pending with operation"
// +kubebuilder:printcolumn:name="FailedNodes",type=string,JSONPath=.status.failedNodes,priority=0,description="nodes are failed to finish operation"
// +kubebuilder:printcolumn:name="CanceledNodes",type=string,JSONPath=.status.canceledNodes,priority=0,description="nodes are canceled with operation"
// +kubebuilder:printcolumn:name="SucceededNodes",type=string,JSONPath=.status.SucceededNodes,priority=0,description="nodes finish operation successfully"
// +kubebuilder:printcolumn:name="ObservedGeneration",type=string,JSONPath=.status.observedGeneration,priority=0,description="observed generation of custom resource"
// +kubebuilder:printcolumn:name="CreationTimestamp",type=date,JSONPath=.metadata.creationTimestamp,priority=0
// +kubebuilder:printcolumn:name="DeletionTimestamp",type=date,JSONPath=.metadata.deletionTimestamp,priority=0

// NodeOperationRequest is the schema for collecting node operation request and reporting operation progress
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NodeOperationRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeOperationRequestSpec   `json:"spec,omitempty"`
	Status NodeOperationRequestStatus `json:"status,omitempty"`
}

// NodeOperationRequestSpec defines the desired state of NodeOperationRequest
type NodeOperationRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	NodeNames         []string          `json:"nodeNames,omitempty"`
	NodeLabelSelector map[string]string `json:"nodeLabelSelector,omitempty"`
	// +kubebuilder:validation:Enum=cycling;reboot
	Action Action `json:"action"`

	CyclingActionDetails CyclingActionDetails `json:"cyclingActionDetails"`
	// +kubebuilder:validation:XIntOrString
	MaxUnavailable       intstr.IntOrString   `json:"maxUnavailable"`
	NodeEvictionSettings NodeEvictionSettings `json:"nodeEvictionSettings"`
	IsPaused             bool                 `json:"isPaused"`
	// +kubebuilder:validation:Enum=internal;external
	// +kubebuilder:default=external
	CreationSource CreationSource `json:"creationSource"`
}

type NodeOperationResult struct {
	NodeName      string `json:"nodeName"`
	WorkRequestId string `json:"workRequestId"`
}

type NodeOperationRequestState string

const (
	NodeOperationRequestStateNew        NodeOperationRequestState = "New"
	NodeOperationRequestStateInProgress NodeOperationRequestState = "InProgress"
	NodeOperationRequestStateSuccessful NodeOperationRequestState = "Successful"
	NodeOperationRequestStateFailed     NodeOperationRequestState = "Failed"
	NodeOperationRequestStateCanceled   NodeOperationRequestState = "Canceled"
	NodeOperationRequestStatePaused     NodeOperationRequestState = "Paused"
)

// NodeOperationRequestStatus defines the observed state of NodeOperationRequest
type NodeOperationRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=New;InProgress;Successful;Failed;Canceled;Paused
	// +kubebuilder:default=New
	NodeOperationRequestState NodeOperationRequestState `json:"nodeOperationRequestState"`

	NodeCandidates        []string `json:"nodeCandidates"`
	NumberSucceededNodes  int      `json:"numberSucceededNodes"`
	NumberFailedNodes     int      `json:"numberFailedNodes"`
	NumberInProgressNodes int      `json:"numberInProgressNodes"`
	NumberPendingNodes    int      `json:"numberPendingNodes"`

	PendingNodes   []string              `json:"pendingNodes"`
	FailedNodes    []NodeOperationResult `json:"failedNodes"`
	CanceledNodes  []string              `json:"canceledNodes"`
	SucceededNodes []NodeOperationResult `json:"succeededNodes"`

	TerminalTime       metav1.Time `json:"terminalTime"`
	PausedTime         metav1.Time `json:"pausedTime"`
	ObservedGeneration int         `json:"observedGeneration"`
	Hashes             []string    `json:"hashes"`
}

type CycleMode string

const (
	BootVolumeReplaceMode CycleMode = "bootVolumeReplace"
)

type CyclingActionDetails struct {
	KubernetesVersion   string              `json:"kubernetesVersion"`
	NodeMetaData        []map[string]string `json:"nodeMetaData"`
	ImageId             string              `json:"imageId"`
	BootVolumeSizeInGBs int                 `json:"bootVolumeSizeInGBs"`
	SshPublicKey        string              `json:"sshPublicKey"`
	IsCycleInSyncNode   bool                `json:"isCycleInSyncNode"`
	// +kubebuilder:validation:Enum=bootVolumeReplace
	// +kubebuilder:default=bootVolumeReplace
	CycleMode CycleMode `json:"cycleMode"`
}

type NodeEvictionSettings struct {
	// +kubebuilder:validation:Minimum=0
	EvictionGracePeriod             int  `json:"evictionGracePeriod"`
	IsForceActionAfterGraceDuration bool `json:"isForceActionAfterGraceDuration"`
}

type CreationSource string

const (
	CreationSourceExternal CreationSource = "external"
	CreationSourceInternal CreationSource = "internal"
)

type Action string

const (
	CyclingAction Action = "cycling"
	RebootAction  Action = "reboot"
)

// +kubebuilder:object:root=true

// NodeOperationRequestList contains a list of NodeOperationRequest
type NodeOperationRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeOperationRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeOperationRequest{}, &NodeOperationRequestList{})
}
