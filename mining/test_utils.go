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

// PrepareBlockForTest generates a block with the proper merkle roots, coinbase transaction etc. This function is used for test purposes only
func PrepareBlockForTest(dag *blockdag.BlockDAG, params *dagconfig.Params, parentHashes []*daghash.Hash, transactions []*wire.MsgTx, forceTransactions bool) (*wire.MsgBlock, error) {
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

	OpTrueAddr, err := OpTrueAddress(params.Prefix)
	if err != nil {
		return nil, err
	}

	template, err := blockTemplateGenerator.NewBlockTemplate(OpTrueAddr)
	if err != nil {
		return nil, err
	}

	// In order of creating deterministic coinbase tx ids.
	err = blockTemplateGenerator.UpdateExtraNonce(template.Block, dag.Height()+1, GenerateDeterministicExtraNonceForTest())
	if err != nil {
		return nil, err
	}

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
	if forceTransactions && len(txsToAdd) > 0 {
		for _, tx := range txsToAdd {
			template.Block.Transactions = append(template.Block.Transactions, tx)
		}
	}
	updateHeaderFields := forceTransactions && len(txsToAdd) > 0
	if updateHeaderFields {
		utilTxs := make([]*util.Tx, len(template.Block.Transactions))
		for i, tx := range template.Block.Transactions {
			utilTxs[i] = util.NewTx(tx)
		}
		template.Block.Header.HashMerkleRoot = blockdag.BuildHashMerkleTreeStore(utilTxs).Root()

		template.Block.Header.UTXOCommitment, err = blockTemplateGenerator.buildUTXOCommitment(template.Block.Transactions, dag.Height()+1)
		if err != nil {
			return nil, err
		}
	}
	return template.Block, nil
}

// GenerateDeterministicExtraNonceForTest returns a unique deterministic extra nonce for coinbase data, in order to create unique coinbase transactions.
func GenerateDeterministicExtraNonceForTest() uint64 {
	extraNonceForTest++
	return extraNonceForTest
}

func OpTrueAddress(prefix util.Bech32Prefix) (util.Address, error) {
	return util.NewAddressScriptHash(blockdag.OpTrueScript, prefix)
}

var extraNonceForTest = uint64(0)
