package transactionid

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// cmp compares two transaction IDs and returns:
//
//   -1 if a <  b
//    0 if a == b
//   +1 if a >  b
//
func cmp(a, b *externalapi.DomainTransactionID) int {
	// We compare the transaction IDs backwards because Hash is stored as a little endian byte array.
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

// Less returns true iff transaction ID a is less than hash b
func Less(a, b *externalapi.DomainTransactionID) bool {
	return cmp(a, b) < 0
}
