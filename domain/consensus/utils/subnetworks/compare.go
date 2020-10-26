package subnetworks

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func cmp(a, b externalapi.DomainSubnetworkID) int {
	// We compare the hashes backwards because Hash is stored as a little endian byte array.
	for i := externalapi.DomainSubnetworkIDSize - 1; i >= 0; i-- {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}
	return 0
}

// Less returns true iff id a is less than id b
func Less(a, b externalapi.DomainSubnetworkID) bool {
	return cmp(a, b) < 0
}
