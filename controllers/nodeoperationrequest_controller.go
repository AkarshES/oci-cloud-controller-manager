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

package controllers

import (
	"context"
	"errors"
	norv1beta1 "github.com/oracle/oci-cloud-controller-manager/api/node-cycling/v1beta1"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errNodeCandidatesEmpty     = errors.New("node candidates are empty")
	errNodeCandidatesConflict  = errors.New("node candidates are conflicted")
	errFailedToFetchNode       = errors.New("fail to get the node")
	errProviderIdMissingOnNode = errors.New("missing provider id for node")
)

// NodeOperationRequestReconciler reconciles a NodeOperationRequest object
type NodeOperationRequestReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	kubeClient kubernetes.Interface
	OCIClient  ociclient.Interface
	config     *providercfg.Config
}

// +kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nodeoperationrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nodeoperationrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oci.oraclecloud.com,resources=nodeoperationrequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodesOperationRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *NodeOperationRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	var nor norv1beta1.NodeOperationRequest
	if err := r.Get(ctx, req.NamespacedName, &nor); err != nil {
		log.Error(err, "unable to fetch NodeOperationRequest")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// nor is new
	if len(nor.Status.NodeOperationRequestState) != 0 && nor.Status.NodeOperationRequestState == norv1beta1.NodeOperationRequestStateNew {

	}

	// nor is in progress
	// nor is terminal status: successful / failed

	return ctrl.Result{}, nil
}

// cyclingNode initiates a cycling operation for a specified node within a cluster.
// It takes the node ID, cluster ID, and a NodeOperationRequest object as input.
// The function returns the work request ID associated with the cycling operation.
//
// Parameters:
// - nodeId: A string representing the unique identifier of the node to be cycled.
// - clusterId: A string representing the unique identifier of the cluster where the node resides.
// - nor: An instance of norv1beta1.NodeOperationRequest containing additional details for the cycling operation.
//
// Returns:
// - A string representing the work request ID associated with the cycling operation.
// - An error indicating any issues encountered during the cycling operation; otherwise, returns nil.
func (r *NodeOperationRequestReconciler) cyclingNode(nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) (string, error) {
	workRequestId, err := r.OCIClient.ContainerEngine().CycleClusterNode(context.Background(), nodeId, clusterId, nor)
	return workRequestId, err
}

// rebootNode initiates a reboot operation for a specified node within a cluster.
// It takes the node ID, cluster ID, and a NodeOperationRequest object as input.
// The function returns the work request ID associated with the reboot operation.
//
// Parameters:
// - nodeId: A string representing the unique identifier of the node to be rebooted.
// - clusterId: A string representing the unique identifier of the cluster where the node resides.
// - nor: An instance of norv1beta1.NodeOperationRequest containing additional details for the reboot operation.
//
// Returns:
// - A string representing the work request ID associated with the reboot operation.
// - An error indicating any issues encountered during the reboot operation; otherwise, returns nil.
func (r *NodeOperationRequestReconciler) rebootNode(nodeId string, clusterId string, nor norv1beta1.NodeOperationRequest) (string, error) {
	workRequestId, err := r.OCIClient.ContainerEngine().RebootClusterNode(context.Background(), nodeId, clusterId, nor)
	return workRequestId, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeOperationRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&norv1beta1.NodeOperationRequest{}).
		Complete(r)
}

// LookupNodeProviderID retrieves the provider ID of a specified node using the Kubernetes API.
// It takes the node name
// The function returns the provider ID of the node and an error if any issues occur during the retrieval process.
//
// Parameters:
// - kubeClient: An instance of kubernetes.Interface representing the Kubernetes client.
// - nodeName: A string representing the name of the node whose provider ID is to be retrieved.
//
// Returns:
// - A string representing the provider ID of the node.
// - An error indicating any issues encountered during the retrieval process; otherwise, returns nil.
func LookupNodeProviderID(kubeClient kubernetes.Interface, nodeName string) (string, error) {
	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return "", errFailedToFetchNode
	}
	if node.Spec.ProviderID == "" {
		return "", errProviderIdMissingOnNode
	}
	return node.Spec.ProviderID, nil
}
