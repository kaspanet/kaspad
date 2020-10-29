package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
)

func DbSubnetworkIdToDomainSubnetworkID(dbSubnetworkId *DbSubnetworkId) (*externalapi.DomainSubnetworkID, error) {
	return subnetworks.FromBytes(dbSubnetworkId.SubnetworkId)
}

func DomainSubnetworkIDToDbSubnetworkId(domainSubnetworkID *externalapi.DomainSubnetworkID) *DbSubnetworkId {
	return &DbSubnetworkId{SubnetworkId: domainSubnetworkID[:]}
}
