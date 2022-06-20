package txindex

import (
	"encoding/binary"

	"github.com/kaspanet/kaspad/domain/consensus/database/binaryserialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

var txMergeIndexBucket = database.MakeBucket([]byte("tx-merge-index"))
var txAcceptedIndexBuced = database.MakeBucket([]byte("tx-accepted-index"))
var virtualParentsKey = database.MakeBucket([]byte("")).Key([]byte("tx-index-virtual-parents"))
var ghostDagBlocksKey = database.MakeBucket([]byte("")).Key([]byte("tx-index-ghostdagblocks"))

type txIndexStore struct {
	database database.Database
	toAddMerge map[externalapi.DomainHash][]*externalapi.DomainTransactionID
	toRemoveMerge map[externalapi.DomainHash][]*externalapi.DomainTransactionID
	toAddAccepted map[externalapi.DomainHash][]*externalapi.DomainTransactionID
	toRemoveAccepted map[externalapi.DomainHash][]*externalapi.DomainTransactionID
	virtualParents []*externalapi.DomainHash
	ghostdagBlocks []*externalapi.DomainHash
}

func newTXIndexStore(database database.Database) *txIndexStore {
	return &txIndexStore{
		database: database,
		toAddMerge:  make(map[externalapi.DomainHash][]*externalapi.DomainTransactionID),
		toRemoveMerge: make(map[externalapi.DomainHash][]*externalapi.DomainTransactionID),
		toAddAccepted: make(map[externalapi.DomainHash][]*externalapi.DomainTransactionID),
		toRemoveAccepted: make(map[externalapi.DomainHash][]*externalapi.DomainTransactionID),
		virtualParents: nil,
		ghostdagBlocks: nil,
	}	
}

func (tis *txIndexStore) addMerged(txIDs []*externalapi.DomainTransactionID, mergingBlockHash *externalapi.DomainHash) {
	log.Tracef("Adding %d Txs from mergingBlockHash %s", len(txIDs), mergingBlockHash.String())
	if _, found := tis.toRemoveMerge[*mergingBlockHash]; found {
		delete(tis.toRemoveMerge, *mergingBlockHash)
	}
	tis.toAddMerge[*mergingBlockHash] = txIDs
}

func (tis *txIndexStore) removeMerged(txIDs []*externalapi.DomainTransactionID, mergingBlockHash *externalapi.DomainHash) {
	log.Tracef("Removing %d Txs from mergingBlockHash %s", len(txIDs), mergingBlockHash.String())
	if _, found := tis.toAddMerge[*mergingBlockHash]; found {
		delete(tis.toAddMerge, *mergingBlockHash)
	}
	tis.toRemoveMerge[*mergingBlockHash] = txIDs
}

func (tis *txIndexStore) addAccepted(txIDs []*externalapi.DomainTransactionID, acceptingBlockHash *externalapi.DomainHash) {
	log.Tracef("Adding %d Txs from acceptingBlockHash %s", len(txIDs), acceptingBlockHash.String())
	if _, found := tis.toRemoveAccepted[*acceptingBlockHash]; found {
		delete(tis.toRemoveAccepted, *acceptingBlockHash)
	}
	tis.toAddAccepted[*acceptingBlockHash] = txIDs
}

func (tis *txIndexStore) removeAccepted(txIDs []*externalapi.DomainTransactionID, acceptingBlockHash *externalapi.DomainHash) {
	log.Tracef("Removing %d Txs from acceptingBlockHash %s", len(txIDs), acceptingBlockHash.String())
	if _, found := tis.toAddAccepted[*acceptingBlockHash]; found {
		delete(tis.toAddAccepted, *acceptingBlockHash)
	}
	tis.toRemoveMerge[*acceptingBlockHash] = txIDs
}

func (tis *txIndexStore) discardMerged() {
	tis.toAddMerge =  make(map[externalapi.DomainHash][]*externalapi.DomainTransactionID)
	tis.toRemoveMerge =  make(map[externalapi.DomainHash][]*externalapi.DomainTransactionID)
	tis.virtualParents = nil
}

func (tis *txIndexStore) discardAccepted() {
	tis.toAddAccepted =  make(map[externalapi.DomainHash][]*externalapi.DomainTransactionID)
	tis.toRemoveAccepted = make(map[externalapi.DomainHash][]*externalapi.DomainTransactionID)
	tis.ghostdagBlocks = nil
}

func (tis *txIndexStore) discardAll() {
	tis.discardAccepted()
	tis.discardMerged()
}

func (tis *txIndexStore) removeAll() error {
	tis.removeAccepted()
	tis.removeAll()
	return nil
}


func (tis *txIndexStore) commitAll() error {
	tis.commitAccepted()
	tis.commitMerged()
	return nil
}

func (tis *txIndexStore) commitMerged() error {
	if tis.isAnythingMergingStaged() {
		return errors.Errorf("cannot commit merging TxIds while merge staging isn't empty")
	}
	return nil
}

func (tis *txIndexStore) commitAccepted() error {
	return nil
}


func (tis *txIndexStore) convertTxIDToKey(bucket *database.Bucket, txID *externalapi.DomainTransactionID) *database.Key {
	return bucket.Key(txID.ByteSlice())
}

func (tis *txIndexStore) updateVirtualParents(virtualParents []*externalapi.DomainHash) {
	tis.virtualParents = virtualParents
}

func (tis *txIndexStore) updateGhostDagBlocks(ghostdagBlocks []*externalapi.DomainHash) {
	tis.ghostdagBlocks = ghostdagBlocks
}


func (tis *txIndexStore) convertKeyToTxID(key *database.Key) (*externalapi.DomainTransactionID, error) {
	serializedTxID := key.Suffix()
	return externalapi.NewDomainTransactionIDFromByteSlice(serializedTxID)
}

func (tis *txIndexStore) stagedAcceptingData() error {
	return nil
}

func (tis *txIndexStore) stagedMergingData() error {
	return nil
}

func (tis *txIndexStore) stagedData() error {
	return nil
}

func (tis *txIndexStore) isAnythingStaged() bool {
	return tis.isAnythingAcceptingStaged() || tis.isAnythingMergingStaged()
}

func (tis *txIndexStore) isAnythingAcceptingStaged() bool {
	return len(tis.toAddAccepted) > 0 || len(tis.toRemoveAccepted) > 0 
}

func (tis *txIndexStore) isAnythingMergingStaged() bool {
	return len(tis.toAddMerge) > 0 || len(tis.ToRemoveMerge) > 0 
}

func (tis *txIndexStore) getTxAcceptingBlockHash(scriptPublicKey *externalapi.ScriptPublicKey) (externalapi.DomainHash, error) {
	if tis.isAnythingAcceptingStaged() {
		return nil, errors.Errorf("cannot get utxo outpoint entry pairs while staging isn't empty")
	}
	return nil, nil
}

func (tis *txIndexStore) getTxMergeBlockHash(scriptPublicKey *externalapi.ScriptPublicKey) (externalapi.DomainHash, error) {
	if tis.isAnythingMergingMergingStaged() {
		return nil, errors.Errorf("cannot get utxo outpoint entry pairs while staging isn't empty")
	}
	return nil, nil
}

func (tis *txIndexStore) getTxBlockHashes(scriptPublicKey *externalapi.ScriptPublicKey) (externalapi.DomainHash, error) {
	if tis.isAnythingStaged() {
		return nil, errors.Errorf("cannot get utxo outpoint entry pairs while staging isn't empty")
	}
	return nil, nil
}

func (tis *txIndexStore) deleteAccepptingData() error {
	return nil
}

func (tis *txIndexStore) deleteMergingData() error {
	return nil
}

func (tis *txIndexStore) deleteAll() error {
	tis.deleteAccepptingData()
	tis.deleteMergingData()
	return nil
}

func (tis *txIndexStore) resetAcceptingData() error {
	return nil
}

func (tis *txIndexStore) resetMergingData() error {
	return nil
}

func (tis *txIndexStore) resetAll() error {
	tis.resetAcceptingData()
	tis.resetMergingData()
	return nil
}
