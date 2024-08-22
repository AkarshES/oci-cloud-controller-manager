package oci

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/workqueue"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
)

var (
	bvOcid  = "ocid1.volume.aaaaa..."
	fssOCid = "ocid1.filesystem.aaaaa..."
	pvList  = map[string]*v1.PersistentVolume{
		"not-oci-csi-fvd-pv": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "not-oci-csi-fvd-pv",
			},
			Spec: v1.PersistentVolumeSpec{
				StorageClassName: "oci-bv",
			},
		},
		"oci-csi-bv-pv": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "oci-csi-bv-pv",
			},
			Spec: v1.PersistentVolumeSpec{
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:       bvCSIDriverName,
						VolumeHandle: bvOcid,
					},
				},
			},
			Status: v1.PersistentVolumeStatus{
				Phase: v1.VolumeAvailable,
			},
		},
		"oci-csi-bv-pv-2": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "oci-csi-bv-pv-2",
			},
			Spec: v1.PersistentVolumeSpec{
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:       bvCSIDriverName,
						VolumeHandle: bvOcid,
					},
				},
			},
			Status: v1.PersistentVolumeStatus{
				Phase: v1.VolumeFailed,
			},
		},
		"oci-csi-fss-pv": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "oci-csi-fss-pv",
			},
			Spec: v1.PersistentVolumeSpec{
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:       fssCSIDriverName,
						VolumeHandle: fssOCid,
					},
				},
			},
			Status: v1.PersistentVolumeStatus{
				Phase: v1.VolumeAvailable,
			},
		},
		"oci-csi-fvd-pv": {
			ObjectMeta: metav1.ObjectMeta{
				Name: bvOcid,
			},
			Spec: v1.PersistentVolumeSpec{
				PersistentVolumeSource: v1.PersistentVolumeSource{
					FlexVolume: &v1.FlexPersistentVolumeSource{
						Driver: fvdDriverName,
					},
				},
			},
			Status: v1.PersistentVolumeStatus{
				Phase: v1.VolumeAvailable,
			},
		},
		"csi-bv-pv": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "csi-bv-pv",
			},
			Spec: v1.PersistentVolumeSpec{
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:       "csidriver",
						VolumeHandle: bvOcid,
					},
				},
			},
			Status: v1.PersistentVolumeStatus{
				Phase: v1.VolumeAvailable,
			},
		},
		"csi-fss-pv": {
			ObjectMeta: metav1.ObjectMeta{
				Name: "csi-fss-pv",
			},
			Spec: v1.PersistentVolumeSpec{
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:       "fssCsiDriver",
						VolumeHandle: fssOCid,
					},
				},
			},
			Status: v1.PersistentVolumeStatus{
				Phase: v1.VolumeAvailable,
			},
		},
		"fvd-pv": {
			ObjectMeta: metav1.ObjectMeta{
				Name: bvOcid,
			},
			Spec: v1.PersistentVolumeSpec{
				PersistentVolumeSource: v1.PersistentVolumeSource{
					FlexVolume: &v1.FlexPersistentVolumeSource{
						Driver: "fvdDriver",
					},
				},
			},
			Status: v1.PersistentVolumeStatus{
				Phase: v1.VolumeAvailable,
			},
		},
	}
)

type mockServiceError struct {
	StatusCode   int
	Code         string
	Message      string
	OpcRequestID string
}

func (m mockServiceError) GetHTTPStatusCode() int {
	return m.StatusCode
}

func (m mockServiceError) GetMessage() string {
	return m.Message
}

func (m mockServiceError) GetCode() string {
	return m.Code
}

func (m mockServiceError) GetOpcRequestID() string {
	return m.OpcRequestID
}
func (m mockServiceError) Error() string {
	return m.Message
}

