/*
Copyright 2022.

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

type VNICAddress struct {
	VNICID        *string       `json:"vnicId,omitempty"`
	MACAddress    *string       `json:"macAddress,omitempty"`
	RouterIP      *string       `json:"routerIp,omitempty"`
	RouterIPs     []RouterIP    `json:"routerIps,omitempty"`
	Addresses     []*string     `json:"addresses,omitempty"`
	PodAddresses  []PodAddress  `json:"podAddresses,omitempty"`
	SubnetCidr    *string       `json:"subnetCidr,omitempty"`
	SubnetCidrs   []SubnetCidr  `json:"subnetCidrs,omitempty"`
	HostAddress   *string       `json:"hostAddress,omitempty"`
	HostAddresses []HostAddress `json:"hostAddresses,omitempty"`
}

type RouterIP struct {
	V4 *string `json:"v4,omitempty"`
	V6 *string `json:"v6,omitempty"`
}

type PodAddress struct {
	V4 *string `json:"v4,omitempty"`
	V6 *string `json:"v6,omitempty"`
}

type SubnetCidr struct {
	V4 *string `json:"v4,omitempty"`
	V6 *string `json:"v6,omitempty"`
}

type HostAddress struct {
	V4 *string `json:"v4,omitempty"`
	V6 *string `json:"v6,omitempty"`
}

// NativePodNetworkSpec defines the desired state of NativePodNetwork
type NativePodNetworkSpec struct {
	MaxPodCount             *int      `json:"maxPodCount,omitempty"`
	PodSubnetIds            []*string `json:"podSubnetIds,omitempty"`
	Id                      *string   `json:"id,omitempty"`
	NetworkSecurityGroupIds []*string `json:"networkSecurityGroupIds,omitempty"`
	IPFamilies              []*string `json:"ipFamilies,omitempty"`
}

// NativePodNetworkStatus defines the observed state of NativePodNetwork
type NativePodNetworkStatus struct {
	State  *string       `json:"state,omitempty"`
	Reason *string       `json:"reason,omitempty"`
	VNICs  []VNICAddress `json:"vnics,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NativePodNetwork is the Schema for the nativepodnetworks API
type NativePodNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NativePodNetworkSpec   `json:"spec,omitempty"`
	Status NativePodNetworkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NativePodNetworkList contains a list of NativePodNetwork
type NativePodNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NativePodNetwork `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NativePodNetwork{}, &NativePodNetworkList{})
}
