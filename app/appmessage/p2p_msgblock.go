// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// defaultTransactionAlloc is the default size used for the backing array
// for transactions. The transaction array will dynamically grow as needed, but
// this figure is intended to provide enough space for the number of
// transactions in the vast majority of blocks without needing to grow the
// backing array multiple times.
const defaultTransactionAlloc = 2048

// MaxMassAcceptedByBlock is the maximum total transaction mass a block may accept.
const MaxMassAcceptedByBlock = 10000000

// MaxMassPerTx is the maximum total mass a transaction may have.
const MaxMassPerTx = MaxMassAcceptedByBlock / 2

// MaxTxPerBlock is the maximum number of transactions that could
// possibly fit into a block.
const MaxTxPerBlock = (MaxMassAcceptedByBlock / minTxPayload) + 1

// MaxBlockParents is the maximum allowed number of parents for block.
const MaxBlockParents = 10

// TxLoc holds locator data for the offset and length of where a transaction is
// located within a MsgBlock data buffer.
type TxLoc struct {
	TxStart int
	TxLen   int
}

// MsgBlock implements the Message interface and represents a kaspa
// block message. It is used to deliver block and transaction information in
// response to a getdata message (MsgGetData) for a given block hash.
type MsgBlock struct {
	baseMessage
	Header       MsgBlockHeader
	Transactions []*MsgTx
}

// AddTransaction adds a transaction to the message.
func (msg *MsgBlock) AddTransaction(tx *MsgTx) {
	msg.Transactions = append(msg.Transactions, tx)
}

// ClearTransactions removes all transactions from the message.
func (msg *MsgBlock) ClearTransactions() {
	msg.Transactions = make([]*MsgTx, 0, defaultTransactionAlloc)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgBlock) Command() MessageCommand {
	return CmdBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgBlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxMessagePayload
}

// ConvertToPartial clears out all the payloads of the subnetworks that are
// incompatible with the given subnetwork ID.
// Note: this operation modifies the block in place.
func (msg *MsgBlock) ConvertToPartial(subnetworkID *externalapi.DomainSubnetworkID) {
	for _, tx := range msg.Transactions {
		if !tx.SubnetworkID.Equal(subnetworkID) {
			tx.Payload = []byte{}
		}
	}
}

// NewMsgBlock returns a new kaspa block message that conforms to the
// Message interface. See MsgBlock for details.
func NewMsgBlock(blockHeader *MsgBlockHeader) *MsgBlock {
	return &MsgBlock{
		Header:       *blockHeader,
		Transactions: make([]*MsgTx, 0, defaultTransactionAlloc),
	}
}
