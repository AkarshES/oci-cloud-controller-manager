package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/oracle/oci-go-sdk/v65/common"
	oke "github.com/oracle/oci-go-sdk/v65/containerengine"
)

func (f *Framework) GetUpgradeTestingCluster(k8sVersion string) string {
	clusterList := f.ListClusters()
	Logf("Searching for a pre-existing cluster with required k8s version...")

	clusterId := ""
	for _, clusterSummary := range clusterList {
		Logf("ClusterName: '%s', ClusterID: '%s', k8sVersion: '%s', LifecycleState: %v", *clusterSummary.Name, *clusterSummary.Id, *clusterSummary.KubernetesVersion, clusterSummary.LifecycleState)
		if *clusterSummary.KubernetesVersion == k8sVersion && clusterSummary.LifecycleState == oke.ClusterLifecycleStateActive {
			Logf("Found a cluster with required kubernetes version.\n clusterId: '%s' \n clusterName: '%s'", *clusterSummary.Id, *clusterSummary.Name)
			Logf("Checking Architecture...")

			if f.Architecture == "AMD" && strings.Contains(*clusterSummary.Name, "AMD") {
				Logf("Architecture = AMD")
				clusterId = *clusterSummary.Id
				break
			} else if f.Architecture == "ARM" && strings.Contains(*clusterSummary.Name, "ARM") {
				Logf("Architecture = ARM")
				clusterId = *clusterSummary.Id
				break
			}
		}
	}

	return clusterId
}

// GetClusterOptions requires a clusterOptionId as input (must be "all" right now)
func (f *Framework) GetClusterOptions(clusterOptionId string) oke.ClusterOptions {
	ctx, cancel := context.WithTimeout(f.context, f.timeout)
	defer cancel()
	Logf("GetClusterOptions")

	// TODO: remove in the future when something other than "all" is allowed
	id := clusterOptionId
	if id != "all" {
		Logf("GetClusterOptions : clusterOptionId : '%s' overridden with 'all'", clusterOptionId)
		id = "all"
	}
	// retries added because of the occasional network error
	timeout := 2 * time.Minute
	var tmperr error
	var resp oke.ClusterOptions
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(Poll) {
		response, err := f.clustersClient.GetClusterOptions(ctx, oke.GetClusterOptionsRequest{
			ClusterOptionId: &id,
		})
		if err == nil {
			Expect(err).NotTo(HaveOccurred())
			Logf("ClusterOptions : '%#v'", response.ClusterOptions)
			return response.ClusterOptions
		}

		Logf("Error received on GetClusterOptions : '%s', Retrying", err)
		resp = response.ClusterOptions
		tmperr = err
	}
	Logf("Timeout waiting for GetClusterOptions \n")
	Expect(tmperr).NotTo(HaveOccurred())
	return resp
}

// Temporary limit variable to increase the page size
// TODO: Must add paging code

// ListClusters lists all clusters in the configured compartment1.
func (f *Framework) ListClusters() []oke.ClusterSummary {
	ctx, cancel := context.WithTimeout(f.context, f.timeout)
	limit := 500
	defer cancel()

	var tmperr error
	start := time.Now()
	var page *string
	resp := make([]oke.ClusterSummary, 0)
	timeout := 2 * time.Minute
	for {
		response, err := f.clustersClient.ListClusters(ctx, oke.ListClustersRequest{
			CompartmentId: &f.Compartment1,
			Limit:         &limit,
			Page:          page,
		})
		if err != nil {
			Logf("Error received on ListClusters: '%s', Retrying", err)
			tmperr = err
		} else {
			for _, cl := range response.Items {
				resp = append(resp, cl)
			}

			page = response.OpcNextPage
			if page == nil {
				break
			} else {
				Logf("received page token, continue calling ListClusters()")
			}
		}

		if time.Since(start) > timeout {
			Logf("Timeout waiting for ListClusters \n")
			Expect(tmperr).NotTo(HaveOccurred())
		}
	}

	return resp
}

// GetClusterSummary returns the specified cluster summary.
func (f *Framework) GetClusterSummary(clusterID string) *oke.ClusterSummary {
	clusterSummaries := f.ListClusters()
	for _, clusterSummary := range clusterSummaries {
		if *clusterSummary.Id == clusterID {
			return &clusterSummary
		}
	}
	return nil
}

