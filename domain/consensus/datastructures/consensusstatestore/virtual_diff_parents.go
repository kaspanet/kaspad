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
		return externalapi.CloneHashes(css.virtualDiffParentsStaging), nil
	}

	if css.virtualDiffParentsCache != nil {
		return externalapi.CloneHashes(css.virtualDiffParentsCache), nil
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
	return externalapi.CloneHashes(virtualDiffParents), nil
}

func (css *consensusStateStore) StageVirtualDiffParents(tipHashes []*externalapi.DomainHash) {
	css.virtualDiffParentsStaging = externalapi.CloneHashes(tipHashes)
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

	// Note: we don't discard the staging here since that's
	// being done at the end of Commit()
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
