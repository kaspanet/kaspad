// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"io"

	"github.com/kaspanet/kaspad/app/appmessage"
)

// TxIndexUnknown is the value returned for a transaction index that is unknown.
// This is typically because the transaction has not been inserted into a block
// yet.
const TxIndexUnknown = -1

// Tx defines a kaspa transaction that provides easier and more efficient
// manipulation of raw transactions. It also memoizes the hash for the
// transaction on its first access so subsequent accesses don't have to repeat
// the relatively expensive hashing operations.
type Tx struct {
	msgTx   *appmessage.MsgTx                // Underlying MsgTx
	txHash  *externalapi.DomainHash          // Cached transaction hash
	txID    *externalapi.DomainTransactionID // Cached transaction ID
	txIndex int                              // Position within a block or TxIndexUnknown
}

// MsgTx returns the underlying appmessage.MsgTx for the transaction.
func (t *Tx) MsgTx() *appmessage.MsgTx {
	// Return the cached transaction.
	return t.msgTx
}

// Hash returns the hash of the transaction. This is equivalent to
// calling TxHash on the underlying appmessage.MsgTx, however it caches the
// result so subsequent calls are more efficient.
func (t *Tx) Hash() *externalapi.DomainHash {
	// Return the cached hash if it has already been generated.
	if t.txHash != nil {
		return t.txHash
	}

	// Cache the hash and return it.
	hash := t.msgTx.TxHash()
	t.txHash = hash
	return hash
}

// ID returns the id of the transaction. This is equivalent to
// calling TxID on the underlying appmessage.MsgTx, however it caches the
// result so subsequent calls are more efficient.
func (t *Tx) ID() *externalapi.DomainTransactionID {
	// Return the cached hash if it has already been generated.
	if t.txID != nil {
		return t.txID
	}

	// Cache the hash and return it.
	id := t.msgTx.TxID()
	t.txID = id
	return id
}

// Index returns the saved index of the transaction within a block. This value
// will be TxIndexUnknown if it hasn't already explicitly been set.
func (t *Tx) Index() int {
	return t.txIndex
}

// SetIndex sets the index of the transaction in within a block.
func (t *Tx) SetIndex(index int) {
	t.txIndex = index
}

// IsCoinBase determines whether or not a transaction is a coinbase. A coinbase
// is a special transaction created by miners that has no inputs. This is
// represented in the block dag by a transaction with a single input that has
// a previous output transaction index set to the maximum value along with a
// zero hash.
func (t *Tx) IsCoinBase() bool {
	return t.MsgTx().IsCoinBase()
}

// NewTx returns a new instance of a kaspa transaction given an underlying
// appmessage.MsgTx. See Tx.
func NewTx(msgTx *appmessage.MsgTx) *Tx {
	return &Tx{
		msgTx:   msgTx,
		txIndex: TxIndexUnknown,
	}
}

// NewTxFromBytes returns a new instance of a kaspa transaction given the
// serialized bytes. See Tx.
func NewTxFromBytes(serializedTx []byte) (*Tx, error) {
	br := bytes.NewReader(serializedTx)
	return NewTxFromReader(br)
}

// NewTxFromReader returns a new instance of a kaspa transaction given a
// Reader to deserialize the transaction. See Tx.
func NewTxFromReader(r io.Reader) (*Tx, error) {
	// Deserialize the bytes into a MsgTx.
	var msgTx appmessage.MsgTx
	err := msgTx.Deserialize(r)
	if err != nil {
		return nil, err
	}

	t := Tx{
		msgTx:   &msgTx,
		txIndex: TxIndexUnknown,
	}
	return &t, nil
}
