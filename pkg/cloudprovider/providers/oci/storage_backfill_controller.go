package oci

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/workqueue"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
)

const (
	sbBaseRetryDelay     = 50 * time.Second
	sbMaxRetryDelay      = 300 * time.Second
	pollingInterval      = 360 * time.Second // 6 mins
	SbWorkerInitialDelay = 3 * time.Second
	sbClientTimeout      = 150 * time.Second
)

var SbWorkerDelay time.Duration

const (
	fssCSIDriverName              = "fss.csi.oraclecloud.com"
	bvCSIDriverName               = "blockvolume.csi.oraclecloud.com"
	fvdDriverName                 = "oracle/oci"
	ProvisionerSecretKey          = "volume.kubernetes.io/provisioner-deletion-secret-name"
	ProvisionerSecretNamespaceKey = "volume.kubernetes.io/provisioner-deletion-secret-namespace"
)

type PvStorageType string

const (
	BV  PvStorageType = "BV"
	FSS PvStorageType = "FSS"
)

type StorageBackfillController struct {
	kubeClient   clientset.Interface
	ociClient    client.Interface
	logger       *zap.SugaredLogger
	metricPusher *metrics.MetricPusher
	config       *providercfg.Config
	pvLister     corelisters.PersistentVolumeLister
	queue        workqueue.RateLimitingInterface
}

type genericVolume struct {
	id               string
	definedTags      map[string]map[string]interface{}
	metricNamePrefix string
	pvStorageType    PvStorageType
	freeformTags     map[string]string
	displayName      *string
	kmsKeyId         *string
}

func NewStorageBackfillController(
	kubeClient clientset.Interface,
	ociClient client.Interface,
	logger *zap.SugaredLogger,
	metricPusher *metrics.MetricPusher,
	config *providercfg.Config,
	pvLister corelisters.PersistentVolumeLister,
) *StorageBackfillController {
	controllerName := "storage-backfill-controller"

	return &StorageBackfillController{
		kubeClient:   kubeClient,
		ociClient:    ociClient,
		logger:       logger.With("controller", controllerName),
		metricPusher: metricPusher,
		config:       config,
		pvLister:     pvLister,
		queue:        workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(sbBaseRetryDelay, sbMaxRetryDelay)),
	}
}

func (sb *StorageBackfillController) Run(stopCh chan struct{}) {
	defer utilruntime.HandleCrash()
	defer sb.queue.ShutDown()
	defer close(stopCh)

	var sbcPollChannel = make(chan struct{})

	SbWorkerDelay = SbWorkerInitialDelay
	sb.logger.Info("Starting storage backfill controller")

	go wait.Until(sb.worker, sbBaseRetryDelay, stopCh)
	sb.pusher()
	go sb.pollWorkQueueEmpty(sbcPollChannel)
	<-sbcPollChannel
	close(sbcPollChannel)
	sb.logger.Info("Stopping storage backfill controller")
}

// pusher list all the persitent volume objects and queues up only the persistent volumes
// using CSI and FVD storage driver
func (sb *StorageBackfillController) pusher() {
	sb.logger.Infof("starting pusher")
	pvs, err := sb.pvLister.List(labels.Everything())
	if err != nil {
		sb.logger.With(zap.Error(err)).Error("unable to list persistent volumes")
		return
	}

	for _, pv := range pvs {
		sb.logger.Infof("checking if pv is eligible for processing %s", pv.Name)
		if !sb.pvNeedProcessing(pv) {
			continue
		}

		sb.queue.Add(pv.Name)
	}
}

// pollWorkQueueEmpty polls for controller queue size to send signal
// to the channel when the queue is empty
func (sb *StorageBackfillController) pollWorkQueueEmpty(sbcPollChannel chan struct{}) {
	sb.logger.Infof("Starting the poller...")
	wait.PollUntil(pollingInterval, func() (done bool, err error) {
		sb.logger.Infof("checking the queue size. current size is %d", sb.queue.Len())
		if sb.queue.Len() > 0 {
			return false, nil
		}
		sb.logger.Infof("it's an empty queue! sending signal")
		sbcPollChannel <- struct{}{}
		return true, nil
	}, sbcPollChannel)
}

