// base32.go defines consts to help with conforming with https://tools.ietf.org/html/rfc4648
package ocid

const (
	EncodedGroupElements = 8 // number of 5-bit octets in a group within encoded data
	DecodedGroupElements = 5 // number of 8-bit characters in a group within decoded data

	EncodedElementBits = 8 // number of bits in an encoded character
	DecodedElementBits = 5 // number of bits in a decoded octet
)
