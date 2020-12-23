// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"math"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TestTx tests the MsgTx API.
func TestTx(t *testing.T) {
	pver := ProtocolVersion

	txIDStr := "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"
	txID, err := transactionid.FromString(txIDStr)
	if err != nil {
		t.Errorf("NewTxIDFromStr: %v", err)
	}

	// Ensure the command is expected value.
	wantCmd := MessageCommand(6)
	msg := NewNativeMsgTx(1, nil, nil)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgAddresses: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	wantPayload := uint32(1024 * 1024 * 32)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Ensure we get the same transaction outpoint data back out.
	// NOTE: This is a block hash and made up index, but we're only
	// testing package functionality.
	prevOutIndex := uint32(1)
	prevOut := NewOutpoint(txID, prevOutIndex)
	if !prevOut.TxID.Equal(txID) {
		t.Errorf("NewOutpoint: wrong ID - got %v, want %v",
			spew.Sprint(&prevOut.TxID), spew.Sprint(txID))
	}
	if prevOut.Index != prevOutIndex {
		t.Errorf("NewOutpoint: wrong index - got %v, want %v",
			prevOut.Index, prevOutIndex)
	}
	prevOutStr := fmt.Sprintf("%s:%d", txID.String(), prevOutIndex)
	if s := prevOut.String(); s != prevOutStr {
		t.Errorf("Outpoint.String: unexpected result - got %v, "+
			"want %v", s, prevOutStr)
	}

	// Ensure we get the same transaction input back out.
	sigScript := []byte{0x04, 0x31, 0xdc, 0x00, 0x1b, 0x01, 0x62}
	txIn := NewTxIn(prevOut, sigScript, constants.MaxTxInSequenceNum)
	if !reflect.DeepEqual(&txIn.PreviousOutpoint, prevOut) {
		t.Errorf("NewTxIn: wrong prev outpoint - got %v, want %v",
			spew.Sprint(&txIn.PreviousOutpoint),
			spew.Sprint(prevOut))
	}
	if !bytes.Equal(txIn.SignatureScript, sigScript) {
		t.Errorf("NewTxIn: wrong signature script - got %v, want %v",
			spew.Sdump(txIn.SignatureScript),
			spew.Sdump(sigScript))
	}

	// Ensure we get the same transaction output back out.
	txValue := uint64(5000000000)
	scriptPubKey := []byte{
		0x41, // OP_DATA_65
		0x04, 0xd6, 0x4b, 0xdf, 0xd0, 0x9e, 0xb1, 0xc5,
		0xfe, 0x29, 0x5a, 0xbd, 0xeb, 0x1d, 0xca, 0x42,
		0x81, 0xbe, 0x98, 0x8e, 0x2d, 0xa0, 0xb6, 0xc1,
		0xc6, 0xa5, 0x9d, 0xc2, 0x26, 0xc2, 0x86, 0x24,
		0xe1, 0x81, 0x75, 0xe8, 0x51, 0xc9, 0x6b, 0x97,
		0x3d, 0x81, 0xb0, 0x1c, 0xc3, 0x1f, 0x04, 0x78,
		0x34, 0xbc, 0x06, 0xd6, 0xd6, 0xed, 0xf6, 0x20,
		0xd1, 0x84, 0x24, 0x1a, 0x6a, 0xed, 0x8b, 0x63,
		0xa6, // 65-byte signature
		0xac, // OP_CHECKSIG
	}
	txOut := NewTxOut(txValue, scriptPubKey)
	if txOut.Value != txValue {
		t.Errorf("NewTxOut: wrong scriptPubKey - got %v, want %v",
			txOut.Value, txValue)

	}
	if !bytes.Equal(txOut.ScriptPubKey, scriptPubKey) {
		t.Errorf("NewTxOut: wrong scriptPubKey - got %v, want %v",
			spew.Sdump(txOut.ScriptPubKey),
			spew.Sdump(scriptPubKey))
	}

	// Ensure transaction inputs are added properly.
	msg.AddTxIn(txIn)
	if !reflect.DeepEqual(msg.TxIn[0], txIn) {
		t.Errorf("AddTxIn: wrong transaction input added - got %v, want %v",
			spew.Sprint(msg.TxIn[0]), spew.Sprint(txIn))
	}

	// Ensure transaction outputs are added properly.
	msg.AddTxOut(txOut)
	if !reflect.DeepEqual(msg.TxOut[0], txOut) {
		t.Errorf("AddTxIn: wrong transaction output added - got %v, want %v",
			spew.Sprint(msg.TxOut[0]), spew.Sprint(txOut))
	}

	// Ensure the copy produced an identical transaction message.
	newMsg := msg.Copy()
	if !reflect.DeepEqual(newMsg, msg) {
		t.Errorf("Copy: mismatched tx messages - got %v, want %v",
			spew.Sdump(newMsg), spew.Sdump(msg))
	}
}

