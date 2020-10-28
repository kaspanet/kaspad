package utxodiffstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var utxoDiffBucket = dbkeys.MakeBucket([]byte("utxo-diffs"))
var utxoDiffChildBucket = dbkeys.MakeBucket([]byte("utxo-diff-children"))

// utxoDiffStore represents a store of UTXODiffs
type utxoDiffStore struct {
	utxoDiffStaging      map[externalapi.DomainHash]*model.UTXODiff
	utxoDiffChildStaging map[externalapi.DomainHash]*externalapi.DomainHash
}

// New instantiates a new UTXODiffStore
func New() model.UTXODiffStore {
	return &utxoDiffStore{
		utxoDiffStaging:      make(map[externalapi.DomainHash]*model.UTXODiff),
		utxoDiffChildStaging: make(map[externalapi.DomainHash]*externalapi.DomainHash),
	}
}

// Stage stages the given utxoDiff for the given blockHash
func (uds *utxoDiffStore) Stage(blockHash *externalapi.DomainHash, utxoDiff *model.UTXODiff, utxoDiffChild *externalapi.DomainHash) {
	uds.utxoDiffStaging[*blockHash] = utxoDiff
	uds.utxoDiffChildStaging[*blockHash] = utxoDiffChild
}

func (uds *utxoDiffStore) IsStaged() bool {
	return len(uds.utxoDiffStaging) != 0
}

func (uds *utxoDiffStore) Discard() {
	uds.utxoDiffStaging = make(map[externalapi.DomainHash]*model.UTXODiff)
	uds.utxoDiffChildStaging = make(map[externalapi.DomainHash]*externalapi.DomainHash)
}

func (uds *utxoDiffStore) Commit(dbTx model.DBTransaction) error {
	for hash, utxoDiff := range uds.utxoDiffStaging {
		err := dbTx.Put(uds.utxoDiffHashAsKey(&hash), uds.serializeUTXODiff(utxoDiff))
		if err != nil {
			return err
		}
	}
	for hash, utxoDiffChild := range uds.utxoDiffChildStaging {
		err := dbTx.Put(uds.utxoDiffHashAsKey(&hash), uds.serializeUTXODiffChild(utxoDiffChild))
		if err != nil {
			return err
		}
	}

	uds.Discard()
	return nil
}

// UTXODiff gets the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) UTXODiff(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.UTXODiff, error) {
	if utxoDiff, ok := uds.utxoDiffStaging[*blockHash]; ok {
		return utxoDiff, nil
	}

	utxoDiffBytes, err := dbContext.Get(uds.utxoDiffHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return uds.deserializeUTXODiff(utxoDiffBytes)
}

// UTXODiffChild gets the utxoDiff child associated with the given blockHash
func (uds *utxoDiffStore) UTXODiffChild(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	if utxoDiffChild, ok := uds.utxoDiffChildStaging[*blockHash]; ok {
		return utxoDiffChild, nil
	}

	utxoDiffChildBytes, err := dbContext.Get(uds.utxoDiffChildHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	utxoDiffChild, err := uds.deserializeUTXODiffChild(utxoDiffChildBytes)
	if err != nil {
		return nil, err
	}
	return utxoDiffChild, nil
}

// Delete deletes the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) Delete(dbTx model.DBTransaction, blockHash *externalapi.DomainHash) error {
	if uds.IsStaged() {
		if _, ok := uds.utxoDiffChildStaging[*blockHash]; ok {
			delete(uds.utxoDiffChildStaging, *blockHash)
		}
		if _, ok := uds.utxoDiffStaging[*blockHash]; ok {
			delete(uds.utxoDiffStaging, *blockHash)
		}
		return nil
	}
	err := dbTx.Delete(uds.utxoDiffHashAsKey(blockHash))
	if err != nil {
		return err
	}
	return dbTx.Delete(uds.utxoDiffChildHashAsKey(blockHash))
}

func (uds *utxoDiffStore) utxoDiffHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return utxoDiffBucket.Key(hash[:])
}

func (uds *utxoDiffStore) utxoDiffChildHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return utxoDiffChildBucket.Key(hash[:])
}

func (uds *utxoDiffStore) serializeUTXODiff(utxoDiff *model.UTXODiff) []byte {
	panic("implement me")
}

func (uds *utxoDiffStore) deserializeUTXODiff(utxoDiffBytes []byte) (*model.UTXODiff, error) {
	panic("implement me")
}

func (uds *utxoDiffStore) serializeUTXODiffChild(utxoDiffChild *externalapi.DomainHash) []byte {
	panic("implement me")
}

func (uds *utxoDiffStore) deserializeUTXODiffChild(utxoDiffChildBytes []byte) (*externalapi.DomainHash, error) {
	panic("implement me")
}
