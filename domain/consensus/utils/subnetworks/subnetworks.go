package subnetworks

import "github.com/kaspanet/kaspad/domain/consensus/model"

var (
	// SubnetworkIDNative is the default subnetwork ID which is used for transactions without related payload data
	SubnetworkIDNative = model.DomainSubnetworkID{}

	// SubnetworkIDCoinbase is the subnetwork ID which is used for the coinbase transaction
	SubnetworkIDCoinbase = model.DomainSubnetworkID{1}

	// SubnetworkIDRegistry is the subnetwork ID which is used for adding new sub networks to the registry
	SubnetworkIDRegistry = model.DomainSubnetworkID{2}
)

// IsBuiltIn returns true if the subnetwork is a built in subnetwork, which
// means all nodes, including partial nodes, must validate it, and its transactions
// always use 0 gas.
func IsBuiltIn(id model.DomainSubnetworkID) bool {
	return id == SubnetworkIDCoinbase || id == SubnetworkIDRegistry
}

// IsBuiltInOrNative returns true if the subnetwork is the native or a built in subnetwork,
// see IsBuiltIn for further details
func IsBuiltInOrNative(id model.DomainSubnetworkID) bool {
	return id == SubnetworkIDNative || IsBuiltIn(id)
}
