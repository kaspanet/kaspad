package externalapi

// DomainSubnetworkID is the domain representation of a Subnetwork ID
type DomainSubnetworkID [20]byte

var (
	// SubnetworkIDNative is the default subnetwork ID which is used for transactions without related payload data
	SubnetworkIDNative = &DomainSubnetworkID{}

	// SubnetworkIDCoinbase is the subnetwork ID which is used for the coinbase transaction
	SubnetworkIDCoinbase = &DomainSubnetworkID{1}

	// SubnetworkIDRegistry is the subnetwork ID which is used for adding new sub networks to the registry
	SubnetworkIDRegistry = &DomainSubnetworkID{2}
)