func (sb *StorageBackfillController) worker() {
	for sb.processNextWorkItem() {
	}
}

func (sb *StorageBackfillController) processNextWorkItem() bool {
	time.Sleep(SbWorkerDelay)
	key, quit := sb.queue.Get()
	if quit {
		sb.logger.Infof("quit called. returning..")
		return false
	}
	defer sb.queue.Done(key)
	sb.logger.Infof("processing %s", key.(string))
	err := sb.backfill(key.(string))

	if err == nil {
		sb.queue.Forget(key)
		SbWorkerDelay = SbWorkerInitialDelay
		return true
	}

	// do not requeue & log an error message incase of tagging failure.
	sb.logger.With(zap.Error(err), "pvName", key.(string)).Errorf("error backfilling the persistent volume %s. will be defering to next sync", key.(string))
	return true
}

// backfill performs GET on BV/FSS to check for presence of the OKE system tags
// and read OKE system tags from the config to invokes UPDATE to add OKE system tags, otherwise
func (sb *StorageBackfillController) backfill(pvName string) error {
	startTime := time.Now()
	volume := genericVolume{}
	logger := sb.logger.With("persistentVolume", pvName)
	var metricName string

	pv, err := sb.pvLister.Get(pvName)
	if err != nil {
		logger.With(zap.Error(err)).Warnf("failed to get persistent volume %s", pvName)
		return err
	}

	volume.pvStorageType = GetStorageType(pv)

	if volume.pvStorageType == FSS {
		volume.metricNamePrefix = "FSS_" + util.SystemTagErrTypePrefix

		volumeHandle := csi_util.ValidateFssId(pv.Spec.PersistentVolumeSource.CSI.VolumeHandle)
		if volumeHandle.FilesystemOcid == "" {
			return fmt.Errorf("unable to get the FilesystemOcid from pv")
		}

		volume.id = volumeHandle.FilesystemOcid

		logger = logger.With("volumeId", volume.id)

		fss, err := sb.getFss(volume.id)

		if err != nil {
			logger.With(zap.Error(err)).Errorf("failed to get FSS for id %s", volume.id)
			sb.sendSBFailureMetric(volume, metrics.FSSUpdate, startTime, err)
			return err
		}

		volume.definedTags = fss.DefinedTags
		volume.freeformTags = fss.FreeformTags
		volume.kmsKeyId = fss.KmsKeyId

		if !systemTagsExists(sb.logger, fss.SystemTags, sb.config) && fss.LifecycleState == filestorage.FileSystemLifecycleStateActive {
			logger.With("fss", "%v", *fss).Infof("detected FSS without OKE system tags. proceeding to add")
			err = sb.addSystemTagToVolume(logger, volume)
			if err != nil {
				logger.With(zap.Error(err)).Warnf("updateFileSystem didn't succeed. unable to add oke system tags")
				sb.sendSBFailureMetric(volume, metrics.FSSUpdate, startTime, err)
				return err
			}
			logger.Infof("Successfully added oke system tags to fss")

		}
	}

	if volume.pvStorageType == BV {
		volume.metricNamePrefix = "BV_" + util.SystemTagErrTypePrefix
		metricName = metrics.PVUpdate
		volume.id, err = sb.getBlockVolumeOcidFromPV(pv)
		if err != nil {
			return err
		}
		logger = logger.With("volumeId", volume.id)

		bv, err := sb.getBv(volume.id)
		if err != nil {
			logger.With(zap.Error(err)).Errorf("failed to get BV for id %s", volume.id)
			sb.sendSBFailureMetric(volume, metricName, startTime, err)
			return err
		}

		volume.definedTags = bv.DefinedTags
		volume.freeformTags = bv.FreeformTags
		volume.displayName = bv.DisplayName

		if !systemTagsExists(sb.logger, bv.SystemTags, sb.config) && bv.LifecycleState == core.VolumeLifecycleStateAvailable {
			logger.With("bv", "%v", *bv).Infof("detected block volume without OKE system tags. proceeding to add")
			err = sb.addSystemTagToVolume(logger, volume)
			if err != nil {
				logger.With(zap.Error(err)).Warnf("updateBlockVolume didn't succeed. unable to add oke system tags")
				sb.sendSBFailureMetric(volume, metricName, startTime, err)
				return err
			}
			logger.Infof("sucessfully added oke system tags")
		}
	}
	return nil
}

