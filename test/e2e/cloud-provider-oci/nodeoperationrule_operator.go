package e2e

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/oracle/oci-cloud-controller-manager/api/node-cycling/v1beta1"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/strings/slices"
	"time"
)

const (
	NodeOperationRuleCRDName            = "nodeoperationrules.oci.oraclecloud.com"
	NodeOperationLabel                  = "oke.oraclecloud.com/node_operation"
	RebootNodeOperationLabelValue       = "reboot"
	ReplaceBVRNodeOperationLabelValue   = "replace-bvr"
	RebootNodeCancelOperationLabelValue = "reboot-cancel"
	RebootCustomResourceYaml            = "nor_operator_reboot.yaml"
	ReplaceBVRCustomResourceYaml        = "nor_operator_bvr.yaml"
	RebootCancelCustomResourceYaml      = "nor_operator_reboot_cancel.yaml"
	RebootNodeCheckInterval             = 400 * time.Second
	ReplaceBVRCheckInterval             = 1000 * time.Second
	RebootCancelCheckInterval           = 120 * time.Second
)

var _ = Describe("Node operation tests", func() {
	f := framework.NewDefaultFramework("noroperator")
	Context("[cloudprovider][ccm][noroperator]", func() {
		tests := []struct {
			OperationType          string
			CustomResourceType     string
			NORStatusCheckInterval time.Duration
		}{
			{RebootNodeOperationLabelValue, RebootCustomResourceYaml, RebootNodeCheckInterval},
			{ReplaceBVRNodeOperationLabelValue, ReplaceBVRCustomResourceYaml, ReplaceBVRCheckInterval},
			{RebootNodeCancelOperationLabelValue, RebootCancelCustomResourceYaml, RebootCancelCheckInterval},
		}

		for index, entry := range tests {
			testName := "Should be able to perform node operation " + entry.OperationType + " successfully"
			It(testName, func() {
				checkNodeOperationRuleCRD(f)
				nodes := getEligibleNodes(f)
				defer func() {
					// recover from panic if one occurred.
					if recover() != nil {
						Skip("Skipping test " + entry.OperationType + "node operation due to insufficient ready nodes")
					}
				}()
				testNodeOperation(f, entry.OperationType, entry.CustomResourceType, entry.NORStatusCheckInterval, nodes[index])
				cleanUpCustomResources(entry.CustomResourceType)
				cleanUpLabels(nodes[index].Name)
			})
		}
	})
})

func getEligibleNodes(f *framework.CloudProviderFramework) []v1.Node {
	var err error
	var nodes *v1.NodeList
	var eligibleNodes []v1.Node
	framework.Logf("Fetching nodes from the cluster")
	nodes, err = f.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		framework.Failf("failed to fetch nodes from the cluster: %v", err)
	}
	for _, node := range nodes.Items {
		if !validateNodeHasRequiredLabels(node) {
			eligibleNodes = append(eligibleNodes, node)
		} else {
			framework.Logf("node already has node operation rule label, will not process node %s", node.Name)
		}
	}
	if eligibleNodes == nil {
		Skip("Skipping the test case as no nodes were available")
	}
	return eligibleNodes
}

func testNodeOperation(f *framework.CloudProviderFramework, operationType string, customResourceType string, statusCheckInterval time.Duration, node v1.Node) {
	var nodeName string
	nodeName = addNodeLabelSelectorsForNodeOperation(f, node, operationType)
	createCustomResource(customResourceType)
	ticker := time.NewTicker(statusCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if operationType == RebootNodeCancelOperationLabelValue {
				time.Sleep(30 * time.Second)
				cleanUpLabels(nodeName)
			}
			isOperationCompleted, isOperationCancelled := checkNORStatus(operationType, nodeName)
			if isOperationCompleted {
				framework.Logf("node operation %s completed within time window", operationType)
				return
			} else if isOperationCancelled {
				framework.Logf("node operation %s cancelled within time window", operationType)
				return
			} else {
				framework.Failf("node operation %s didnt complete within time window", operationType)
			}
		}
	}
}

func checkNodeOperationRuleCRD(f *framework.CloudProviderFramework) {
	var err error
	framework.Logf("Checking NodeOperationRule CRD exists")
	_, err = f.CRDClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), NodeOperationRuleCRDName, metav1.GetOptions{})
	if err != nil {
		Skip("Skipping test because fetching NodeOperationRule CRD encountered an error")
	}
}

func addNodeLabelSelectorsForNodeOperation(f *framework.CloudProviderFramework, node v1.Node, labelValue string) string {
	var err error
	var nodePatchBytes []byte
	framework.Logf("Adding node label selectors for node operation")
	nodePatchBytes = []byte(fmt.Sprintf("{\"metadata\": {\"labels\": {\"%s\":\"%s\"}}}", NodeOperationLabel, labelValue))
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err := f.ClientSet.CoreV1().Nodes().Patch(context.Background(), node.Name, types.StrategicMergePatchType, nodePatchBytes, metav1.PatchOptions{})
		return err
	})

	if err != nil {
		framework.Failf("error in applying patch for node %v", err.Error())
	}
	framework.Logf("Successfully added label selectors for node operation on %s", node.Name)
	return node.Name
}

func validateNodeHasRequiredLabels(node v1.Node) bool {
	_, isNodeOperationLabel := node.ObjectMeta.Labels[NodeOperationLabel]
	if isNodeOperationLabel {
		return true
	}
	return false
}

func createCustomResource(filename string) {
	_, err := framework.RunKubectl("create", "-f", "../node-operator-rules/"+filename)
	if err != nil {
		framework.Failf("unable to create nor custom resource %v", err)
	}
}

func cleanUpCustomResources(resourceType string) {
	_, err := framework.RunKubectl("delete", "-f", "../node-operator-rules/"+resourceType)
	if err != nil {
		framework.Logf("unable to delete nor custom resource or doesnt exist")
	}
}

func cleanUpLabels(nodeName string) {
	_, err := framework.RunKubectl("label", "node", nodeName, NodeOperationLabel+"-")
	if err != nil {
		framework.Logf("unable to delete label on node %s due to %v", nodeName, err)
	}
}

func checkNORStatus(norName string, nodeName string) (bool, bool) {
	response, err := framework.RunKubectl("get", "nor", norName, "-o", "yaml")
	if err != nil {
		framework.Logf("unable to fetch nor %s", norName)
	}
	var nor v1beta1.NodeOperationRule
	var isNodeOperationSuccessful = false
	var isNodeOperationCancelled = false
	var succeededNodes []string
	var cancelledNodes []string

	err = yaml.Unmarshal([]byte(response), &nor)
	if err != nil {
		framework.Failf("Error unmarshaling YAML: %v\n", err)
	}
	for _, succeededNode := range nor.Status.SucceededNodes {
		succeededNodes = append(succeededNodes, succeededNode.NodeName)
	}
	framework.Logf("Succeeded nodes are %s ", succeededNodes)
	for _, cancelledNode := range nor.Status.CanceledNodes {
		cancelledNodes = append(cancelledNodes, cancelledNode.NodeName)
	}
	framework.Logf("Cancelled nodes are %s ", cancelledNodes)
	isNodeOperationSuccessful = slices.Contains(succeededNodes, nodeName)
	isNodeOperationCancelled = slices.Contains(cancelledNodes, nodeName)
	return isNodeOperationSuccessful, isNodeOperationCancelled
}
