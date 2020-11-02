package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (c consensusStateStore) StageVirtualUTXODiff(virtualUTXODiff *model.UTXODiff) {
	panic("implement me")
}

func (c consensusStateStore) UTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, error) {
	panic("implement me")
}

func (c consensusStateStore) HasUTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (bool, error) {
	panic("implement me")
}
