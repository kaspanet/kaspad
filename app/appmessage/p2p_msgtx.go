// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/util/binaryserializer"
)

const (
	// TxVersion is the current latest supported transaction version.
	TxVersion = 1

	// MaxTxInSequenceNum is the maximum sequence number the sequence field
	// of a transaction input can be.
	MaxTxInSequenceNum uint64 = math.MaxUint64

	// MaxPrevOutIndex is the maximum index the index field of a previous
	// outpoint can be.
	MaxPrevOutIndex uint32 = 0xffffffff

	// SequenceLockTimeDisabled is a flag that if set on a transaction
	// input's sequence number, the sequence number will not be interpreted
	// as a relative locktime.
	SequenceLockTimeDisabled = 1 << 31

	// SequenceLockTimeIsSeconds is a flag that if set on a transaction
	// input's sequence number, the relative locktime has units of 512
	// seconds.
	SequenceLockTimeIsSeconds = 1 << 22

	// SequenceLockTimeMask is a mask that extracts the relative locktime
	// when masked against the transaction input sequence number.
	SequenceLockTimeMask = 0x0000ffff

	// SequenceLockTimeGranularity is the defined time based granularity
	// for milliseconds-based relative time locks. When converting from milliseconds
	// to a sequence number, the value is right shifted by this amount,
	// therefore the granularity of relative time locks in 524288 or 2^19
	// seconds. Enforced relative lock times are multiples of 524288 milliseconds.
	SequenceLockTimeGranularity = 19

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
	// Value 8 bytes + Varint for ScriptPubKey length 1 byte.
	MinTxOutPayload = 9

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

	// freeListMaxScriptSize is the size of each buffer in the free list
	// that	is used for deserializing scripts from the appmessage before they are
	// concatenated into a single contiguous buffers. This value was chosen
	// because it is slightly more than twice the size of the vast majority
	// of all "standard" scripts. Larger scripts are still deserialized
	// properly as the free list will simply be bypassed for them.
	freeListMaxScriptSize = 512

	// freeListMaxItems is the number of buffers to keep in the free list
	// to use for script deserialization. This value allows up to 100
	// scripts per transaction being simultaneously deserialized by 125
	// peers. Thus, the peak usage of the free list is 12,500 * 512 =
	// 6,400,000 bytes.
	freeListMaxItems = 12500
)

// txEncoding is a bitmask defining which transaction fields we
// want to encode and which to ignore.
type txEncoding uint8

const (
	txEncodingFull txEncoding = 0

	txEncodingExcludePayload txEncoding = 1 << iota

	txEncodingExcludeSignatureScript
)

// scriptFreeList defines a free list of byte slices (up to the maximum number
// defined by the freeListMaxItems constant) that have a cap according to the
// freeListMaxScriptSize constant. It is used to provide temporary buffers for
// deserializing scripts in order to greatly reduce the number of allocations
// required.
//
// The caller can obtain a buffer from the free list by calling the Borrow
// function and should return it via the Return function when done using it.
type scriptFreeList chan []byte

// Borrow returns a byte slice from the free list with a length according the
// provided size. A new buffer is allocated if there are any items available.
//
// When the size is larger than the max size allowed for items on the free list
// a new buffer of the appropriate size is allocated and returned. It is safe
// to attempt to return said buffer via the Return function as it will be
// ignored and allowed to go the garbage collector.
func (c scriptFreeList) Borrow(size uint64) []byte {
	if size > freeListMaxScriptSize {
		return make([]byte, size)
	}

	var buf []byte
	select {
	case buf = <-c:
	default:
		buf = make([]byte, freeListMaxScriptSize)
	}
	return buf[:size]
}

// Return puts the provided byte slice back on the free list when it has a cap
// of the expected length. The buffer is expected to have been obtained via
// the Borrow function. Any slices that are not of the appropriate size, such
// as those whose size is greater than the largest allowed free list item size
// are simply ignored so they can go to the garbage collector.
func (c scriptFreeList) Return(buf []byte) {
	// Ignore any buffers returned that aren't the expected size for the
	// free list.
	if cap(buf) != freeListMaxScriptSize {
		return
	}

	// Return the buffer to the free list when it's not full. Otherwise let
	// it be garbage collected.
	select {
	case c <- buf:
	default:
		// Let it go to the garbage collector.
	}
}

