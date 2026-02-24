package ocid

import (
	"encoding/base32"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// consts...
const (
	RealmDefault             = "oc1"
	DefaultEntityEncodedSize = 60
	EntityVersionV0          = "\x00"             // from IdV2.java
	EntityNoADBytes          = "\x00\x00\x00\x00" // from IdV2.java
	EntityJavaHeader         = EntityVersionV0 + EntityNoADBytes
)

// vars...
var (
	ErrEmptyEntityType = errors.New("invalid empty entity type")

	MinimumEntityEncodedSize = ((len(EntityJavaHeader) +
		int(reflect.TypeOf(ShortIDVersionV1).Size())) * EncodedElementBits) / DecodedElementBits
)

// assert that the impl meets the interface
var _ OCID = &OCIDV2{}

// OCIDV2 defines the ocid type at version v2
type OCIDV2 struct {
	ID             string         // serialized ocid, e.g. "ocid1.cluster.oc1.iad.aaaaaaaaaeyweztbmuydaobwgjrtszdbgq3gmytdmfrtentdgczwcnbsme2t"
	version        string         // ocid version, e.g. "ocid1"
	entityType     string         // entity type, e.g. "cluster"
	realm          string         // ocid realm, e.g. "oc1"
	region         string         // entity region, e.g. "iad"
	extensions     []string       // extensions, e.g. ["aannb3f4aac"]
	entityDecoded  string         // decoded entity, e.g. "\x00\x00\x00\x00\x00\x011bfae00862c9da46fbcac26c0\xb3a42a5"
	entityEncoded  string         // encoded entity, e.g. "aaaaaaaaaeyweztbmuydaobwgjrtszdbgq3gmytdmfrtentdgczwcnbsme2t"
	shortIDVersion ShortIDVersion // version of the short id, e.g. 0x1
	shortIDPrefix  uint8          // prefix for the short id, e.g. 'c'
}

// NewOCIDV2 parses a new ocid v2 into its parts and returns a pointer to the OCIDV2 type
func NewOCIDV2(ocidv2 string) (*OCIDV2, error) {
	parts := strings.Split(ocidv2, ".")
	count := len(parts)
	if l := count; l < 5 {
		return nil, fmt.Errorf("invalid number of parts of %d in ocid", l)
	}

	if ver := parts[0]; ver != OCIDVersionV2.String() {
		return nil, fmt.Errorf("invalid ocid version of %s", ver)
	}
	if typ := parts[1]; len(typ) < 1 {
		return nil, fmt.Errorf("invalid empty ocid type")
	}
	if l := len(parts[count-1]); l < MinimumEntityEncodedSize {
		return nil, fmt.Errorf("invalid ocid entity with length %d does not meet minimum of %d", l, MinimumEntityEncodedSize)
	}

	var extensions []string
	if count > 5 {
		extensions = parts[4 : count-1]
	}

	o := &OCIDV2{
		ID:            ocidv2,
		version:       parts[0],
		entityType:    parts[1],
		realm:         parts[2],
		region:        parts[3],
		extensions:    extensions,
		entityEncoded: parts[count-1],
	}

	// decode using base32 to maintain compatibility with:
	// https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/core/browse/src/main/java/com/oracle/pic/commons/id/IdV2.java
	decoded, err := decodeDataV2(o.entityEncoded)
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

// GenerateOCIDV2Options defines the options to the GenerateOCIDV2 function
type GenerateOCIDV2Options struct {
	entityEncodedSize int
	realm             string
	region            string
	shortIDVersion    ShortIDVersion
	shortIDPrefix     uint8
}

// GenerateOCIDV2OptionsFunc defines the function to set the options
type GenerateOCIDV2OptionsFunc func(o *GenerateOCIDV2Options)

// EntityEncodedSizeV2 assigns an override for the entity encoded size
func EntityEncodedSizeV2(entityEncodedSize int) GenerateOCIDV2OptionsFunc {
	return func(o *GenerateOCIDV2Options) {
		o.entityEncodedSize = entityEncodedSize
	}
}

// RealmV2 assigns an override for the realm
func RealmV2(realm string) GenerateOCIDV2OptionsFunc {
	return func(o *GenerateOCIDV2Options) {
		o.realm = realm
	}
}

// RegionV2 assigns an override for the region
func RegionV2(region string) GenerateOCIDV2OptionsFunc {
	return func(o *GenerateOCIDV2Options) {
		o.region = region
	}
}

// ShortIDVersionV2 assigns an override for the short id version
func ShortIDVersionV2(shortIDVersion ShortIDVersion) GenerateOCIDV2OptionsFunc {
	return func(o *GenerateOCIDV2Options) {
		o.shortIDVersion = shortIDVersion
	}
}

// ShortIDPrefixV2 assigns an override for the short id prefix letter a-zA-Z
func ShortIDPrefixV2(letter uint8) GenerateOCIDV2OptionsFunc {
	return func(o *GenerateOCIDV2Options) {
		o.shortIDPrefix = letter
	}
}

// GenerateOCIDV2 generates a new ocid v2 and returns a pointer to the OCIDV2 type
func GenerateOCIDV2(typ string, opts ...GenerateOCIDV2OptionsFunc) (*OCIDV2, error) {
	options := &GenerateOCIDV2Options{
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

	if err := validateEntityTypeV2(typ); err != nil {
		return nil, err
	}

	if options.shortIDPrefix == 0 {
		options.shortIDPrefix = typ[0]
	}

	raw, encoded := generateEntityV2(options.entityEncodedSize, options.shortIDVersion, options.shortIDPrefix)
	o := &OCIDV2{
		version:        OCIDVersionV2.String(),
		entityType:     typ,
		realm:          options.realm,
		region:         options.region,
		entityDecoded:  raw,
		entityEncoded:  encoded,
		shortIDVersion: options.shortIDVersion,
		shortIDPrefix:  options.shortIDPrefix,
	}
	o.ID = serializeV2(o.version, o.entityType, o.realm, o.region, o.entityEncoded)

	return o, nil
}

// String returns the serialized ocid value
func (o *OCIDV2) String() string {
	return o.ID
}

// Version returns the version of the ocid
func (o *OCIDV2) Version() string {
	return o.version
}

// Type returns the type of the ocid
func (o *OCIDV2) Type() string {
	return o.entityType
}

// Region returns the region of the ocid
func (o *OCIDV2) Region() string {
	return o.region
}

// Region returns the realm of the ocid
func (o *OCIDV2) Realm() string {
	return o.realm
}

// Extensions returns the extensions of the ocid
func (o *OCIDV2) Extensions() []string {
	return o.extensions
}

// EntityEncoded returns the encoded entity of the ocid
func (o *OCIDV2) EntityEncoded() string {
	return o.entityEncoded
}

// ShortID returns the short id of the entity.  It is extracted from the end
// of the encoded entity part of the ocid.  It is composed of a one letter
// short id prefix followed by Size-1 alphanumeric characters.
func (o *OCIDV2) ShortID() string {
	return getShortID(o.shortIDVersion, o.shortIDPrefix, o.entityEncoded)
}

// serializeV2 serializes the parts of the ocid into a string
func serializeV2(ver, typ, realm, region, entity string) string {
	return strings.Join([]string{ver, typ, realm, region, entity}, ".")
}

// generateEntityV2 generates the entity string given the encoded size
func generateEntityV2(encodedSize int, shortIDVersion ShortIDVersion, shortIDPrefix uint8) (string /*raw*/, string /*encoded*/) {
	// encode using base32 to maintain compatibility with
	// https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/core/browse/src/main/java/com/oracle/pic/commons/id/IdV2.java
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

	return string(data), string(encodeDataV2(data)[0:encodedSize])
}

// encodeDataV2 encodes the data using base32
func encodeDataV2(data []byte) string {
	str := string(strings.ToLower(base32.StdEncoding.EncodeToString(data)))
	return strings.Replace(str, "=", "", -1)
}

// decodeDataV2 decodes the encoded data using base32
func decodeDataV2(encoded string) (string, error) {
	if len(encoded) == 0 {
		return "", nil
	}

	// pad to multiples of EncodedGroupElements characters
	var padding string
	if mod := len(encoded) % EncodedGroupElements; mod > 0 {
		padding = strings.Repeat("=", EncodedGroupElements-mod)
	}
	decoded, err := base32.StdEncoding.DecodeString(strings.ToUpper(encoded) + padding)
	if err != nil {
		return "", fmt.Errorf("unable to decode encoded ocid entity '%s' due to error: %s", encoded, err)
	}
	return string(decoded), nil
}

// validateEntityTypeV2 validates that the required typ part is not empty
func validateEntityTypeV2(typ string) error {
	if len(typ) <= 0 {
		return ErrEmptyEntityType
	}
	return nil
}

// replaceAtIndexFromEnd replaces the character at the index from the end
func replaceAtIndexFromEnd(in string, r rune, idx int) string {
	idx = len(in) - idx
	if idx < 0 {
		return in
	}
	out := []rune(in)
	out[idx] = r
	return string(out)
}
