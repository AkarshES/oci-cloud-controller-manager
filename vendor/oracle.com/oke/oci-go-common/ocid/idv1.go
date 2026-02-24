/**
    idv1.go

	Some refs:
	Java implementation: https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/core/browse/core-resources/src/main/java/com/oracle/pic/commons/id/IdV1.java
	OCID specification: https://confluence.oci.oraclecorp.com/display/DEX/OCID+specification
	OCI Resource: https://confluence.oci.oraclecorp.com/pages/viewpage.action?pageId=31610248
*/
package ocid

import (
	"encoding/base32"
	"fmt"
	"strings"
)

var _ OCID = &OCIDV1{} 			  // assert that the impl meets the interface

// OCIDV1 defines the ocid type at version v1
type OCIDV1 struct {
	ID             string         // serialized ocid, e.g. "ocidv1:cluster:oc1:iad:aannb3f4aac:aaaaaaaaafsgentggm4wezrcgrtgcytcmq"
	version        string         // ocid version, e.g. "ocidv1"
	entityType     string         // entity type, e.g. "cluster"
	realm          string         // ocid realm, e.g. "oc1"
	region         string         // entity region, e.g. "iad"
	extensions     []string       // extensions, e.g. ["aannb3f4aac"]
	entityDecoded  string         // decoded entity, e.g. ""
	entityEncoded  string         // encoded entity, e.g. "aaaaaaaaafsgentggm4wezrcgrtgcytcmq"
	shortIDVersion ShortIDVersion // version of the short id, e.g. 0x1
	shortIDPrefix  uint8          // prefix for the short id, e.g. 'c'
}

// NewOCIDV1 parses a new ocid v1 into its parts and returns a pointer to the OCIDV1 type
func NewOCIDV1(ocidv1 string) (*OCIDV1, error) {
	parts := strings.Split(ocidv1, ":")
	count := len(parts)
	if l := count; l < 5 {
		return nil, fmt.Errorf("invalid number of parts of %d in ocid", l)
	}
	if ver := parts[0]; ver != OCIDVersionV1.String() {
		return nil, fmt.Errorf("invalid ocid version: %s", ver)
	}
	if typ := parts[1]; len(typ) < 1 {
		return nil, fmt.Errorf("invalid empty ocid type: %s", typ)
	}
	if l := len(parts[count-1]); l < MinimumEntityEncodedSize {
		return nil, fmt.Errorf("invalid ocid entity with length %d does not meet minimum of %d", l, MinimumEntityEncodedSize)
	}
	var extensions []string
	if count > 5 {
		extensions = parts[4 : count - 1]
	}

	o := &OCIDV1{
		ID:            ocidv1,
		version:       parts[0],
		entityType:    parts[1],
		realm:         parts[2],
		region:        parts[3],
		extensions:    extensions,
		entityEncoded: parts[count - 1],
	}

	// decode using base32 to maintain compatibility with:
	// https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/core/browse/core-resources/src/main/java/com/oracle/pic/commons/id/IdV1.java
	decoded, err := decodeDataV1(o.entityEncoded)
	if err != nil {
		return nil, fmt.Errorf("unable to decode ocid entity due to error: %s", err)
	}
	o.entityDecoded = string(decoded)
	o.shortIDVersion, o.shortIDPrefix, err = decodeShortIDPrefix(
		o.entityType, o.entityDecoded, o.entityEncoded, len(EntityJavaHeader))
	if err != nil {
		return nil, fmt.Errorf("unable to decode ocid short id prefix due to error: %s", err)
	}
	return o, nil
}

// GenerateOCIDV1Options defines the options to the GenerateOCIDV1 function
type GenerateOCIDV1Options struct {
	entityEncodedSize int
	realm             string
	region            string
	shortIDVersion    ShortIDVersion
	shortIDPrefix     uint8
}
// GenerateOCIDV1OptionsFunc defines the function to set the options
type GenerateOCIDV1OptionsFunc func(o *GenerateOCIDV1Options)
// EntityEncodedSizeV1 assigns an override for the entity encoded size
func EntityEncodedSizeV1(entityEncodedSize int) GenerateOCIDV1OptionsFunc {
	return func(o *GenerateOCIDV1Options) {
		o.entityEncodedSize = entityEncodedSize
	}
}
// RealmV1 assigns an override for the realm
func RealmV1(realm string) GenerateOCIDV1OptionsFunc {
	return func(o *GenerateOCIDV1Options) {
		o.realm = realm
	}
}
// RegionV1 assigns an override for the region
func RegionV1(region string) GenerateOCIDV1OptionsFunc {
	return func(o *GenerateOCIDV1Options) {
		o.region = region
	}
}
// ShortIDVersionV1 assigns an override for the short id version
func ShortIDVersion_V1(shortIDVersion ShortIDVersion) GenerateOCIDV1OptionsFunc {
	return func(o *GenerateOCIDV1Options) {
		o.shortIDVersion = shortIDVersion
	}
}
// ShortIDPrefixV1 assigns an override for the short id prefix letter a-zA-Z
func ShortIDPrefixV1(letter uint8) GenerateOCIDV1OptionsFunc {
	return func(o *GenerateOCIDV1Options) {
		o.shortIDPrefix = letter
	}
}

