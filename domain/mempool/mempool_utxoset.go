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

func (mpus *mempoolUTXOSet) utxoEntryByOutpoint(outpoint appmessage.Outpoint) (entry *blockdag.UTXOEntry, isInPool bool, exists bool) {
	entry, exists = mpus.dag.GetUTXOEntry(outpoint)
	if !exists {
		entry, exists := mpus.poolUnspentOutputs[outpoint]
		if !exists {
			return nil, false, false
		}
		return entry, true, true
	}
	return entry, false, true
}

// addTx adds a transaction to the mempool UTXO set. It assumes that it doesn't double spend another transaction
// in the mempool, and that its outputs doesn't exist in the mempool UTXO set, and returns error otherwise.
func (mpus *mempoolUTXOSet) addTx(tx *util.Tx) error {
	msgTx := tx.MsgTx()
	for _, txIn := range msgTx.TxIn {
		if existingTx, exists := mpus.transactionByPreviousOutpoint[txIn.PreviousOutpoint]; exists {
			return errors.Errorf("outpoint %s is already used by %s", txIn.PreviousOutpoint, existingTx.ID())
		}
		mpus.transactionByPreviousOutpoint[txIn.PreviousOutpoint] = tx
	}

	for i, txOut := range msgTx.TxOut {
		outpoint := appmessage.NewOutpoint(tx.ID(), uint32(i))
		if _, exists := mpus.poolUnspentOutputs[*outpoint]; exists {
			return errors.Errorf("outpoint %s already exists", outpoint)
		}
		mpus.poolUnspentOutputs[*outpoint] = blockdag.NewUTXOEntry(txOut, false, blockdag.UnacceptedBlueScore)
	}
	return nil
}

// removeTx removes a transaction to the mempool UTXO set.
// Note: it doesn't re-add its previous outputs to the mempool UTXO set.
func (mpus *mempoolUTXOSet) removeTx(tx *util.Tx) error {
	msgTx := tx.MsgTx()
	for _, txIn := range msgTx.TxIn {
		if _, exists := mpus.transactionByPreviousOutpoint[txIn.PreviousOutpoint]; !exists {
			return errors.Errorf("outpoint %s doesn't exist", txIn.PreviousOutpoint)
		}
		delete(mpus.transactionByPreviousOutpoint, txIn.PreviousOutpoint)
	}

	for i := range msgTx.TxOut {
		outpoint := appmessage.NewOutpoint(tx.ID(), uint32(i))
		if _, exists := mpus.poolUnspentOutputs[*outpoint]; !exists {
			return errors.Errorf("outpoint %s doesn't exist", outpoint)
		}
		delete(mpus.poolUnspentOutputs, *outpoint)
	}

	return nil
}

func (mpus *mempoolUTXOSet) poolTransactionBySpendingOutpoint(outpoint appmessage.Outpoint) (*util.Tx, bool) {
	tx, exists := mpus.transactionByPreviousOutpoint[outpoint]
	return tx, exists
}

func (mpus *mempoolUTXOSet) transactionRelatedUTXOEntries(tx *util.Tx) (spentUTXOEntries []*blockdag.UTXOEntry, parentsInPool []*appmessage.Outpoint, missingParents []*daghash.TxID) {
	msgTx := tx.MsgTx()
	spentUTXOEntries = make([]*blockdag.UTXOEntry, len(msgTx.TxIn))
	missingParents = make([]*daghash.TxID, 0)
	parentsInPool = make([]*appmessage.Outpoint, 0)

	isOrphan := false
	for i, txIn := range msgTx.TxIn {
		entry, isInPool, exists := mpus.utxoEntryByOutpoint(txIn.PreviousOutpoint)
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