// pvNeedProcessing checks if the PV is using OCI CSI/FVD driver and PV using the workload identity should be skipped. Currently only FSS supports workload identity.
func (sb *StorageBackfillController) pvNeedProcessing(pv *v1.PersistentVolume) bool {
	return sb.isOciCSIorFVDStorage(pv) &&
		!util.StorageClassWorkloadIdentityCheck(pv.GetObjectMeta().GetAnnotations(), ProvisionerSecretKey, ProvisionerSecretNamespaceKey) &&
		sb.isPvPhaseEligble(pv)
}

func (sb *StorageBackfillController) isPvPhaseEligble(pv *v1.PersistentVolume) bool {
	switch phase := pv.Status.Phase; phase {
	case v1.VolumeAvailable,
		v1.VolumeBound,
		v1.VolumeReleased:
		return true
	default:
		return false
	}
}

func (sb *StorageBackfillController) isOciCSIorFVDStorage(pv *v1.PersistentVolume) bool {
	pvSource := pv.Spec.PersistentVolumeSource

	switch {
	case pvSource.CSI != nil:
		if pvSource.CSI.Driver == bvCSIDriverName || pvSource.CSI.Driver == fssCSIDriverName {
			return true
		}
	case pvSource.FlexVolume != nil:
		if pvSource.FlexVolume.Driver == fvdDriverName {
			return true
		}
	}
	return false
}

func (sb *StorageBackfillController) getBlockVolumeOcidFromPV(pv *v1.PersistentVolume) (string, error) {
	pvSource := pv.Spec.PersistentVolumeSource

	// assuming it is OCI CSI/FVD drivers as the check is done in queueing
	if pvSource.CSI != nil {
		return pvSource.CSI.VolumeHandle, nil
	}
	if pvSource.FlexVolume != nil {
		return pv.Name, nil
	}
	return "", fmt.Errorf("unable to get the block volume ocid from pv")
}

func GetStorageType(pv *v1.PersistentVolume) PvStorageType {
	pvSource := pv.Spec.PersistentVolumeSource
	var volumeType PvStorageType

	switch {
	case pvSource.CSI != nil:
		switch pvSource.CSI.Driver {
		case bvCSIDriverName:
			volumeType = BV
		case fssCSIDriverName:
			volumeType = FSS
		}
	case pvSource.FlexVolume != nil:
		if pvSource.FlexVolume.Driver == fvdDriverName {
			volumeType = BV
		}
	}
	return volumeType
}

func (sb *StorageBackfillController) getBv(id string) (*core.Volume, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	return sb.ociClient.BlockStorage().GetVolume(ctx, id)
}

func (sb *StorageBackfillController) getFss(id string) (*filestorage.FileSystem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()
	return sb.ociClient.FSS(nil).GetFileSystem(ctx, id)
}

