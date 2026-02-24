// A shortid is composed of a prefix and the last X characters of the encoded
// entity part of an ocid.  The number of characters for X is determined by
// the short id version.  The prefix is embedded in the ocid entity for short
// id version v1.  The prefix is the first letter of the ocid type for short
// id version of "none".  Look at the examples in TestOCID_New.
package ocid

import (
	"bytes"
	"fmt"
)

// ShortIDVersion determines the version of the short id
type ShortIDVersion byte

// consts...
const (
	ShortIDVersionNone ShortIDVersion = 0
	ShortIDVersionV1   ShortIDVersion = 1

	DefaultShortIDPrefixV1 = uint8('z')
)

// ShortIDVersion_name provides name given the value of the short id version
var ShortIDVersion_name = map[byte]string{
	0: "none",
	1: "v1",
}

// ShortIDVersion_value provides the value given the name of the short id version
var ShortIDVersion_value = map[string]byte{
	"none": 0,
	"v1":   1,
}

// ShortIDVersion_size provides the size of the short id given the value of the short id version
var ShortIDVersion_size = map[byte]int{
	0: 11,
	1: 11,
}

// String returns the name of the short id version
func (x ShortIDVersion) String() string {
	if _, ok := ShortIDVersion_name[byte(x)]; !ok {
		return ShortIDVersion_name[byte(ShortIDVersionNone)]
	}
	return ShortIDVersion_name[byte(x)]
}

// Size returns the size of the short id given the short id version
func (x ShortIDVersion) Size() int {
	if _, ok := ShortIDVersion_size[byte(x)]; !ok {
		return ShortIDVersion_size[byte(ShortIDVersionNone)]
	}
	return ShortIDVersion_size[byte(x)]
}

// ByteToShortIDVersion converts a byte to a short id version type
func ByteToShortIDVersion(v byte) ShortIDVersion {
	if _, ok := ShortIDVersion_name[v]; !ok {
		return ShortIDVersionNone
	}
	return ShortIDVersion(v)
}

// getShortID returns the short id from the given version, prefix, and long id
func getShortID(shortIDVersion ShortIDVersion, prefixLetter uint8, longID string) string {
	prefix := fmt.Sprintf("%c", prefixLetter)
	shortIDSize := shortIDVersion.Size()
	if shortIDSize-1 > len(longID) {
		return prefix + longID
	}

	index := len(longID) - ShortIDVersionV1.Size() + 1
	return prefix + string(longID[index:])
}

// decodeShortIDPrefix returns the shortid version, the short id prefix, and possible error
// by inspecting the ocid type, decoded data, encoded data, and desired short id index in
// the encoded data.
func decodeShortIDPrefix(typ, decoded, encoded string, decodedShortIDVersionIndex int) (ShortIDVersion, uint8, error) {
	getTypePrefix := func() byte {
		if len(typ) > 0 {
			return typ[0]
		}
		return DefaultShortIDPrefixV1
	}
	if decodedShortIDVersionIndex >= len(decoded) {
		return ShortIDVersionNone, getTypePrefix(), nil
	}

	shortIDVersion := ByteToShortIDVersion(decoded[decodedShortIDVersionIndex])
	shortIDSize := shortIDVersion.Size()
	if shortIDSize > len(encoded) {
		return shortIDVersion, 0, fmt.Errorf("invalid length of encoded entity to decode short id prefix")
	}

	if shortIDVersion == ShortIDVersionNone {
		return shortIDVersion, getTypePrefix(), nil
	}

	encodedIndex := len(encoded) - shortIDSize
	return shortIDVersion, encoded[encodedIndex], nil
}

