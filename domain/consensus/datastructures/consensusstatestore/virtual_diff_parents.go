package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (c consensusStateStore) StageVirtualDiffParents(virtualDiffParents []*externalapi.DomainHash) error {
	panic("implement me")
}

func (c consensusStateStore) VirtualDiffParents(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}
