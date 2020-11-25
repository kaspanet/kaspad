package consensusstatestore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var virtualDiffParentsKey = dbkeys.MakeBucket().Key([]byte("virtual-diff-parents"))

func (css *consensusStateStore) VirtualDiffParents(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	if css.virtualDiffParentsStaging != nil {
		return css.cloneVirtualDiffParents(css.virtualDiffParentsStaging)
	}

	if css.virtualDiffParentsCache != nil {
		return css.cloneVirtualDiffParents(css.virtualDiffParentsCache)
	}

	virtualDiffParentsBytes, err := dbContext.Get(virtualDiffParentsKey)
	if err != nil {
		return nil, err
	}

	virtualDiffParents, err := css.deserializeVirtualDiffParents(virtualDiffParentsBytes)
	if err != nil {
		return nil, err
	}
	css.virtualDiffParentsCache = virtualDiffParents
	return css.cloneVirtualDiffParents(virtualDiffParents)
}

func (css *consensusStateStore) StageVirtualDiffParents(tipHashes []*externalapi.DomainHash) error {
	clone, err := css.cloneVirtualDiffParents(tipHashes)
	if err != nil {
		return err
	}

	css.virtualDiffParentsStaging = clone
	return nil
}

func (css *consensusStateStore) commitVirtualDiffParents(dbTx model.DBTransaction) error {
	if css.virtualDiffParentsStaging == nil {
		return nil
	}

	virtualDiffParentsBytes, err := css.serializeVirtualDiffParents(css.virtualDiffParentsStaging)
	if err != nil {
		return err
	}
	err = dbTx.Put(virtualDiffParentsKey, virtualDiffParentsBytes)
	if err != nil {
		return err
	}
	css.virtualDiffParentsCache = css.virtualDiffParentsStaging

	return nil
}

func (css *consensusStateStore) serializeVirtualDiffParents(virtualDiffParentsBytes []*externalapi.DomainHash) ([]byte, error) {
	virtualDiffParents := serialization.VirtualDiffParentsToDBHeaderVirtualDiffParents(virtualDiffParentsBytes)
	return proto.Marshal(virtualDiffParents)
}

func (css *consensusStateStore) deserializeVirtualDiffParents(virtualDiffParentsBytes []byte) ([]*externalapi.DomainHash,
	error) {

	dbVirtualDiffParents := &serialization.DbVirtualDiffParents{}
	err := proto.Unmarshal(virtualDiffParentsBytes, dbVirtualDiffParents)
	if err != nil {
		return nil, err
	}

	return serialization.DBVirtualDiffParentsToVirtualDiffParents(dbVirtualDiffParents)
}

func (css *consensusStateStore) cloneVirtualDiffParents(virtualDiffParents []*externalapi.DomainHash,
) ([]*externalapi.DomainHash, error) {

	serialized, err := css.serializeVirtualDiffParents(virtualDiffParents)
	if err != nil {
		return nil, err
	}

	return css.deserializeVirtualDiffParents(serialized)
}
