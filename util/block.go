// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/coinbasepayload"
	"github.com/kaspanet/kaspad/util/mstime"
	"io"

	"github.com/kaspanet/kaspad/app/appmessage"
)

// OutOfRangeError describes an error due to accessing an element that is out
// of range.
type OutOfRangeError string

const (
	// CoinbaseTransactionIndex is the index of the coinbase transaction in every block
	CoinbaseTransactionIndex = 0
)

// Error satisfies the error interface and prints human-readable errors.
func (e OutOfRangeError) Error() string {
	return string(e)
}

// Block defines a kaspa block that provides easier and more efficient
// manipulation of raw blocks. It also memoizes hashes for the block and its
// transactions on their first access so subsequent accesses don't have to
// repeat the relatively expensive hashing operations.
type Block struct {
	// Underlying MsgBlock
	msgBlock *appmessage.MsgBlock

	// Serialized bytes for the block. This is used only internally, and .Hash() should be used anywhere.
	serializedBlock []byte

	// Cached block hash. This is used only internally, and .Hash() should be used anywhere.
	blockHash *externalapi.DomainHash

	// Transactions. This is used only internally, and .Transactions() should be used anywhere.
	transactions []*Tx

	// ALL wrapped transactions generated
	txnsGenerated bool

	// Blue score. This is used only internally, and .BlueScore() should be used anywhere.
	blueScore *uint64
}

// MsgBlock returns the underlying appmessage.MsgBlock for the Block.
func (b *Block) MsgBlock() *appmessage.MsgBlock {
	// Return the cached block.
	return b.msgBlock
}

// Bytes returns the serialized bytes for the Block. This is equivalent to
// calling Serialize on the underlying appmessage.MsgBlock, however it caches the
// result so subsequent calls are more efficient.
func (b *Block) Bytes() ([]byte, error) {
	// Return the cached serialized bytes if it has already been generated.
	if len(b.serializedBlock) != 0 {
		return b.serializedBlock, nil
	}

	// Serialize the MsgBlock.
	w := bytes.NewBuffer(make([]byte, 0, b.msgBlock.SerializeSize()))
	err := b.msgBlock.Serialize(w)
	if err != nil {
		return nil, err
	}
	serializedBlock := w.Bytes()

	// Cache the serialized bytes and return them.
	b.serializedBlock = serializedBlock
	return serializedBlock, nil
}

// Hash returns the block identifier hash for the Block. This is equivalent to
// calling BlockHash on the underlying appmessage.MsgBlock, however it caches the
// result so subsequent calls are more efficient.
func (b *Block) Hash() *externalapi.DomainHash {
	// Return the cached block hash if it has already been generated.
	if b.blockHash != nil {
		return b.blockHash
	}

	// Cache the block hash and return it.
	hash := b.msgBlock.BlockHash()
	b.blockHash = hash
	return hash
}

// Tx returns a wrapped transaction (util.Tx) for the transaction at the
// specified index in the Block. The supplied index is 0 based. That is to
// say, the first transaction in the block is txNum 0. This is nearly
// equivalent to accessing the raw transaction (appmessage.MsgTx) from the
// underlying appmessage.MsgBlock, however the wrapped transaction has some helpful
// properties such as caching the hash so subsequent calls are more efficient.
func (b *Block) Tx(txNum int) (*Tx, error) {
	// Ensure the requested transaction is in range.
	numTx := uint64(len(b.msgBlock.Transactions))
	if txNum < 0 || uint64(txNum) > numTx {
		str := fmt.Sprintf("transaction index %d is out of range - max %d",
			txNum, numTx-1)
		return nil, OutOfRangeError(str)
	}

	// Generate slice to hold all of the wrapped transactions if needed.
	if len(b.transactions) == 0 {
		b.transactions = make([]*Tx, numTx)
	}

	// Return the wrapped transaction if it has already been generated.
	if b.transactions[txNum] != nil {
		return b.transactions[txNum], nil
	}

	// Generate and cache the wrapped transaction and return it.
	newTx := NewTx(b.msgBlock.Transactions[txNum])
	newTx.SetIndex(txNum)
	b.transactions[txNum] = newTx
	return newTx, nil
}

