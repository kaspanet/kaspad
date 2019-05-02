// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"crypto/rand"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/davecgh/go-spew/spew"
)

// TestMerkleBlock tests the MsgMerkleBlock API.
func TestMerkleBlock(t *testing.T) {
	pver := ProtocolVersion

	// Block 1 header.
	parentHashes := blockOne.Header.ParentHashes
	hashMerkleRoot := blockOne.Header.HashMerkleRoot
	idMerkleRoot := blockOne.Header.IDMerkleRoot
	bits := blockOne.Header.Bits
	nonce := blockOne.Header.Nonce
	bh := NewBlockHeader(1, parentHashes, hashMerkleRoot, idMerkleRoot, bits, nonce)

	// Ensure the command is expected value.
	wantCmd := "merkleblock"
	msg := NewMsgMerkleBlock(bh)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgBlock: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	// Num addresses (varInt) + max allowed addresses.
	wantPayload := uint32(1000000)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Load maxTxPerBlock hashes
	data := make([]byte, 32)
	for i := 0; i < maxTxPerBlock; i++ {
		rand.Read(data)
		hash, err := daghash.NewHash(data)
		if err != nil {
			t.Errorf("NewHash failed: %v\n", err)
			return
		}

		if err = msg.AddTxHash(hash); err != nil {
			t.Errorf("AddTxHash failed: %v\n", err)
			return
		}
	}

	// Add one more Tx to test failure.
	rand.Read(data)
	hash, err := daghash.NewHash(data)
	if err != nil {
		t.Errorf("NewHash failed: %v\n", err)
		return
	}

	if err = msg.AddTxHash(hash); err == nil {
		t.Errorf("AddTxHash succeeded when it should have failed")
		return
	}

	// Test encode with latest protocol version.
	var buf bytes.Buffer
	err = msg.BtcEncode(&buf, pver)
	if err != nil {
		t.Errorf("encode of MsgMerkleBlock failed %v err <%v>", msg, err)
	}

	// Test decode with latest protocol version.
	readmsg := MsgMerkleBlock{}
	err = readmsg.BtcDecode(&buf, pver)
	if err != nil {
		t.Errorf("decode of MsgMerkleBlock failed [%v] err <%v>", buf, err)
	}

	// Force extra hash to test maxTxPerBlock.
	msg.Hashes = append(msg.Hashes, hash)
	err = msg.BtcEncode(&buf, pver)
	if err == nil {
		t.Errorf("encode of MsgMerkleBlock succeeded with too many " +
			"tx hashes when it should have failed")
		return
	}

	// Force too many flag bytes to test maxFlagsPerMerkleBlock.
	// Reset the number of hashes back to a valid value.
	msg.Hashes = msg.Hashes[len(msg.Hashes)-1:]
	msg.Flags = make([]byte, maxFlagsPerMerkleBlock+1)
	err = msg.BtcEncode(&buf, pver)
	if err == nil {
		t.Errorf("encode of MsgMerkleBlock succeeded with too many " +
			"flag bytes when it should have failed")
		return
	}
}

// TestMerkleBlockCrossProtocol tests the MsgMerkleBlock API when encoding with
// the latest protocol version and decoding with BIP0031Version.
func TestMerkleBlockCrossProtocol(t *testing.T) {
	// Block 1 header.
	parentHashes := blockOne.Header.ParentHashes
	hashMerkleRoot := blockOne.Header.HashMerkleRoot
	idMerkleRoot := blockOne.Header.IDMerkleRoot
	bits := blockOne.Header.Bits
	nonce := blockOne.Header.Nonce
	bh := NewBlockHeader(1, parentHashes, hashMerkleRoot, idMerkleRoot, bits, nonce)

	msg := NewMsgMerkleBlock(bh)

	// Encode with latest protocol version.
	var buf bytes.Buffer
	err := msg.BtcEncode(&buf, ProtocolVersion)
	if err != nil {
		t.Errorf("encode of NewMsgFilterLoad failed %v err <%v>", msg,
			err)
	}
}

