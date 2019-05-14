package mining

// This file functions are not considered safe for regular use, and should be used for test purposes only.

import (
	"fmt"
	"time"

	"github.com/daglabs/btcd/dagconfig"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

// fakeTxSource is a simple implementation of TxSource interface
type fakeTxSource struct {
	txDescs []*TxDesc
}

func (txs *fakeTxSource) LastUpdated() time.Time {
	return time.Unix(0, 0)
}

func (txs *fakeTxSource) MiningDescs() []*TxDesc {
	return txs.txDescs
}

func (txs *fakeTxSource) HaveTransaction(txID *daghash.TxID) bool {
	for _, desc := range txs.txDescs {
		if *desc.Tx.ID() == *txID {
			return true
		}
	}
	return false
}

// PrepareBlockForTest generates a block with the proper merkle roots, coinbase/fee transactions etc. This function is used for test purposes only
func PrepareBlockForTest(dag *blockdag.BlockDAG, params *dagconfig.Params, parentHashes []*daghash.Hash, transactions []*wire.MsgTx, forceTransactions bool, coinbaseOutputs uint64) (*wire.MsgBlock, error) {
	newVirtual, err := blockdag.GetVirtualFromParentsForTest(dag, parentHashes)
	if err != nil {
		return nil, err
	}
	oldVirtual := blockdag.SetVirtualForTest(dag, newVirtual)
	defer blockdag.SetVirtualForTest(dag, oldVirtual)
	policy := Policy{
		BlockMaxSize:      50000,
		BlockPrioritySize: 750000,
		TxMinFreeFee:      util.Amount(0),
	}

	txSource := &fakeTxSource{
		txDescs: make([]*TxDesc, len(transactions)),
	}

	for i, tx := range transactions {
		txSource.txDescs[i] = &TxDesc{
			Tx: util.NewTx(tx),
		}
	}

	blockTemplateGenerator := NewBlkTmplGenerator(&policy,
		params, txSource, dag, blockdag.NewMedianTime(), txscript.NewSigCache(100000))

	template, err := blockTemplateGenerator.NewBlockTemplate(nil)
	if err != nil {
		return nil, err
	}

	// In order of creating deterministic coinbase tx ids.
	blockTemplateGenerator.UpdateExtraNonce(template.Block, dag.Height()+1, GenerateDeterministicExtraNonceForTest())

	txsToAdd := make([]*wire.MsgTx, 0)
	for _, tx := range transactions {
		found := false
		for _, blockTx := range template.Block.Transactions {
			if blockTx.TxHash().IsEqual(tx.TxHash()) {
				found = true
				break
			}
		}
		if !found {
			if !forceTransactions {
				return nil, fmt.Errorf("tx %s wasn't found in the block", tx.TxHash())
			}
			txsToAdd = append(txsToAdd, tx)
		}
	}
	if coinbaseOutputs != 1 {
		cb := template.Block.Transactions[0]
		originalValue := cb.TxOut[0].Value
		pkScript := cb.TxOut[0].PkScript
		cb.TxOut = make([]*wire.TxOut, coinbaseOutputs)
		if coinbaseOutputs != 0 {
			newOutValue := originalValue / coinbaseOutputs
			for i := uint64(0); i < coinbaseOutputs; i++ {
				cb.TxOut[i] = &wire.TxOut{
					Value:    newOutValue,
					PkScript: pkScript,
				}
			}
		}
	}
	if forceTransactions && len(txsToAdd) > 0 {
		for _, tx := range txsToAdd {
			template.Block.Transactions = append(template.Block.Transactions, tx)
		}
	}
	updateMerkleRoots := coinbaseOutputs != 1 || (forceTransactions && len(txsToAdd) > 0)
	if updateMerkleRoots {
		utilTxs := make([]*util.Tx, len(template.Block.Transactions))
		for i, tx := range template.Block.Transactions {
			utilTxs[i] = util.NewTx(tx)
		}
		template.Block.Header.HashMerkleRoot = blockdag.BuildHashMerkleTreeStore(utilTxs).Root()
		template.Block.Header.IDMerkleRoot = blockdag.BuildIDMerkleTreeStore(utilTxs).Root()
	}
	return template.Block, nil
}

// GenerateDeterministicExtraNonceForTest returns a unique deterministic extra nonce for coinbase data, in order to create unique coinbase transactions.
func GenerateDeterministicExtraNonceForTest() uint64 {
	extraNonceForTest++
	return extraNonceForTest
}

var extraNonceForTest = uint64(0)
