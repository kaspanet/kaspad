// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/util/daghash"
)

// TestGetBlockInvs tests the MsgGetBlockInvs API.
func TestGetBlockInvs(t *testing.T) {
	pver := ProtocolVersion

	hashStr := "000000000002e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	lowHash, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	hashStr = "3ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"
	highHash, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// Ensure we get the same data back out.
	msg := NewMsgGetBlockInvs(lowHash, highHash)
	if !msg.HighHash.IsEqual(highHash) {
		t.Errorf("NewMsgGetBlockInvs: wrong high hash - got %v, want %v",
			msg.HighHash, highHash)
	}

	// Ensure the command is expected value.
	wantCmd := "getblockinvs"
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgGetBlockInvs: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is low hash (32 bytes) + high hash (32 bytes).
	wantPayload := uint32(64)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}
}

// TestGetBlockInvsWire tests the MsgGetBlockInvs wire encode and decode for various
// numbers of block locator hashes and protocol versions.
func TestGetBlockInvsWire(t *testing.T) {
	hashStr := "2710f40c87ec93d010a6fd95f42c59a2cbacc60b18cf6b7957535"
	lowHash, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	hashStr = "3ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"
	highHash, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// MsgGetBlocks message with no start or high hash.
	noStartOrStop := NewMsgGetBlockInvs(&daghash.Hash{}, &daghash.Hash{})
	noStartOrStopEncoded := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Low hash
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // High hash
	}

	// MsgGetBlockInvs message with a low hash and a high hash.
	withLowAndHighHash := NewMsgGetBlockInvs(lowHash, highHash)
	withLowAndHighHashEncoded := []byte{
		0x35, 0x75, 0x95, 0xb7, 0xf6, 0x8c, 0xb1, 0x60,
		0xcc, 0xba, 0x2c, 0x9a, 0xc5, 0x42, 0x5f, 0xd9,
		0x6f, 0x0a, 0x01, 0x3d, 0xc9, 0x7e, 0xc8, 0x40,
		0x0f, 0x71, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // Low hash
		0x06, 0xe5, 0x33, 0xfd, 0x1a, 0xda, 0x86, 0x39,
		0x1f, 0x3f, 0x6c, 0x34, 0x32, 0x04, 0xb0, 0xd2,
		0x78, 0xd4, 0xaa, 0xec, 0x1c, 0x0b, 0x20, 0xaa,
		0x27, 0xba, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, // High hash
	}

	tests := []struct {
		in   *MsgGetBlockInvs // Message to encode
		out  *MsgGetBlockInvs // Expected decoded message
		buf  []byte           // Wire encoding
		pver uint32           // Protocol version for wire encoding
	}{
		// Latest protocol version with no block locators.
		{
			noStartOrStop,
			noStartOrStop,
			noStartOrStopEncoded,
			ProtocolVersion,
		},

		// Latest protocol version with multiple block locators.
		{
			withLowAndHighHash,
			withLowAndHighHash,
			withLowAndHighHashEncoded,
			ProtocolVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode the message to wire format.
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

		// Decode the message from wire format.
		var msg MsgGetBlockInvs
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

// TestGetBlockInvsWireErrors performs negative tests against wire encode and
// decode of MsgGetBlockInvs to confirm error paths work correctly.
func TestGetBlockInvsWireErrors(t *testing.T) {
	// Set protocol inside getheaders message.
	pver := ProtocolVersion

	hashStr := "2710f40c87ec93d010a6fd95f42c59a2cbacc60b18cf6b7957535"
	lowHash, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	hashStr = "3ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"
	highHash, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// MsgGetBlockInvs message with multiple block locators and a high hash.
	baseGetBlocks := NewMsgGetBlockInvs(lowHash, highHash)
	baseGetBlocksEncoded := []byte{
		0x35, 0x75, 0x95, 0xb7, 0xf6, 0x8c, 0xb1, 0x60,
		0xcc, 0xba, 0x2c, 0x9a, 0xc5, 0x42, 0x5f, 0xd9,
		0x6f, 0x0a, 0x01, 0x3d, 0xc9, 0x7e, 0xc8, 0x40,
		0x0f, 0x71, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // Low hash
		0x06, 0xe5, 0x33, 0xfd, 0x1a, 0xda, 0x86, 0x39,
		0x1f, 0x3f, 0x6c, 0x34, 0x32, 0x04, 0xb0, 0xd2,
		0x78, 0xd4, 0xaa, 0xec, 0x1c, 0x0b, 0x20, 0xaa,
		0x27, 0xba, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, // High hash
	}

	tests := []struct {
		in       *MsgGetBlockInvs // Value to encode
		buf      []byte           // Wire encoding
		pver     uint32           // Protocol version for wire encoding
		max      int              // Max size of fixed buffer to induce errors
		writeErr error            // Expected write error
		readErr  error            // Expected read error
	}{
		// Force error in low hash.
		{baseGetBlocks, baseGetBlocksEncoded, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in high hash.
		{baseGetBlocks, baseGetBlocksEncoded, pver, 32, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := test.in.KaspaEncode(w, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.writeErr) {
			t.Errorf("KaspaEncode #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// For errors which are not of type MessageError, check them for
		// equality. If the error is a MessageError, check only if it's
		// the expected type.
		if msgErr := &(MessageError{}); !errors.As(err, &msgErr) {
			if err != test.writeErr {
				t.Errorf("KaspaEncode #%d wrong error got: %v, "+
					"want: %v", i, err, test.writeErr)
				continue
			}
		}

		// Decode from wire format.
		var msg MsgGetBlockInvs
		r := newFixedReader(test.max, test.buf)
		err = msg.KaspaDecode(r, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.readErr) {
			t.Errorf("KaspaDecode #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}

		// For errors which are not of type MessageError, check them for
		// equality. If the error is a MessageError, check only if it's
		// the expected type.
		if msgErr := &(MessageError{}); !errors.As(err, &msgErr) {
			if err != test.readErr {
				t.Errorf("KaspaDecode #%d wrong error got: %v, "+
					"want: %v", i, err, test.readErr)
				continue
			}
		}
	}
}
