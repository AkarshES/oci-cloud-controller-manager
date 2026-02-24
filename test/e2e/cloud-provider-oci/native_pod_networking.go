/*
Copyright 2026 Oracle and/or its affiliates.

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

package e2e

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	oke "github.com/oracle/oci-go-sdk/v65/containerengine"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("Custom Node Name", func() {
	baseName := "custom-node-name"
	f := sharedfw.NewDefaultFramework(baseName)

	BeforeEach(func() {
		nodes := sharedfw.GetReadySchedulableVirtualNodesOrDie(f.ClientSet)
		if len(nodes.Items) != 0 {
			Skip("Skipping test since virtual nodes exist in the cluster.")
		}
	})

	Context("[cloudprovider][ccm][npn][lb][nodepool-update][custom-node-name]", func() {
		It("should update nodepool with custom node names and preserve LB reachability", func() {
			if setupF.CniType != oke.ClusterPodNetworkOptionDetailsCniTypeOciVcnIpNative {
				Skip("Skipping test since cluster is not using OCI_VCN_IP_NATIVE (NPN_CNI)")
			}

			npList := setupF.ListNodePools(sharedfw.ClusterID)
			if len(npList) == 0 {
				sharedfw.Failf("NodePool not found.")
			}
			if npList[0].Id == nil {
				sharedfw.Failf("NodePool OCID is nil.")
			}
			nodePool := setupF.GetNodePool(*npList[0].Id)

			customHostnamePrefix := "custom"
			customHostnameSuffix := fmt.Sprintf("%05d", time.Now().UnixNano()%100000)
			userData := fmt.Sprintf("#!/bin/bash\n"+
				"HOST_ID=$(hostname | awk -F'-' '{print $NF}')\n"+
				"NODE_NAME=\"%s-%s-${HOST_ID}\"\n"+
				"curl --fail -H \"Authorization: Bearer Oracle\" -L0 http://169.254.169.254/opc/v2/instance/metadata/oke_init_script | base64 --decode > /var/run/oke-init.sh\n"+
				"bash /var/run/oke-init.sh --kubelet-extra-args \"--hostname-override=${NODE_NAME}\"\n", customHostnamePrefix, customHostnameSuffix)
			encodedUserData := base64.StdEncoding.EncodeToString([]byte(userData))

			nodeMetadata := map[string]string{}
			for key, value := range setupF.NodeMetadata {
				nodeMetadata[key] = value
			}
			nodeMetadata["user_data"] = encodedUserData

			By("updating nodepool metadata with custom cloud-init and recycling nodes")
			setupF.UpdateNodePoolMetadata(nodePool, nodeMetadata)
			nodePool = setupF.GetNodePool(*nodePool.Id)
			var expectedNodes int
			if sharedfw.IsVersion2NodePool(nodePool) {
				expectedNodes = *nodePool.NodeConfigDetails.Size
			} else {
				expectedNodes = *nodePool.QuantityPerSubnet * len(nodePool.SubnetIds)
			}

			zeroNodes := 0
			By(fmt.Sprintf("waiting for the nodepool to scale to %d nodes", zeroNodes))
			setupF.ScaleNodePool(nodePool, zeroNodes)
			err := wait.PollImmediate(10*time.Second, 20*time.Minute, func() (bool, error) {
				nodePool = setupF.GetNodePool(*nodePool.Id)
				totalNodeCount := 0
				for _, node := range nodePool.Nodes {
					if node.LifecycleState != oke.NodeLifecycleStateDeleted && node.LifecycleState != oke.NodeLifecycleStateDeleting {
						totalNodeCount++
					}
				}
				sharedfw.Logf("Active nodes after scale-down: %d", totalNodeCount)
				return totalNodeCount == 0, nil
			})
			if err != nil {
				sharedfw.Failf("Timed out waiting for nodepool to scale down to zero active nodes: %v", err)
			}

			By(fmt.Sprintf("waiting for the nodepool to scale to %d nodes", expectedNodes))
			setupF.ScaleNodePool(nodePool, expectedNodes)
			sharedfw.WaitReadySchedulableNodesOrDie(f.ClientSet, expectedNodes)
			By("waiting for nodes to register with the custom hostname")

			// Check if the nodes have expected hostnames
			err = wait.PollImmediate(10*time.Second, 15*time.Minute, func() (bool, error) {
				nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet)
				customNamedNodes := 0
				for _, node := range nodes.Items {
					if strings.HasPrefix(node.Name, customHostnamePrefix+"-") {
						customNamedNodes++
					}
				}
				sharedfw.Logf("Custom named nodes: %d/%d", customNamedNodes, expectedNodes)
				return customNamedNodes >= expectedNodes, nil
			})
			if err != nil {
				sharedfw.Failf("Timed out waiting for custom node names to appear: %v", err)
			}

			ns := f.Namespace.Name
			serviceName := fmt.Sprintf("custom-node-name-lb")
			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)
			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			loadBalancerLagTimeout := sharedfw.LoadBalancerLagTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
			}

			By("creating a load balancer service and pod")
			tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeLoadBalancer
				s.Spec.Ports = []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)}}
			})
			jig.RunOrFail(ns, nil)

			By("waiting for the TCP service to have a load balancer")
			tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)
			svcPort := int(tcpService.Spec.Ports[0].Port)
			ingressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", ingressIP)

			By("hitting the TCP service's LoadBalancer")
			jig.TestReachableHTTP(false, ingressIP, svcPort, loadBalancerLagTimeout)

			By("changing TCP service back to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
			})
			jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, ingressIP, svcPort, loadBalancerCreateTimeout)
		})
	})
})
