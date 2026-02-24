package ocid

// OCIDVersion determines the ocid version
type OCIDVersion uint8

// OCIDVersion definitions should maintain binary compatibility with
// https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/core/browse/src/main/java/com/oracle/pic/commons/id/IdV2.java
const (
	OCIDVersionV0 OCIDVersion = 0
	OCIDVersionV1 OCIDVersion = 1
	OCIDVersionV2 OCIDVersion = 2
)

// OCIDVersion_name provides name given the value of the ocid version
var OCIDVersion_name = map[uint8]string{
	0: "ocidv0",
	1: "ocidv1",
	2: "ocid1",
}

// OCIDVersion_value provides the value given the name of the ocid version
var OCIDVersion_value = map[string]uint8{
	"ocidv0": 0,
	"ocidv1": 1,
	"ocid1":  2,
}

// String returns the name of the ocid version
func (x OCIDVersion) String() string {
	return OCIDVersion_name[uint8(x)]
}