// TestMerkleBlockWire tests the MsgMerkleBlock wire encode and decode for
// various numbers of transaction hashes and protocol versions.
func TestMerkleBlockWire(t *testing.T) {
	tests := []struct {
		in   *MsgMerkleBlock // Message to encode
		out  *MsgMerkleBlock // Expected decoded message
		buf  []byte          // Wire encoding
		pver uint32          // Protocol version for wire encoding
	}{
		// Latest protocol version.
		{
			&merkleBlockOne, &merkleBlockOne, merkleBlockOneBytes, ProtocolVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode the message to wire format.
		var buf bytes.Buffer
		err := test.in.BtcEncode(&buf, test.pver)
		if err != nil {
			t.Errorf("BtcEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("BtcEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the message from wire format.
		var msg MsgMerkleBlock
		rbuf := bytes.NewReader(test.buf)
		err = msg.BtcDecode(rbuf, test.pver)
		if err != nil {
			t.Errorf("BtcDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&msg, test.out) {
			t.Errorf("BtcDecode #%d\n got: %s want: %s", i,
				spew.Sdump(&msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestMerkleBlockWireErrors performs negative tests against wire encode and
// decode of MsgBlock to confirm error paths work correctly.
func TestMerkleBlockWireErrors(t *testing.T) {
	pver := ProtocolVersion

	tests := []struct {
		in       *MsgMerkleBlock // Value to encode
		buf      []byte          // Wire encoding
		pver     uint32          // Protocol version for wire encoding
		max      int             // Max size of fixed buffer to induce errors
		writeErr error           // Expected write error
		readErr  error           // Expected read error
	}{
		// Force error in version.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in num prev hashes.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 4, io.ErrShortWrite, io.EOF},
		// Force error in prev block hash #1.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 5, io.ErrShortWrite, io.EOF},
		// Force error in prev block hash #2.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 37, io.ErrShortWrite, io.EOF},
		// Force error in hash merkle root.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 69, io.ErrShortWrite, io.EOF},
		// Force error in hash merkle root.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 101, io.ErrShortWrite, io.EOF},
		// Force error in timestamp.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 133, io.ErrShortWrite, io.EOF},
		// Force error in difficulty bits.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 141, io.ErrShortWrite, io.EOF},
		// Force error in header nonce.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 145, io.ErrShortWrite, io.EOF},
		// Force error in transaction count.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 153, io.ErrShortWrite, io.EOF},
		// Force error in num hashes.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 157, io.ErrShortWrite, io.EOF},
		// Force error in hashes.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 158, io.ErrShortWrite, io.EOF},
		// Force error in num flag bytes.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 190, io.ErrShortWrite, io.EOF},
		// Force error in flag bytes.
		{&merkleBlockOne, merkleBlockOneBytes, pver, 191, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := test.in.BtcEncode(w, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.writeErr) {
			t.Errorf("BtcEncode #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// For errors which are not of type MessageError, check them for
		// equality.
		if _, ok := err.(*MessageError); !ok {
			if err != test.writeErr {
				t.Errorf("BtcEncode #%d wrong error got: %v, "+
					"want: %v", i, err, test.writeErr)
				continue
			}
		}

		// Decode from wire format.
		var msg MsgMerkleBlock
		r := newFixedReader(test.max, test.buf)
		err = msg.BtcDecode(r, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.readErr) {
			t.Errorf("BtcDecode #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}

		// For errors which are not of type MessageError, check them for
		// equality.
		if _, ok := err.(*MessageError); !ok {
			if err != test.readErr {
				t.Errorf("BtcDecode #%d wrong error got: %v, "+
					"want: %v", i, err, test.readErr)
				continue
			}
		}
	}
}

// TestMerkleBlockOverflowErrors performs tests to ensure encoding and decoding
// merkle blocks that are intentionally crafted to use large values for the
// number of hashes and flags are handled properly.  This could otherwise
// potentially be used as an attack vector.
func TestMerkleBlockOverflowErrors(t *testing.T) {
	pver := ProtocolVersion

	// Create bytes for a merkle block that claims to have more than the max
	// allowed tx hashes.
	var buf bytes.Buffer
	WriteVarInt(&buf, maxTxPerBlock+1)
	numHashesOffset := 157
	exceedMaxHashes := make([]byte, numHashesOffset)
	copy(exceedMaxHashes, merkleBlockOneBytes[:numHashesOffset])
	exceedMaxHashes = append(exceedMaxHashes, buf.Bytes()...)

	// Create bytes for a merkle block that claims to have more than the max
	// allowed flag bytes.
	buf.Reset()
	WriteVarInt(&buf, maxFlagsPerMerkleBlock+1)
	numFlagBytesOffset := 190
	exceedMaxFlagBytes := make([]byte, numFlagBytesOffset)
	copy(exceedMaxFlagBytes, merkleBlockOneBytes[:numFlagBytesOffset])
	exceedMaxFlagBytes = append(exceedMaxFlagBytes, buf.Bytes()...)

	tests := []struct {
		buf  []byte // Wire encoding
		pver uint32 // Protocol version for wire encoding
		err  error  // Expected error
	}{
		// Block that claims to have more than max allowed hashes.
		{exceedMaxHashes, pver, &MessageError{}},
		// Block that claims to have more than max allowed flag bytes.
		{exceedMaxFlagBytes, pver, &MessageError{}},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Decode from wire format.
		var msg MsgMerkleBlock
		r := bytes.NewReader(test.buf)
		err := msg.BtcDecode(r, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("BtcDecode #%d wrong error got: %v, want: %v",
				i, err, reflect.TypeOf(test.err))
			continue
		}
	}
}

// merkleBlockOne is a merkle block created from block one of the block chain
// where the first transaction matches.
var merkleBlockOne = MsgMerkleBlock{
	Header: BlockHeader{
		Version:      1,
		ParentHashes: []*daghash.Hash{mainNetGenesisHash, simNetGenesisHash},
		HashMerkleRoot: &daghash.Hash{
			0x98, 0x20, 0x51, 0xfd, 0x1e, 0x4b, 0xa7, 0x44,
			0xbb, 0xbe, 0x68, 0x0e, 0x1f, 0xee, 0x14, 0x67,
			0x7b, 0xa1, 0xa3, 0xc3, 0x54, 0x0b, 0xf7, 0xb1,
			0xcd, 0xb6, 0x06, 0xe8, 0x57, 0x23, 0x3e, 0x0e,
		},
		IDMerkleRoot: exampleIDMerkleRoot,
		Timestamp:    time.Unix(0x4966bc61, 0), // 2009-01-08 20:54:25 -0600 CST
		Bits:         0x1d00ffff,               // 486604799
		Nonce:        0x9962e301,               // 2573394689
	},
	Transactions: 1,
	Hashes: []*daghash.Hash{
		(*daghash.Hash)(&[daghash.HashSize]byte{
			0x98, 0x20, 0x51, 0xfd, 0x1e, 0x4b, 0xa7, 0x44,
			0xbb, 0xbe, 0x68, 0x0e, 0x1f, 0xee, 0x14, 0x67,
			0x7b, 0xa1, 0xa3, 0xc3, 0x54, 0x0b, 0xf7, 0xb1,
			0xcd, 0xb6, 0x06, 0xe8, 0x57, 0x23, 0x3e, 0x0e,
		}),
	},
	Flags: []byte{0x80},
}

// merkleBlockOneBytes is the serialized bytes for a merkle block created from
// block one of the block chain where the first transaction matches.
var merkleBlockOneBytes = []byte{
	0x01, 0x00, 0x00, 0x00, // Version 1
	0x02,                                           // NumParentBlocks
	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, // mainNetGenesisHash
	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
	0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a, // simNetGenesisHash
	0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
	0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
	0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
	0x98, 0x20, 0x51, 0xfd, 0x1e, 0x4b, 0xa7, 0x44, // MerkleRoot
	0xbb, 0xbe, 0x68, 0x0e, 0x1f, 0xee, 0x14, 0x67,
	0x7b, 0xa1, 0xa3, 0xc3, 0x54, 0x0b, 0xf7, 0xb1,
	0xcd, 0xb6, 0x06, 0xe8, 0x57, 0x23, 0x3e, 0x0e,
	0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63, // IDMerkleRoot
	0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
	0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C,
	0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
	0x61, 0xbc, 0x66, 0x49, 0x00, 0x00, 0x00, 0x00, // Timestamp
	0xff, 0xff, 0x00, 0x1d, // Bits
	0x01, 0xe3, 0x62, 0x99, 0x00, 0x00, 0x00, 0x00, // Fake Nonce. TODO: (Ori) Replace to a real nonce
	0x01, 0x00, 0x00, 0x00, // TxnCount
	0x01, // Num hashes
	0x98, 0x20, 0x51, 0xfd, 0x1e, 0x4b, 0xa7, 0x44,
	0xbb, 0xbe, 0x68, 0x0e, 0x1f, 0xee, 0x14, 0x67,
	0x7b, 0xa1, 0xa3, 0xc3, 0x54, 0x0b, 0xf7, 0xb1,
	0xcd, 0xb6, 0x06, 0xe8, 0x57, 0x23, 0x3e, 0x0e, // Hash
	0x01, // Num flag bytes
	0x80, // Flags
}
