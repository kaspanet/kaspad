// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"strconv"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

const (
	// MaxPrevOutIndex is the maximum index the index field of a previous
	// outpoint can be.
	MaxPrevOutIndex uint32 = 0xffffffff

	// defaultTxInOutAlloc is the default size used for the backing array for
	// transaction inputs and outputs. The array will dynamically grow as needed,
	// but this figure is intended to provide enough space for the number of
	// inputs and outputs in a typical transaction without needing to grow the
	// backing array multiple times.
	defaultTxInOutAlloc = 15

	// minTxInPayload is the minimum payload size for a transaction input.
	// PreviousOutpoint.TxID + PreviousOutpoint.Index 4 bytes + Varint for
	// SignatureScript length 1 byte + Sequence 4 bytes.
	minTxInPayload = 9 + externalapi.DomainHashSize

	// maxTxInPerMessage is the maximum number of transactions inputs that
	// a transaction which fits into a message could possibly have.
	maxTxInPerMessage = (MaxMessagePayload / minTxInPayload) + 1

	// MinTxOutPayload is the minimum payload size for a transaction output.
	// Value 8 bytes + version 2 bytes + Varint for ScriptPublicKey length 1 byte.
	MinTxOutPayload = 11

	// maxTxOutPerMessage is the maximum number of transactions outputs that
	// a transaction which fits into a message could possibly have.
	maxTxOutPerMessage = (MaxMessagePayload / MinTxOutPayload) + 1

	// minTxPayload is the minimum payload size for a transaction. Note
	// that any realistically usable transaction must have at least one
	// input or output, but that is a rule enforced at a higher layer, so
	// it is intentionally not included here.
	// Version 4 bytes + Varint number of transaction inputs 1 byte + Varint
	// number of transaction outputs 1 byte + LockTime 4 bytes + min input
	// payload + min output payload.
	minTxPayload = 10
)

// Outpoint defines a kaspa data type that is used to track previous
// transaction outputs.
type Outpoint struct {
	TxID  externalapi.DomainTransactionID
	Index uint32
}

// NewOutpoint returns a new kaspa transaction outpoint point with the
// provided hash and index.
func NewOutpoint(txID *externalapi.DomainTransactionID, index uint32) *Outpoint {
	return &Outpoint{
		TxID:  *txID,
		Index: index,
	}
}

// String returns the Outpoint in the human-readable form "txID:index".
func (o Outpoint) String() string {
	// Allocate enough for ID string, colon, and 10 digits. Although
	// at the time of writing, the number of digits can be no greater than
	// the length of the decimal representation of maxTxOutPerMessage, the
	// maximum message payload may increase in the future and this
	// optimization may go unnoticed, so allocate space for 10 decimal
	// digits, which will fit any uint32.
	buf := make([]byte, 2*externalapi.DomainHashSize+1, 2*externalapi.DomainHashSize+1+10)
	copy(buf, o.TxID.String())
	buf[2*externalapi.DomainHashSize] = ':'
	buf = strconv.AppendUint(buf, uint64(o.Index), 10)
	return string(buf)
}

// TxIn defines a kaspa transaction input.
type TxIn struct {
	PreviousOutpoint Outpoint
	SignatureScript  []byte
	Sequence         uint64
}

// NewTxIn returns a new kaspa transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxIn(prevOut *Outpoint, signatureScript []byte, sequence uint64) *TxIn {
	return &TxIn{
		PreviousOutpoint: *prevOut,
		SignatureScript:  signatureScript,
		Sequence:         sequence,
	}
}

// TxOut defines a kaspa transaction output.
type TxOut struct {
	Value        uint64
	ScriptPubKey *externalapi.ScriptPublicKey
}

// NewTxOut returns a new kaspa transaction output with the provided
// transaction value and public key script.
func NewTxOut(value uint64, scriptPubKey *externalapi.ScriptPublicKey) *TxOut {
	return &TxOut{
		Value:        value,
		ScriptPubKey: scriptPubKey,
	}
}

