package e2e

import (
	. "github.com/onsi/ginkgo"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("Post Upgrade testing", func() {
	f := framework.NewDefaultFramework("post-upgrade")
	f.SkipNamespaceCreation = true
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