// Transactions returns a slice of wrapped transactions (util.Tx) for all
// transactions in the Block. This is nearly equivalent to accessing the raw
// transactions (appmessage.MsgTx) in the underlying appmessage.MsgBlock, however it
// instead provides easy access to wrapped versions (util.Tx) of them.
func (b *Block) Transactions() []*Tx {
	// Return transactions if they have ALL already been generated. This
	// flag is necessary because the wrapped transactions are lazily
	// generated in a sparse fashion.
	if b.txnsGenerated {
		return b.transactions
	}

	// Generate slice to hold all of the wrapped transactions if needed.
	if len(b.transactions) == 0 {
		b.transactions = make([]*Tx, len(b.msgBlock.Transactions))
	}

	// Generate and cache the wrapped transactions for all that haven't
	// already been done.
	for i, tx := range b.transactions {
		if tx == nil {
			newTx := NewTx(b.msgBlock.Transactions[i])
			newTx.SetIndex(i)
			b.transactions[i] = newTx
		}
	}

	b.txnsGenerated = true
	return b.transactions
}

// TxHash returns the hash for the requested transaction number in the Block.
// The supplied index is 0 based. That is to say, the first transaction in the
// block is txNum 0. This is equivalent to calling TxHash on the underlying
// appmessage.MsgTx, however it caches the result so subsequent calls are more
// efficient.
func (b *Block) TxHash(txNum int) (*externalapi.DomainHash, error) {
	// Attempt to get a wrapped transaction for the specified index. It
	// will be created lazily if needed or simply return the cached version
	// if it has already been generated.
	tx, err := b.Tx(txNum)
	if err != nil {
		return nil, err
	}

	// Defer to the wrapped transaction which will return the cached hash if
	// it has already been generated.
	return tx.Hash(), nil
}

// TxLoc returns the offsets and lengths of each transaction in a raw block.
// It is used to allow fast indexing into transactions within the raw byte
// stream.
func (b *Block) TxLoc() ([]appmessage.TxLoc, error) {
	rawMsg, err := b.Bytes()
	if err != nil {
		return nil, err
	}
	rbuf := bytes.NewBuffer(rawMsg)

	var mblock appmessage.MsgBlock
	txLocs, err := mblock.DeserializeTxLoc(rbuf)
	if err != nil {
		return nil, err
	}
	return txLocs, err
}

// IsGenesis returns whether or not this block is the genesis block.
func (b *Block) IsGenesis() bool {
	return b.MsgBlock().Header.IsGenesis()
}

// CoinbaseTransaction returns this block's coinbase transaction
func (b *Block) CoinbaseTransaction() *Tx {
	return b.Transactions()[CoinbaseTransactionIndex]
}

// Timestamp returns this block's timestamp
func (b *Block) Timestamp() mstime.Time {
	return b.msgBlock.Header.Timestamp
}

// BlueScore returns this block's blue score.
func (b *Block) BlueScore() (uint64, error) {
	if b.blueScore == nil {
		blueScore, _, _, err := coinbasepayload.DeserializeCoinbasePayload(b.CoinbaseTransaction().MsgTx())
		if err != nil {
			return 0, err
		}
		b.blueScore = &blueScore
	}
	return *b.blueScore, nil
}

// NewBlock returns a new instance of a kaspa block given an underlying
// appmessage.MsgBlock. See Block.
func NewBlock(msgBlock *appmessage.MsgBlock) *Block {
	return &Block{
		msgBlock: msgBlock,
	}
}

// NewBlockFromBytes returns a new instance of a kaspa block given the
// serialized bytes. See Block.
func NewBlockFromBytes(serializedBlock []byte) (*Block, error) {
	br := bytes.NewReader(serializedBlock)
	b, err := NewBlockFromReader(br)
	if err != nil {
		return nil, err
	}
	b.serializedBlock = serializedBlock
	return b, nil
}

// NewBlockFromReader returns a new instance of a kaspa block given a
// Reader to deserialize the block. See Block.
func NewBlockFromReader(r io.Reader) (*Block, error) {
	// Deserialize the bytes into a MsgBlock.
	var msgBlock appmessage.MsgBlock
	err := msgBlock.Deserialize(r)
	if err != nil {
		return nil, err
	}

	b := Block{
		msgBlock: &msgBlock,
	}
	return &b, nil
}

// NewBlockFromBlockAndBytes returns a new instance of a kaspa block given
// an underlying appmessage.MsgBlock and the serialized bytes for it. See Block.
func NewBlockFromBlockAndBytes(msgBlock *appmessage.MsgBlock, serializedBlock []byte) *Block {
	return &Block{
		msgBlock:        msgBlock,
		serializedBlock: serializedBlock,
	}
}
