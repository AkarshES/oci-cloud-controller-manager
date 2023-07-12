package e2e

import (
	. "github.com/onsi/ginkgo"

	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/block"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"

	v1 "k8s.io/api/core/v1"
)

const (
	replicas int32 = 3
)

var _ = Describe("Pre Upgrade testing", func() {
	f := framework.NewDefaultFramework("pre-upgrade")
	f.SkipNamespaceCreation = true
	Context("[pre-upgrade]", func() {
		It("Should be possible to create a statefulset with persistent volume claim for a block storage (PVC)", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "fvp-preupgrade-e2e-test")

			scName := f.CreateStorageClassOrFail(framework.ClassOCI, core.ProvisionerNameDefault, nil, pvcJig.Labels, "", false, "Delete", nil)
			serviceName := pvcJig.CreateService(setupF.UpgradeTestingNamespace)
			pvcJig.CreateAndAwaitStatefulSet("csi-app-oci-0", serviceName, scName, framework.MinVolumeBlock, setupF.AdLabel, replicas)
		})

		It("Should be possible to create a persistent volume claim (PVC) for a block storage of Ext3 file system ", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "volume-provisioner-e2e-tests-pvc")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIExt3, core.ProvisionerNameDefault, map[string]string{block.FSType: "ext3"}, pvcJig.Labels, "", false, "Delete", nil)
			serviceName := pvcJig.CreateService(setupF.UpgradeTestingNamespace)
			pvcJig.CreateAndAwaitStatefulSet("csi-app-oci-1", serviceName, scName, framework.MinVolumeBlock, setupF.AdLabel, replicas)
		})
	})

	Context("[pre-upgrade]", func() {
		It("Create statefulset with PVC and POD for CSI.", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-preupgrade-e2e-test")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			serviceName := pvcJig.CreateService(setupF.UpgradeTestingNamespace)
			pvcJig.CreateAndAwaitStatefulSetCSI("csi-app-oci-bv", serviceName, scName, framework.MinVolumeBlock, replicas, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce)
		})
	})

	Context("[pre-upgrade]", func() {
		It("Create PVC with fstype as XFS", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-fstype-xfs")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIXfs, "blockvolume.csi.oraclecloud.com", map[string]string{framework.FstypeKey: "xfs"}, pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			serviceName := pvcJig.CreateService(setupF.UpgradeTestingNamespace)
			pvcJig.CreateAndAwaitStatefulSetCSI("csi-app-oci-bv", serviceName, scName, framework.MaxVolumeBlock, replicas, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce)
		})
		It("Create PVC with fstype as EXT3", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-fstype-ext3")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIExt3, "blockvolume.csi.oraclecloud.com", map[string]string{framework.FstypeKey: "ext3"}, pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			serviceName := pvcJig.CreateService(setupF.UpgradeTestingNamespace)
			pvcJig.CreateAndAwaitStatefulSetCSI("csi-app-oci-bv", serviceName, scName, framework.MaxVolumeBlock, replicas, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce)
		})
	})

	Context("[pre-upgrade]", func() {
		It("Basic Create Statefulset with PVC and POD for CSI-FSS", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-preupgrade-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			serviceName := pvcJig.CreateService(setupF.UpgradeTestingNamespace)
			pvcJig.CreateAndAwaitStatefulSetDynamicFss("csi-app-fss-dyn", serviceName, scName, framework.MinVolumeBlock, replicas)
		})
	})
})
