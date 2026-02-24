// Package ocid implements an ocid generator for the OKE service.
//
// Example OCIDs:
//   ocid1.compartment.oc1..aaaaaaaa3oxemcfr554eh75lg5xr4mkmzetbjjuwk6qlehfehpnzfr4oey2q
//   ocid1.image.oc1.iad.aaaaaaaaawy2hh3nreaesyqcdp4m6csg4lwen6ya2njgiyjeu5sodiahlaxq
//   ocid1.subnet.oc1.iad.aaaaaaaas6fctnyttrjzaozxeqbo7w32esy6nfaedqurbnba2q6k2zn7oqia
//   ocid1.instance.oc1.iad.abuwcljsyikzy2kj43aneuqdo22xmpum2i2g4bhjy6w5erzn64yvulcdgvgq
//   ocidv1:tenancy:oc1:phx:1458753575596:aaaaaaaavary4yqe4ljpv5wzp74eflkwpu
//
// For more information check OCI Confluence:
// https://confluence.oci.oraclecorp.com/display/DEX/OCID+specification

/**
 * From https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/core/browse/src/main/java/com/oracle/pic/commons/id/IdV2.java
 *
 * V2 format of ID.
 *
 * Concretely representing identifiers adds considerations such as versioning,
 * what part of an identifier is transparent or opaque, the syntactic structure
 * of the identifier etc. The structure of our identifiers is therefore as
 * follows:
 *
 * <ocid>.<entity-type>.<realm>.<region>(.future-extensibility).<entity-type-specific-id>
 *
 * - ocid
 *   All ids begin with ocid as a means of describing the type of id. This
 *   also acts as a version number. This field is always required.
 *
 * - entity-type
 *   This is the type of entity. "instance" for a compute instance, "volume"
 *   for a block storage volume. Entity types must be centrally managed to
 *   ensure no collisions. This field is always required.
 *
 * - realm
 *   The realm the entity lives in.  This will be "oc1".  This
 *   field is always required. Intuitively this represents the root of a
 *   containment hierarchy for a set of regions that share users, and some
 *   other entities that would otherwise be global. There might, for example,
 *   be a public realm and a gov realm. Realms are globally unique.
 *
 * - region
 *   The region the entity lives in. For Phoenix, this will be "phx". For
 *   regional or AD entitys, region is required. If this field is not
 *   applicable, it should be left blank. Regions are unique within a realm.
 *
 * - future-extensibility
 *   Reserved for future changes to the identifier format. Not required.
 *
 * - entity-type-specific-id
 *   This part of the id, by itself, must be globally unique. This part of the
 *   id may contain structure of its own. This part of the id may contain
 *   obfuscated or encrypted opaque data for routing. The internal AD for
 *   AD-local entities is contained in this part.
 *
 */

package ocid

import (
	"fmt"
	"strings"
)

// OCID defines the ocid interface
type OCID interface {
	String() string        // serialized ocid, e.g. "ocid1.cluster.oc1.iad.aaaaaaaablahdeadbeef1"
	Version() string       // ocid version, e.g. "ocid1"
	Type() string          // entity type, e.g. "cluster"
	Realm() string         // ocid realm, e.g. "oc1"
	Region() string        // entity region, e.g. "iad"
	Extensions() []string  // extensions
	EntityEncoded() string // encoded entity
	ShortID() string       // short id of the entity, e.g. "cdeadbeef1"
}

// New parses a new ocid into its parts
func New(serialized string) (OCID, error) {
	delimiter := " "
	if strings.HasPrefix(serialized, "ocid1") {
		delimiter = "." // OCIDV2
	}
	if strings.HasPrefix(serialized, "ocidv1") {
		delimiter = ":" // OCIDV1
	}

	parts := strings.Split(serialized, delimiter)
	if l := len(parts); l < 1 {
		return nil, fmt.Errorf("invalid number of parts of %d in ocid", l)
	}
	switch parts[0] {
	case OCIDVersionV2.String():
		return NewOCIDV2(serialized)
	case OCIDVersionV1.String():
		return NewOCIDV1(serialized)
	}

	return nil, fmt.Errorf("unsupported ocid version %s", parts[0])
}