// Create the concurrent safe free list to use for script deserialization. As
// previously described, this free list is maintained to significantly reduce
// the number of allocations.
var scriptPool scriptFreeList = make(chan []byte, freeListMaxItems)

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

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction input.
func (t *TxIn) SerializeSize() int {
	return t.serializeSize(txEncodingFull)
}

func (t *TxIn) serializeSize(encodingFlags txEncoding) int {
	// Outpoint ID 32 bytes + Outpoint Index 4 bytes + Sequence 8 bytes +
	// serialized varint size for the length of SignatureScript +
	// SignatureScript bytes.
	return 44 + serializeSignatureScriptSize(t.SignatureScript, encodingFlags)
}

func serializeSignatureScriptSize(signatureScript []byte, encodingFlags txEncoding) int {
	if encodingFlags&txEncodingExcludeSignatureScript != txEncodingExcludeSignatureScript {
		return VarIntSerializeSize(uint64(len(signatureScript))) +
			len(signatureScript)
	}
	return VarIntSerializeSize(0)
}

// NewTxIn returns a new kaspa transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxIn(prevOut *Outpoint, signatureScript []byte) *TxIn {
	return &TxIn{
		PreviousOutpoint: *prevOut,
		SignatureScript:  signatureScript,
		Sequence:         MaxTxInSequenceNum,
	}
}

