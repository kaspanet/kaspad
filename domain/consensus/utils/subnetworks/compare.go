package subnetworks

import "github.com/kaspanet/kaspad/domain/consensus/model"

func cmp(a, b model.DomainSubnetworkID) int {
	// We compare the hashes backwards because Hash is stored as a little endian byte array.
	for i := model.SubnetworkIDSize - 1; i >= 0; i-- {
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
func Less(a, b model.DomainSubnetworkID) bool {
	return cmp(a, b) < 0
}