// MsgTx implements the Message interface and represents a kaspa tx message.
// It is used to deliver transaction information in response to a getdata
// message (MsgGetData) for a given transaction.
//
// Use the AddTxIn and AddTxOut functions to build up the list of transaction
// inputs and outputs.
type MsgTx struct {
	baseMessage
	Version      uint16
	TxIn         []*TxIn
	TxOut        []*TxOut
	LockTime     uint64
	SubnetworkID externalapi.DomainSubnetworkID
	Gas          uint64
	PayloadHash  externalapi.DomainHash
	Payload      []byte
}

// AddTxIn adds a transaction input to the message.
func (msg *MsgTx) AddTxIn(ti *TxIn) {
	msg.TxIn = append(msg.TxIn, ti)
}

// AddTxOut adds a transaction output to the message.
func (msg *MsgTx) AddTxOut(to *TxOut) {
	msg.TxOut = append(msg.TxOut, to)
}

// IsCoinBase determines whether or not a transaction is a coinbase transaction. A coinbase
// transaction is a special transaction created by miners that distributes fees and block subsidy
// to the previous blocks' miners, and to specify the scriptPubKey that will be used to pay the current
// miner in future blocks. Each input of the coinbase transaction should set index to maximum
// value and reference the relevant block id, instead of previous transaction id.
func (msg *MsgTx) IsCoinBase() bool {
	// A coinbase transaction must have subnetwork id SubnetworkIDCoinbase
	return msg.SubnetworkID == subnetworks.SubnetworkIDCoinbase
}

// TxHash generates the Hash for the transaction.
func (msg *MsgTx) TxHash() *externalapi.DomainHash {
	return consensushashing.TransactionHash(MsgTxToDomainTransaction(msg))
}

// TxID generates the Hash for the transaction without the signature script, gas and payload fields.
func (msg *MsgTx) TxID() *externalapi.DomainTransactionID {
	return consensushashing.TransactionID(MsgTxToDomainTransaction(msg))
}

// Copy creates a deep copy of a transaction so that the original does not get
// modified when the copy is manipulated.
func (msg *MsgTx) Copy() *MsgTx {
	// Create new tx and start by copying primitive values and making space
	// for the transaction inputs and outputs.
	newTx := MsgTx{
		Version:      msg.Version,
		TxIn:         make([]*TxIn, 0, len(msg.TxIn)),
		TxOut:        make([]*TxOut, 0, len(msg.TxOut)),
		LockTime:     msg.LockTime,
		SubnetworkID: msg.SubnetworkID,
		Gas:          msg.Gas,
		PayloadHash:  msg.PayloadHash,
	}

	if msg.Payload != nil {
		newTx.Payload = make([]byte, len(msg.Payload))
		copy(newTx.Payload, msg.Payload)
	}

	// Deep copy the old TxIn data.
	for _, oldTxIn := range msg.TxIn {
		// Deep copy the old previous outpoint.
		oldOutpoint := oldTxIn.PreviousOutpoint
		newOutpoint := Outpoint{}
		newOutpoint.TxID = oldOutpoint.TxID
		newOutpoint.Index = oldOutpoint.Index

		// Deep copy the old signature script.
		var newScript []byte
		oldScript := oldTxIn.SignatureScript
		oldScriptLen := len(oldScript)
		if oldScriptLen > 0 {
			newScript = make([]byte, oldScriptLen)
			copy(newScript, oldScript[:oldScriptLen])
		}

		// Create new txIn with the deep copied data.
		newTxIn := TxIn{
			PreviousOutpoint: newOutpoint,
			SignatureScript:  newScript,
			Sequence:         oldTxIn.Sequence,
		}

		// Finally, append this fully copied txin.
		newTx.TxIn = append(newTx.TxIn, &newTxIn)
	}

	// Deep copy the old TxOut data.
	for _, oldTxOut := range msg.TxOut {
		// Deep copy the old ScriptPublicKey
		var newScript externalapi.ScriptPublicKey
		oldScript := oldTxOut.ScriptPubKey
		oldScriptLen := len(oldScript.Script)
		if oldScriptLen > 0 {
			newScript = externalapi.ScriptPublicKey{Script: make([]byte, oldScriptLen), Version: oldScript.Version}
			copy(newScript.Script, oldScript.Script[:oldScriptLen])
		}

		// Create new txOut with the deep copied data and append it to
		// new Tx.
		newTxOut := TxOut{
			Value:        oldTxOut.Value,
			ScriptPubKey: &newScript,
		}
		newTx.TxOut = append(newTx.TxOut, &newTxOut)
	}

	return &newTx
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgTx) Command() MessageCommand {
	return CmdTx
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgTx) MaxPayloadLength(pver uint32) uint32 {
	return MaxMessagePayload
}