// TestTxHash tests the ability to generate the hash of a transaction accurately.
func TestTxHashAndID(t *testing.T) {
	txHash1Str := "c2ac1e792c5c49260103ad9f86caf749d431958b7c7e5e5129346ceab8b709cf"
	txID1Str := "47ce12a5ee5727cf97c0481eebedad0d80646b743305b0921a2403f1836f8b37"
	wantTxID1, err := transactionid.FromString(txID1Str)
	if err != nil {
		t.Fatalf("NewTxIDFromStr: %v", err)
	}
	wantTxHash1, err := transactionid.FromString(txHash1Str)
	if err != nil {
		t.Fatalf("NewTxIDFromStr: %v", err)
	}

	// A coinbase transaction
	txIn := &TxIn{
		PreviousOutpoint: Outpoint{
			TxID:  externalapi.DomainTransactionID{},
			Index: math.MaxUint32,
		},
		SignatureScript: []byte{0x04, 0x31, 0xdc, 0x00, 0x1b, 0x01, 0x62},
		Sequence:        math.MaxUint64,
	}
	txOut := &TxOut{
		Value: 5000000000,
		ScriptPubKey: []byte{
			0x41, // OP_DATA_65
			0x04, 0xd6, 0x4b, 0xdf, 0xd0, 0x9e, 0xb1, 0xc5,
			0xfe, 0x29, 0x5a, 0xbd, 0xeb, 0x1d, 0xca, 0x42,
			0x81, 0xbe, 0x98, 0x8e, 0x2d, 0xa0, 0xb6, 0xc1,
			0xc6, 0xa5, 0x9d, 0xc2, 0x26, 0xc2, 0x86, 0x24,
			0xe1, 0x81, 0x75, 0xe8, 0x51, 0xc9, 0x6b, 0x97,
			0x3d, 0x81, 0xb0, 0x1c, 0xc3, 0x1f, 0x04, 0x78,
			0x34, 0xbc, 0x06, 0xd6, 0xd6, 0xed, 0xf6, 0x20,
			0xd1, 0x84, 0x24, 0x1a, 0x6a, 0xed, 0x8b, 0x63,
			0xa6, // 65-byte signature
			0xac, // OP_CHECKSIG
		},
	}
	tx1 := NewSubnetworkMsgTx(1, []*TxIn{txIn}, []*TxOut{txOut}, &subnetworks.SubnetworkIDCoinbase, 0, nil)

	// Ensure the hash produced is expected.
	tx1Hash := tx1.TxHash()
	if *tx1Hash != (externalapi.DomainHash)(*wantTxHash1) {
		t.Errorf("TxHash: wrong hash - got %v, want %v",
			spew.Sprint(tx1Hash), spew.Sprint(wantTxID1))
	}

	// Ensure the TxID for coinbase transaction is the same as TxHash.
	tx1ID := tx1.TxID()
	if !tx1ID.Equal(wantTxID1) {
		t.Errorf("TxID: wrong ID - got %v, want %v",
			spew.Sprint(tx1ID), spew.Sprint(wantTxID1))
	}

	hash2Str := "6b769655a1420022e4690a4f7bb9b1c381185ebbefe3070351f06fb573a0600c"
	wantHash2, err := hashes.FromString(hash2Str)
	if err != nil {
		t.Errorf("NewTxIDFromStr: %v", err)
		return
	}

	id2Str := "af916032e271adaaa21f02bee4b44db2cca4dad9149dcaebc188009c7313ec68"
	wantID2, err := transactionid.FromString(id2Str)
	if err != nil {
		t.Errorf("NewTxIDFromStr: %v", err)
		return
	}
	payload := []byte{1, 2, 3}
	txIns := []*TxIn{{
		PreviousOutpoint: Outpoint{
			Index: 0,
			TxID:  externalapi.DomainTransactionID{1, 2, 3},
		},
		SignatureScript: []byte{
			0x49, 0x30, 0x46, 0x02, 0x21, 0x00, 0xDA, 0x0D, 0xC6, 0xAE, 0xCE, 0xFE, 0x1E, 0x06, 0xEF, 0xDF,
			0x05, 0x77, 0x37, 0x57, 0xDE, 0xB1, 0x68, 0x82, 0x09, 0x30, 0xE3, 0xB0, 0xD0, 0x3F, 0x46, 0xF5,
			0xFC, 0xF1, 0x50, 0xBF, 0x99, 0x0C, 0x02, 0x21, 0x00, 0xD2, 0x5B, 0x5C, 0x87, 0x04, 0x00, 0x76,
			0xE4, 0xF2, 0x53, 0xF8, 0x26, 0x2E, 0x76, 0x3E, 0x2D, 0xD5, 0x1E, 0x7F, 0xF0, 0xBE, 0x15, 0x77,
			0x27, 0xC4, 0xBC, 0x42, 0x80, 0x7F, 0x17, 0xBD, 0x39, 0x01, 0x41, 0x04, 0xE6, 0xC2, 0x6E, 0xF6,
			0x7D, 0xC6, 0x10, 0xD2, 0xCD, 0x19, 0x24, 0x84, 0x78, 0x9A, 0x6C, 0xF9, 0xAE, 0xA9, 0x93, 0x0B,
			0x94, 0x4B, 0x7E, 0x2D, 0xB5, 0x34, 0x2B, 0x9D, 0x9E, 0x5B, 0x9F, 0xF7, 0x9A, 0xFF, 0x9A, 0x2E,
			0xE1, 0x97, 0x8D, 0xD7, 0xFD, 0x01, 0xDF, 0xC5, 0x22, 0xEE, 0x02, 0x28, 0x3D, 0x3B, 0x06, 0xA9,
			0xD0, 0x3A, 0xCF, 0x80, 0x96, 0x96, 0x8D, 0x7D, 0xBB, 0x0F, 0x91, 0x78,
		},
		Sequence: math.MaxUint64,
	}}
	txOuts := []*TxOut{
		{
			Value: 244623243,
			ScriptPubKey: []byte{
				0x76, 0xA9, 0x14, 0xBA, 0xDE, 0xEC, 0xFD, 0xEF, 0x05, 0x07, 0x24, 0x7F, 0xC8, 0xF7, 0x42, 0x41,
				0xD7, 0x3B, 0xC0, 0x39, 0x97, 0x2D, 0x7B, 0x88, 0xAC,
			},
		},
		{
			Value: 44602432,
			ScriptPubKey: []byte{
				0x76, 0xA9, 0x14, 0xC1, 0x09, 0x32, 0x48, 0x3F, 0xEC, 0x93, 0xED, 0x51, 0xF5, 0xFE, 0x95, 0xE7,
				0x25, 0x59, 0xF2, 0xCC, 0x70, 0x43, 0xF9, 0x88, 0xAC,
			},
		},
	}
	tx2 := NewSubnetworkMsgTx(1, txIns, txOuts, &externalapi.DomainSubnetworkID{1, 2, 3}, 0, payload)

	// Ensure the hash produced is expected.
	tx2Hash := tx2.TxHash()
	if !tx2Hash.Equal(wantHash2) {
		t.Errorf("TxHash: wrong hash - got %v, want %v",
			spew.Sprint(tx2Hash), spew.Sprint(wantHash2))
	}

	// Ensure the TxID for coinbase transaction is the same as TxHash.
	tx2ID := tx2.TxID()
	if !tx2ID.Equal(wantID2) {
		t.Errorf("TxID: wrong ID - got %v, want %v",
			spew.Sprint(tx2ID), spew.Sprint(wantID2))
	}

	if tx2ID.Equal((*externalapi.DomainTransactionID)(tx2Hash)) {
		t.Errorf("tx2ID and tx2Hash shouldn't be the same for non-coinbase transaction with signature and/or payload")
	}

	tx2.TxIn[0].SignatureScript = []byte{}
	newTx2Hash := tx2.TxHash()
	if *tx2ID == (externalapi.DomainTransactionID)(*newTx2Hash) {
		t.Errorf("tx2ID and newTx2Hash should not be the same even for transaction with an empty signature")
	}
}
