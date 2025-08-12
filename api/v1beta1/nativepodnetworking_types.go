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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type VNICAddress struct {
	VNICID           *string           `json:"vnicId,omitempty"`
	MACAddress       *string           `json:"macAddress,omitempty"`
	RouterIP         *string           `json:"routerIp,omitempty"`
	RouterIPs        []RouterIP        `json:"routerIps,omitempty"`
	Addresses        []*string         `json:"addresses,omitempty"`
	PodAddresses     []PodAddress      `json:"podAddresses,omitempty"`
	SubnetCidr       *string           `json:"subnetCidr,omitempty"`
	SubnetCidrs      []SubnetCidr      `json:"subnetCidrs,omitempty"`
	HostAddress      *string           `json:"hostAddress,omitempty"`
	HostAddresses    []HostAddress     `json:"hostAddresses,omitempty"`
	NicConfiguration *NicConfiguration `json:"nicConfiguration,omitempty"`
}

type NicConfiguration struct {
	IpCount                 *int     `json:"ipCount,omitempty"`
	IpFamilies              []string `json:"ipFamilies,omitempty"`
	SubnetId                *string  `json:"subnetId,omitempty"`
	NetworkSecurityGroupIDs []string `json:"networkSecurityGroupIds,omitempty"`
	ApplicationResources    []string `json:"applicationResources,omitempty"`
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

type Ipv6AddressIpv6SubnetCidrPairDetail struct {
	Ipv6Address    string `json:"ipv6Address,omitempty"`
	Ipv6SubnetCidr string `json:"ipv6SubnetCidr,omitempty"`
}

type CreateVnicDetails struct {
	AssignIpv6Ip                         bool                                  `json:"assignIpv6Ip,omitempty"`
	AssignPublicIp                       bool                                  `json:"assignPublicIp,omitempty"`
	DefinedTags                          map[string]map[string]string          `json:"definedTags,omitempty"`
	DisplayName                          string                                `json:"displayName,omitempty"`
	Ipv6AddressIpv6SubnetCidrPairDetails []Ipv6AddressIpv6SubnetCidrPairDetail `json:"ipv6AddressIpv6SubnetCidrPairDetails,omitempty"`
	NsgIds                               []string                              `json:"nsgIds,omitempty"`
	SecurityAttributes                   map[string]apiextensionsv1.JSON       `json:"securityAttributes,omitempty"`
	SkipSourceDestCheck                  bool                                  `json:"skipSourceDestCheck,omitempty"`
	SubnetId                             string                                `json:"subnetId,omitempty"`
	ApplicationResources                 []string                              `json:"applicationResources,omitempty"`
	IpCount                              int                                   `json:"ipCount,omitempty"`
	FreeformTags                         map[string]string                     `json:"freeformTags,omitempty"`
	IpFamilies                           []string                              `json:"ipFamilies,omitempty"`
}

type SecondaryVnic struct {
	CreateVnicDetails CreateVnicDetails `json:"createVnicDetails,omitempty"`
	DisplayName       string            `json:"displayName,omitempty"`
	NicIndex          int               `json:"nicIndex,omitempty"`
}

// NativePodNetworkSpec defines the desired state of NativePodNetwork
type NativePodNetworkSpec struct {
	MaxPodCount             *int            `json:"maxPodCount,omitempty"`
	PodSubnetIds            []*string       `json:"podSubnetIds,omitempty"`
	Id                      *string         `json:"id,omitempty"`
	NetworkSecurityGroupIds []*string       `json:"networkSecurityGroupIds,omitempty"`
	IPFamilies              []*string       `json:"ipFamilies,omitempty"`
	SecondaryVnics          []SecondaryVnic `json:"secondaryVnics,omitempty"`
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