func (sb *StorageBackfillController) addSystemTagToVolume(logger *zap.SugaredLogger, gv genericVolume) error {
	var gvDefinedTagRequest, okeSystemTagFromConfig map[string]map[string]interface{}
	ctx, cancel := context.WithTimeout(context.Background(), sbClientTimeout)
	defer cancel()
	okeSystemTagFromConfig = getResourceTrackingSystemTagsFromConfig(sb.logger, sb.config.Tags)
	if okeSystemTagFromConfig == nil {
		return fmt.Errorf("oke system tag is not found in the cloud config")
	}

	if _, exists := okeSystemTagFromConfig[OkeSystemTagNamesapce]; !exists {
		return fmt.Errorf("oke system tag namespace is not found in the cloud config")
	}

	if gv.definedTags != nil {
		gvDefinedTagRequest = gv.definedTags
	}
	gvDefinedTagRequest[OkeSystemTagNamesapce] = okeSystemTagFromConfig[OkeSystemTagNamesapce]
	if len(gvDefinedTagRequest) > MaxDefinedTagPerResource {
		return fmt.Errorf(MaxDefinedTagErrMessage, "volume")
	}

	var err error
	var fss *filestorage.FileSystem
	var bv *core.Volume

	switch gv.pvStorageType {
	case FSS:
		fssUpdateDetails := filestorage.UpdateFileSystemDetails{
			FreeformTags: gv.freeformTags,
			DefinedTags:  gvDefinedTagRequest,
			KmsKeyId:     gv.kmsKeyId,
		}

		fss, err = sb.ociClient.FSS(nil).UpdateFileSystem(ctx, fssUpdateDetails, gv.id)
		if err != nil {
			checkForRateLimit(logger, err, "UpdateFileSystem")
			return err
		}
		logger.Infof("updated file system request to add oke system tag is successful")
		fss, err = sb.ociClient.FSS(nil).AwaitFileSystemActive(ctx, logger, *fss.Id)
		if err != nil {
			return err
		}
		if !systemTagsExists(sb.logger, fss.SystemTags, sb.config) {
			return fmt.Errorf("validation of oke system tags after update file system operation has failed. File system details: %v", *fss)
		}
	case BV:
		bvUpdateDetails := core.UpdateVolumeDetails{
			FreeformTags: gv.freeformTags,
			DefinedTags:  gvDefinedTagRequest,
			DisplayName:  gv.displayName,
		}

		bv, err = sb.ociClient.BlockStorage().UpdateVolume(ctx, gv.id, bvUpdateDetails)
		if err != nil {
			checkForRateLimit(logger, err, "UpdateVolume")
			return err
		}
		logger.Info("updated volume request to add oke system tag is successful")
		bv, err = sb.ociClient.BlockStorage().AwaitVolumeAvailableORTimeout(ctx, *bv.Id)
		if err != nil {
			return err
		}
		if !systemTagsExists(sb.logger, bv.SystemTags, sb.config) {
			return fmt.Errorf("validation of oke system tags after update volume operation has failed. volume details: %v", *bv)
		}
	default:
		return fmt.Errorf("Unknown PV storage type: %v", gv.pvStorageType)
	}

	return nil
}

func checkForRateLimit(logger *zap.SugaredLogger, err error, requestType string) {
	var ociServiceError common.ServiceError
	if errors.As(err, &ociServiceError) {
		if ociServiceError.GetHTTPStatusCode() == http.StatusTooManyRequests &&
			ociServiceError.GetCode() == client.HTTP429TooManyRequestsCode {
			SbWorkerDelay = SbWorkerDelay * 2
			logger.Infof("rate limited for the %s request. updated sleep delay to %v", requestType, SbWorkerDelay)
		}
	}
}

func (sb *StorageBackfillController) sendSBFailureMetric(volume genericVolume, metricName string, startTime time.Time, err error) {
	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ComponentDimension] = util.GetComponentForMetricDimension(util.GetError(err), volume.metricNamePrefix)
	dimensionsMap[metrics.ResourceOCIDDimension] = volume.id
	metrics.SendMetricData(sb.metricPusher, metricName, time.Since(startTime).Seconds(), dimensionsMap)
}

// compares resource and config system tags. Returns true if all false if systemTags matches.
func systemTagsExists(logger *zap.SugaredLogger, systemTags map[string]map[string]interface{}, config *providercfg.Config) bool {
	if systemTags == nil {
		return false
	}

	okeSystemTagFromConfig := getResourceTrackingSystemTagsFromConfig(logger, config.Tags)
	if okeSystemTagFromConfig == nil {
		return false
	}

	if okeSystemTag, okeSystemTagNsExists := systemTags[OkeSystemTagNamesapce]; okeSystemTagNsExists {
		return reflect.DeepEqual(okeSystemTag, okeSystemTagFromConfig[OkeSystemTagNamesapce])
	}
	return false
}
