package hashes

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// cmp compares two hashes and returns:
//
//   -1 if a <  b
//    0 if a == b
//   +1 if a >  b
//
func cmp(a, b *externalapi.DomainHash) int {
	aBytes := a.ByteArray()
	bBytes := b.ByteArray()
	// We compare the hashes backwards because Hash is stored as a little endian byte array.
	for i := externalapi.DomainHashSize - 1; i >= 0; i-- {
		switch {
		case aBytes[i] < bBytes[i]:
			return -1
		case aBytes[i] > bBytes[i]:
			return 1
		}
	}
	return 0
}

// Less returns true iff hash a is less than hash b
func Less(a, b *externalapi.DomainHash) bool {
	return cmp(a, b) < 0
}
