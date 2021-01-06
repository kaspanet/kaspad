package mempool

import (
	"math"

	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

const unacceptedBlueScore = math.MaxUint64

func newMempoolUTXOSet() *mempoolUTXOSet {
	return &mempoolUTXOSet{
		transactionByPreviousOutpoint: make(map[consensusexternalapi.DomainOutpoint]*consensusexternalapi.DomainTransaction),
		poolUnspentOutputs:            make(map[consensusexternalapi.DomainOutpoint]consensusexternalapi.UTXOEntry),
	}
}

type mempoolUTXOSet struct {
	transactionByPreviousOutpoint map[consensusexternalapi.DomainOutpoint]*consensusexternalapi.DomainTransaction
	poolUnspentOutputs            map[consensusexternalapi.DomainOutpoint]consensusexternalapi.UTXOEntry
}

// Populate UTXO Entries in the transaction, to allow chained txs.
func (mpus *mempoolUTXOSet) populateUTXOEntries(tx *consensusexternalapi.DomainTransaction) (parentsInPool []consensusexternalapi.DomainOutpoint) {
	for _, txIn := range tx.Inputs {
		if utxoEntry, exists := mpus.poolUnspentOutputs[txIn.PreviousOutpoint]; exists {
			txIn.UTXOEntry = utxoEntry
			parentsInPool = append(parentsInPool, txIn.PreviousOutpoint)
		}
	}
	return parentsInPool
}

func (mpus *mempoolUTXOSet) checkExists(tx *consensusexternalapi.DomainTransaction) bool {
	// Check if it was already spent.
	for _, txIn := range tx.Inputs {
		if _, exists := mpus.transactionByPreviousOutpoint[txIn.PreviousOutpoint]; exists {
			return true
		}
	}

	// Check if it creates an already existing UTXO
	outpoint := consensusexternalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(tx)}
	for i := range tx.Outputs {
		outpoint.Index = uint32(i)
		if _, exists := mpus.poolUnspentOutputs[outpoint]; exists {
			return true
		}
	}
	return false
}

// addTx adds a transaction to the mempool UTXO set. It assumes that it doesn't double spend another transaction
// in the mempool, and that its outputs doesn't exist in the mempool UTXO set, and returns error otherwise.
func (mpus *mempoolUTXOSet) addTx(tx *consensusexternalapi.DomainTransaction) error {
	for _, txIn := range tx.Inputs {
		if existingTx, exists := mpus.transactionByPreviousOutpoint[txIn.PreviousOutpoint]; exists {
			return errors.Errorf("outpoint %s is already used by %s", txIn.PreviousOutpoint, consensushashing.TransactionID(existingTx))
		}
		mpus.transactionByPreviousOutpoint[txIn.PreviousOutpoint] = tx
	}

	for i, txOut := range tx.Outputs {
		outpoint := consensusexternalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(tx), Index: uint32(i)}
		if _, exists := mpus.poolUnspentOutputs[outpoint]; exists {
			return errors.Errorf("outpoint %s already exists", outpoint)
		}
		mpus.poolUnspentOutputs[outpoint] =
			utxo.NewUTXOEntry(txOut.Value, txOut.ScriptPublicKey, false, unacceptedBlueScore)
	}
	return nil
}

// removeTx removes a transaction to the mempool UTXO set.
// Note: it doesn't re-add its previous outputs to the mempool UTXO set.
func (mpus *mempoolUTXOSet) removeTx(tx *consensusexternalapi.DomainTransaction) error {
	for _, txIn := range tx.Inputs {
		if _, exists := mpus.transactionByPreviousOutpoint[txIn.PreviousOutpoint]; !exists {
			return errors.Errorf("outpoint %s doesn't exist", txIn.PreviousOutpoint)
		}
		delete(mpus.transactionByPreviousOutpoint, txIn.PreviousOutpoint)
	}

	outpoint := consensusexternalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(tx)}
	for i := range tx.Outputs {
		outpoint.Index = uint32(i)
		if _, exists := mpus.poolUnspentOutputs[outpoint]; !exists {
			return errors.Errorf("outpoint %s doesn't exist", outpoint)
		}
		delete(mpus.poolUnspentOutputs, outpoint)
	}

	return nil
}

func (mpus *mempoolUTXOSet) poolTransactionBySpendingOutpoint(outpoint consensusexternalapi.DomainOutpoint) (*consensusexternalapi.DomainTransaction, bool) {
	tx, exists := mpus.transactionByPreviousOutpoint[outpoint]
	return tx, exists
}
