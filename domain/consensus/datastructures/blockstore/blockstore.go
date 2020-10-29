package blockstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var bucket = dbkeys.MakeBucket([]byte("blocks"))

// blockStore represents a store of blocks
type blockStore struct {
	staging  map[externalapi.DomainHash]*externalapi.DomainBlock
	toDelete map[externalapi.DomainHash]struct{}
}

// New instantiates a new BlockStore
func New() model.BlockStore {
	return &blockStore{
		staging:  make(map[externalapi.DomainHash]*externalapi.DomainBlock),
		toDelete: make(map[externalapi.DomainHash]struct{}),
	}
}

// Stage stages the given block for the given blockHash
func (bms *blockStore) Stage(blockHash *externalapi.DomainHash, block *externalapi.DomainBlock) {
	bms.staging[*blockHash] = block
}

func (bms *blockStore) IsStaged() bool {
	return len(bms.staging) != 0 || len(bms.toDelete) != 0
}

func (bms *blockStore) Discard() {
	bms.staging = make(map[externalapi.DomainHash]*externalapi.DomainBlock)
	bms.toDelete = make(map[externalapi.DomainHash]struct{})
}

func (bms *blockStore) Commit(dbTx model.DBTransaction) error {
	for hash, block := range bms.staging {
		blockBytes, err := bms.serializeBlock(block)
		if err != nil {
			return err
		}
		err = dbTx.Put(bms.hashAsKey(&hash), blockBytes)
		if err != nil {
			return err
		}
	}

	for hash, _ := range bms.toDelete {
		err := dbTx.Delete(bms.hashAsKey(&hash))
		if err != nil {
			return err
		}
	}

	bms.Discard()
	return nil
}

// Block gets the block associated with the given blockHash
func (bms *blockStore) Block(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	if block, ok := bms.staging[*blockHash]; ok {
		return block, nil
	}

	blockBytes, err := dbContext.Get(bms.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return bms.deserializeBlock(blockBytes)
}

// HasBlock returns whether a block with a given hash exists in the store.
func (bms *blockStore) HasBlock(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := bms.staging[*blockHash]; ok {
		return true, nil
	}

	exists, err := dbContext.Has(bms.hashAsKey(blockHash))
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Blocks gets the blocks associated with the given blockHashes
func (bms *blockStore) Blocks(dbContext model.DBReader, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error) {
	blocks := make([]*externalapi.DomainBlock, len(blockHashes))
	for i, hash := range blockHashes {
		var err error
		blocks[i], err = bms.Block(dbContext, hash)
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

// Delete deletes the block associated with the given blockHash
func (bms *blockStore) Delete(blockHash *externalapi.DomainHash) {
	if _, ok := bms.staging[*blockHash]; ok {
		delete(bms.staging, *blockHash)
		return
	}
	bms.toDelete[*blockHash] = struct{}{}
}

func (bms *blockStore) serializeBlock(block *externalapi.DomainBlock) ([]byte, error) {
	dbBlock := serialization.DomainBlockToDbBlock(block)
	return proto.Marshal(dbBlock)
}

func (bms *blockStore) deserializeBlock(blockBytes []byte) (*externalapi.DomainBlock, error) {
	dbBlock := &serialization.DbBlock{}
	err := proto.Unmarshal(blockBytes, dbBlock)
	if err != nil {
		return nil, err
	}
	return serialization.DbBlockToDomainBlock(dbBlock)
}

func (bms *blockStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}
