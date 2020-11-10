package consensusstatestore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
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

	return c.deserializeVirtualDiffParents(virtualDiffParentsBytes)
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
	virtualDiffParentsBytes, err := c.serializeVirtualDiffParents(c.stagedVirtualDiffParents)
	if err != nil {
		return err
	}

	err = dbTx.Put(virtualDiffParentsKey, virtualDiffParentsBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *consensusStateStore) serializeVirtualDiffParents(virtualDiffParentsBytes []*externalapi.DomainHash) ([]byte, error) {
	virtualDiffParents := serialization.VirtualDiffParentsToDBHeaderVirtualDiffParents(virtualDiffParentsBytes)
	return proto.Marshal(virtualDiffParents)
}

func (c *consensusStateStore) deserializeVirtualDiffParents(virtualDiffParentsBytes []byte) ([]*externalapi.DomainHash,
	error) {

	dbVirtualDiffParents := &serialization.DbVirtualDiffParents{}
	err := proto.Unmarshal(virtualDiffParentsBytes, dbVirtualDiffParents)
	if err != nil {
		return nil, err
	}

	return serialization.DBVirtualDiffParentsToVirtualDiffParents(dbVirtualDiffParents)
}

func (c *consensusStateStore) cloneVirtualDiffParents(virtualDiffParents []*externalapi.DomainHash,
) ([]*externalapi.DomainHash, error) {

	serialized, err := c.serializeVirtualDiffParents(virtualDiffParents)
	if err != nil {
		return nil, err
	}

	return c.deserializeVirtualDiffParents(serialized)
}
