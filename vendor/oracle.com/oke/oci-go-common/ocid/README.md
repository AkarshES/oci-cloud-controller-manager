# Go OCID package

Package ocid provides generate and parse utilities for the OCI ocids.

## Usage

To parse an ocid, create a new OCID type supplying the ocid.  To generate
an ocid, call the appropriate version of GenerateOCID supplying the ocid
type, region, and desired encoded size.  Then call object methods as the
example (from examples/main.go) below illustrates.

```
package main

import (
	"fmt"

	"oracle.com/oke/oci-go-common/ocid"
)

func ExampleParser() {
	id := "ocid1.instance.oc1.iad.abuwcljsyikzy2kj43aneuqdo22xmpum2i2g4bhjy6w5erzn64yvulcdgvgq"
	ot, err := ocid.New(id)
	if err != nil {
		panic(fmt.Sprintf("unable to use ocid %s due to error: %s\n", id, err))
	}

	fmt.Printf("Parsed -- ID: %s\n", ot.String())
	fmt.Printf("  ShortID: %s\n", ot.ShortID())
	fmt.Printf("  Version: %s\n", ot.Version())
	fmt.Printf("  Type: %s\n", ot.Type())
	fmt.Printf("  Region: %s\n", ot.Region())
	fmt.Println("")
}

func ExampleGenerator() {
	ot, err := ocid.GenerateOCIDV2("cluster", ocid.RegionV2("iad"), ocid.EntityEncodedSizeV2(60))
	if err != nil {
		panic(fmt.Sprintf("unable to generate ocid due to error: %s\n", err))
	}

	fmt.Printf("Generated ID: %s\n", ot.String())
	fmt.Printf("  EntityEncoded: %s\n", ot.EntityEncoded())
	fmt.Printf("  ShortID: %s\n", ot.ShortID())
	fmt.Printf("  Version: %s\n", ot.Version())
	fmt.Printf("  Type: %s\n", ot.Type())
	fmt.Printf("  Region: %s\n", ot.Region())
	fmt.Println("")
}

func main() {
	ExampleParser()

	ExampleGenerator()
}
```

# References
 [Oracle Cloud Infrastructure Resource Identifiers](https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm)
 [OCID Specification] (https://confluence.oci.oraclecorp.com/display/~ffkuo/DRAFT+-+OCID+specification)
 [OCID Java Source] (https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/core/browse/src/main/java/com/oracle/pic/commons/id/IdV2.java)
