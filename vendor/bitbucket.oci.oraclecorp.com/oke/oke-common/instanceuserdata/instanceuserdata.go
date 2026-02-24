package instanceuserdata

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/regions.htm
// https://docs.us-phoenix-1.oraclecloud.com/Content/Compute/Tasks/gettingmetadata.htm
/*
	$ curl -sL http://169.254.169.254/opc/v1/instance/
	{
	  "availabilityDomain" : "hSoJ:US-ASHBURN-AD-1",
	  "compartmentId" : "ocid1.compartment.oc1..aaaaaaaalfj7z4p5km7zel6qtdudqcfndhkyfyhuw3cptb2b5pq2mrugnz5q",
	  "displayName" : "bastion-ad1-0",
	  "id" : "ocid1.instance.oc1.iad.abuwcljt5e4yeh5jayhverhtnvsxjvlg3imevahj4xoipchptbtvzmir3ciq",
	  "image" : "ocid1.image.oc1.iad.aaaaaaaalcitrz3qydbhcpcehohib6lnlv3pmalgqa6vp3brsnhinanlp7pa",
	  "metadata" : {
	    "role" : "bastion",
	    "ssh_authorized_keys" : "ssh-rsa AAAAB3NzaC1yc200000DELETED00000C491EZ5VqBe1W5ZGD"
	  },
	  "region" : "iad",
	  "shape" : "VM.Standard1.1",
	  "state" : "Provisioning",
	  "timeCreated" : 1508508226337
	}

	$ curl -sL http://169.254.169.254/opc/v1/instance/
	{
	  "availabilityDomain" : "zkJl:EU-FRANKFURT-1-AD-2",
	  "canonicalRegionName" : "eu-frankfurt-1",
	  "compartmentId" : "ocid1.compartment.oc1..aaaaaaaaq6uvckviwnnv3uk2khfka4bn254hwjcwwmcdvq5dhiveffthne4a",
	  "displayName" : "oke-c3dcndghfsg-nytcn3bmjqt-snlir3kex6q-0",
	  "faultDomain" : "FAULT-DOMAIN-3",
	  "id" : "ocid1.instance.oc1.eu-frankfurt-1.abtheljtdx6dfd2zy2sm3gjih63agevbweimpzzzksdxhgb3vjsnlwku7jwq",
	  "image" : "ocid1.image.oc1.eu-frankfurt-1.aaaaaaaabfxzgyg2gbwxgkfeir3xfqtkdey3g6k6cmri7wuost35vnrmo4zq",
	  "metadata" : {
	    ...
	    "oke-ad" : "zkJl:EU-FRANKFURT-1-AD-2",
	    ...
	  },
	  "region" : "eu-frankfurt-1",
	  "shape" : "VM.Standard1.2",
	  "state" : "Running",
	  "timeCreated" : 1539981415685
	}
*/

var regionName string
var canonicalRegionName string
var realmDomain string
var airportCode string

var regionNameOnce sync.Once

const instanceUserdataURL = "http://169.254.169.254/opc/v2/instance/"

const DefaultRegionalOCIR = "iad.ocir.io"

// Metadata type here, so we can do customized marshal and unmarshal
// we only meant to parse what can be converted to string into metadata
// This should be the same as https://github.com/oracle/oci-go-sdk/v65/blob/master/core/launch_instance_details.go#L150
// For non string metadata, it should go to a different type (ExtendedMetadata), which we are not loading here
type Metadata map[string]string

// UnmarshalJSON is used to skip those value not string
func (m *Metadata) UnmarshalJSON(data []byte) error {

	var all map[string]interface{}

	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	for k, v := range all {
		if vstr, ok := v.(string); ok {
			(*m)[k] = vstr
		}
	}
	return nil
}

// Object is a type for marshalling all of the returned data from an instance's metadata endpoint
type Object struct {
	AvailabilityDomain  string   `json:"availabilityDomain"`
	FaultDomain         string   `json:"faultDomain"`
	CompartmentID       string   `json:"compartmentId"`
	DisplayName         string   `json:"displayName"`
	ID                  string   `json:"id"`
	Image               string   `json:"image"`
	Metadata            Metadata `json:"metadata"`
	RegionInfo          Metadata `json:"regionInfo"`
	RegionKey           string   `json:"region"`
	CanonicalRegionName string   `json:"canonicalRegionName"`
	Shape               string   `json:"shape"`
	State               string   `json:"state"`
	TimeCreated         int64    `json:"timeCreated"`
}

// RegionName expands the 3 letter IATA region code, useful for
// resolving region URLs on oraclecloud.com
func (i *Object) RegionName() string {
	switch i.RegionKey {
	case "iad":
		return "us-ashburn-1"
	case "phx":
		return "us-phoenix-1"
	case "fra":
		return "eu-frankfurt-1"
	case "lhr":
		return "uk-london-1"
	default:
		return i.RegionKey
	}
}

// AirportCode parses our IMDS object
// Will check if the RegionInfo is available then will look into oke_region and finally will tokenize ocid
func (i *Object) AirportCode() {
	if iataCode, ok := i.RegionInfo["regionKey"]; ok {
		airportCode = strings.ToLower(iataCode)
		return
	}
	if len(strings.Split(i.ID, ".")) > 3 {
		airportCode = strings.Split(i.ID, ".")[3]
		return
	}
	log.Warning("AirportCode() failed to resolved regionInfo or ocid - returning empty string")
}

