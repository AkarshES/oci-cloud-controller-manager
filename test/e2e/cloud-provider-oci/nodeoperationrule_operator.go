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
	"time"
)

const (
	NodeOperationRuleCRDName          = "nodeoperationrules.oci.oraclecloud.com"
	NodeOperationLabel                = "oke.oraclecloud.com/node_operation"
	RebootNodeOperationLabelValue     = "reboot"
	ReplaceBVRNodeOperationLabelValue = "bvr"
	RebootCustomResourceYaml          = "nor_operator_reboot.yaml"
	ReplaceBVRCustomResourceYaml      = "nor_operator_bvr.yaml"
	RebootNodeCheckInterval           = 400 * time.Second
	ReplaceBVRCheckInterval           = 900 * time.Second
)

var _ = Describe("Node operation tests", func() {
	f := framework.NewDefaultFramework("noroperator")
	var nodeNames []string
	BeforeEach(func() {
		checkNodeOperationRuleCRD(f)
	})
	AfterEach(func() {
		cleanUpCustomResources()
		cleanUpLabels(nodeNames)
	})
	Context("[cloudprovider][ccm][noroperator]", func() {
		It("Test node operation for reboot action performed successfully", func() {
			nodeNames = addNodeLabelSelectorsForNodeOperation(f, RebootNodeOperationLabelValue)
			createCustomResource(RebootCustomResourceYaml)
			ticker := time.NewTicker(RebootNodeCheckInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					isOperationCompleted, _ := checkNORStatus(RebootNodeOperationLabelValue, nodeNames)
					if isOperationCompleted {
						framework.Logf("reboot node operations completed within time window")
						return
					} else {
						framework.Failf("reboot node operations didnt complete within time window")
					}
				}
			}
		})
		It("Test node operation for replace bvr action performed successfully", func() {
			nodeNames = addNodeLabelSelectorsForNodeOperation(f, ReplaceBVRNodeOperationLabelValue)
			createCustomResource(ReplaceBVRCustomResourceYaml)
			ticker := time.NewTicker(ReplaceBVRCheckInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					isOperationCompleted, _ := checkNORStatus(ReplaceBVRNodeOperationLabelValue, nodeNames)
					if isOperationCompleted {
						framework.Logf("replace bvr node operations completed within time window")
						return
					} else {
						framework.Failf("replace bvr node operations didnt complete within time window")
					}
				}
			}
		})
		It("Test node operation for reboot action initiated and cancel the reboot operation in-flight", func() {
			nodeNames = addNodeLabelSelectorsForNodeOperation(f, RebootNodeOperationLabelValue)
			createCustomResource(RebootCustomResourceYaml)
			ticker := time.NewTicker(120 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					time.Sleep(30 * time.Second)
					cleanUpLabels(nodeNames)
					_, isOperationCancelled := checkNORStatus(RebootNodeOperationLabelValue, nodeNames)
					if isOperationCancelled {
						framework.Logf("reboot node operations cancelled within time window")
						return
					} else {
						framework.Failf("reboot node operations didnt cancel within time window")
					}
				}
			}
		})
	})
})

func checkNodeOperationRuleCRD(f *framework.CloudProviderFramework) {
	var err error
	framework.Logf("Checking NodeOperationRule CRD exists")
	_, err = f.CRDClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), NodeOperationRuleCRDName, metav1.GetOptions{})
	if err != nil {
		Skip("Skipping test because fetching NodeOperationRule CRD encountered an error")
	}
}

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

func addNodeLabelSelectorsForNodeOperation(f *framework.CloudProviderFramework, labelValue string) []string {
	var err error
	var nodePatchBytes []byte
	framework.Logf("Adding node label selectors for node operation")
	var nodes []v1.Node
	var nodeNames []string
	nodes = getEligibleNodes(f)
	for _, node := range nodes {
		nodePatchBytes = []byte(fmt.Sprintf("{\"metadata\": {\"labels\": {\"%s\":\"%s\"}}}", NodeOperationLabel, labelValue))
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			_, err := f.ClientSet.CoreV1().Nodes().Patch(context.Background(), node.Name, types.StrategicMergePatchType, nodePatchBytes, metav1.PatchOptions{})
			return err
		})

		if err != nil {
			framework.Failf("error in applying patch for node %v", err.Error())
		}
		nodeNames = append(nodeNames, node.Name)
		framework.Logf("Successfully added label selectors for node operation on %s", node.Name)
	}
	return nodeNames
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

func cleanUpCustomResources() {
	var customResources = []string{RebootCustomResourceYaml, ReplaceBVRCustomResourceYaml}
	for _, resource := range customResources {
		_, err := framework.RunKubectl("delete", "-f", "../node-operator-rules/"+resource)
		if err != nil {
			framework.Logf("unable to delete nor custom resource or doesnt exist")
		}
	}
}

func cleanUpLabels(nodeNames []string) {
	for _, nodeName := range nodeNames {
		_, err := framework.RunKubectl("label", "node", nodeName, NodeOperationLabel+"-")
		if err != nil {
			framework.Logf("unable to delete label on node %s due to %v", nodeName, err)
		}
	}
}

func checkNORStatus(norName string, nodeNames []string) (bool, bool) {
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
	isNodeOperationSuccessful = sameStringSlice(nodeNames, succeededNodes)
	isNodeOperationCancelled = sameStringSlice(nodeNames, cancelledNodes)
	return isNodeOperationSuccessful, isNodeOperationCancelled
}

func sameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		diff[_x]++
	}
	for _, _y := range y {
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y]--
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}