// GetClusterSummaryByName returns the specified cluster summary.
func (f *Framework) GetClusterSummaryByName(clusterName string) *oke.ClusterSummary {
	clusterSummaries := f.ListClusters()
	for _, clusterSummary := range clusterSummaries {
		if *clusterSummary.Name == clusterName &&
			clusterSummary.LifecycleState == oke.ClusterSummaryLifecycleStateActive {
			return &clusterSummary
		}
	}
	return nil
}

// GetCluster returns the cluster with the specified clusterID.
func (f *Framework) GetCluster(clusterID string) oke.Cluster {
	response, err := f.GetClusterWithResponse(clusterID)
	if err != nil {
		Expect(err).NotTo(HaveOccurred())
	}
	return response.Cluster
}

// GetClusterWithResponse returns the raw GetCluster response for the specified clusterID.
func (f *Framework) GetClusterWithResponse(clusterID string) (response oke.GetClusterResponse, err error) {
	ctx, cancel := context.WithTimeout(f.context, f.timeout)
	defer cancel()
	// retries added because of the occasional network error
	timeout := 2 * time.Minute
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(Poll) {
		response, err = f.clustersClient.GetCluster(ctx, oke.GetClusterRequest{
			ClusterId: &clusterID,
		})
		if err == nil {
			Expect(err).NotTo(HaveOccurred())
			return
		} else {
			Logf("Error received on GetCluster: '%s', Retrying", err)
		}
	}
	Logf("Timeout waiting for GetCluster \n")
	return
}

// TestOptions passes the test related options that are to be used when doing Cluster operations.
type TestOptions struct {
	// ExpectedError specifies the ServiceError that is expected to be returned during a Cluster operation.
	ExpectedError common.ServiceError
}

// ClusterCreateConfig contains values that can be specified when creating a cluster.
type ClusterCreateConfig struct {
	// Name of the cluster. If empty a unique name will be generated.
	Name string
	// Version of the cluster. If empty the default version will be used.
	Version string
	// ServiceCIDR of the cluster. If empty the default value will be used.
	ServiceCIDR string
	// PodCIDR of the cluster. If empty the default value will be used.
	PodCIDR string
	// Options contain the test related options that are to be used when doing Cluster create operations.
	Options TestOptions

	// KMSKeyID is the optional KMS Key ocid to use for encrypting customer Kubernetes Secrets.
	KMSKeyID string
	// Delete all clusters with the same name if present before creating new
	DeleteClusterWithSameName bool
}

// CreateClusterWithOptions create a default OKE cluster with given TestOptions. No nodepools are created.
func (f *Framework) CreateClusterWithOptions(options TestOptions) string {
	return f.CreateClusterFromConfig(&ClusterCreateConfig{
		Version: f.OkeClusterK8sVersion,
		Options: options,
	})
}

// CreateClusterWithResponse create a default OKE cluster with given TestOptions and returns raw CreateCluster response. No nodepools are created.
func (f *Framework) CreateClusterWithResponse(options TestOptions) (response oke.CreateClusterResponse, err error) {
	cfg := &ClusterCreateConfig{
		Version: f.OkeClusterK8sVersion,
		Options: options,
	}

	response, err = f.createClusterFromConfig(cfg)
	return
}

// DeleteClusterWithOptions deletes the specified cluster and returns the raw DeleteCluster response.
func (f *Framework) DeleteClusterWithResponse(clusterID string) (response oke.DeleteClusterResponse, err error) {
	Logf("Deleting cluster, clusterID: '%s'", clusterID)
	return f.clustersClient.DeleteCluster(f.context, oke.DeleteClusterRequest{
		ClusterId: &clusterID,
	})
}

// CreateCluster create a default OKE cluster with unique name. No nodepools
// are created.
func (f *Framework) CreateCluster() string {
	clusterName := "TestCluster-" + UniqueID()
	if isPreUpgradeBool {
		if f.Architecture == "AMD" {
			clusterName = "Cluster-AMD-" + UniqueID()
		} else if f.Architecture == "ARM" {
			clusterName = "Cluster-ARM-" + UniqueID()
		}
	}
	return f.CreateClusterNamed(clusterName)
}