// IsSubnetworkCompatible return true iff subnetworkID is one or more of the following:
// 1. The SupportsAll subnetwork (full node)
// 2. The native subnetwork
// 3. The transaction's subnetwork
func (msg *MsgTx) IsSubnetworkCompatible(subnetworkID *externalapi.DomainSubnetworkID) bool {
	return subnetworkID == nil ||
		subnetworkID.Equal(&subnetworks.SubnetworkIDNative) ||
		subnetworkID.Equal(&msg.SubnetworkID)
}

// newMsgTx returns a new tx message that conforms to the Message interface.
//
// All fields except version and gas has default values if nil is passed:
// txIn, txOut - empty arrays
// payload - an empty payload
//
// The payload hash is calculated automatically according to provided payload.
// Also, the lock time is set to zero to indicate the transaction is valid
// immediately as opposed to some time in future.
func newMsgTx(version uint16, txIn []*TxIn, txOut []*TxOut, subnetworkID *externalapi.DomainSubnetworkID,
	gas uint64, payload []byte, lockTime uint64) *MsgTx {

	if txIn == nil {
		txIn = make([]*TxIn, 0, defaultTxInOutAlloc)
	}

	if txOut == nil {
		txOut = make([]*TxOut, 0, defaultTxInOutAlloc)
	}

	var payloadHash externalapi.DomainHash
	if *subnetworkID != subnetworks.SubnetworkIDNative {
		payloadHash = *hashes.PayloadHash(payload)
	}

	return &MsgTx{
		Version:      version,
		TxIn:         txIn,
		TxOut:        txOut,
		SubnetworkID: *subnetworkID,
		Gas:          gas,
		PayloadHash:  payloadHash,
		Payload:      payload,
		LockTime:     lockTime,
	}
}

// NewNativeMsgTx returns a new tx message in the native subnetwork
func NewNativeMsgTx(version uint16, txIn []*TxIn, txOut []*TxOut) *MsgTx {
	return newMsgTx(version, txIn, txOut, &subnetworks.SubnetworkIDNative, 0, nil, 0)
}

// NewSubnetworkMsgTx returns a new tx message in the specified subnetwork with specified gas and payload
func NewSubnetworkMsgTx(version uint16, txIn []*TxIn, txOut []*TxOut, subnetworkID *externalapi.DomainSubnetworkID,
	gas uint64, payload []byte) *MsgTx {

	return newMsgTx(version, txIn, txOut, subnetworkID, gas, payload, 0)
}

// NewNativeMsgTxWithLocktime returns a new tx message in the native subnetwork with a locktime.
//
// See newMsgTx for further documntation of the parameters
func NewNativeMsgTxWithLocktime(version uint16, txIn []*TxIn, txOut []*TxOut, locktime uint64) *MsgTx {
	return newMsgTx(version, txIn, txOut, &subnetworks.SubnetworkIDNative, 0, nil, locktime)
}

// NewRegistryMsgTx creates a new MsgTx that registers a new subnetwork
func NewRegistryMsgTx(version uint16, txIn []*TxIn, txOut []*TxOut, gasLimit uint64) *MsgTx {
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint64(payload, gasLimit)

	return NewSubnetworkMsgTx(version, txIn, txOut, &subnetworks.SubnetworkIDRegistry, 0, payload)
}
