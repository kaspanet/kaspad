// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"bytes"
	"fmt"
	"io"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/util/daghash"
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
	Header       BlockHeader
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

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding blocks stored to disk, such as in a database, as
// opposed to decoding blocks from the appmessage.
func (msg *MsgBlock) KaspaDecode(r io.Reader, pver uint32) error {
	err := readBlockHeader(r, pver, &msg.Header)
	if err != nil {
		return err
	}

	txCount, err := ReadVarInt(r)
	if err != nil {
		return err
	}

	// Prevent more transactions than could possibly fit into a block.
	// It would be possible to cause memory exhaustion and panics without
	// a sane upper bound on this count.
	if txCount > MaxTxPerBlock {
		str := fmt.Sprintf("too many transactions to fit into a block "+
			"[count %d, max %d]", txCount, MaxTxPerBlock)
		return messageError("MsgBlock.KaspaDecode", str)
	}

	msg.Transactions = make([]*MsgTx, 0, txCount)
	for i := uint64(0); i < txCount; i++ {
		tx := MsgTx{}
		err := tx.KaspaDecode(r, pver)
		if err != nil {
			return err
		}
		msg.Transactions = append(msg.Transactions, &tx)
	}

	return nil
}

// Deserialize decodes a block from r into the receiver using a format that is
// suitable for long-term storage such as a database while respecting the
// Version field in the block. This function differs from KaspaDecode in that
// KaspaDecode decodes from the kaspa appmessage protocol as it was sent across the
// network. The appmessage encoding can technically differ depending on the protocol
// version and doesn't even really need to match the format of a stored block at
// all. As of the time this comment was written, the encoded block is the same
// in both instances, but there is a distinct difference and separating the two
// allows the API to be flexible enough to deal with changes.
func (msg *MsgBlock) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the appmessage encoding
	// at protocol version 0 and the stable long-term storage format. As
	// a result, make use of KaspaDecode.
	return msg.KaspaDecode(r, 0)
}

// DeserializeTxLoc decodes r in the same manner Deserialize does, but it takes
// a byte buffer instead of a generic reader and returns a slice containing the
// start and length of each transaction within the raw data that is being
// deserialized.
func (msg *MsgBlock) DeserializeTxLoc(r *bytes.Buffer) ([]TxLoc, error) {
	fullLen := r.Len()

	// At the current time, there is no difference between the appmessage encoding
	// at protocol version 0 and the stable long-term storage format. As
	// a result, make use of existing appmessage protocol functions.
	err := readBlockHeader(r, 0, &msg.Header)
	if err != nil {
		return nil, err
	}

	txCount, err := ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	// Prevent more transactions than could possibly fit into a block.
	// It would be possible to cause memory exhaustion and panics without
	// a sane upper bound on this count.
	if txCount > MaxTxPerBlock {
		str := fmt.Sprintf("too many transactions to fit into a block "+
			"[count %d, max %d]", txCount, MaxTxPerBlock)
		return nil, messageError("MsgBlock.DeserializeTxLoc", str)
	}

	// Deserialize each transaction while keeping track of its location
	// within the byte stream.
	msg.Transactions = make([]*MsgTx, 0, txCount)
	txLocs := make([]TxLoc, txCount)
	for i := uint64(0); i < txCount; i++ {
		txLocs[i].TxStart = fullLen - r.Len()
		tx := MsgTx{}
		err := tx.Deserialize(r)
		if err != nil {
			return nil, err
		}
		msg.Transactions = append(msg.Transactions, &tx)
		txLocs[i].TxLen = (fullLen - r.Len()) - txLocs[i].TxStart
	}

	return txLocs, nil
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding blocks to be stored to disk, such as in a
// database, as opposed to encoding blocks for the appmessage.
func (msg *MsgBlock) KaspaEncode(w io.Writer, pver uint32) error {
	err := writeBlockHeader(w, pver, &msg.Header)
	if err != nil {
		return err
	}

	err = WriteVarInt(w, uint64(len(msg.Transactions)))
	if err != nil {
		return err
	}

	for _, tx := range msg.Transactions {
		err = tx.KaspaEncode(w, pver)
		if err != nil {
			return err
		}
	}

	return nil
}

// Serialize encodes the block to w using a format that suitable for long-term
// storage such as a database while respecting the Version field in the block.
// This function differs from KaspaEncode in that KaspaEncode encodes the block to
// the kaspa appmessage protocol in order to be sent across the network. The appmessage
// encoding can technically differ depending on the protocol version and doesn't
// even really need to match the format of a stored block at all. As of the
// time this comment was written, the encoded block is the same in both
// instances, but there is a distinct difference and separating the two allows
// the API to be flexible enough to deal with changes.
func (msg *MsgBlock) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the appmessage encoding
	// at protocol version 0 and the stable long-term storage format. As
	// a result, make use of KaspaEncode.
	return msg.KaspaEncode(w, 0)
}

// SerializeSize returns the number of bytes it would take to serialize the
// block.
func (msg *MsgBlock) SerializeSize() int {
	// Block header bytes + Serialized varint size for the number of
	// transactions.
	n := msg.Header.SerializeSize() + VarIntSerializeSize(uint64(len(msg.Transactions)))

	for _, tx := range msg.Transactions {
		n += tx.SerializeSize()
	}

	return n
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

// BlockHash computes the block identifier hash for this block.
func (msg *MsgBlock) BlockHash() *daghash.Hash {
	return msg.Header.BlockHash()
}

// ConvertToPartial clears out all the payloads of the subnetworks that are
// incompatible with the given subnetwork ID.
// Note: this operation modifies the block in place.
func (msg *MsgBlock) ConvertToPartial(subnetworkID *externalapi.DomainSubnetworkID) {
	for _, tx := range msg.Transactions {
		if tx.SubnetworkID != *subnetworkID {
			tx.Payload = []byte{}
		}
	}
}

// NewMsgBlock returns a new kaspa block message that conforms to the
// Message interface. See MsgBlock for details.
func NewMsgBlock(blockHeader *BlockHeader) *MsgBlock {
	return &MsgBlock{
		Header:       *blockHeader,
		Transactions: make([]*MsgTx, 0, defaultTransactionAlloc),
	}
}
