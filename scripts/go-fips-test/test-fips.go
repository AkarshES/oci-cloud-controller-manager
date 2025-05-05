package main

import (
	"crypto/fips140"
	"fmt"
)

func main() {
	fmt.Printf("FIPS enabled %v\n", fips140.Enabled())
}
