package hashes

import (
	"encoding/hex"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// NewHashFromStr creates a Hash from a hash string. The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the Hash.
func FromString(hash string) (*externalapi.DomainHash, error) {
	ret := new(externalapi.DomainHash)
	err := decode(ret, hash)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// decode decodes the byte-reversed hexadecimal string encoding of a Hash to a
// destination.
func decode(dst *externalapi.DomainHash, src string) error {
	// Return error if hash string is too long.
	if len(src) != externalapi.DomainHashSize {
		return errors.Errorf("hash string length is %d, while it should be be %d bytes",
			len(src), externalapi.DomainHashSize)
	}

	// Hex decoder expects the hash to be a multiple of two. When not, pad
	// with a leading zero.
	var srcBytes []byte
	if len(src)%2 == 0 {
		srcBytes = []byte(src)
	} else {
		srcBytes = make([]byte, 1+len(src))
		srcBytes[0] = '0'
		copy(srcBytes[1:], src)
	}

	// Hex decode the source bytes to a temporary destination.
	var reversedHash externalapi.DomainHash
	_, err := hex.Decode(reversedHash[externalapi.DomainHashSize-hex.DecodedLen(len(srcBytes)):], srcBytes)
	if err != nil {
		return errors.Wrap(err, "couldn't decode hash hex")
	}

	// Reverse copy from the temporary hash to destination. Because the
	// temporary was zeroed, the written result will be correctly padded.
	for i, b := range reversedHash[:externalapi.DomainHashSize/2] {
		dst[i], dst[externalapi.DomainHashSize-1-i] = reversedHash[externalapi.DomainHashSize-1-i], b
	}

	return nil
}
