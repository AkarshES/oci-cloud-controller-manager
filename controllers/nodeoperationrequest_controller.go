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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errNodeCandidatesEmpty    = errors.New("node candidates are empty")
	errNodeCandidatesConflict = errors.New("node candidates are conflicted")
)

// NodeOperationRequestReconciler reconciles a NodeOperationRequest object
type NodeOperationRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

func invokeNodeCycling(nor norv1beta1.NodeOperationRequest) (string, error) {

	return "", nil
}

func invokeNodeReboot(nor norv1beta1.NodeOperationRequest) (string, error) {
	return "", nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeOperationRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&norv1beta1.NodeOperationRequest{}).
		Complete(r)
}
