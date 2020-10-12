package hashes

import "github.com/kaspanet/kaspad/domain/consensus/model"

// cmp compares two hashes and returns:
//
//   -1 if a <  b
//    0 if a == b
//   +1 if a >  b
//
func cmp(a, b *model.DomainHash) int {
	// We compare the hashes backwards because Hash is stored as a little endian byte array.
	for i := model.HashSize - 1; i >= 0; i-- {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}
	return 0
}

// Less returns true iff hash a is less than hash b
func Less(a, b *model.DomainHash) bool {
	return cmp(a, b) < 0
}
