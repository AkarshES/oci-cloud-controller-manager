package ociclient

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"bitbucket.oci.oraclecorp.com/oke/oke-common/instanceuserdata"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/keymanagement"
	"oracle.com/oke/oci-go-common/ocid"
)

// KMSEndpointTemplate is an endpoint URL template for generating vault-specific KMS OCI clients
type KMSEndpointTemplate string

const (
	// KMSManagementEndpointTemplate is the template for a vault-specific KMS Management endpoint URL
	KMSManagementEndpointTemplate KMSEndpointTemplate = "https://%v-management.kms.{region}.oci.{secondLevelDomain}"
	// KMSCryptoEndpointTemplate is the template for a vault-specific KMS Crypto endpoint URL
	KMSCryptoEndpointTemplate KMSEndpointTemplate = "https://%v-crypto.kms.{region}.oci.{secondLevelDomain}"
)

var (
	// for regions need to use OLD KMS related template
	RegionsWithOldKMSEndpoints = map[common.Region]struct{}{
		common.RegionSEA:           {},
		common.RegionPHX:           {},
		common.RegionIAD:           {},
		common.RegionFRA:           {},
		common.RegionLHR:           {},
		common.RegionCAToronto1:    {},
		common.RegionAPSeoul1:      {},
		common.RegionAPTokyo1:      {},
		common.RegionAPMumbai1:     {},
		common.RegionEUZurich1:     {},
		common.RegionSASaopaulo1:   {},
		common.RegionAPSydney1:     {},
		common.RegionMEJeddah1:     {},
		common.RegionEUAmsterdam1:  {},
		common.RegionAPMelbourne1:  {},
		common.RegionAPOsaka1:      {},
		common.RegionCAMontreal1:   {},
		common.RegionUSLangley1:    {},
		common.RegionUSLuke1:       {},
		common.RegionUSGovAshburn1: {},
		common.RegionUSGovChicago1: {},
		common.RegionUSGovPhoenix1: {},
		common.RegionUKGovLondon1:  {},
	}
)

func (c *client) GetKey(ctx context.Context, keyID string) (*keymanagement.Key, error) {
	endpoint, err := GetKmsEndpointFromKeyID(keyID, KMSManagementEndpointTemplate, instanceuserdata.GetEnvCanonicalRegionName())
	if err != nil {
		return nil, err
	}

	kmsManagementClient, err := keymanagement.NewKmsManagementClientWithConfigurationProvider(c.provider, endpoint)
	if err != nil {
		return nil, err
	}

	configureSDKBaseClient(&kmsManagementClient.BaseClient, c.config.RequestSigner, c.config.RequestInterceptor, c.config.RequestDispatcher, "")

	req := keymanagement.GetKeyRequest{
		KeyId:           &keyID,
		RequestMetadata: c.requestMetadata,
	}

	resp, err := kmsManagementClient.GetKey(ctx, req)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.Key, nil
}

func (c *client) GetKeyVersion(ctx context.Context, keyID string, keyVersionID string) (*keymanagement.KeyVersion, error) {
	endpoint, err := GetKmsEndpointFromKeyID(keyID, KMSManagementEndpointTemplate, instanceuserdata.GetEnvCanonicalRegionName())
	if err != nil {
		return nil, err
	}

	kmsManagementClient, err := keymanagement.NewKmsManagementClientWithConfigurationProvider(c.provider, endpoint)
	if err != nil {
		return nil, err
	}

	configureSDKBaseClient(&kmsManagementClient.BaseClient, c.config.RequestSigner, c.config.RequestInterceptor, c.config.RequestDispatcher, "")

	req := keymanagement.GetKeyVersionRequest{
		KeyId:           &keyID,
		KeyVersionId:    &keyVersionID,
		RequestMetadata: c.requestMetadata,
	}

	resp, err := kmsManagementClient.GetKeyVersion(ctx, req)
	if err != nil {
		return nil, NewOCIClientError(resp.OpcRequestId, err)
	}

	return &resp.KeyVersion, nil
}

// GetKmsEndpointFromKeyID gets the vault-specific endpoint URL from the KMS Key OCID string
func GetKmsEndpointFromKeyID(keyID string, endpointTemplate KMSEndpointTemplate, currentRegionName string) (string, error) {
	keyOCID, err := ocid.NewOCIDV2(keyID)
	if err != nil {
		return "", errors.New("invalid ocid: " + err.Error())
	}

	vaultPrefix, err := getKmsVaultPrefixFromKeyOCID(keyOCID)
	if err != nil {
		return "", err
	}

	common.EnableInstanceMetadataServiceLookup()
	currentRegion := common.StringToRegion(strings.ToLower(currentRegionName))
	ocidRegion := common.StringToRegion(strings.ToLower(keyOCID.Region()))

	if currentRegion != ocidRegion {
		return "", errors.New("the vault and key do not belong to same region")
	}

	if _, ok := RegionsWithOldKMSEndpoints[currentRegion]; ok {
		endpointTemplate = KMSEndpointTemplate(strings.Replace(string(endpointTemplate), "oci.{secondLevelDomain}", "{secondLevelDomain}", -1))
	}

	serviceEndpointTemplate := fmt.Sprintf(string(endpointTemplate), vaultPrefix)
	endpoint := currentRegion.EndpointForTemplate("kms", serviceEndpointTemplate)

	return endpoint, nil
}

func getKmsVaultPrefixFromKeyOCID(keyOCID *ocid.OCIDV2) (string, error) {
	extensions := keyOCID.Extensions()
	if extensions == nil || len(extensions) == 0 {
		return "", errors.New("KMS Ocid must have vaultPrefix as first extension")
	}

	return extensions[0], nil
}
