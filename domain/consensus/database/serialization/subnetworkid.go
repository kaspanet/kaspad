package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func DbSubnetworkIdToDomainSubnetworkID(dbSubnetworkId *DbSubnetworkId) (*externalapi.DomainSubnetworkID, error) {
	panic("implement me")
}

func DomainSubnetworkIDToDbSubnetworkId(domainSubnetworkID *externalapi.DomainSubnetworkID) *DbSubnetworkId {
	return &DbSubnetworkId{SubnetworkId: domainSubnetworkID[:]}
}
