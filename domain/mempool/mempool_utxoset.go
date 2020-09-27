package mempool

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

func newMempoolUTXOSet(dag *blockdag.BlockDAG) *mempoolUTXOSet {
	return &mempoolUTXOSet{
		transactionByPreviousOutpoint: make(map[appmessage.Outpoint]*util.Tx),
		poolUnspentOutputs:            make(map[appmessage.Outpoint]*blockdag.UTXOEntry),
		dag:                           dag,
	}
}

type mempoolUTXOSet struct {
	transactionByPreviousOutpoint map[appmessage.Outpoint]*util.Tx
	poolUnspentOutputs            map[appmessage.Outpoint]*blockdag.UTXOEntry
	dag                           *blockdag.BlockDAG
}

func (mpSet *mempoolUTXOSet) utxoEntryByOutpoint(outpoint appmessage.Outpoint) (entry *blockdag.UTXOEntry, isInPool bool, exists bool) {
	entry, exists = mpSet.dag.GetUTXOEntry(outpoint)
	if !exists {
		entry, exists := mpSet.poolUnspentOutputs[outpoint]
		if !exists {
			return nil, false, false
		}
		return entry, true, true
	}
	return entry, false, true
}

// addTx adds a transaction to the mempool UTXO set. It assumes that it doesn't double spend another transaction
// in the mempool, and that its outputs doesn't exist in the mempool UTXO set, and returns error otherwise.
func (mpSet *mempoolUTXOSet) addTx(tx *util.Tx) error {
	msgTx := tx.MsgTx()
	for _, txIn := range msgTx.TxIn {
		if existingTx, exists := mpSet.transactionByPreviousOutpoint[txIn.PreviousOutpoint]; exists {
			return errors.Errorf("outpoint %s is already used by %s", txIn.PreviousOutpoint, existingTx.ID())
		}
		mpSet.transactionByPreviousOutpoint[txIn.PreviousOutpoint] = tx
	}

	for i, txOut := range msgTx.TxOut {
		outpoint := appmessage.NewOutpoint(tx.ID(), uint32(i))
		if _, exists := mpSet.poolUnspentOutputs[*outpoint]; exists {
			return errors.Errorf("outpoint %s already exists", outpoint)
		}
		mpSet.poolUnspentOutputs[*outpoint] = blockdag.NewUTXOEntry(txOut, false, blockdag.UnacceptedBlueScore)
	}
	return nil
}

// removeTx removes a transaction to the mempool UTXO set.
// Note: it doesn't re-add its previous outputs to the mempool UTXO set.
func (mpSet *mempoolUTXOSet) removeTx(tx *util.Tx) error {
	msgTx := tx.MsgTx()
	for _, txIn := range msgTx.TxIn {
		if _, exists := mpSet.transactionByPreviousOutpoint[txIn.PreviousOutpoint]; !exists {
			return errors.Errorf("outpoint %s doesn't exist", txIn.PreviousOutpoint)
		}
		delete(mpSet.transactionByPreviousOutpoint, txIn.PreviousOutpoint)
	}

	for i := range msgTx.TxOut {
		outpoint := appmessage.NewOutpoint(tx.ID(), uint32(i))
		if _, exists := mpSet.poolUnspentOutputs[*outpoint]; !exists {
			return errors.Errorf("outpoint %s doesn't exist", outpoint)
		}
		delete(mpSet.poolUnspentOutputs, *outpoint)
	}

	return nil
}

func (mpSet *mempoolUTXOSet) poolTransactionBySpendingOutpoint(outpoint appmessage.Outpoint) (*util.Tx, bool) {
	tx, exists := mpSet.transactionByPreviousOutpoint[outpoint]
	return tx, exists
}

func (mp *mempoolUTXOSet) transactionRelatedUTXOEntries(tx *util.Tx) (spentUTXOEntries []*blockdag.UTXOEntry, parentsInPool []*appmessage.Outpoint, missingParents []*daghash.TxID) {
	msgTx := tx.MsgTx()
	spentUTXOEntries = make([]*blockdag.UTXOEntry, len(msgTx.TxIn))
	missingParents = make([]*daghash.TxID, 0)
	parentsInPool = make([]*appmessage.Outpoint, 0)

	isOrphan := false
	for i, txIn := range msgTx.TxIn {
		entry, isInPool, exists := mp.utxoEntryByOutpoint(txIn.PreviousOutpoint)
		if !exists {
			isOrphan = true
			missingParents = append(missingParents, &txIn.PreviousOutpoint.TxID)
		}

		if isOrphan {
			continue
		}

		if isInPool {
			parentsInPool = append(parentsInPool, &txIn.PreviousOutpoint)
		}

		spentUTXOEntries[i] = entry
	}

	if isOrphan {
		return nil, nil, missingParents
	}

	return spentUTXOEntries, parentsInPool, nil
}
