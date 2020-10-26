package externalapi

// SubnetworkIDSize is the size of the array used to store subnetwork IDs.
const SubnetworkIDSize = 20

// DomainSubnetworkID is the domain representation of a Subnetwork ID
type DomainSubnetworkID [SubnetworkIDSize]byte
