package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

var virtualDiffParentsKey = dbkeys.MakeBucket().Key([]byte("virtual-diff-parents"))

func (c *consensusStateStore) VirtualDiffParents(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if c.stagedVirtualDiffParents != nil {
		return c.stagedVirtualDiffParents, nil
	}

	virtualDiffParentsBytes, err := dbContext.Get(virtualDiffParentsKey)
	if err != nil {
		return nil, err
	}

	return hashes.DeserializeHashSlice(virtualDiffParentsBytes)
}

func (c *consensusStateStore) StageVirtualDiffParents(tipHashes []*externalapi.DomainHash) {
	c.stagedVirtualDiffParents = tipHashes
}

func (c *consensusStateStore) commitVirtualDiffParents(dbTx model.DBTransaction) error {
	virtualDiffParentsBytes := hashes.SerializeHashSlice(c.stagedVirtualDiffParents)

	err := dbTx.Put(virtualDiffParentsKey, virtualDiffParentsBytes)
	if err != nil {
		return err
	}

	return nil
}