// TxOut defines a kaspa transaction output.
type TxOut struct {
	Value        uint64
	ScriptPubKey []byte
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (t *TxOut) SerializeSize() int {
	// Value 8 bytes + serialized varint size for the length of ScriptPubKey +
	// ScriptPubKey bytes.
	return 8 + VarIntSerializeSize(uint64(len(t.ScriptPubKey))) + len(t.ScriptPubKey)
}

// NewTxOut returns a new kaspa transaction output with the provided
// transaction value and public key script.
func NewTxOut(value uint64, scriptPubKey []byte) *TxOut {
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
	Version      int32
	TxIn         []*TxIn
	TxOut        []*TxOut
	LockTime     uint64
	SubnetworkID externalapi.DomainSubnetworkID
	Gas          uint64
	PayloadHash  *externalapi.DomainHash
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
	return hashserialization.TransactionHash(MsgTxToDomainTransaction(msg))
}

// TxID generates the Hash for the transaction without the signature script, gas and payload fields.
func (msg *MsgTx) TxID() *externalapi.DomainTransactionID {
	return hashserialization.TransactionID(MsgTxToDomainTransaction(msg))
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
		// Deep copy the old ScriptPubKey
		var newScript []byte
		oldScript := oldTxOut.ScriptPubKey
		oldScriptLen := len(oldScript)
		if oldScriptLen > 0 {
			newScript = make([]byte, oldScriptLen)
			copy(newScript, oldScript[:oldScriptLen])
		}

		// Create new txOut with the deep copied data and append it to
		// new Tx.
		newTxOut := TxOut{
			Value:        oldTxOut.Value,
			ScriptPubKey: newScript,
		}
		newTx.TxOut = append(newTx.TxOut, &newTxOut)
	}

	return &newTx
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding transactions stored to disk, such as in a
// database, as opposed to decoding transactions from the appmessage.
func (msg *MsgTx) KaspaDecode(r io.Reader, pver uint32) error {
	version, err := binaryserializer.Uint32(r, littleEndian)
	if err != nil {
		return err
	}
	msg.Version = int32(version)

	count, err := ReadVarInt(r)
	if err != nil {
		return err
	}

	// Prevent more input transactions than could possibly fit into a
	// message. It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if count > uint64(maxTxInPerMessage) {
		str := fmt.Sprintf("too many input transactions to fit into "+
			"max message size [count %d, max %d]", count,
			maxTxInPerMessage)
		return messageError("MsgTx.KaspaDecode", str)
	}

	// returnScriptBuffers is a closure that returns any script buffers that
	// were borrowed from the pool when there are any deserialization
	// errors. This is only valid to call before the final step which
	// replaces the scripts with the location in a contiguous buffer and
	// returns them.
	returnScriptBuffers := func() {
		for _, txIn := range msg.TxIn {
			if txIn == nil || txIn.SignatureScript == nil {
				continue
			}
			scriptPool.Return(txIn.SignatureScript)
		}
		for _, txOut := range msg.TxOut {
			if txOut == nil || txOut.ScriptPubKey == nil {
				continue
			}
			scriptPool.Return(txOut.ScriptPubKey)
		}
	}

	// Deserialize the inputs.
	var totalScriptSize uint64
	txIns := make([]TxIn, count)
	msg.TxIn = make([]*TxIn, count)
	for i := uint64(0); i < count; i++ {
		// The pointer is set now in case a script buffer is borrowed
		// and needs to be returned to the pool on error.
		ti := &txIns[i]
		msg.TxIn[i] = ti
		err = readTxIn(r, pver, msg.Version, ti)
		if err != nil {
			returnScriptBuffers()
			return err
		}
		totalScriptSize += uint64(len(ti.SignatureScript))
	}

	count, err = ReadVarInt(r)
	if err != nil {
		returnScriptBuffers()
		return err
	}

	// Prevent more output transactions than could possibly fit into a
	// message. It would be possible to cause memory exhaustion and panics
	// without a sane upper bound on this count.
	if count > uint64(maxTxOutPerMessage) {
		returnScriptBuffers()
		str := fmt.Sprintf("too many output transactions to fit into "+
			"max message size [count %d, max %d]", count,
			maxTxOutPerMessage)
		return messageError("MsgTx.KaspaDecode", str)
	}

	// Deserialize the outputs.
	txOuts := make([]TxOut, count)
	msg.TxOut = make([]*TxOut, count)
	for i := uint64(0); i < count; i++ {
		// The pointer is set now in case a script buffer is borrowed
		// and needs to be returned to the pool on error.
		to := &txOuts[i]
		msg.TxOut[i] = to
		err = readTxOut(r, pver, msg.Version, to)
		if err != nil {
			returnScriptBuffers()
			return err
		}
		totalScriptSize += uint64(len(to.ScriptPubKey))
	}

	lockTime, err := binaryserializer.Uint64(r, littleEndian)
	msg.LockTime = lockTime
	if err != nil {
		returnScriptBuffers()
		return err
	}

	_, err = io.ReadFull(r, msg.SubnetworkID[:])
	if err != nil {
		returnScriptBuffers()
		return err
	}

	if msg.SubnetworkID != subnetworks.SubnetworkIDNative {
		msg.Gas, err = binaryserializer.Uint64(r, littleEndian)
		if err != nil {
			returnScriptBuffers()
			return err
		}

		var payloadHash externalapi.DomainHash
		err = ReadElement(r, &payloadHash)
		if err != nil {
			returnScriptBuffers()
			return err
		}
		msg.PayloadHash = &payloadHash

		payloadLength, err := ReadVarInt(r)
		if err != nil {
			returnScriptBuffers()
			return err
		}

		msg.Payload = make([]byte, payloadLength)
		_, err = io.ReadFull(r, msg.Payload)
		if err != nil {
			returnScriptBuffers()
			return err
		}
	}

	// Create a single allocation to house all of the scripts and set each
	// input signature script and output public key script to the
	// appropriate subslice of the overall contiguous buffer. Then, return
	// each individual script buffer back to the pool so they can be reused
	// for future deserializations. This is done because it significantly
	// reduces the number of allocations the garbage collector needs to
	// track, which in turn improves performance and drastically reduces the
	// amount of runtime overhead that would otherwise be needed to keep
	// track of millions of small allocations.
	//
	// NOTE: It is no longer valid to call the returnScriptBuffers closure
	// after these blocks of code run because it is already done and the
	// scripts in the transaction inputs and outputs no longer point to the
	// buffers.
	var offset uint64
	scripts := make([]byte, totalScriptSize)
	for i := 0; i < len(msg.TxIn); i++ {
		// Copy the signature script into the contiguous buffer at the
		// appropriate offset.
		signatureScript := msg.TxIn[i].SignatureScript
		copy(scripts[offset:], signatureScript)

		// Reset the signature script of the transaction input to the
		// slice of the contiguous buffer where the script lives.
		scriptSize := uint64(len(signatureScript))
		end := offset + scriptSize
		msg.TxIn[i].SignatureScript = scripts[offset:end:end]
		offset += scriptSize

		// Return the temporary script buffer to the pool.
		scriptPool.Return(signatureScript)
	}
	for i := 0; i < len(msg.TxOut); i++ {
		// Copy the public key script into the contiguous buffer at the
		// appropriate offset.
		scriptPubKey := msg.TxOut[i].ScriptPubKey
		copy(scripts[offset:], scriptPubKey)

		// Reset the public key script of the transaction output to the
		// slice of the contiguous buffer where the script lives.
		scriptSize := uint64(len(scriptPubKey))
		end := offset + scriptSize
		msg.TxOut[i].ScriptPubKey = scripts[offset:end:end]
		offset += scriptSize

		// Return the temporary script buffer to the pool.
		scriptPool.Return(scriptPubKey)
	}

	return nil
}

