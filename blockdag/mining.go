package blockdag

import (
	"bytes"
	"encoding/binary"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/wire"
	"time"
)

// BlockForMining returns a block with the given transactions
// that points to the current DAG tips, that is valid from
// all aspects except proof of work.
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

	utxoWithTransactions, err := dag.UTXOSet().WithTransactions(msgBlock.Transactions, UnacceptedBlueScore, false)
	if err != nil {
		return nil, err
	}
	utxoCommitment := utxoWithTransactions.Multiset().Hash()

	msgBlock.Header = wire.BlockHeader{
		Version:              nextBlockVersion,
		ParentHashes:         dag.TipHashes(),
		HashMerkleRoot:       hashMerkleTree.Root(),
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       utxoCommitment,
		Timestamp:            blockTimestamp,
		Bits:                 requiredDifficulty,
	}

	return &msgBlock, nil
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
func (dag *BlockDAG) NextBlockMinimumTime() time.Time {
	return dag.CalcPastMedianTime().Add(time.Second)
}

// NextBlockTime returns a valid block time for the
// next block that will point to the existing DAG tips.
func (dag *BlockDAG) NextBlockTime() time.Time {
	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks. Thus, choose the maximum between the
	// current time and one second after the past median time. The current
	// timestamp is truncated to a second boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// second.
	newTimestamp := dag.Now()
	minTimestamp := dag.NextBlockMinimumTime()
	if newTimestamp.Before(minTimestamp) {
		newTimestamp = minTimestamp
	}

	return newTimestamp
}
