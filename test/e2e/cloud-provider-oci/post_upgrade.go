package e2e

import (
	"context"
	. "github.com/onsi/ginkgo"
	cloudprovider "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Post Upgrade testing", func() {
	f := framework.NewDefaultFramework("post-upgrade")
	f.SkipNamespaceCreation = true
	Context("[post-upgrade][system-tags]", func() {
		It("Validate presence of oke system tags on storage resources", func() {
			if !setupF.AddOkeSystemTags {
				Skip("skip system tag backfill testing")
			}
			// List PVs
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "post-upgrade-system-tag")
			pvs, err := pvcJig.KubeClient.CoreV1().PersistentVolumes().List(context.Background(), v1.ListOptions{})
			sharedfw.ExpectNoError(err)

			// Get ocids from PV
			for _, pv := range pvs.Items {
				if cloudprovider.GetStorageType(&pv) == cloudprovider.BV {
					id := pvcJig.GetOcidFromPV(pv)
					bv, err := f.Client.BlockStorage().GetVolume(context.Background(), id)
					sharedfw.ExpectNoError(err)
					if setupF.AddOkeSystemTags && !sharedfw.HasOkeSystemTags(bv.SystemTags) {
						sharedfw.Failf("block volume %s is expected to have the system tags", *bv.Id)
					}
				}

			}
		})
	})
	Context("[post-upgrade]", func() {
		It("Checking the status of pre-existing statefulsets", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "post-upgrade-testing")
			pvcJig.ValidateExistingResources()
		})

		It("Restart pre-existing statefulsets", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "post-upgrade-testing")
			pvcJig.RestartExistingResources()
			f.CleanupUpgradeTestingNamespace()
		})
	})
})
