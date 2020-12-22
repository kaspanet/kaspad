package hashes

import (
	"encoding/hex"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// FromString creates a DomainHash from a hash string. The string should be
// the hexadecimal string of a hash, but any missing characters
// result in zero padding at the end of the Hash.
func FromString(hash string) (*externalapi.DomainHash, error) {
	ret := new(externalapi.DomainHash)
	err := decode(ret, hash)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// decode decodes the hexadecimal string encoding of a Hash to a destination.
func decode(dst *externalapi.DomainHash, src string) error {
	expectedSrcLength := externalapi.DomainHashSize * 2
	// Return error if hash string is too long.
	if len(src) != expectedSrcLength {
		return errors.Errorf("hash string length is %d, while it should be be %d",
			len(src), expectedSrcLength)
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

	// Hex decode the source bytes
	_, err := hex.Decode(dst[externalapi.DomainHashSize-hex.DecodedLen(len(srcBytes)):], srcBytes)
	if err != nil {
		return errors.Wrap(err, "couldn't decode hash hex")
	}
	return nil
}

// ToStrings converts a slice of hashes into a slice of the corresponding strings
func ToStrings(hashes []*externalapi.DomainHash) []string {
	strings := make([]string, len(hashes))
	for i, hash := range hashes {
		strings[i] = hash.String()
	}
	return strings
}
