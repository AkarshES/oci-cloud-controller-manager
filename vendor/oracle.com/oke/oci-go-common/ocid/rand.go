package ocid

import (
	"crypto/rand"
	"fmt"
)

// Rand gives back a random string from a crypto secure source in the OS
var Rand func(int) string = randProduction

func randProduction(size int) string {
	b := make([]byte, size/2)
	_, err := rand.Read(b)
	if err != nil {
		// failure is highly unlikely, but we didn't ignore the error
		panic(fmt.Sprintf("failed to read /dev/urandom: %s", err))
	}
	return fmt.Sprintf("%x", b)
}

func init() {
	Rand(DefaultEntityEncodedSize)
}
