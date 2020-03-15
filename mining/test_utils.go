package mining

// This file functions are not considered safe for regular use, and should be used for test purposes only.

import (
	"github.com/pkg/errors"
	"time"

	"github.com/kaspanet/kaspad/dagconfig"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
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
		BlockMaxMass: 50000,
	}

	txSource := &fakeTxSource{
		txDescs: make([]*TxDesc, len(transactions)),
	}

	for i, tx := range transactions {
		txSource.txDescs[i] = &TxDesc{
			Tx:  util.NewTx(tx),
			Fee: 1,
		}
	}

	blockTemplateGenerator := NewBlkTmplGenerator(&policy,
		params, txSource, dag, blockdag.NewTimeSource(), txscript.NewSigCache(100000))

	OpTrueAddr, err := OpTrueAddress(params.Prefix)
	if err != nil {
		return nil, err
	}

	template, err := blockTemplateGenerator.NewBlockTemplate(OpTrueAddr)
	if err != nil {
		return nil, err
	}

	// In order of creating deterministic coinbase tx ids.
	err = blockTemplateGenerator.UpdateExtraNonce(template.Block, GenerateDeterministicExtraNonceForTest())
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
				return nil, errors.Errorf("tx %s wasn't found in the block", tx.TxHash())
			}
			txsToAdd = append(txsToAdd, tx)
		}
	}
	if forceTransactions && len(txsToAdd) > 0 {
		template.Block.Transactions = append(template.Block.Transactions, txsToAdd...)
	}
	updateHeaderFields := forceTransactions && len(txsToAdd) > 0
	if updateHeaderFields {
		utilTxs := make([]*util.Tx, len(template.Block.Transactions))
		for i, tx := range template.Block.Transactions {
			utilTxs[i] = util.NewTx(tx)
		}
		template.Block.Header.HashMerkleRoot = blockdag.BuildHashMerkleTreeStore(utilTxs).Root()

		template.Block.Header.UTXOCommitment, err = blockTemplateGenerator.buildUTXOCommitment(template.Block.Transactions)
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

// OpTrueAddress returns an address pointing to a P2SH anyone-can-spend script
func OpTrueAddress(prefix util.Bech32Prefix) (util.Address, error) {
	return util.NewAddressScriptHash(blockdag.OpTrueScript, prefix)
}

var extraNonceForTest = uint64(0)
