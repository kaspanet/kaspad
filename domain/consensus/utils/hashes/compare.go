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
	// We compare the hashes backwards because Hash is stored as a little endian byte array.
	for i := externalapi.DomainHashSize - 1; i >= 0; i-- {
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
func Less(a, b *externalapi.DomainHash) bool {
	return cmp(a, b) < 0
}

// LessTransactionID returns true iff transactionID a is less then transactionID b
func LessTransactionID(a, b *externalapi.DomainTransactionID) bool {
	return Less((*externalapi.DomainHash)(a), (*externalapi.DomainHash)(b))
}
