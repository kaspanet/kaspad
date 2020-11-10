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

func (c *consensusStateStore) StageVirtualDiffParents(tipHashes []*externalapi.DomainHash) error {
	clone, err := c.cloneVirtualDiffParents(tipHashes)
	if err != nil {
		return err
	}

	c.stagedVirtualDiffParents = clone
	return nil
}

func (c *consensusStateStore) commitVirtualDiffParents(dbTx model.DBTransaction) error {
	virtualDiffParentsBytes := hashes.SerializeHashSlice(c.stagedVirtualDiffParents)

	err := dbTx.Put(virtualDiffParentsKey, virtualDiffParentsBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *consensusStateStore) serializeVirtualDiffParents(virtualDiffParentsBytes []*externalapi.DomainHash) ([]byte, error) {
	panic("unimplemented")
}

func (c *consensusStateStore) deserializeVirtualDiffParents(virtualDiffParentsBytes []byte) ([]*externalapi.DomainHash,
	error) {

	panic("unimplemented")
}

func (c *consensusStateStore) cloneVirtualDiffParents(virtualDiffParents []*externalapi.DomainHash,
) ([]*externalapi.DomainHash, error) {

	serialized, err := c.serializeVirtualDiffParents(virtualDiffParents)
	if err != nil {
		return nil, err
	}

	return c.deserializeVirtualDiffParents(serialized)
}
