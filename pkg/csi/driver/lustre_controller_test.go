package driver

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	lustre "github.com/oracle/oci-go-sdk/v65/lustrefilestorage"
	"go.uber.org/zap"
)

// Unit test: happy path, verify exactly one capability and it is CREATE_DELETE_VOLUME
func TestLustreController_ControllerGetCapabilities_SingleCreateDeleteOnly(t *testing.T) {
	d := &LustreControllerDriver{}

	resp, err := d.ControllerGetCapabilities(context.Background(), &csi.ControllerGetCapabilitiesRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	if len(resp.Capabilities) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(resp.Capabilities))
	}
	cap := resp.Capabilities[0]
	if cap.GetRpc() == nil {
		t.Fatalf("expected RPC capability, got nil")
	}
	if got := cap.GetRpc().GetType(); got != csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME {
		t.Fatalf("expected capability CREATE_DELETE_VOLUME, got %v", got)
	}
}

// Unit test: method should be safe to call with nil request (it doesn't use the request)
func TestLustreController_ControllerGetCapabilities_NilRequest(t *testing.T) {
	d := &LustreControllerDriver{}
	resp, err := d.ControllerGetCapabilities(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || len(resp.Capabilities) == 0 {
		t.Fatalf("expected capabilities in response")
	}
}

// Unit test: method should not dereference the receiver; nil receiver should not panic
func TestLustreController_ControllerGetCapabilities_NilReceiver(t *testing.T) {
	var d *LustreControllerDriver // nil receiver
	// Call should not panic because implementation doesn't use the receiver fields
	resp, err := d.ControllerGetCapabilities(context.Background(), &csi.ControllerGetCapabilitiesRequest{})
	if err != nil {
		t.Fatalf("unexpected error with nil receiver: %v", err)
	}
	if resp == nil || len(resp.Capabilities) != 1 {
		t.Fatalf("expected exactly one capability with nil receiver")
	}
	if resp.Capabilities[0].GetRpc().GetType() != csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME {
		t.Fatalf("expected CREATE_DELETE_VOLUME with nil receiver")
	}
}

type MockOCILustreFileStorageClient struct {
	client util.MockOCILustreFileStorageClient
}

func (m MockOCILustreFileStorageClient) CreateLustreFileSystem(ctx context.Context, details lustre.CreateLustreFileSystemDetails) (*lustre.LustreFileSystem, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockOCILustreFileStorageClient) GetLustreFileSystem(ctx context.Context, id string) (*lustre.LustreFileSystem, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockOCILustreFileStorageClient) ListLustreFileSystems(ctx context.Context, compartmentID, ad, displayName string) ([]lustre.LustreFileSystemSummary, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockOCILustreFileStorageClient) DeleteLustreFileSystem(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (m MockOCILustreFileStorageClient) AwaitLustreFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*lustre.LustreFileSystem, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockOCILustreFileStorageClient) AwaitLustreFileSystemDeleted(ctx context.Context, logger *zap.SugaredLogger, id string) error {
	//TODO implement me
	panic("implement me")
}

func (m MockOCILustreFileStorageClient) ListWorkRequests(ctx context.Context, compartmentID, resourceID string) ([]lustre.WorkRequestSummary, error) {
	return nil, nil
}
func (m MockOCILustreFileStorageClient) ListWorkRequestErrors(ctx context.Context, workRequestID string) ([]lustre.WorkRequestError, error) {
	return nil, nil
}
