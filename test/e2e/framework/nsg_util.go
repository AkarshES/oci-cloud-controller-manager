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
	"fmt"
	"sync"
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

// HasValidSinglePortRulesAfterPortChangeNSG checks the counts of 'single port' security rules
// expectMultipleRules parameter relaxes single rule check if required.
func HasValidSinglePortRulesAfterPortChangeNSG(oci client.Interface, nsgId string, oldPort, newPort int, direction core.SecurityRuleDirectionEnum, expectRulesCount int) bool {
	if nsgId != "" {
		numOldPortRules := CountSinglePortRules(oci, nsgId, oldPort, direction)
		numNewPortRules := CountSinglePortRules(oci, nsgId, newPort, direction)
		if numOldPortRules != 0 {
			fmt.Printf("Rule %s on NSG for old port still present: (oldPort: %d, newPort: %d) (numOldPortRules: %d, numNewPortRules: %d)\n", string(direction), oldPort, newPort, numOldPortRules, numNewPortRules)
			return false
		}
		if numNewPortRules != expectRulesCount {
			fmt.Printf("Rule %s on NSG for new port did not match: (oldPort: %d, newPort: %d) (numOldPortRules: %d, numNewPortRules: %d, expectedNumNewPortRules: %d)\n", string(direction), oldPort, newPort, numOldPortRules, numNewPortRules, expectRulesCount)
			return false
		}
	}
	return true
}

// WaitForSinglePortRulesAfterPortChangeOrFailNSG waits for the expected rules to be added and validates
// that the rule on the old port is removed and the rule on the new port is added
func WaitForSinglePortRulesAfterPortChangeOrFailNSG(oci client.Interface, nsgId string, oldPort, newPort int, direction core.SecurityRuleDirectionEnum, expectRulesCount int) {
	for start := time.Now(); time.Since(start) < 120*time.Second; {
		valid := HasValidSinglePortRulesAfterPortChangeNSG(oci, nsgId, oldPort, newPort, direction, expectRulesCount)
		if !valid {
			time.Sleep(5 * time.Second)
		} else {
			return
		}
	}
	Failf("Failed: ValidSinglePortRulesAfterPortChangeOrDie Rule %s on NSG for old port still present: oldPort: %d, newPort: %d)", string(direction), oldPort, newPort)
}

// A simple in-process critical section mechanism for E2E tests that mutate
// shared backend NSGs. This uses a global mutex and assumes E2Es run once per
// cluster within a single test runner process time duration.

// Gate mutex protecting the backend NSG critical section.
var backendNSGLock sync.Mutex

// Re-entrancy tracking for the NSG critical section so the same logical holder
// can acquire the lock multiple times (nested) without deadlocking.
var backendNSGStateMu sync.Mutex
var backendNSGOwner string
var backendNSGRecurse int

// AcquireBackendNSGCriticalSection attempts to acquire a global, in-process
// mutex guarding mutations to shared backend NSGs. If the mutex cannot be
// acquired within the framework's ServiceTestTimeout, an error is returned so
// the calling E2E can fail fast instead of hanging indefinitely.
//
// holder is only used for logging/debugging.
func AcquireBackendNSGCriticalSection(holder string) error {
	deadline := time.Now().Add(ServiceTestTimeout)

	// Re-entrant fast path: if the same holder already owns the lock, just
	// bump recursion and return immediately.
	backendNSGStateMu.Lock()
	if backendNSGOwner == holder && backendNSGRecurse > 0 {
		backendNSGRecurse++
		depth := backendNSGRecurse
		backendNSGStateMu.Unlock()
		Logf("Reused backend NSG lock as %q (depth=%d)", holder, depth)
		return nil
	}
	backendNSGStateMu.Unlock()

	// Try fast-path once before falling back to polling
	if backendNSGLock.TryLock() {
		backendNSGStateMu.Lock()
		backendNSGOwner = holder
		backendNSGRecurse = 1
		backendNSGStateMu.Unlock()
		Logf("Acquired backend NSG lock as %q", holder)
		return nil
	}

	// Poll until timeout
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for now := range ticker.C {
		// Allow re-entrant reuse while waiting
		backendNSGStateMu.Lock()
		if backendNSGOwner == holder && backendNSGRecurse > 0 {
			backendNSGRecurse++
			depth := backendNSGRecurse
			backendNSGStateMu.Unlock()
			Logf("Reused backend NSG lock as %q (depth=%d)", holder, depth)
			return nil
		}
		backendNSGStateMu.Unlock()
		if backendNSGLock.TryLock() {
			backendNSGStateMu.Lock()
			backendNSGOwner = holder
			backendNSGRecurse = 1
			backendNSGStateMu.Unlock()
			Logf("Acquired backend NSG lock as %q", holder)
			return nil
		}
		if now.After(deadline) {
			return fmt.Errorf("timeout acquiring backend NSG lock for %q after %v", holder, ServiceTestTimeout)
		}
		// keep waiting
	}
	return fmt.Errorf("internal error: backend NSG acquire ticker stopped unexpectedly for %q", holder)
}

// ReleaseBackendNSGCriticalSection unlocks the global mutex. Parameters are
// kept for API compatibility but the client is ignored.
func ReleaseBackendNSGCriticalSection(holder string) {
	// Manage recursion and ownership; only unlock the gate on final release.
	backendNSGStateMu.Lock()
	switch {
	case backendNSGOwner == holder && backendNSGRecurse > 1:
		backendNSGRecurse--
		depth := backendNSGRecurse
		backendNSGStateMu.Unlock()
		Logf("Released backend NSG lock (decrement) by %q (depth=%d)", holder, depth)
		return
	case backendNSGOwner == holder && backendNSGRecurse == 1:
		backendNSGOwner = ""
		backendNSGRecurse = 0
		backendNSGStateMu.Unlock()
		Logf("Released backend NSG lock held by %q", holder)
		backendNSGLock.Unlock()
		return
	default:
		// Unexpected release from a non-owner. Reset defensively to avoid deadlocks.
		prevOwner := backendNSGOwner
		prevDepth := backendNSGRecurse
		backendNSGOwner = ""
		backendNSGRecurse = 0
		backendNSGStateMu.Unlock()
		Logf("Warning: Release called by %q, but lock owner was %q (depth=%d). Forcing unlock.", holder, prevOwner, prevDepth)
		backendNSGLock.Unlock()
		return
	}
}
