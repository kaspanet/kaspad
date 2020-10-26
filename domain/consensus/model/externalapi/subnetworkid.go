package externalapi

// DomainSubnetworkIDSize is the size of the array used to store subnetwork IDs.
const DomainSubnetworkIDSize = 20

// DomainSubnetworkID is the domain representation of a Subnetwork ID
type DomainSubnetworkID [DomainSubnetworkIDSize]byte
