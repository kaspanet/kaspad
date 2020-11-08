package externalapi

import "encoding/hex"

// DomainSubnetworkIDSize is the size of the array used to store subnetwork IDs.
const DomainSubnetworkIDSize = 20

// DomainSubnetworkID is the domain representation of a Subnetwork ID
type DomainSubnetworkID [DomainSubnetworkIDSize]byte

// String stringifies a subnetwork ID.
func (id DomainSubnetworkID) String() string {
	for i := 0; i < DomainSubnetworkIDSize/2; i++ {
		id[i], id[DomainSubnetworkIDSize-1-i] = id[DomainSubnetworkIDSize-1-i], id[i]
	}
	return hex.EncodeToString(id[:])
}