// Deserialize decodes a transaction from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field in the transaction. This function differs from KaspaDecode
// in that KaspaDecode decodes from the kaspa appmessage protocol as it was sent
// across the network. The appmessage encoding can technically differ depending on
// the protocol version and doesn't even really need to match the format of a
// stored transaction at all. As of the time this comment was written, the
// encoded transaction is the same in both instances, but there is a distinct
// difference and separating the two allows the API to be flexible enough to
// deal with changes.
func (msg *MsgTx) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the appmessage encoding
	// at protocol version 0 and the stable long-term storage format. As
	// a result, make use of KaspaDecode.
	return msg.KaspaDecode(r, 0)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding transactions to be stored to disk, such as in a
// database, as opposed to encoding transactions for the appmessage.
func (msg *MsgTx) KaspaEncode(w io.Writer, pver uint32) error {
	return msg.encode(w, pver, txEncodingFull)
}

func (msg *MsgTx) encode(w io.Writer, pver uint32, encodingFlags txEncoding) error {
	err := binaryserializer.PutUint32(w, littleEndian, uint32(msg.Version))
	if err != nil {
		return err
	}

	count := uint64(len(msg.TxIn))
	err = WriteVarInt(w, count)
	if err != nil {
		return err
	}

	for _, ti := range msg.TxIn {
		err = writeTxIn(w, pver, msg.Version, ti, encodingFlags)
		if err != nil {
			return err
		}
	}

	count = uint64(len(msg.TxOut))
	err = WriteVarInt(w, count)
	if err != nil {
		return err
	}

	for _, to := range msg.TxOut {
		err = WriteTxOut(w, pver, msg.Version, to)
		if err != nil {
			return err
		}
	}

	err = binaryserializer.PutUint64(w, littleEndian, msg.LockTime)
	if err != nil {
		return err
	}

	_, err = w.Write(msg.SubnetworkID[:])
	if err != nil {
		return err
	}

	if msg.SubnetworkID != subnetworks.SubnetworkIDNative {
		if subnetworks.IsBuiltIn(msg.SubnetworkID) && msg.Gas != 0 {
			str := "Transactions from built-in should have 0 gas"
			return messageError("MsgTx.KaspaEncode", str)
		}

		err = binaryserializer.PutUint64(w, littleEndian, msg.Gas)
		if err != nil {
			return err
		}

		err = WriteElement(w, msg.PayloadHash)
		if err != nil {
			return err
		}

		if encodingFlags&txEncodingExcludePayload != txEncodingExcludePayload {
			err = WriteVarInt(w, uint64(len(msg.Payload)))
			w.Write(msg.Payload)
		} else {
			err = WriteVarInt(w, 0)
		}
		if err != nil {
			return err
		}
	} else if msg.Payload != nil {
		str := "Transactions from native subnetwork should have <nil> payload"
		return messageError("MsgTx.KaspaEncode", str)
	} else if msg.PayloadHash != nil {
		str := "Transactions from native subnetwork should have <nil> payload hash"
		return messageError("MsgTx.KaspaEncode", str)
	} else if msg.Gas != 0 {
		str := "Transactions from native subnetwork should have 0 gas"
		return messageError("MsgTx.KaspaEncode", str)
	}

	return nil
}

// Serialize encodes the transaction to w using a format that suitable for
// long-term storage such as a database while respecting the Version field in
// the transaction. This function differs from KaspaEncode in that KaspaEncode
// encodes the transaction to the kaspa appmessage protocol in order to be sent
// across the network. The appmessage encoding can technically differ depending on
// the protocol version and doesn't even really need to match the format of a
// stored transaction at all. As of the time this comment was written, the
// encoded transaction is the same in both instances, but there is a distinct
// difference and separating the two allows the API to be flexible enough to
// deal with changes.
func (msg *MsgTx) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the appmessage encoding
	// at protocol version 0 and the stable long-term storage format. As
	// a result, make use of KaspaEncode.
	return msg.KaspaEncode(w, 0)
}

func (msg *MsgTx) serialize(w io.Writer, encodingFlags txEncoding) error {
	// At the current time, there is no difference between the appmessage encoding
	// at protocol version 0 and the stable long-term storage format. As
	// a result, make use of `encode`.
	return msg.encode(w, 0, encodingFlags)
}

