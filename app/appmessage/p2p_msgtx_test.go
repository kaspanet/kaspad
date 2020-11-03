// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"

	"github.com/davecgh/go-spew/spew"
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
	if prevOut.TxID != *txID {
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
	txIn := NewTxIn(prevOut, sigScript)
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
	txID1Str := "a3d29c39bfb578235e4813cc8138a9ba10def63acad193a7a880159624840d7f"
	wantTxID1, err := transactionid.FromString(txID1Str)
	if err != nil {
		t.Errorf("NewTxIDFromStr: %v", err)
		return
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
	if *tx1Hash != (externalapi.DomainHash)(*wantTxID1) {
		t.Errorf("TxHash: wrong hash - got %v, want %v",
			spew.Sprint(tx1Hash), spew.Sprint(wantTxID1))
	}

	// Ensure the TxID for coinbase transaction is the same as TxHash.
	tx1ID := tx1.TxID()
	if *tx1ID != *wantTxID1 {
		t.Errorf("TxID: wrong ID - got %v, want %v",
			spew.Sprint(tx1ID), spew.Sprint(wantTxID1))
	}

	hash2Str := "c84f3009b337aaa3adeb2ffd41010d5f62dd773ca25b39c908a77da91f87b729"
	wantHash2, err := hashes.FromString(hash2Str)
	if err != nil {
		t.Errorf("NewTxIDFromStr: %v", err)
		return
	}

	id2Str := "7c919f676109743a1271a88beeb43849a6f9cc653f6082e59a7266f3df4802b9"
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
	if *tx2Hash != *wantHash2 {
		t.Errorf("TxHash: wrong hash - got %v, want %v",
			spew.Sprint(tx2Hash), spew.Sprint(wantHash2))
	}

	// Ensure the TxID for coinbase transaction is the same as TxHash.
	tx2ID := tx2.TxID()
	if *tx2ID != *wantID2 {
		t.Errorf("TxID: wrong ID - got %v, want %v",
			spew.Sprint(tx2ID), spew.Sprint(wantID2))
	}

	if *tx2ID == (externalapi.DomainTransactionID)(*tx2Hash) {
		t.Errorf("tx2ID and tx2Hash shouldn't be the same for non-coinbase transaction with signature and/or payload")
	}

	tx2.TxIn[0].SignatureScript = []byte{}
	newTx2Hash := tx2.TxHash()
	if *tx2ID != (externalapi.DomainTransactionID)(*newTx2Hash) {
		t.Errorf("tx2ID and newTx2Hash should be the same for transaction with an empty signature")
	}
}

// TestTxEncoding tests the MsgTx appmessage encode and decode for various numbers
// of transaction inputs and outputs and protocol versions.
func TestTxEncoding(t *testing.T) {
	// Empty tx message.
	noTx := NewNativeMsgTx(1, nil, nil)
	noTxEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x00,                                           // Varint for number of input transactions
		0x00,                                           // Varint for number of output transactions
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Lock time
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // Sub Network ID
	}

	tests := []struct {
		in   *MsgTx // Message to encode
		out  *MsgTx // Expected decoded message
		buf  []byte // Encoded value
		pver uint32 // Protocol version for appmessage encoding
	}{
		// Latest protocol version with no transactions.
		{
			noTx,
			noTx,
			noTxEncoded,
			ProtocolVersion,
		},

		// Latest protocol version with multiple transactions.
		{
			multiTx,
			multiTx,
			multiTxEncoded,
			ProtocolVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode the message to appmessage format.
		var buf bytes.Buffer
		err := test.in.KaspaEncode(&buf, test.pver)
		if err != nil {
			t.Errorf("KaspaEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("KaspaEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the message from appmessage format.
		var msg MsgTx
		rbuf := bytes.NewReader(test.buf)
		err = msg.KaspaDecode(rbuf, test.pver)
		if err != nil {
			t.Errorf("KaspaDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&msg, test.out) {
			t.Errorf("KaspaDecode #%d\n got: %s want: %s", i,
				spew.Sdump(&msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestTxEncodingErrors performs negative tests against appmessage encode and decode
// of MsgTx to confirm error paths work correctly.
func TestTxEncodingErrors(t *testing.T) {
	pver := ProtocolVersion

	tests := []struct {
		in       *MsgTx // Value to encode
		buf      []byte // Encoded value
		pver     uint32 // Protocol version for appmessage encoding
		max      int    // Max size of fixed buffer to induce errors
		writeErr error  // Expected write error
		readErr  error  // Expected read error
	}{
		// Force error in version.
		{multiTx, multiTxEncoded, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in number of transaction inputs.
		{multiTx, multiTxEncoded, pver, 4, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block hash.
		{multiTx, multiTxEncoded, pver, 5, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block output index.
		{multiTx, multiTxEncoded, pver, 37, io.ErrShortWrite, io.EOF},
		// Force error in transaction input signature script length.
		{multiTx, multiTxEncoded, pver, 41, io.ErrShortWrite, io.EOF},
		// Force error in transaction input signature script.
		{multiTx, multiTxEncoded, pver, 42, io.ErrShortWrite, io.EOF},
		// Force error in transaction input sequence.
		{multiTx, multiTxEncoded, pver, 49, io.ErrShortWrite, io.EOF},
		// Force error in number of transaction outputs.
		{multiTx, multiTxEncoded, pver, 57, io.ErrShortWrite, io.EOF},
		// Force error in transaction output value.
		{multiTx, multiTxEncoded, pver, 58, io.ErrShortWrite, io.EOF},
		// Force error in transaction output scriptPubKey length.
		{multiTx, multiTxEncoded, pver, 66, io.ErrShortWrite, io.EOF},
		// Force error in transaction output scriptPubKey.
		{multiTx, multiTxEncoded, pver, 67, io.ErrShortWrite, io.EOF},
		// Force error in transaction output lock time.
		{multiTx, multiTxEncoded, pver, 210, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to appmessage format.
		w := newFixedWriter(test.max)
		err := test.in.KaspaEncode(w, test.pver)
		if !errors.Is(err, test.writeErr) {
			t.Errorf("KaspaEncode #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Decode from appmessage format.
		var msg MsgTx
		r := newFixedReader(test.max, test.buf)
		err = msg.KaspaDecode(r, test.pver)
		if !errors.Is(err, test.readErr) {
			t.Errorf("KaspaDecode #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestTxSerialize tests MsgTx serialize and deserialize.
func TestTxSerialize(t *testing.T) {
	noTx := NewNativeMsgTx(1, nil, nil)
	noTxEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x00,                                           // Varint for number of input transactions
		0x00,                                           // Varint for number of output transactions
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Lock time
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // Sub Network ID
	}

	registryTx := NewRegistryMsgTx(1, nil, nil, 16)
	registryTxEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x00,                                           // Varint for number of input transactions
		0x00,                                           // Varint for number of output transactions
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Lock time
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // Sub Network ID
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Gas
		0x77, 0x56, 0x36, 0xb4, 0x89, 0x32, 0xe9, 0xa8,
		0xbb, 0x67, 0xe6, 0x54, 0x84, 0x36, 0x93, 0x8d,
		0x9f, 0xc5, 0x62, 0x49, 0x79, 0x5c, 0x0d, 0x0a,
		0x86, 0xaf, 0x7c, 0x5d, 0x54, 0x45, 0x4c, 0x4b, // Payload hash
		0x08,                                           // Payload length varint
		0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Payload / Gas limit
	}

	subnetworkTx := NewSubnetworkMsgTx(1, nil, nil, &externalapi.DomainSubnetworkID{0xff}, 5, []byte{0, 1, 2})

	subnetworkTxEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x00,                                           // Varint for number of input transactions
		0x00,                                           // Varint for number of output transactions
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Lock time
		0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // Sub Network ID
		0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Gas
		0x35, 0xf9, 0xf2, 0x93, 0x0e, 0xa3, 0x44, 0x61,
		0x88, 0x22, 0x79, 0x5e, 0xee, 0xc5, 0x68, 0xae,
		0x67, 0xab, 0x29, 0x87, 0xd8, 0xb1, 0x9e, 0x45,
		0x91, 0xe1, 0x05, 0x27, 0xba, 0xa1, 0xdf, 0x3d, // Payload hash
		0x03,             // Payload length varint
		0x00, 0x01, 0x02, // Payload
	}

	tests := []struct {
		name             string
		in               *MsgTx // Message to encode
		out              *MsgTx // Expected decoded message
		buf              []byte // Serialized data
		scriptPubKeyLocs []int  // Expected output script locations
	}{
		// No transactions.
		{
			"noTx",
			noTx,
			noTx,
			noTxEncoded,
			nil,
		},

		// Registry Transaction.
		{
			"registryTx",
			registryTx,
			registryTx,
			registryTxEncoded,
			nil,
		},

		// Sub Network Transaction.
		{
			"subnetworkTx",
			subnetworkTx,
			subnetworkTx,
			subnetworkTxEncoded,
			nil,
		},

		// Multiple transactions.
		{
			"multiTx",
			multiTx,
			multiTx,
			multiTxEncoded,
			multiTxScriptPubKeyLocs,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Serialize the transaction.
		var buf bytes.Buffer
		err := test.in.Serialize(&buf)
		if err != nil {
			t.Errorf("Serialize %s: error %v", test.name, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("Serialize %s:\n got: %s want: %s", test.name,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Deserialize the transaction.
		var tx MsgTx
		rbuf := bytes.NewReader(test.buf)
		err = tx.Deserialize(rbuf)
		if err != nil {
			t.Errorf("Deserialize #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&tx, test.out) {
			t.Errorf("Deserialize #%d\n got: %s want: %s", i,
				spew.Sdump(&tx), spew.Sdump(test.out))
			continue
		}

		// Ensure the public key script locations are accurate.
		scriptPubKeyLocs := test.in.ScriptPubKeyLocs()
		if !reflect.DeepEqual(scriptPubKeyLocs, test.scriptPubKeyLocs) {
			t.Errorf("ScriptPubKeyLocs #%d\n got: %s want: %s", i,
				spew.Sdump(scriptPubKeyLocs),
				spew.Sdump(test.scriptPubKeyLocs))
			continue
		}
		for j, loc := range scriptPubKeyLocs {
			wantScriptPubKey := test.in.TxOut[j].ScriptPubKey
			gotScriptPubKey := test.buf[loc : loc+len(wantScriptPubKey)]
			if !bytes.Equal(gotScriptPubKey, wantScriptPubKey) {
				t.Errorf("ScriptPubKeyLocs #%d:%d\n unexpected "+
					"script got: %s want: %s", i, j,
					spew.Sdump(gotScriptPubKey),
					spew.Sdump(wantScriptPubKey))
			}
		}
	}
}

// TestTxSerializeErrors performs negative tests against appmessage encode and decode
// of MsgTx to confirm error paths work correctly.
func TestTxSerializeErrors(t *testing.T) {
	tests := []struct {
		in       *MsgTx // Value to encode
		buf      []byte // Serialized data
		max      int    // Max size of fixed buffer to induce errors
		writeErr error  // Expected write error
		readErr  error  // Expected read error
	}{
		// Force error in version.
		{multiTx, multiTxEncoded, 0, io.ErrShortWrite, io.EOF},
		// Force error in number of transaction inputs.
		{multiTx, multiTxEncoded, 4, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block hash.
		{multiTx, multiTxEncoded, 5, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block output index.
		{multiTx, multiTxEncoded, 37, io.ErrShortWrite, io.EOF},
		// Force error in transaction input signature script length.
		{multiTx, multiTxEncoded, 41, io.ErrShortWrite, io.EOF},
		// Force error in transaction input signature script.
		{multiTx, multiTxEncoded, 42, io.ErrShortWrite, io.EOF},
		// Force error in transaction input sequence.
		{multiTx, multiTxEncoded, 49, io.ErrShortWrite, io.EOF},
		// Force error in number of transaction outputs.
		{multiTx, multiTxEncoded, 57, io.ErrShortWrite, io.EOF},
		// Force error in transaction output value.
		{multiTx, multiTxEncoded, 58, io.ErrShortWrite, io.EOF},
		// Force error in transaction output scriptPubKey length.
		{multiTx, multiTxEncoded, 66, io.ErrShortWrite, io.EOF},
		// Force error in transaction output scriptPubKey.
		{multiTx, multiTxEncoded, 67, io.ErrShortWrite, io.EOF},
		// Force error in transaction output lock time.
		{multiTx, multiTxEncoded, 210, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Serialize the transaction.
		w := newFixedWriter(test.max)
		err := test.in.Serialize(w)
		if !errors.Is(err, test.writeErr) {
			t.Errorf("Serialize #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Deserialize the transaction.
		var tx MsgTx
		r := newFixedReader(test.max, test.buf)
		err = tx.Deserialize(r)
		if !errors.Is(err, test.readErr) {
			t.Errorf("Deserialize #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}

	registryTx := NewSubnetworkMsgTx(1, nil, nil, &subnetworks.SubnetworkIDRegistry, 1, nil)

	w := bytes.NewBuffer(make([]byte, 0, registryTx.SerializeSize()))
	err := registryTx.Serialize(w)
	str := "Transactions from built-in should have 0 gas"
	expectedErr := messageError("MsgTx.KaspaEncode", str)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("TestTxSerializeErrors: expected error %v but got %v", expectedErr, err)
	}

	nativeTx := NewSubnetworkMsgTx(1, nil, nil, &subnetworks.SubnetworkIDNative, 1, nil)
	w = bytes.NewBuffer(make([]byte, 0, registryTx.SerializeSize()))
	err = nativeTx.Serialize(w)

	str = "Transactions from native subnetwork should have 0 gas"
	expectedErr = messageError("MsgTx.KaspaEncode", str)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("TestTxSerializeErrors: expected error %v but got %v", expectedErr, err)
	}

	nativeTx.Gas = 0
	nativeTx.Payload = []byte{1, 2, 3}
	nativeTx.PayloadHash = hashes.HashData(nativeTx.Payload)
	w = bytes.NewBuffer(make([]byte, 0, registryTx.SerializeSize()))
	err = nativeTx.Serialize(w)

	str = "Transactions from native subnetwork should have <nil> payload"
	expectedErr = messageError("MsgTx.KaspaEncode", str)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("TestTxSerializeErrors: expected error %v but got %v", expectedErr, err)
	}
}

// TestTxOverflowErrors performs tests to ensure deserializing transactions
// which are intentionally crafted to use large values for the variable number
// of inputs and outputs are handled properly. This could otherwise potentially
// be used as an attack vector.
func TestTxOverflowErrors(t *testing.T) {
	pver := ProtocolVersion
	txVer := uint32(1)

	tests := []struct {
		buf     []byte // Encoded value
		pver    uint32 // Protocol version for appmessage encoding
		version uint32 // Transaction version
		err     error  // Expected error
	}{
		// Transaction that claims to have ~uint64(0) inputs.
		{
			[]byte{
				0x00, 0x00, 0x00, 0x01, // Version
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // Varint for number of input transactions
			}, pver, txVer, &MessageError{},
		},

		// Transaction that claims to have ~uint64(0) outputs.
		{
			[]byte{
				0x00, 0x00, 0x00, 0x01, // Version
				0x00, // Varint for number of input transactions
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // Varint for number of output transactions
			}, pver, txVer, &MessageError{},
		},

		// Transaction that has an input with a signature script that
		// claims to have ~uint64(0) length.
		{
			[]byte{
				0x00, 0x00, 0x00, 0x01, // Version
				0x01, // Varint for number of input transactions
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Previous output hash
				0xff, 0xff, 0xff, 0xff, // Prevous output index
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // Varint for length of signature script
			}, pver, txVer, &MessageError{},
		},

		// Transaction that has an output with a public key script
		// that claims to have ~uint64(0) length.
		{
			[]byte{
				0x00, 0x00, 0x00, 0x01, // Version
				0x01, // Varint for number of input transactions
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Previous output hash
				0xff, 0xff, 0xff, 0xff, // Prevous output index
				0x00,                                           // Varint for length of signature script
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // Sequence
				0x01,                                           // Varint for number of output transactions
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Transaction amount
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // Varint for length of public key script
			}, pver, txVer, &MessageError{},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Decode from appmessage format.
		var msg MsgTx
		r := bytes.NewReader(test.buf)
		err := msg.KaspaDecode(r, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("KaspaDecode #%d wrong error got: %v, want: %v",
				i, err, reflect.TypeOf(test.err))
			continue
		}

		// Decode from appmessage format.
		r = bytes.NewReader(test.buf)
		err = msg.Deserialize(r)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Deserialize #%d wrong error got: %v, want: %v",
				i, err, reflect.TypeOf(test.err))
			continue
		}
	}
}

// TestTxSerializeSize performs tests to ensure the serialize size for
// various transactions is accurate.
func TestTxSerializeSize(t *testing.T) {
	// Empty tx message.
	noTx := NewNativeMsgTx(1, nil, nil)

	tests := []struct {
		in   *MsgTx // Tx to encode
		size int    // Expected serialized size
	}{
		// No inputs or outpus.
		{noTx, 34},

		// Transcaction with an input and an output.
		{multiTx, 238},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		serializedSize := test.in.SerializeSize()
		if serializedSize != test.size {
			t.Errorf("MsgTx.SerializeSize: #%d got: %d, want: %d", i,
				serializedSize, test.size)
			continue
		}
	}
}

func TestIsSubnetworkCompatible(t *testing.T) {
	testTx := NewSubnetworkMsgTx(1, nil, nil, &externalapi.DomainSubnetworkID{123}, 0, []byte{})
	tests := []struct {
		name           string
		subnetworkID   *externalapi.DomainSubnetworkID
		expectedResult bool
	}{
		{
			name:           "Native subnetwork",
			subnetworkID:   &subnetworks.SubnetworkIDNative,
			expectedResult: true,
		},
		{
			name:           "same subnetwork as test tx",
			subnetworkID:   &externalapi.DomainSubnetworkID{123},
			expectedResult: true,
		},
		{
			name:           "other subnetwork",
			subnetworkID:   &externalapi.DomainSubnetworkID{234},
			expectedResult: false,
		},
	}

	for _, test := range tests {
		result := testTx.IsSubnetworkCompatible(test.subnetworkID)
		if result != test.expectedResult {
			t.Errorf("IsSubnetworkCompatible got unexpected result in test '%s': "+
				"expected: %t, want: %t", test.name, test.expectedResult, result)
		}
	}
}

func TestScriptFreeList(t *testing.T) {
	var list scriptFreeList = make(chan []byte, freeListMaxItems)

	expectedCapacity := 512
	expectedLengthFirst := 12
	expectedLengthSecond := 13

	first := list.Borrow(uint64(expectedLengthFirst))
	if cap(first) != expectedCapacity {
		t.Errorf("MsgTx.TestScriptFreeList: Expected capacity for first %d, but got %d",
			expectedCapacity, cap(first))
	}
	if len(first) != expectedLengthFirst {
		t.Errorf("MsgTx.TestScriptFreeList: Expected length for first %d, but got %d",
			expectedLengthFirst, len(first))
	}
	list.Return(first)

	// Borrow again, and check that the underlying array is re-used for second
	second := list.Borrow(uint64(expectedLengthSecond))
	if cap(second) != expectedCapacity {
		t.Errorf("MsgTx.TestScriptFreeList: Expected capacity for second %d, but got %d",
			expectedCapacity, cap(second))
	}
	if len(second) != expectedLengthSecond {
		t.Errorf("MsgTx.TestScriptFreeList: Expected length for second %d, but got %d",
			expectedLengthSecond, len(second))
	}

	firstArrayAddress := underlyingArrayAddress(first)
	secondArrayAddress := underlyingArrayAddress(second)

	if firstArrayAddress != secondArrayAddress {
		t.Errorf("First underlying array is at address %d and second at address %d, "+
			"which means memory was not re-used", firstArrayAddress, secondArrayAddress)
	}

	list.Return(second)

	// test for buffers bigger than freeListMaxScriptSize
	expectedCapacityBig := freeListMaxScriptSize + 1
	expectedLengthBig := expectedCapacityBig
	big := list.Borrow(uint64(expectedCapacityBig))

	if cap(big) != expectedCapacityBig {
		t.Errorf("MsgTx.TestScriptFreeList: Expected capacity for second %d, but got %d",
			expectedCapacityBig, cap(big))
	}
	if len(big) != expectedLengthBig {
		t.Errorf("MsgTx.TestScriptFreeList: Expected length for second %d, but got %d",
			expectedLengthBig, len(big))
	}

	list.Return(big)

	// test there's no crash when channel is full because borrowed too much
	buffers := make([][]byte, freeListMaxItems+1)
	for i := 0; i < freeListMaxItems+1; i++ {
		buffers[i] = list.Borrow(1)
	}
	for i := 0; i < freeListMaxItems+1; i++ {
		list.Return(buffers[i])
	}
}

func underlyingArrayAddress(buf []byte) uint64 {
	return uint64((*reflect.SliceHeader)(unsafe.Pointer(&buf)).Data)
}

// multiTx is a MsgTx with an input and output and used in various tests.
var multiTxIns = []*TxIn{
	{
		PreviousOutpoint: Outpoint{
			TxID:  externalapi.DomainTransactionID{},
			Index: 0xffffffff,
		},
		SignatureScript: []byte{
			0x04, 0x31, 0xdc, 0x00, 0x1b, 0x01, 0x62,
		},
		Sequence: math.MaxUint64,
	},
}
var multiTxOuts = []*TxOut{
	{
		Value: 0x12a05f200,
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
	},
	{
		Value: 0x5f5e100,
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
	},
}
var multiTx = NewNativeMsgTx(1, multiTxIns, multiTxOuts)

// multiTxEncoded is the appmessage encoded bytes for multiTx using protocol version
// 60002 and is used in the various tests.
var multiTxEncoded = []byte{
	0x01, 0x00, 0x00, 0x00, // Version
	0x01, // Varint for number of input transactions
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Previous output hash
	0xff, 0xff, 0xff, 0xff, // Prevous output index
	0x07,                                     // Varint for length of signature script
	0x04, 0x31, 0xdc, 0x00, 0x1b, 0x01, 0x62, // Signature script
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // Sequence
	0x02,                                           // Varint for number of output transactions
	0x00, 0xf2, 0x05, 0x2a, 0x01, 0x00, 0x00, 0x00, // Transaction amount
	0x43, // Varint for length of scriptPubKey
	0x41, // OP_DATA_65
	0x04, 0xd6, 0x4b, 0xdf, 0xd0, 0x9e, 0xb1, 0xc5,
	0xfe, 0x29, 0x5a, 0xbd, 0xeb, 0x1d, 0xca, 0x42,
	0x81, 0xbe, 0x98, 0x8e, 0x2d, 0xa0, 0xb6, 0xc1,
	0xc6, 0xa5, 0x9d, 0xc2, 0x26, 0xc2, 0x86, 0x24,
	0xe1, 0x81, 0x75, 0xe8, 0x51, 0xc9, 0x6b, 0x97,
	0x3d, 0x81, 0xb0, 0x1c, 0xc3, 0x1f, 0x04, 0x78,
	0x34, 0xbc, 0x06, 0xd6, 0xd6, 0xed, 0xf6, 0x20,
	0xd1, 0x84, 0x24, 0x1a, 0x6a, 0xed, 0x8b, 0x63,
	0xa6,                                           // 65-byte signature
	0xac,                                           // OP_CHECKSIG
	0x00, 0xe1, 0xf5, 0x05, 0x00, 0x00, 0x00, 0x00, // Transaction amount
	0x43, // Varint for length of scriptPubKey
	0x41, // OP_DATA_65
	0x04, 0xd6, 0x4b, 0xdf, 0xd0, 0x9e, 0xb1, 0xc5,
	0xfe, 0x29, 0x5a, 0xbd, 0xeb, 0x1d, 0xca, 0x42,
	0x81, 0xbe, 0x98, 0x8e, 0x2d, 0xa0, 0xb6, 0xc1,
	0xc6, 0xa5, 0x9d, 0xc2, 0x26, 0xc2, 0x86, 0x24,
	0xe1, 0x81, 0x75, 0xe8, 0x51, 0xc9, 0x6b, 0x97,
	0x3d, 0x81, 0xb0, 0x1c, 0xc3, 0x1f, 0x04, 0x78,
	0x34, 0xbc, 0x06, 0xd6, 0xd6, 0xed, 0xf6, 0x20,
	0xd1, 0x84, 0x24, 0x1a, 0x6a, 0xed, 0x8b, 0x63,
	0xa6,                                           // 65-byte signature
	0xac,                                           // OP_CHECKSIG
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Lock time
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, // Sub Network ID
}

// multiTxScriptPubKeyLocs is the location information for the public key scripts
// located in multiTx.
var multiTxScriptPubKeyLocs = []int{67, 143}
