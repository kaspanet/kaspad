package blockstatusstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var bucket = dbkeys.MakeBucket([]byte("block-status"))

// blockStatusStore represents a store of BlockStatuses
type blockStatusStore struct {
	staging map[externalapi.DomainHash]model.BlockStatus
}

// New instantiates a new BlockStatusStore
func New() model.BlockStatusStore {
	return &blockStatusStore{
		staging: make(map[externalapi.DomainHash]model.BlockStatus),
	}
}

// Stage stages the given blockStatus for the given blockHash
func (bss *blockStatusStore) Stage(blockHash *externalapi.DomainHash, blockStatus model.BlockStatus) {
	bss.staging[*blockHash] = blockStatus
}

func (bss *blockStatusStore) IsStaged() bool {
	return len(bss.staging) != 0
}

func (bss *blockStatusStore) Discard() {
	bss.staging = make(map[externalapi.DomainHash]model.BlockStatus)
}

func (bss *blockStatusStore) Commit(dbTx model.DBTransaction) error {
	for hash, status := range bss.staging {
		err := dbTx.Put(bss.hashAsKey(&hash), bss.serializeBlockStatus(status))
		if err != nil {
			return err
		}
	}

	bss.Discard()
	return nil
}

// Get gets the blockStatus associated with the given blockHash
func (bss *blockStatusStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (model.BlockStatus, error) {
	if status, ok := bss.staging[*blockHash]; ok {
		return status, nil
	}

	statusBytes, err := dbContext.Get(bss.hashAsKey(blockHash))
	if err != nil {
		return 0, err
	}

	return bss.deserializeHeader(statusBytes)
}

// Exists returns true if the blockStatus for the given blockHash exists
func (bss *blockStatusStore) Exists(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := bss.staging[*blockHash]; ok {
		return true, nil
	}

	exists, err := dbContext.Has(bss.hashAsKey(blockHash))
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (bms *blockStatusStore) serializeBlockStatus(status model.BlockStatus) []byte {
	panic("implement me")
}

func (bms *blockStatusStore) deserializeHeader(statusBytes []byte) (model.BlockStatus, error) {
	panic("implement me")
}

func (bms *blockStatusStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}
