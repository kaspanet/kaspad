package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

var virtualDiffParentsBucket = dbkeys.MakeBucket([]byte("virtual-diff-parents"))
var virtualDiffParentsKey = virtualDiffParentsBucket.Key([]byte("virtual-diff-parents"))

func (c consensusStateStore) VirtualDiffParents(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if c.stagedVirtualDiffParents != nil {
		return c.stagedVirtualDiffParents, nil
	}

	virtualDiffParentsBytes, err := dbContext.Get(virtualDiffParentsKey)
	if err != nil {
		return nil, err
	}

	return hashes.DeserializeHashSlice(virtualDiffParentsBytes)
}

func (c consensusStateStore) StageVirtualDiffParents(tipHashes []*externalapi.DomainHash) error {
	c.stagedVirtualDiffParents = tipHashes

	return nil
}

func (c consensusStateStore) commitVirtualDiffParents(dbTx model.DBTransaction) error {
	virtualDiffParentsBytes := hashes.SerializeHashSlice(c.stagedVirtualDiffParents)

	err := dbTx.Put(virtualDiffParentsKey, virtualDiffParentsBytes)
	if err != nil {
		return err
	}

	c.stagedVirtualDiffParents = nil
	return nil
}