// CreateClusterNamed create a default OKE cluster named clusterName with specified
// k8sVersion, as determined by the OKE release upon which we are running. No nodepools
// are created.
func (f *Framework) CreateClusterNamed(clusterName string) string {
	return f.CreateClusterFromConfig(&ClusterCreateConfig{
		Name:    clusterName,
		Version: f.OkeClusterK8sVersion,
	})
}

// CreateCluster create a default OKE cluster with the default name and the specified version. No nodepools
// are created.
func (f *Framework) CreateClusterVersioned(version string) string {
	clusterName := "TestCluster-" + UniqueID()
	return f.CreateClusterFromConfig(&ClusterCreateConfig{
		Name:    clusterName,
		Version: version,
	})
}

func (f *Framework) getClusterIDFromWR(wrResponse oke.GetWorkRequestResponse) string {
	for _, resource := range wrResponse.Resources {
		Logf("EntityType:  '%s'", *resource.EntityType)
		if *resource.EntityType == "cluster" || *resource.EntityType == "clusterdev" || *resource.EntityType == "clusterinteg" {
			return *resource.Identifier
		}
	}
	return ""
}

func (f *Framework) cleanupClusterAsErrorExpected(wrId string, err common.ServiceError) {
	wrResponse := f.GetWorkRequest(wrId)
	clusterId := f.getClusterIDFromWR(wrResponse)
	if clusterId != "" {
		f.DeleteCluster(clusterId, false)
	}
	Failf("Expected error: %v but create cluster succeeded.", err)
}

// createClusterFromConfig creates a cluster from the given config and returns the raw CreateCluster response.
func (f *Framework) createClusterFromConfig(cfg *ClusterCreateConfig) (response oke.CreateClusterResponse, err error) {
	if cfg == nil {
		cfg = &ClusterCreateConfig{}
	}
	clusterName := cfg.Name

	if clusterName == "" {
		clusterName = "TestCluster-" + UniqueID()
	}

	version := cfg.Version
	if version == "" {
		version = f.OkeClusterK8sVersion
	}

	// TODO Remove when cluster name lengths > 32 cause an error
	if len(clusterName) > 32 {
		Logf("The cluster Name: '%s' has len %d.\n", clusterName, len(clusterName))
		clusterName = fmt.Sprintf("%.32s", clusterName)
		Logf("Truncated clusterName: '%s' has len %d.\n", clusterName, len(clusterName))
	}

	var subnets []string
	if os.Getenv("USE_REGIONALSUBNET") == "true" {
		subnets = []string{f.Lbrgnsubnet}
	} else {
		if f.IsOAD() {
			subnets = []string{f.LbSubnet1}
		} else {
			subnets = []string{f.LbSubnet1}
		}

	}

	// Delete if cluster is present with the same cluster name
	// This is important for PostUpgrade test which identifies the PreUpgrade cluster by name "UpgradeTestCluster"
	if cfg.DeleteClusterWithSameName {
		for {
			cl := f.GetClusterSummaryByName(clusterName)

			if cl == nil {
				break
			}
			f.DeleteCluster(*cl.Id, true)
		}
	}

	var ipFamilies []oke.ClusterCreateOptionsIpFamiliesEnum
	if f.ClusterIPFamily != "" {
		for _, ipFamily := range strings.Split(f.ClusterIPFamily, ",") {
			if clusterIPFamilyEnum, exists := oke.GetMappingClusterCreateOptionsIpFamiliesEnum(ipFamily); exists {
				ipFamilies = append(ipFamilies, clusterIPFamilyEnum)
			}
		}
	} else {
		ipFamilies = append(ipFamilies, oke.ClusterCreateOptionsIpFamiliesIpv4)
	}

	var clusterPodNetworkOptions []oke.ClusterPodNetworkOptionDetails
	if cniTypeEnum == oke.ClusterPodNetworkOptionDetailsCniTypeOciVcnIpNative {
		clusterPodNetworkOptions = []oke.ClusterPodNetworkOptionDetails{oke.OciVcnIpNativeClusterPodNetworkOptionDetails{}}
	} else {
		clusterPodNetworkOptions = []oke.ClusterPodNetworkOptionDetails{oke.FlannelOverlayClusterPodNetworkOptionDetails{}}
	}
	request := oke.CreateClusterRequest{
		CreateClusterDetails: oke.CreateClusterDetails{
			Name:                     &clusterName,
			CompartmentId:            &f.Compartment1,
			VcnId:                    &f.Vcn,
			KubernetesVersion:        &version,
			ClusterPodNetworkOptions: clusterPodNetworkOptions,
			Options: &oke.ClusterCreateOptions{
				IpFamilies:         ipFamilies,
				ServiceLbSubnetIds: subnets,
				KubernetesNetworkConfig: &oke.KubernetesNetworkConfig{
					ServicesCidr: &cfg.ServiceCIDR,
					PodsCidr:     &cfg.PodCIDR,
				},
			},
			Type: f.ClusterType,
		},
	}

	isPublicIpEnabled := true
	if f.Architecture == "ARM" || (f.Architecture == "AMD" && cniTypeEnum == oke.ClusterPodNetworkOptionDetailsCniTypeOciVcnIpNative) {
		request.CreateClusterDetails.EndpointConfig = &oke.CreateClusterEndpointConfigDetails{
			SubnetId:          &f.K8sSubnet,
			NsgIds:            strings.Split(f.NsgOCIDS, ","),
			IsPublicIpEnabled: &isPublicIpEnabled,
		}
	}

	if cfg.KMSKeyID != "" {
		request.KmsKeyId = &cfg.KMSKeyID
	}

	requestB, _ := json.MarshalIndent(request, "", "  ")
	Logf("Creating cluster, name: '%s' with request: %s", clusterName, requestB)

	ctx, cancel := context.WithTimeout(f.context, f.timeout)
	defer cancel()
	response, err = f.clustersClient.CreateCluster(ctx, request)

	if err != nil {
		Logf("Error on create cluster err: '%s'", err)
	}

	if cfg.Options.ExpectedError != nil {
		if err != nil {
			checkForExpectedError(err, cfg.Options.ExpectedError)
			return
		}
		// Cluster create request was expected to throw an error but it got accepted. Hence wait for the cluster
		// create work request to complete and clean up the cluster.
		defer f.cleanupClusterAsErrorExpected(*response.OpcWorkRequestId, cfg.Options.ExpectedError)
	} else {
		Expect(err).NotTo(HaveOccurred())
	}

	return
}

