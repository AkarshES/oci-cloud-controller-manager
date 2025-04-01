/*
Copyright 2018 The Kubernetes Authors.

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

package framework

import (
	"context"
	"strings"
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/core"
)

// CountSinglePortRules counts the number of 'single port' (non-ranged)
func CountSinglePortRules(oci client.Interface, nsgId string, port int, direction core.SecurityRuleDirectionEnum) int {
	count := 0
	if oci != nil && nsgId != "" {
		_, _, err := oci.Networking(nil).GetNetworkSecurityGroup(context.Background(), nsgId)
		if err != nil {
			Failf("Could not obtain nsg: %v", err)
		}
		response, err := oci.Networking(nil).ListNetworkSecurityGroupSecurityRules(context.Background(), nsgId,
			core.ListNetworkSecurityGroupSecurityRulesDirectionEnum(direction))
		filteredRules := []core.SecurityRule{}
		for _, rule := range response {
			if rule.TcpOptions != nil && rule.TcpOptions.DestinationPortRange != nil &&
				*rule.TcpOptions.DestinationPortRange.Max == port &&
				*rule.TcpOptions.DestinationPortRange.Min == port {
				filteredRules = append(filteredRules, rule)
			}
		}
		count = len(filteredRules)
	}
	return count
}

// HasValidSinglePortRulesAfterPortChangeNSG checks the counts of 'single port'
func HasValidSinglePortRulesAfterPortChangeNSG(oci client.Interface, nsgId string, oldPort, newPort int, direction core.SecurityRuleDirectionEnum) bool {
	if nsgId != "" {
		numOldPortRules := CountSinglePortRules(oci, nsgId, oldPort, direction)
		numNewPortRules := CountSinglePortRules(oci, nsgId, newPort, direction)
		if numOldPortRules != 0 {
			return false
		}
		if numNewPortRules == 0 {
			return false
		}
	}
	return true
}

// WaitForSinglePortRulesAfterPortChangeOrFailNSG waits for the expected rules to be added and validates
// that the rule on the old port is removed and the rule on the new port is added
func WaitForSinglePortRulesAfterPortChangeOrFailNSG(oci client.Interface, nsgId string, oldPort, newPort int, direction core.SecurityRuleDirectionEnum) {
	for start := time.Now(); time.Since(start) < 190*time.Second; {
		valid := HasValidSinglePortRulesAfterPortChangeNSG(oci, nsgId, oldPort, newPort, direction)
		if !valid {
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
	Failf("Failed: ValidSinglePortRulesAfterPortChangeOrDie Rule %s on NSG %s for old port still present: oldPort: %d, newPort: %d)", string(direction), nsgId, oldPort, newPort)
}

// Helper function to get the NSGs with a specified tag for a node
func GetNSGsWithTagForNode(client client.Interface, compartmentId, nodeOcid, tagKey, tagValue string) ([]string, error) {
	ctx := context.TODO()

	// Get VNICs for the instance
	vnicAttachments, err := client.Compute().ListVnicAttachments(ctx, compartmentId, nodeOcid)
	if err != nil {
		return nil, err
	}

	var nsgIds []string
	// For each VNIC, get associated NSGs and check tags
	for _, attach := range vnicAttachments {
		vnic, err := client.Networking(nil).GetVNIC(ctx, *attach.VnicId)
		if err != nil {
			return nil, err
		}

		// Get NSGs for this VNIC
		if vnic.NsgIds != nil && len(vnic.NsgIds) > 0 {
			for _, nsgId := range vnic.NsgIds {
				// Get the NSG details to check tags
				nsg, _, err := client.Networking(nil).GetNetworkSecurityGroup(ctx, nsgId)
				if err != nil {
					return nil, err
				}

				// Check if the NSG has the required tag
				if nsg.FreeformTags != nil {
					if val, ok := nsg.FreeformTags[tagKey]; ok && val == tagValue {
						nsgIds = append(nsgIds, nsgId)
					}
				}
			}
		}
	}

	return nsgIds, nil
}

// Helper function to verify no rules exist for a service in an NSG
func VerifyNoRulesForService(client client.Interface, nsgId, serviceName string) error {
	ctx := context.TODO()

	// Check ingress rules
	ingressRules, err := client.Networking(nil).ListNetworkSecurityGroupSecurityRules(
		ctx, nsgId, core.ListNetworkSecurityGroupSecurityRulesDirectionIngress)
	if err != nil {
		return err
	}

	// Check egress rules
	egressRules, err := client.Networking(nil).ListNetworkSecurityGroupSecurityRules(
		ctx, nsgId, core.ListNetworkSecurityGroupSecurityRulesDirectionEgress)
	if err != nil {
		return err
	}

	// Verify no ingress rules for the service exist
	for _, rule := range ingressRules {
		if rule.Description != nil && strings.Contains(*rule.Description, serviceName) {
			Failf("Found ingress rule for service %s in NSG %s: %s",
				serviceName, nsgId, *rule.Description)
		}
	}

	// Verify no egress rules for the service exist
	for _, rule := range egressRules {
		if rule.Description != nil && strings.Contains(*rule.Description, serviceName) {
			Failf("Found egress rule for service %s in NSG %s: %s",
				serviceName, nsgId, *rule.Description)
		}
	}

	return nil
}
