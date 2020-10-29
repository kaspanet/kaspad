package serialization

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func DomainOutpointToDbOutpoint(domainOutpoint *externalapi.DomainOutpoint) *DbOutpoint {
	panic("implement me")
}

func DbOutpointToDomainOutpoint(dbOutpoint *DbOutpoint) (*externalapi.DomainOutpoint, error) {
	panic("implement me")
}
