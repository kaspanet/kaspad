package blockdag

import (
	"bytes"
	"encoding/binary"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/wire"
	"time"
)

// BlockForMining returns a block with the given transactions
// that points to the current DAG tips, that is valid from
// all aspects except proof of work.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) BlockForMining(transactions []*util.Tx) (*wire.MsgBlock, error) {
	blockTimestamp := dag.NextBlockTime()
	requiredDifficulty := dag.NextRequiredDifficulty(blockTimestamp)

	// Calculate the next expected block version based on the state of the
	// rule change deployments.
	nextBlockVersion, err := dag.CalcNextBlockVersion()
	if err != nil {
		return nil, err
	}

	// Create a new block ready to be solved.
	hashMerkleTree := BuildHashMerkleTreeStore(transactions)
	acceptedIDMerkleRoot, err := dag.NextAcceptedIDMerkleRootNoLock()
	if err != nil {
		return nil, err
	}
	var msgBlock wire.MsgBlock
	for _, tx := range transactions {
		msgBlock.AddTransaction(tx.MsgTx())
	}

	multiset, err := dag.NextBlockMultiset()
	if err != nil {
		return nil, err
	}

	msgBlock.Header = wire.BlockHeader{
		Version:              nextBlockVersion,
		ParentHashes:         dag.TipHashes(),
		HashMerkleRoot:       hashMerkleTree.Root(),
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       (*daghash.Hash)(multiset.Finalize()),
		Timestamp:            blockTimestamp,
		Bits:                 requiredDifficulty,
	}

	return &msgBlock, nil
}

// NextBlockMultiset returns the multiset of an assumed next block
// built on top of the current tips.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) NextBlockMultiset() (*secp256k1.MultiSet, error) {
	_, selectedParentPastUTXO, txsAcceptanceData, err := dag.pastUTXO(&dag.virtual.blockNode)
	if err != nil {
		return nil, err
	}

	return dag.virtual.blockNode.calcMultiset(dag, txsAcceptanceData, selectedParentPastUTXO)
}

// CoinbasePayloadExtraData returns coinbase payload extra data parameter
// which is built from extra nonce and coinbase flags.
func CoinbasePayloadExtraData(extraNonce uint64, coinbaseFlags string) ([]byte, error) {
	extraNonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(extraNonceBytes, extraNonce)
	w := &bytes.Buffer{}
	_, err := w.Write(extraNonceBytes)
	if err != nil {
		return nil, err
	}
	_, err = w.Write([]byte(coinbaseFlags))
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// NextCoinbaseFromAddress returns a coinbase transaction for the
// next block with the given address and extra data in its payload.
func (dag *BlockDAG) NextCoinbaseFromAddress(payToAddress util.Address, extraData []byte) (*util.Tx, error) {
	coinbasePayloadScriptPubKey, err := txscript.PayToAddrScript(payToAddress)
	if err != nil {
		return nil, err
	}
	coinbaseTx, err := dag.NextBlockCoinbaseTransactionNoLock(coinbasePayloadScriptPubKey, extraData)
	if err != nil {
		return nil, err
	}
	return coinbaseTx, nil
}

// NextBlockMinimumTime returns the minimum allowed timestamp for a block building
// on the end of the DAG. In particular, it is one second after
// the median timestamp of the last several blocks per the DAG consensus
// rules.
func (dag *BlockDAG) NextBlockMinimumTime() mstime.Time {
	return dag.CalcPastMedianTime().Add(time.Second)
}

// NextBlockTime returns a valid block time for the
// next block that will point to the existing DAG tips.
func (dag *BlockDAG) NextBlockTime() mstime.Time {
	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks. Thus, choose the maximum between the
	// current time and one second after the past median time. The current
	// timestamp is truncated to a millisecond boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// millisecond.
	newTimestamp := dag.Now()
	minTimestamp := dag.NextBlockMinimumTime()
	if newTimestamp.Before(minTimestamp) {
		newTimestamp = minTimestamp
	}

	return newTimestamp
}