// SerializeSize returns the number of bytes it would take to serialize
// the transaction.
func (msg *MsgTx) SerializeSize() int {
	return msg.serializeSize(txEncodingFull)
}

// SerializeSize returns the number of bytes it would take to serialize
// the transaction.
func (msg *MsgTx) serializeSize(encodingFlags txEncoding) int {
	// Version 4 bytes + LockTime 8 bytes + SubnetworkID 20
	// bytes + Serialized varint size for the number of transaction
	// inputs and outputs.
	n := 32 + VarIntSerializeSize(uint64(len(msg.TxIn))) +
		VarIntSerializeSize(uint64(len(msg.TxOut)))

	if msg.SubnetworkID != subnetworks.SubnetworkIDNative {
		// Gas 8 bytes
		n += 8

		// PayloadHash
		n += externalapi.DomainHashSize

		// Serialized varint size for the length of the payload
		if encodingFlags&txEncodingExcludePayload != txEncodingExcludePayload {
			n += VarIntSerializeSize(uint64(len(msg.Payload)))
			n += len(msg.Payload)
		} else {
			n += VarIntSerializeSize(0)
		}
	}

	for _, txIn := range msg.TxIn {
		n += txIn.serializeSize(encodingFlags)
	}

	for _, txOut := range msg.TxOut {
		n += txOut.SerializeSize()
	}

	return n
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

// ScriptPubKeyLocs returns a slice containing the start of each public key script
// within the raw serialized transaction. The caller can easily obtain the
// length of each script by using len on the script available via the
// appropriate transaction output entry.
func (msg *MsgTx) ScriptPubKeyLocs() []int {
	numTxOut := len(msg.TxOut)
	if numTxOut == 0 {
		return nil
	}

	// The starting offset in the serialized transaction of the first
	// transaction output is:
	//
	// Version 4 bytes + serialized varint size for the number of
	// transaction inputs and outputs + serialized size of each transaction
	// input.
	n := 4 + VarIntSerializeSize(uint64(len(msg.TxIn))) +
		VarIntSerializeSize(uint64(numTxOut))

	for _, txIn := range msg.TxIn {
		n += txIn.SerializeSize()
	}

	// Calculate and set the appropriate offset for each public key script.
	scriptPubKeyLocs := make([]int, numTxOut)
	for i, txOut := range msg.TxOut {
		// The offset of the script in the transaction output is:
		//
		// Value 8 bytes + serialized varint size for the length of
		// ScriptPubKey.
		n += 8 + VarIntSerializeSize(uint64(len(txOut.ScriptPubKey)))
		scriptPubKeyLocs[i] = n
		n += len(txOut.ScriptPubKey)
	}

	return scriptPubKeyLocs
}

// IsSubnetworkCompatible return true iff subnetworkID is one or more of the following:
// 1. The SupportsAll subnetwork (full node)
// 2. The native subnetwork
// 3. The transaction's subnetwork
func (msg *MsgTx) IsSubnetworkCompatible(subnetworkID *externalapi.DomainSubnetworkID) bool {
	return subnetworkID == nil ||
		*subnetworkID == subnetworks.SubnetworkIDNative ||
		*subnetworkID == msg.SubnetworkID
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
func newMsgTx(version int32, txIn []*TxIn, txOut []*TxOut, subnetworkID *externalapi.DomainSubnetworkID,
	gas uint64, payload []byte, lockTime uint64) *MsgTx {

	if txIn == nil {
		txIn = make([]*TxIn, 0, defaultTxInOutAlloc)
	}

	if txOut == nil {
		txOut = make([]*TxOut, 0, defaultTxInOutAlloc)
	}

	var payloadHash *externalapi.DomainHash
	if *subnetworkID != subnetworks.SubnetworkIDNative {
		payloadHash = hashes.HashData(payload)
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
func NewNativeMsgTx(version int32, txIn []*TxIn, txOut []*TxOut) *MsgTx {
	return newMsgTx(version, txIn, txOut, &subnetworks.SubnetworkIDNative, 0, nil, 0)
}

// NewSubnetworkMsgTx returns a new tx message in the specified subnetwork with specified gas and payload
func NewSubnetworkMsgTx(version int32, txIn []*TxIn, txOut []*TxOut, subnetworkID *externalapi.DomainSubnetworkID,
	gas uint64, payload []byte) *MsgTx {

	return newMsgTx(version, txIn, txOut, subnetworkID, gas, payload, 0)
}

// NewNativeMsgTxWithLocktime returns a new tx message in the native subnetwork with a locktime.
//
// See newMsgTx for further documntation of the parameters
func NewNativeMsgTxWithLocktime(version int32, txIn []*TxIn, txOut []*TxOut, locktime uint64) *MsgTx {
	return newMsgTx(version, txIn, txOut, &subnetworks.SubnetworkIDNative, 0, nil, locktime)
}

// NewRegistryMsgTx creates a new MsgTx that registers a new subnetwork
func NewRegistryMsgTx(version int32, txIn []*TxIn, txOut []*TxOut, gasLimit uint64) *MsgTx {
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint64(payload, gasLimit)

	return NewSubnetworkMsgTx(version, txIn, txOut, &subnetworks.SubnetworkIDRegistry, 0, payload)
}

// readOutpoint reads the next sequence of bytes from r as an Outpoint.
func readOutpoint(r io.Reader, pver uint32, version int32, op *Outpoint) error {
	_, err := io.ReadFull(r, op.TxID[:])
	if err != nil {
		return err
	}

	op.Index, err = binaryserializer.Uint32(r, littleEndian)
	return err
}

// writeOutpoint encodes op to the kaspa protocol encoding for an Outpoint
// to w.
func writeOutpoint(w io.Writer, pver uint32, version int32, op *Outpoint) error {
	_, err := w.Write(op.TxID[:])
	if err != nil {
		return err
	}

	return binaryserializer.PutUint32(w, littleEndian, op.Index)
}

// readScript reads a variable length byte array that represents a transaction
// script. It is encoded as a varInt containing the length of the array
// followed by the bytes themselves. An error is returned if the length is
// greater than the passed maxAllowed parameter which helps protect against
// memory exhaustion attacks and forced panics through malformed messages. The
// fieldName parameter is only used for the error message so it provides more
// context in the error.
func readScript(r io.Reader, pver uint32, maxAllowed uint32, fieldName string) ([]byte, error) {
	count, err := ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	// Prevent byte array larger than the max message size. It would
	// be possible to cause memory exhaustion and panics without a sane
	// upper bound on this count.
	if count > uint64(maxAllowed) {
		str := fmt.Sprintf("%s is larger than the max allowed size "+
			"[count %d, max %d]", fieldName, count, maxAllowed)
		return nil, messageError("readScript", str)
	}

	b := scriptPool.Borrow(count)
	_, err = io.ReadFull(r, b)
	if err != nil {
		scriptPool.Return(b)
		return nil, err
	}
	return b, nil
}

// readTxIn reads the next sequence of bytes from r as a transaction input
// (TxIn).
func readTxIn(r io.Reader, pver uint32, version int32, ti *TxIn) error {
	err := readOutpoint(r, pver, version, &ti.PreviousOutpoint)
	if err != nil {
		return err
	}

	ti.SignatureScript, err = readScript(r, pver, MaxMessagePayload,
		"transaction input signature script")
	if err != nil {
		return err
	}

	return ReadElement(r, &ti.Sequence)
}

// writeTxIn encodes ti to the kaspa protocol encoding for a transaction
// input (TxIn) to w.
func writeTxIn(w io.Writer, pver uint32, version int32, ti *TxIn, encodingFlags txEncoding) error {
	err := writeOutpoint(w, pver, version, &ti.PreviousOutpoint)
	if err != nil {
		return err
	}

	if encodingFlags&txEncodingExcludeSignatureScript != txEncodingExcludeSignatureScript {
		err = WriteVarBytes(w, pver, ti.SignatureScript)
	} else {
		err = WriteVarBytes(w, pver, []byte{})
	}
	if err != nil {
		return err
	}

	return binaryserializer.PutUint64(w, littleEndian, ti.Sequence)
}

// readTxOut reads the next sequence of bytes from r as a transaction output
// (TxOut).
func readTxOut(r io.Reader, pver uint32, version int32, to *TxOut) error {
	err := ReadElement(r, &to.Value)
	if err != nil {
		return err
	}

	to.ScriptPubKey, err = readScript(r, pver, MaxMessagePayload,
		"transaction output public key script")
	return err
}

// WriteTxOut encodes to into the kaspa protocol encoding for a transaction
// output (TxOut) to w.
func WriteTxOut(w io.Writer, pver uint32, version int32, to *TxOut) error {
	err := binaryserializer.PutUint64(w, littleEndian, uint64(to.Value))
	if err != nil {
		return err
	}

	return WriteVarBytes(w, pver, to.ScriptPubKey)
}