func Test_sbc_pusher(t *testing.T) {

	testCases := []struct {
		name      string
		pvs       []*v1.PersistentVolume
		itemExits bool
	}{
		{
			name: "no OCI CSI/FVD pvs",
			pvs: []*v1.PersistentVolume{
				pvList["not-oci-csi-fvd-pv"],
			},
			itemExits: false,
		},
		{
			name: "oci csi bv pv",
			pvs: []*v1.PersistentVolume{
				pvList["oci-csi-bv-pv"],
			},
			itemExits: true,
		},
		{
			name: "oci csi bv pv 2",
			pvs: []*v1.PersistentVolume{
				pvList["oci-csi-bv-pv-2"],
			},
			itemExits: false,
		},
		{
			name: "oci csi fss pv",
			pvs: []*v1.PersistentVolume{
				pvList["oci-csi-fss-pv"],
			},
			itemExits: false,
		},
		{
			name: "oci csi fvd pv",
			pvs: []*v1.PersistentVolume{
				pvList["oci-csi-fvd-pv"],
			},
			itemExits: true,
		},
		{
			name: "csi bv pv",
			pvs: []*v1.PersistentVolume{
				pvList["csi-bv-pv"],
			},
			itemExits: false,
		},
		{
			name: "csi fss pv",
			pvs: []*v1.PersistentVolume{
				pvList["csi-fss-pv"],
			},
			itemExits: false,
		},
		{
			name: "fvd pv",
			pvs: []*v1.PersistentVolume{
				pvList["fvd-pv"],
			},
			itemExits: false,
		},
	}

	sbc := &StorageBackfillController{
		logger: zap.S(),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sbc.queue = workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(sbBaseRetryDelay, sbMaxRetryDelay))
			sbc.pvLister = &mockPersistentVolumeLister{
				persistentVolumes: tc.pvs,
			}
			sbc.pusher()
			if sbc.queue.Len() > 0 != tc.itemExits {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.itemExits, !tc.itemExits)
			}
		})
	}
}
func Test_getStorageType(t *testing.T) {
	testCases := []struct {
		name                string
		pv                  *v1.PersistentVolume
		expectedStorageType PvStorageType
	}{
		{
			name:                "no OCI CSI/FVD pvs",
			pv:                  pvList["not-oci-csi-fvd-pv"],
			expectedStorageType: "",
		},
		{
			name:                "oci csi bv pv",
			pv:                  pvList["oci-csi-bv-pv"],
			expectedStorageType: BV,
		},
		{
			name:                "oci csi fss pv",
			pv:                  pvList["oci-csi-fss-pv"],
			expectedStorageType: FSS,
		},
		{
			name:                "oci csi fvd pv",
			pv:                  pvList["oci-csi-fvd-pv"],
			expectedStorageType: BV,
		},
		{
			name:                "csi bv pv",
			pv:                  pvList["csi-bv-pv"],
			expectedStorageType: "",
		},
		{
			name:                "csi fss pv",
			pv:                  pvList["csi-fss-pv"],
			expectedStorageType: "",
		},
		{
			name:                "fvd pv",
			pv:                  pvList["fvd-pv"],
			expectedStorageType: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPVType := GetStorageType(tc.pv)

			if actualPVType != tc.expectedStorageType {
				t.Errorf("expected %v but got %v", tc.expectedStorageType, actualPVType)
			}

		})
	}
}

func Test_getBlockVolumeOcidFromPV(t *testing.T) {
	testCases := []struct {
		name     string
		pv       *v1.PersistentVolume
		expected string
		err      error
	}{
		{
			name:     "no OCI CSI/FVD pvs",
			pv:       pvList["not-oci-csi-fvd-pv"],
			expected: "",
			err:      errors.New("unable to get the block volume ocid from pv"),
		},
		{
			name:     "oci csi bv pv",
			pv:       pvList["oci-csi-bv-pv"],
			expected: bvOcid,
			err:      nil,
		},
		{
			name:     "oci csi fss pv",
			pv:       pvList["oci-csi-fss-pv"],
			expected: fssOCid,
			err:      nil,
		},
		{
			name:     "oci csi fvd pv",
			pv:       pvList["oci-csi-fvd-pv"],
			expected: bvOcid,
			err:      nil,
		},
		{
			name:     "csi bv pv",
			pv:       pvList["csi-bv-pv"],
			expected: bvOcid,
			err:      nil,
		},
		{
			name:     "csi fss pv",
			pv:       pvList["csi-fss-pv"],
			expected: fssOCid,
			err:      nil,
		},
		{
			name:     "fvd pv",
			pv:       pvList["fvd-pv"],
			expected: bvOcid,
			err:      nil,
		},
	}
	sbc := &StorageBackfillController{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOcid, err := sbc.getBlockVolumeOcidFromPV(tc.pv)
			if err != nil {
				if !assertError(err, tc.err) {
					t.Errorf("expected error %s but got %s", tc.err, err)
				}
			}
			if actualOcid != tc.expected {
				t.Errorf("expected %v but got %v", tc.expected, actualOcid)
			}

		})
	}

}