// RealmDomainComponent parses our IMDS object
// Will check if the RegionInfo is available then will look into oke_region and finally will tokenize ocid
func (i *Object) RealmDomainComponent() {
	if tld, ok := i.RegionInfo["realmDomainComponent"]; ok {
		realmDomain = tld
		return
	}
	log.Warning("RealmDomainComponent() failed to resolved regionInfo - returning empty string")
}

// GetEnvRealmDomain will return the region name for the
// current running environment. It defaults to oraclecloud.com when it
// has not been set.
func GetEnvRealmDomain() string {
	regionNameOnce.Do(getMetadataFunc)
	return realmDomain
}

// GetEnvAirportCode will return the IATA airport code for the
// current running environment. It defaults to iad when it
// has not been set.
func GetEnvAirportCode() string {
	regionNameOnce.Do(getMetadataFunc)
	return airportCode
}

// GetEnvRegionName will return the region name for the
// current running environment. It defaults to us-ashburn-1 when it
// has not been set.
func GetEnvRegionName() string {
	regionNameOnce.Do(getMetadataFunc)
	return regionName
}

// GetEnvRegionKey will return the region name for the
// current running environment. It defaults to us-ashburn-1 when it
// has not been set. This is an alias.
func GetEnvRegionKey() string {
	return GetEnvRegionName()
}

// GetEnvCanonicalRegionName will return the canonical region name for the
// current running environment. It defaults to us-ashburn-1 when it
// has not been set.
func GetEnvCanonicalRegionName() string {
	regionNameOnce.Do(getMetadataFunc)
	return canonicalRegionName
}

// GetObject calls an instance's metadata endpoint and returns the response as Object
func GetObject() (*Object, error) {
	data := &Object{
		// this is necessary to make Metadata not to be nil
		Metadata:   Metadata{},
		RegionInfo: Metadata{},
	}

	req, err := http.NewRequest("GET", instanceUserdataURL, nil)
	if err != nil {
		return data, err
	}

	body, err := QueryInstanceMetadata(req)
	if err != nil {
		return data, err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}

	return data, nil
}

// QueryInstanceMetadata makes querying the instance metadata simpler by providing fallback to v1 logic in case the call to v2 fails
// Note that upon successful response, we only return the body contents as a string, so the caller doesn't need to close the body
func QueryInstanceMetadata(req *http.Request) ([]byte, error) {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	fields := log.Fields{"url": req.URL.String()}

	if !strings.Contains(req.URL.Host, "169.254.169.254") {
		log.WithFields(fields).Warn("Found unexpected host in the metadata url")
	}
	if !strings.Contains(req.URL.Path, "/opc/v2") {
		log.WithFields(fields).Warn("Attempting to query the instance metadata with a non /opc/v2 url")
	}
	if !strings.Contains(req.Header.Get("Authorization"), "Bearer Oracle") {
		req.Header.Add("Authorization", "Bearer Oracle")
	}
	v2resp, err := httpClient.Do(req)
	useV1 := false

	if err != nil {
		useV1 = true
		log.WithFields(fields).WithError(err).Warn("Unable to query instance metadata v2 endpoint. Falling back to v1 endpoint")
	}
	if v2resp != nil {
		if v2resp.StatusCode >= 200 && v2resp.StatusCode < 300 {
			log.WithFields(fields).WithField("status", v2resp.StatusCode).Info("Query Instance metadata v2 endpoint succeeded")
		} else {
			useV1 = true
			defer v2resp.Body.Close()
			log.WithFields(fields).WithField("status", v2resp.StatusCode).Warn("Unable to query instance metadata v2 endpoint. Falling back to v1 endpoint")
		}
	}

	resp := v2resp
	if useV1 {
		newReq := *req
		// based on https://confluence.oci.oraclecorp.com/pages/viewpage.action?spaceKey=C3E&title=Instance+Metadata%3A+Updating+Your+Code+to+Support+IMDS+v2
		// all calls /opc/v1 endpoints have been moved to /opc/v2 except:
		// /latest/attributes
		// /openstack/2013-10-17/user_data
		// /openstack/2013-10-17/meta_data.json
		// We don't use any of these though, so we can blindly rewrite urls from /v2 to /v1 in terms of fallback logic
		newPath := strings.Replace(req.URL.Path, "/opc/v2", "/opc/v1", 1)
		newReq.URL.Path = newPath

		v1resp, err := httpClient.Do(&newReq)
		if err != nil {
			return nil, errors.Wrap(err, "unable to query metadata v1 endpoint after falling back from v2 endpoint")
		}
		resp = v1resp
	}
	// read the response body in this function so the caller doesn't have to remember to close the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read body from instance metadata response")
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// return the response just in case the caller wants to do something with it
		return nil, errors.Errorf("call to instance metadata returned non-success response: %q %q %q", req.URL.String(), resp.Status, body)
	}
	return body, nil
}

func getMetadataFunc() {
	retrieveRegionData(GetObject)
}

// retrieveRegionData sets our variables with region specific information
// Pass in an function that calls an instance's metadata endpoint and returns the response as Object
func retrieveRegionData(getObjectFunc func() (*Object, error)) {
	data, err := getObjectFunc()
	if err != nil {
		log.WithError(err).Error("could not get instance metadata on startup")
	}
	regionName = data.RegionName()
	canonicalRegionName = data.CanonicalRegionName
	data.RealmDomainComponent()
	data.AirportCode()
}