// CreateClusterFromConfig create a default OKE cluster named clusterName with the specified version.
// No nodepools are created.
func (f *Framework) CreateClusterFromConfig(cfg *ClusterCreateConfig) string {
	response, err := f.createClusterFromConfig(cfg)
	if err != nil && cfg.Options.ExpectedError != nil {
		Logf("Ignoring the cluster create error and returning an empty cluster ID, as the test case expects an error.", err)
		return ""
	}
	return f.waitForClusterCreation(response)
}

func (f *Framework) waitForClusterCreation(response oke.CreateClusterResponse) string {
	workRequestID := *response.OpcWorkRequestId
	Logf("Create Cluster workRequestID : '%s'", workRequestID)

	// Waits for the workRequest.Status to complete, then
	// waits for the cluster.lifecycle state to go ACTIVE
	wrSucceded := false
	timeout := 60 * time.Minute
	clusterID = ""
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(Poll) {
		if wrSucceded == false {
			wrResponse := f.GetWorkRequest(workRequestID)
			Expect(wrResponse.WorkRequest.OperationType).To(Equal(oke.WorkRequestOperationTypeClusterCreate))
			Logf("createCluster workRequest '%s'; current state '%s'", workRequestID, wrResponse.WorkRequest.Status)
			if wrResponse.WorkRequest.Status == oke.WorkRequestStatusSucceeded {
				wrSucceded = true
				clusterID = f.getClusterIDFromWR(wrResponse)
				Logf("Setting clusterID: '%s'", clusterID)
			} else {
				Expect(wrResponse.WorkRequest.Status).To(SatisfyAny(
					Equal(oke.WorkRequestStatusAccepted),
					Equal(oke.WorkRequestStatusInProgress)))
				Expect(wrResponse.WorkRequest.Status).NotTo(SatisfyAny(
					Equal(oke.WorkRequestStatusFailed),
					Equal(oke.WorkRequestStatusCanceling),
					Equal(oke.WorkRequestStatusCanceled)))
			}
		} else {
			cluster := f.GetCluster(clusterID)
			Logf("createCluster ID '%s'; current state '%s'", clusterID, cluster.LifecycleState)
			if cluster.LifecycleState == oke.ClusterLifecycleStateActive {
				return *cluster.Id
			}

			Expect(cluster.LifecycleState).To(SatisfyAny(
				Equal(oke.ClusterLifecycleStateCreating),
				Equal(oke.ClusterLifecycleStateUpdating)))

			Expect(cluster.LifecycleState).NotTo(SatisfyAny(
				Equal(oke.ClusterLifecycleStateFailed),
				Equal(oke.ClusterLifecycleStateDeleting),
				Equal(oke.ClusterLifecycleStateDeleted)))
		}
	}
	Failf("Timeout waiting for cluster '%s' to reach state: '%s'\n", clusterID, oke.ClusterSummaryLifecycleStateActive)
	return ""
}

