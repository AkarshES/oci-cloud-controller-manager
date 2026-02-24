package instanceuserdata

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strings"

	"github.com/pkg/errors"

)

const instanceCertURL = "http://169.254.169.254/opc/v2/identity/cert.pem"

func TenancyID() (string, error) {
	req, err := http.NewRequest("GET", instanceCertURL, nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to create http request")
	}

	certificatePemRaw, err := QueryInstanceMetadata(req)
	if err != nil {
		return "", errors.Wrap(err, "unable to fetch instance cert url")
	}

	var block *pem.Block
	block, _ = pem.Decode(certificatePemRaw)
	if block == nil {
		return "", errors.New("failed to parse the new certificate, not valid pem data")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse the new certificate")
	}

	for _, nameAttr := range cert.Subject.Names {
		value := nameAttr.Value.(string)
		if strings.HasPrefix(value, "opc-tenant:") {
			return value[len("opc-tenant:"):], nil
		}
	}

	return "", errors.New("tenant id not found")
}

func extractTenancyIDFromCertificate(cert *x509.Certificate) string {
	for _, nameAttr := range cert.Subject.Names {
		value := nameAttr.Value.(string)
		if strings.HasPrefix(value, "opc-tenant:") {
			return value[len("opc-tenant:"):]
		}
	}
	return ""
}