func Test_doesBVHaveOkeSystemTag(t *testing.T) {
	testCases := []struct {
		name     string
		bv       *core.Volume
		config   *providercfg.Config
		expected bool
	}{
		{
			name: "basic-volume",
			bv: &core.Volume{
				SystemTags: map[string]map[string]interface{}{
					"orcl-containerengine": {
						"Cluster": "ocid1.cluster.aaaa....",
					},
				},
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{
					Common: &providercfg.TagConfig{
						DefinedTags: map[string]map[string]interface{}{
							"orcl-containerengine": {
								"Cluster": "ocid1.cluster.aaaa....",
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "basic-volume-with-system-tag",
			bv: &core.Volume{
				SystemTags: map[string]map[string]interface{}{
					"orcl-free-tier": {
						"Foo": "bar",
					},
				},
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{
					Common: &providercfg.TagConfig{
						DefinedTags: map[string]map[string]interface{}{
							"orcl-containerengine": {
								"Cluster": "ocid1.cluster.aaaa....",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "basic-volume-with-oke-system-tag-ns",
			bv: &core.Volume{
				SystemTags: map[string]map[string]interface{}{
					"orcl-containerengine": {
						"volume": "ocid1.volume.aaaa..",
					},
				},
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{
					Common: &providercfg.TagConfig{
						DefinedTags: map[string]map[string]interface{}{
							"orcl-containerengine": {
								"Cluster": "ocid1.cluster.aaaa....",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "basic-volume-with-oke-system-tag-key-value",
			bv: &core.Volume{
				SystemTags: map[string]map[string]interface{}{
					"orcl-foobar": {
						"Cluster": "ocid1.cluster.aaaa....",
					},
				},
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{
					Common: &providercfg.TagConfig{
						DefinedTags: map[string]map[string]interface{}{
							"orcl-containerengine": {
								"Cluster": "ocid1.cluster.aaaa....",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "oke-system-tag-is-not-present-in-config",
			bv: &core.Volume{
				SystemTags: map[string]map[string]interface{}{
					"orcl-containerengine": {
						"Cluster": "ocid1.cluster.aaaa....",
					},
				},
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{
					Common: nil,
				},
			},
			expected: false,
		},
	}
	sbc := &StorageBackfillController{
		logger: zap.S(),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sbc.config = tc.config
			actual := sbc.doesBVHaveOkeSystemTag(tc.bv)
			if actual != tc.expected {
				t.Errorf("expected %t but got %t", tc.expected, actual)
			}
		})
	}
}

func Test_addBlockVolumeOkeSystemTags(t *testing.T) {
	testCases := []struct {
		name                     string
		bv                       *core.Volume
		config                   *providercfg.Config
		wantErr                  error
		expectedSbWorkerDelayVal time.Duration
	}{
		{
			name: "expect an error oke system tag is not loaded onto config (nil)",
			bv: &core.Volume{
				Id: common.String("sample-id"),
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{},
			},
			wantErr: errors.New("oke system tag is not found in the cloud config"),
		},
		{
			name: "expect an error when defined tags are limits are reached",
			bv: &core.Volume{
				Id:          common.String("sample-id"),
				DefinedTags: map[string]map[string]interface{}{},
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{
					Common: &providercfg.TagConfig{
						DefinedTags: map[string]map[string]interface{}{
							"orcl-containerengine": {
								"Cluster": "ocid1.cluster.aaaa...",
							},
						},
					},
				},
			},
			wantErr: errors.New("max limit of defined tags for volume is reached. skip adding tags. sending metric"),
		},
		{
			name: "expect an error when updateLoadBalancer work request fails",
			bv: &core.Volume{
				Id:          common.String("work-request-fails"),
				DefinedTags: map[string]map[string]interface{}{},
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{
					Common: &providercfg.TagConfig{
						DefinedTags: map[string]map[string]interface{}{
							"orcl-containerengine": {
								"Cluster": "ocid1.cluster.aaaa...",
							},
						},
					},
				},
			},
			wantErr: errors.New("UpdateVolume request failed: internal server error"),
		},
		{
			name: "expect an error and slow down processing rate when updateLoadBalancer work request with 429",
			bv: &core.Volume{
				Id:          common.String("api-returns-too-many-requests"),
				DefinedTags: map[string]map[string]interface{}{},
			},
			config: &providercfg.Config{
				Tags: &providercfg.InitialTags{
					Common: &providercfg.TagConfig{
						DefinedTags: map[string]map[string]interface{}{
							"orcl-containerengine": {
								"Cluster": "ocid1.cluster.aaaa...",
							},
						},
					},
				},
			},
			wantErr:                  errors.New("Too many requests"),
			expectedSbWorkerDelayVal: time.Second * 6,
		},
	}
	sbc := &StorageBackfillController{
		logger: zap.S(),
	}
	SbWorkerDelay = SbWorkerInitialDelay

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sbc.config = tc.config
			sbc.ociClient = &MockOCIClient{}
			if strings.Contains(tc.name, "limit") {
				for i := 1; i <= 64; i++ {
					tc.bv.DefinedTags["ns"+strconv.Itoa(i)] = map[string]interface{}{"key": strconv.Itoa(i)}
				}
			}
			err := sbc.addBlockVolumeOkeSystemTags(tc.bv)
			t.Logf("%v", tc.wantErr)
			t.Logf("%v", err)
			if !assertError(tc.wantErr, err) {
				t.Errorf("expected %v but got %v", tc.wantErr, err)
			}
			if strings.Contains(tc.name, "429") &&
				tc.expectedSbWorkerDelayVal != SbWorkerDelay {
				t.Errorf("expected SbWorkerDelay to be %v but got %v", tc.expectedSbWorkerDelayVal, SbWorkerDelay)
			}
		})
	}
}

func Test_isPvPhaseEligble(t *testing.T) {
	testCases := []struct {
		name     string
		pv       *v1.PersistentVolume
		expected bool
	}{
		{
			name:     "nil pv phase",
			pv:       &v1.PersistentVolume{},
			expected: false,
		},
		{
			name: "eligble pv",
			pv: &v1.PersistentVolume{
				Status: v1.PersistentVolumeStatus{
					Phase: v1.VolumeAvailable,
				},
			},
			expected: true,
		},
		{
			name: "eligble pv-2",
			pv: &v1.PersistentVolume{
				Status: v1.PersistentVolumeStatus{
					Phase: v1.VolumeBound,
				},
			},
			expected: true,
		},
		{
			name: "eligble pv-3",
			pv: &v1.PersistentVolume{
				Status: v1.PersistentVolumeStatus{
					Phase: v1.VolumeReleased,
				},
			},
			expected: true,
		},
		{
			name: "ineligble pv",
			pv: &v1.PersistentVolume{
				Status: v1.PersistentVolumeStatus{
					Phase: v1.VolumePending,
				},
			},
			expected: false,
		},
		{
			name: "ineligble pv-2",
			pv: &v1.PersistentVolume{
				Status: v1.PersistentVolumeStatus{
					Phase: v1.VolumeFailed,
				},
			},
			expected: false,
		},
	}
	sbc := &StorageBackfillController{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := sbc.isPvPhaseEligble(tc.pv)
			if actual != tc.expected {
				t.Errorf("expected %t but got %t", tc.expected, actual)
			}
		})
	}
}

type mockPersistentVolumeLister struct {
	persistentVolumes []*v1.PersistentVolume
}

func (s *mockPersistentVolumeLister) List(selector labels.Selector) (ret []*v1.PersistentVolume, err error) {
	var pvs, allPvs []*v1.PersistentVolume
	if len(s.persistentVolumes) > 0 {
		allPvs = s.persistentVolumes
	}

	for _, pv := range allPvs {
		if selector != nil {
			if selector.Matches(labels.Set(pv.ObjectMeta.GetLabels())) {
				pvs = append(pvs, pv)
			}
		} else {
			pvs = append(pvs, pv)
		}
	}
	return pvs, nil
}

func (s *mockPersistentVolumeLister) Get(name string) (*v1.PersistentVolume, error) {
	if pv, ok := pvList[name]; ok {
		return pv, nil
	}
	return nil, errors.New("pv not found")
}