// GenerateOCIDV1 generates a new ocid v1 and returns a pointer to the OCIDV1 type
func GenerateOCIDV1(typ string, opts ...GenerateOCIDV1OptionsFunc) (*OCIDV1, error) {
	options := &GenerateOCIDV1Options{
		entityEncodedSize: DefaultEntityEncodedSize,
		realm:             RealmDefault,
		region:            "",
		shortIDVersion:    ShortIDVersionV1,
		shortIDPrefix:     0,
	}
	for _, opt := range opts {
		opt(options)
	}
	if min := MinimumEntityEncodedSize + ShortIDVersionV1.Size(); options.entityEncodedSize < min {
		return nil, fmt.Errorf("invalid entity encoded size %d less than minimum size of %d",
			options.entityEncodedSize, min)
	}
	if err := validateEntityTypeV1(typ); err != nil {
		return nil, err
	}
	if options.shortIDPrefix == 0 {
		options.shortIDPrefix = typ[0]
	}

	raw, encoded := generateEntityV1(options.entityEncodedSize, options.shortIDVersion, options.shortIDPrefix)
	o := &OCIDV1{
		version:        OCIDVersionV1.String(),
		entityType:     typ,
		realm:          options.realm,
		region:         options.region,
		entityDecoded:  raw,
		entityEncoded:  encoded,
		shortIDVersion: options.shortIDVersion,
		shortIDPrefix:  options.shortIDPrefix,
	}
	o.ID = serializeV1(o.version, o.entityType, o.realm, o.region, o.entityEncoded)
	return o, nil
}

// String returns the serialized ocid value
func (o *OCIDV1) String() string {
	return o.ID
}

// Version returns the version of the ocid
func (o *OCIDV1) Version() string {
	return o.version
}

// Type returns the type of the ocid
func (o *OCIDV1) Type() string {
	return o.entityType
}

// Region returns the region of the ocid
func (o *OCIDV1) Region() string {
	return o.region
}

// Version returns the realm of the ocid
func (o *OCIDV1) Realm() string {
	return o.realm
}

// Extensions returns the extensions of the ocid
func (o *OCIDV1) Extensions() []string {
	return o.extensions
}

// EntityEncoded returns the encoded entity of the ocid
func (o *OCIDV1) EntityEncoded() string {
	return o.entityEncoded
}

// ShortID returns the short id of the entity.  It is extracted from the end
// of the encoded entity part of the ocid.  It is composed of a one letter
// short id prefix followed by Size-1 alphanumeric characters.
func (o *OCIDV1) ShortID() string {
	return getShortID(o.shortIDVersion, o.shortIDPrefix, o.entityEncoded)
}

// serializeV1 serializes the parts of the ocid into a string
func serializeV1(ver, typ, realm, region, entity string) string {
	return strings.Join([]string{ver, typ, realm, region, entity}, ":")
}

// generateEntityV1 generates the entity string given the encoded size
func generateEntityV1(encodedSize int, shortIDVersion ShortIDVersion, shortIDPrefix uint8) (string /*raw*/, string /*encoded*/) {
	// encode using base32 to maintain compatibility with
	// https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/core/browse/core-resources/src/main/java/com/oracle/pic/commons/id/IdV1.java
	var data []byte
	data = append(data, ([]byte(EntityVersionV0))[:]...)
	data = append(data, ([]byte(EntityNoADBytes))[:]...)

	if encodedSize < MinimumEntityEncodedSize+shortIDVersion.Size() {
		shortIDVersion = ShortIDVersionNone
	}
	data = append(data, ([]byte{byte(shortIDVersion)})[:]...)
	if encodedSize > MinimumEntityEncodedSize+1 {
		rawSize := (encodedSize - MinimumEntityEncodedSize) * DecodedElementBits / EncodedElementBits
		entropy := Rand(rawSize * 2)
		data = append(data, ([]byte(entropy))[:]...)
	}
	if encodedSize > MinimumEntityEncodedSize+2+shortIDVersion.Size() {
		if shortIDVersion != ShortIDVersionNone {
			embedEncodedLetter(data, shortIDPrefix, encodedSize, -shortIDVersion.Size())
		}
	}
	return string(data), string(encodeDataV1(data)[0:encodedSize])
}

// encodeDataV1 encodes the data using base32
func encodeDataV1(data []byte) string {
	str := string(strings.ToLower(base32.StdEncoding.EncodeToString(data)))
	return strings.Replace(str, "=", "", -1)
}

// decodeDataV1 decodes the encoded data using base32
func decodeDataV1(encoded string) (string, error) {
	if len(encoded) == 0 {
		return "", nil
	}
	// pad to multiples of EncodedGroupElements characters
	var padding string
	if mod := len(encoded) % EncodedGroupElements; mod > 0 {
		padding = strings.Repeat("=", EncodedGroupElements - mod)
	}
	decoded, err := base32.StdEncoding.DecodeString(strings.ToUpper(encoded) + padding)
	if err != nil {
		return "", fmt.Errorf("unable to decode ocid entity '%s' due to error: %s", encoded, err)
	}
	return string(decoded), nil
}

// validateEntityTypeV1 validates that the required typ part is not empty
func validateEntityTypeV1(typ string) error {
	if len(typ) <= 0 {
		return ErrEmptyEntityType
	}
	return nil
}
