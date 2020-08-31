package blockdag

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/coinbasepayload"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/util/txsort"
	"github.com/pkg/errors"
)

// The following functions deal with building and validating the coinbase transaction
func (node *blockNode) validateCoinbaseTransaction(dag *BlockDAG, block *util.Block, txsAcceptanceData MultiBlockTxsAcceptanceData) error {
	if node.isGenesis() {
		return nil
	}
	blockCoinbaseTx := block.CoinbaseTransaction().MsgTx()
	_, scriptPubKey, extraData, err := coinbasepayload.DeserializeCoinbasePayload(blockCoinbaseTx)
	if errors.Is(err, coinbasepayload.ErrIncorrectScriptPubKeyLen) {
		return ruleError(ErrBadCoinbaseTransaction, err.Error())
	}
	if err != nil {
		return err
	}
	expectedCoinbaseTransaction, err := node.expectedCoinbaseTransaction(txsAcceptanceData, scriptPubKey, extraData)
	if err != nil {
		return err
	}

	if !expectedCoinbaseTransaction.Hash().IsEqual(block.CoinbaseTransaction().Hash()) {
		return ruleError(ErrBadCoinbaseTransaction, "Coinbase transaction is not built as expected")
	}

	return nil
}

// expectedCoinbaseTransaction returns the coinbase transaction for the current block
func (node *blockNode) expectedCoinbaseTransaction(txsAcceptanceData MultiBlockTxsAcceptanceData, scriptPubKey []byte, extraData []byte) (*util.Tx, error) {
	txIns := []*appmessage.TxIn{}
	txOuts := []*appmessage.TxOut{}

	for _, blue := range node.blues {
		txOut, err := coinbaseOutputForBlueBlock(node.dag, blue, txsAcceptanceData)
		if err != nil {
			return nil, err
		}
		if txOut != nil {
			txOuts = append(txOuts, txOut)
		}
	}
	payload, err := coinbasepayload.SerializeCoinbasePayload(node.blueScore, scriptPubKey, extraData)
	if err != nil {
		return nil, err
	}
	coinbaseTx := appmessage.NewSubnetworkMsgTx(appmessage.TxVersion, txIns, txOuts, subnetworkid.SubnetworkIDCoinbase, 0, payload)
	sortedCoinbaseTx := txsort.Sort(coinbaseTx)
	return util.NewTx(sortedCoinbaseTx), nil
}

// coinbaseOutputForBlueBlock calculates the output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns nil for txOut
func coinbaseOutputForBlueBlock(dag *BlockDAG, blueBlock *blockNode,
	txsAcceptanceData MultiBlockTxsAcceptanceData) (*appmessage.TxOut, error) {

	blockTxsAcceptanceData, ok := txsAcceptanceData.FindAcceptanceData(blueBlock.hash)
	if !ok {
		return nil, errors.Errorf("No txsAcceptanceData for block %s", blueBlock.hash)
	}

	totalFees := uint64(0)

	for _, txAcceptanceData := range blockTxsAcceptanceData.TxAcceptanceData {
		if txAcceptanceData.IsAccepted {
			totalFees += txAcceptanceData.Fee
		}
	}

	totalReward := CalcBlockSubsidy(blueBlock.blueScore, dag.Params) + totalFees

	if totalReward == 0 {
		return nil, nil
	}

	// the ScriptPubKey for the coinbase is parsed from the coinbase payload
	_, scriptPubKey, _, err := coinbasepayload.DeserializeCoinbasePayload(blockTxsAcceptanceData.TxAcceptanceData[0].Tx.MsgTx())
	if err != nil {
		return nil, err
	}

	txOut := &appmessage.TxOut{
		Value:        totalReward,
		ScriptPubKey: scriptPubKey,
	}

	return txOut, nil
}

// NextBlockCoinbaseTransaction prepares the coinbase transaction for the next mined block
//
// This function CAN'T be called with the DAG lock held.
func (dag *BlockDAG) NextBlockCoinbaseTransaction(scriptPubKey []byte, extraData []byte) (*util.Tx, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.NextBlockCoinbaseTransactionNoLock(scriptPubKey, extraData)
}

// NextBlockCoinbaseTransactionNoLock prepares the coinbase transaction for the next mined block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) NextBlockCoinbaseTransactionNoLock(scriptPubKey []byte, extraData []byte) (*util.Tx, error) {
	txsAcceptanceData, err := dag.TxsAcceptedByVirtual()
	if err != nil {
		return nil, err
	}
	return dag.virtual.blockNode.expectedCoinbaseTransaction(txsAcceptanceData, scriptPubKey, extraData)
}
