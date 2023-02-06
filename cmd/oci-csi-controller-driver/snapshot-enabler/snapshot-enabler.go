package snapshotenabler

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	crdclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	csisnapshotter "github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csi-snapshotter"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
	snapshotcontroller "github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/snapshot-controller"
)

const (
	CRDCheckInterval        = 30 * time.Second
)

func Start(csioptions csioptions.CSIOptions, logger *zap.SugaredLogger) {
	config := csisnapshotter.BuildConfig(csioptions)

	crdClient, err := crdclientset.NewForConfig(config)
	if err != nil {
		logger.Errorf("Error building CRD clientset: %s", err.Error())
		os.Exit(1)
	}

	var goRoutinesEnabled = false
	var stopCh chan struct{}

	logger.Infof("Starting ticker to check for CRDs")

	for range time.NewTicker(CRDCheckInterval).C {
		if CheckCRDExists(crdClient){
			if !goRoutinesEnabled {
				logger.Info("CRDs found! Enabling go routines")
				stopCh = make(chan struct{})

				logger.Info("starting csi-snapshotter go routine")
				go csisnapshotter.StartCSISnapshotter(csioptions, stopCh)

				logger.Info("starting snapshot-controller go routine")
				go snapshotcontroller.StartSnapshotController(csioptions, stopCh)

				goRoutinesEnabled = true
			}
		} else {
			if goRoutinesEnabled {
				logger.Info("CRDs not found! Disabling go routines")
				close(stopCh)

				goRoutinesEnabled = false
			}
		}
	}
}

func CheckCRDExists(crdClient *crdclientset.Clientset) bool {
	var err error

	_, err = crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), "volumesnapshots.snapshot.storage.k8s.io", metav1.GetOptions{})
	if err != nil {
		return false
	}

	_, err = crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), "volumesnapshotclasses.snapshot.storage.k8s.io", metav1.GetOptions{})
	if err != nil {
		return false
	}

	_, err = crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), "volumesnapshotcontents.snapshot.storage.k8s.io", metav1.GetOptions{})
	if err != nil {
		return false
	}

	return true
}