// DeleteCluster deletes the specified cluster (and child nodepools).
func (f *Framework) DeleteCluster(clusterID string, waitForDeleted bool) {
	if strings.EqualFold(f.SkipClusterDeletion, "true") {
		Logf("Skipping cluster deletion based on env SKIP_CLUSTER_DELETION.")
		return
	}
	Logf("Deleting cluster, clusterID: '%s'", clusterID)
	// The nodepools are now deleted by the cluster delete
	// Fetch the current OKE cluster objects.
	cluster := f.GetCluster(clusterID)
	clusterSummary := f.GetClusterSummary(clusterID)
	Expect(clusterSummary).ToNot(BeNil())
	Logf("Deleting cluster '%s', initial LifecycleState: '%s'.", *cluster.Name, cluster.LifecycleState)
	Logf("Deleting cluster summary '%s', initial LifecycleState: '%s'.", *clusterSummary.Name, clusterSummary.LifecycleState)
	Expect(fmt.Sprintf("%s", cluster.LifecycleState)).To(Equal(fmt.Sprintf("%s", clusterSummary.LifecycleState)))
	// If not already 'DELETING' or 'DELETED' then issue delete request.
	if cluster.LifecycleState != oke.ClusterLifecycleStateDeleting &&
		cluster.LifecycleState != oke.ClusterLifecycleStateDeleted &&
		clusterSummary.LifecycleState != oke.ClusterSummaryLifecycleStateDeleting &&
		clusterSummary.LifecycleState != oke.ClusterSummaryLifecycleStateDeleted {
		ctx, cancel := context.WithTimeout(f.context, f.timeout)
		defer cancel()
		Logf("Issuing cluster '%s' DeleteCluster request.", *cluster.Name)
		_, err := f.clustersClient.DeleteCluster(ctx, oke.DeleteClusterRequest{
			ClusterId: &clusterID,
		})
		Expect(err).NotTo(HaveOccurred())
	}
	// If specified wait for 'DELETED'.
	if waitForDeleted {
		timeout := 10 * time.Minute
		for start := time.Now(); time.Since(start) < timeout; time.Sleep(Poll) {
			cluster = f.GetCluster(clusterID)
			clusterSummary = f.GetClusterSummary(clusterID)
			Expect(clusterSummary).ToNot(BeNil())
			Logf("Waiting for cluster '%s' '%s' status - ClusterLifeCycleState: '%s', ClusterSummaryLifeCycleState: '%s'",
				*cluster.Name, oke.ClusterLifecycleStateDeleted, cluster.LifecycleState, clusterSummary.LifecycleState)
			if cluster.LifecycleState == oke.ClusterLifecycleStateDeleted &&
				clusterSummary.LifecycleState == oke.ClusterSummaryLifecycleStateDeleted {
				return
			}
		}
		Failf("Timeout waiting for cluster '%s' to reach state: '%s'\n", clusterID, oke.ClusterLifecycleStateDeleted)
		return
	}

	Logf("Not going to wait for cluster '%s' to reach state: '%s'\n", clusterID, oke.ClusterLifecycleStateDeleted)
}
