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
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
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
	fssCSIDriverName = "fss.csi.oraclecloud.com"
	bvCSIDriverName  = "blockvolume.csi.oraclecloud.com"
	fvdDriverName    = "oracle/oci"
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
	dimensionsMap := make(map[string]string)
	var metricName string

	pv, err := sb.pvLister.Get(pvName)
	if err != nil {
		logger.With(zap.Error(err)).Warnf("failed to get persistent volume %s", pvName)
		return err
	}

	volume.pvStorageType = GetStorageType(pv)

	// FSS doesn't support system tags. do not process when the pv is
	//backed by FSS storage
	if volume.pvStorageType == FSS {
		logger.Infof("%s is a FSS storage. not processing the pv", pv.Name)
		return nil
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
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), volume.metricNamePrefix)
			dimensionsMap[metrics.ResourceOCIDDimension] = volume.id
			metrics.SendMetricData(sb.metricPusher, metricName, time.Since(startTime).Seconds(), dimensionsMap)
			return err
		}

		if !sb.doesBVHaveOkeSystemTag(bv) && bv.LifecycleState == core.VolumeLifecycleStateAvailable {
			logger.With("bv", "%v", *bv).Infof("detected block volume without OKE system tags. proceeding to add")
			err = sb.addBlockVolumeOkeSystemTags(bv)
			if err != nil {
				logger.With(zap.Error(err)).Warnf("updateBlockVolume didn't succeed. unable to add oke system tags")
				dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), volume.metricNamePrefix)
				dimensionsMap[metrics.ResourceOCIDDimension] = volume.id
				metrics.SendMetricData(sb.metricPusher, metricName, time.Since(startTime).Seconds(), dimensionsMap)
				return err
			}
			logger.Infof("sucessfully added oke system tags")
		}
	}
	return nil
}

// pvNeedProcessing checks if the PV is using OCI CSI/FVD driver and the storage type is not FSS
func (sb *StorageBackfillController) pvNeedProcessing(pv *v1.PersistentVolume) bool {
	// TODO: PV using the workload identity should be skipped. Currently only FSS is supported.
	return sb.isOciCSIorFVDStorage(pv) &&
		GetStorageType(pv) != FSS &&
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

func (sb *StorageBackfillController) doesBVHaveOkeSystemTag(bv *core.Volume) bool {
	if bv.SystemTags == nil {
		return false
	}
	okeSystemTagFromConfig := getResourceTrackingSystemTagsFromConfig(sb.logger, sb.config.Tags)
	if okeSystemTagFromConfig == nil {
		return false
	}

	if okeSystemTag, okeSystemTagNsExists := bv.SystemTags[OkeSystemTagNamesapce]; okeSystemTagNsExists {
		return reflect.DeepEqual(okeSystemTag, okeSystemTagFromConfig[OkeSystemTagNamesapce])
	}
	return false
}

func (sb *StorageBackfillController) addBlockVolumeOkeSystemTags(bv *core.Volume) error {
	var bvDefinedTagRequest, okeSystemTagFromConfig map[string]map[string]interface{}
	logger := sb.logger.With("volumeId", *bv.Id)
	ctx, cancel := context.WithTimeout(context.Background(), sbClientTimeout)
	defer cancel()
	okeSystemTagFromConfig = getResourceTrackingSystemTagsFromConfig(sb.logger, sb.config.Tags)
	if okeSystemTagFromConfig == nil {
		return fmt.Errorf("oke system tag is not found in the cloud config")
	}

	if _, exists := okeSystemTagFromConfig[OkeSystemTagNamesapce]; !exists {
		return fmt.Errorf("oke system tag namespace is not found in the cloud config")
	}

	if bv.DefinedTags != nil {
		bvDefinedTagRequest = bv.DefinedTags
	}
	bvDefinedTagRequest[OkeSystemTagNamesapce] = okeSystemTagFromConfig[OkeSystemTagNamesapce]
	if len(bvDefinedTagRequest) > MaxDefinedTagPerResource {
		return fmt.Errorf(MaxDefinedTagErrMessage, "volume")
	}

	bvUpdateDetails := core.UpdateVolumeDetails{
		FreeformTags: bv.FreeformTags,
		DefinedTags:  bvDefinedTagRequest,
		DisplayName:  bv.DisplayName,
	}

	bv, err := sb.ociClient.BlockStorage().UpdateVolume(ctx, *bv.Id, bvUpdateDetails)

	if err != nil {
		// check for rate limiting to slow down processing rate
		var ociServiceError common.ServiceError
		if errors.As(err, &ociServiceError) {
			if ociServiceError.GetHTTPStatusCode() == http.StatusTooManyRequests &&
				ociServiceError.GetCode() == client.HTTP429TooManyRequestsCode {
				SbWorkerDelay = SbWorkerDelay * 2
				logger.Infof("rate limited for the UpdateVolume request. updated sleep delay to %v", SbWorkerDelay)
			}
		}
		return err
	}
	logger.Infof("updated volume request to add oke system tag is successful")
	bv, err = sb.ociClient.BlockStorage().AwaitVolumeAvailableORTimeout(ctx, *bv.Id)
	if err != nil {
		return err
	}
	if !sb.doesBVHaveOkeSystemTag(bv) {
		return fmt.Errorf("validation of oke system tags after update volume operation has failed. volume details: %v", *bv)
	}
	return nil
}
