package framework

import (
	"context"
	"fmt"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
)

func (f *CloudProviderFramework) GetFSIdByDisplayName(ctx context.Context, compartmentId, adLocation, pvName string) (string, error) {
	Logf("GetFileSystemSummaryByDisplayName request params")
	Logf("compartmentId: %+v", compartmentId)
	Logf("adLocation: %+v", adLocation)
	Logf("pvName: %+v", pvName)
	_, fsVolumeSummaryList, err := f.Client.FSS(nil).GetFileSystemSummaryByDisplayName(ctx, compartmentId, adLocation, pvName)
	if client.IsNotFound(err) {
		return "", err
	}
	if err != nil {
		return "", err
	}
	if len(fsVolumeSummaryList) == 0 {
		Logf("fsVolumeSummaryList is empty or nil")
		return "", fmt.Errorf("no file system volume found")
	}

	Logf("fsVolumeSummaryList length: %d", len(fsVolumeSummaryList))
	Logf("First volume summary: %+v", fsVolumeSummaryList[0])

	return *fsVolumeSummaryList[0].Id, nil
}

func (f *CloudProviderFramework) GetExportsSetIdByMountTargetId(ctx context.Context, mountTargetId string) (string, error) {
	mountTarget, err := f.Client.FSS(nil).GetMountTarget(ctx, mountTargetId)
	if client.IsNotFound(err) {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return *mountTarget.ExportSetId, nil
}

func (f *CloudProviderFramework) CheckFSVolumeExist(ctx context.Context, fsId string) bool {
	fs, err := f.Client.FSS(nil).GetFileSystem(ctx, fsId)
	if client.IsNotFound(err) {
		return false
	}
	if err != nil {
		return false
	}
	if fs.LifecycleState == filestorage.FileSystemLifecycleStateDeleting || fs.LifecycleState == filestorage.FileSystemLifecycleStateDeleted {
		return false
	}
	return true
}

func (f *CloudProviderFramework) CheckExportExists(ctx context.Context, fsId, exportPath, exportSetId string) bool {
	export, err := f.Client.FSS(nil).FindExport(ctx, fsId, exportPath, exportSetId)
	if client.IsNotFound(err) {
		return false
	}
	if err != nil {
		return false
	}
	if export.LifecycleState == filestorage.ExportSummaryLifecycleStateDeleting || export.LifecycleState == filestorage.ExportSummaryLifecycleStateDeleted {
		return false
	}
	return true
}