// embedEncodedLetter embeds the provided letter into the decode byte slice
// given the encoded slize size and encoded index.  negative encoded index
// values start from the end.
func embedEncodedLetter(decoded []byte, encodedLetter uint8, encodedSize int, encodedIndex int /* zero-based */) error {
	if encodedSize < 0 {
		return fmt.Errorf("invalid encoded size of %d cannot be negative", encodedSize)
	}
	if (encodedLetter < 'a' || encodedLetter > 'z') && (encodedLetter < 'A' || encodedLetter > 'Z') {
		return fmt.Errorf("encoded letter of '%c' is not within: a-zA-Z", encodedLetter)
	}

	if encodedIndex < 0 {
		if -encodedIndex > encodedSize {
			return fmt.Errorf("invalid out of bounds for encoded index %d outside encoded size %d", encodedIndex, encodedSize)
		}

		encodedIndex = encodedSize - -encodedIndex // calculate encodedIndex to be index from left
	} else if encodedIndex >= encodedSize {
		return fmt.Errorf("invalid out of bounds for encoded index %d outside encoded size %d", encodedIndex, encodedSize)
	}

	decodedValue := bytes.ToUpper([]byte{encodedLetter})[0] - 'A'

	const positionsPerEncodedGroup = EncodedGroupElements
	groupsToSkip := encodedIndex / positionsPerEncodedGroup
	encodedGroupPosition := encodedIndex % positionsPerEncodedGroup

	const positionsPerDecodedGroup = DecodedGroupElements
	firstIndex := groupsToSkip * positionsPerDecodedGroup
	if firstIndex+positionsPerDecodedGroup > len(decoded) {
		return fmt.Errorf("input size %d of decoded slice is too small - needs to be size of at least %d",
			len(decoded), firstIndex+positionsPerDecodedGroup)
	}
	switch encodedGroupPosition {
	/* https://tools.ietf.org/html/rfc4648
	   The case for base 32 is shown in the following figure, borrowed from
	   [7].  Each successive character in a base-32 value represents 5
	   successive bits of the underlying octet sequence.  Thus, each group
	   of 8 characters represents a sequence of 5 octets (40 bits).

	                        1          2          3
	             01234567 89012345 67890123 45678901 23456789
	            +--------+--------+--------+--------+--------+
	            |< 1 >< 2| >< 3 ><|.4 >< 5.|>< 6 ><.|7 >< 8 >|
	            +--------+--------+--------+--------+--------+
	                                                    <===> 8th character
	                                              <====> 7th character
	                                         <===> 6th character
	                                   <====> 5th character
	                             <====> 4th character
	                        <===> 3rd character
	                  <====> 2nd character
	             <===> 1st character

	*/

	case 0: // 1st 5-bit octet
		firstIndex += 0
		decoded[firstIndex] &= byte(0x7)
		decoded[firstIndex] |= decodedValue << 3
	case 1: // 2nd 5-bit octet
		firstIndex += 0
		decoded[firstIndex] &= ^byte(0x7)
		decoded[firstIndex] |= decodedValue >> 2
		decoded[firstIndex+1] &= byte(0x3F)
		decoded[firstIndex+1] |= decodedValue << 6
	case 2: // 3rd 5-bit octet
		firstIndex += 1
		decoded[firstIndex] &= ^byte(0x3E)
		decoded[firstIndex] |= decodedValue << 1
	case 3: // 4th 5-bit octet
		firstIndex += 1
		decoded[firstIndex] &= byte(0xFE)
		decoded[firstIndex] |= decodedValue << 1
		decoded[firstIndex+1] &= byte(0x0F)
		decoded[firstIndex+1] |= decodedValue << 4
	case 4: // 5th 5-bit octet
		firstIndex += 2
		decoded[firstIndex] &= byte(0xF0)
		decoded[firstIndex] |= decodedValue >> 1
		decoded[firstIndex+1] &= byte(0x7F)
		decoded[firstIndex+1] |= decodedValue << 7
	case 5: // 6th 5-bit octet
		firstIndex += 3
		decoded[firstIndex] &= byte(0x83)
		decoded[firstIndex] |= decodedValue << 2
	case 6: // 7th 5-bit octet
		firstIndex += 3
		decoded[firstIndex] &= byte(0xFC)
		decoded[firstIndex] |= decodedValue >> 3
		decoded[firstIndex+1] &= byte(0x1F)
		decoded[firstIndex+1] |= decodedValue << 5
	case 7: // 8th 5-bit octet
		firstIndex += 4
		decoded[firstIndex] &= byte(0xE0)
		decoded[firstIndex] |= decodedValue

	default:
		return fmt.Errorf("invalid encodedGroupPosition:%d derived from encodedSize:%d encodedIndex:%d\n",
			encodedGroupPosition, encodedSize, encodedIndex)
	}

	return nil
}
