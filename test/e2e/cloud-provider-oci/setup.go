package e2e

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	oke "github.com/oracle/oci-go-sdk/v65/containerengine"
	"time"
)

var setupF *sharedfw.Framework

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {

	setupF = sharedfw.New()

	sharedfw.Logf("CloudProviderFramework Setup")

	if setupF.EnableCreateCluster {
		createUpgradeTestingNodepool := false
		clusterOCID := ""

		if setupF.IsPostUpgrade {
			clusterOCID = setupF.GetUpgradeTestingCluster(setupF.OkeClusterK8sVersion)
			Expect(clusterOCID).ShouldNot(BeZero())
			sharedfw.Logf("Cluster OCID is %s", clusterOCID)

		} else if setupF.IsPreUpgrade {
			clusterOCID = setupF.GetUpgradeTestingCluster(setupF.OkeClusterK8sVersion)
			if clusterOCID == "" {
				sharedfw.Logf("No cluster with k8s version %s and architecture %s found in Compartment1. Creating new cluster", setupF.OkeClusterK8sVersion, setupF.Architecture)
				createUpgradeTestingNodepool = true
			} else {
				createUpgradeTestingNodepool = !setupF.CheckIfNodepoolExists(clusterOCID)
			}
		}

		if clusterOCID == "" {
			sharedfw.Logf("Creating the cluster...")
			clusterOCID = setupF.CreateCluster()
			Expect(clusterOCID).ShouldNot(BeZero())
			sharedfw.Logf("Cluster OCID is %s", clusterOCID)
		}

		kubeConfig := setupF.CreateClusterKubeconfigContent(clusterOCID)
		Expect(setupF.IsNotJsonFormatStr(kubeConfig)).To(BeTrue())
		Expect(kubeConfig != "").To(BeTrue())

		err := setupF.SaveKubeConfig(kubeConfig)
		Expect(err).NotTo(HaveOccurred())
		sharedfw.Logf("Returned Kubeconfig: \n%s", kubeConfig)

		cloudConfig := setupF.CreateCloudConfig()
		Expect(cloudConfig).ShouldNot(BeNil())

		err = setupF.SaveCloudConfig(cloudConfig)
		Expect(err).Should(BeNil())

		if (!setupF.IsPreUpgrade && !setupF.IsPostUpgrade) || createUpgradeTestingNodepool {
			if !setupF.CreateUhpNodepool {
				var ocpus = float32(1.0)
				var memoryInGBs = float32(6.0)
				var NodeShapeConfig = oke.CreateNodeShapeConfigDetails{
					Ocpus:       &ocpus,
					MemoryInGBs: &memoryInGBs,
				}

				size := 3
				nodepool := setupF.CreateNodePool(clusterOCID, setupF.Compartment1, "Oracle-Linux-7.6",
					setupF.NodeShape, size, setupF.OkeNodePoolK8sVersion,
					[]string{setupF.NodeSubnet, setupF.NodeSubnet, setupF.NodeSubnet},
					NodeShapeConfig)
				Expect(nodepool).ShouldNot(BeNil())
				sharedfw.Logf(" Created cluster %s with nodepool %s ", clusterOCID, *nodepool.Id)
			} else {
				var ocpus = float32(16.0)
				var memoryInGBs = float32(256.0)
				var NodeShapeConfig = oke.CreateNodeShapeConfigDetails{
					Ocpus:       &ocpus,
					MemoryInGBs: &memoryInGBs,
				}

				size := 3

				nodepool := setupF.CreateNodePool(clusterOCID, setupF.Compartment1, "Oracle-Linux-7.6",
					"VM.Standard.E4.Flex", size, setupF.OkeNodePoolK8sVersion,
					[]string{setupF.NodeSubnet, setupF.NodeSubnet, setupF.NodeSubnet},
					NodeShapeConfig)
				Expect(nodepool).ShouldNot(BeNil())
				sharedfw.Logf(" Created cluster %s with nodepool %s ", clusterOCID, *nodepool.Id)
				setupF.EnableBVMPluginOnNodepool(nodepool)
				sharedfw.Logf("Waiting 10 mins for block volume management plugin to be enabled")
				time.Sleep(10 * time.Minute)
			}
			setupF.CrossValidateCluster(clusterOCID, setupF.ValidateChildResources)
		}
	} else {
		sharedfw.Logf("Cluster creation skipped. Running tests with existing cluster.")
	}
	return nil
}, func(data []byte) {})

var _ = ginkgo.SynchronizedAfterSuite(func() {
	sharedfw.Logf("Running AfterSuite actions on all node")
	if !setupF.IsPostUpgrade && !setupF.IsPreUpgrade {
		sharedfw.RunCleanupActions()
		if setupF.EnableCreateCluster {
			setupF.CleanAllWithoutWait()
		}
	}
}, func() {})
