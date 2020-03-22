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
	"github.com/kaspanet/kaspad/util/random"
)

// TestPongLatest tests the MsgPong API against the latest protocol version.
func TestPongLatest(t *testing.T) {
	pver := ProtocolVersion

	nonce, err := random.Uint64()
	if err != nil {
		t.Errorf("random.Uint64: error generating nonce: %v", err)
	}
	msg := NewMsgPong(nonce)
	if msg.Nonce != nonce {
		t.Errorf("NewMsgPong: wrong nonce - got %v, want %v",
			msg.Nonce, nonce)
	}

	// Ensure the command is expected value.
	wantCmd := "pong"
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgPong: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	wantPayload := uint32(8)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Test encode with latest protocol version.
	var buf bytes.Buffer
	err = msg.KaspaEncode(&buf, pver)
	if err != nil {
		t.Errorf("encode of MsgPong failed %v err <%v>", msg, err)
	}

	// Test decode with latest protocol version.
	readmsg := NewMsgPong(0)
	err = readmsg.KaspaDecode(&buf, pver)
	if err != nil {
		t.Errorf("decode of MsgPong failed [%v] err <%v>", buf, err)
	}

	// Ensure nonce is the same.
	if msg.Nonce != readmsg.Nonce {
		t.Errorf("Should get same nonce for protocol version %d", pver)
	}
}

// TestPongCrossProtocol tests the MsgPong API when encoding with the latest
// protocol version and decoding with BIP0031Version.
func TestPongCrossProtocol(t *testing.T) {
	nonce, err := random.Uint64()
	if err != nil {
		t.Errorf("Error generating nonce: %v", err)
	}
	msg := NewMsgPong(nonce)
	if msg.Nonce != nonce {
		t.Errorf("Should get same nonce back out.")
	}

	// Encode with latest protocol version.
	var buf bytes.Buffer
	err = msg.KaspaEncode(&buf, ProtocolVersion)
	if err != nil {
		t.Errorf("encode of MsgPong failed %v err <%v>", msg, err)
	}
}

// TestPongWire tests the MsgPong wire encode and decode for various protocol
// versions.
func TestPongWire(t *testing.T) {
	tests := []struct {
		in   MsgPong // Message to encode
		out  MsgPong // Expected decoded message
		buf  []byte  // Wire encoding
		pver uint32  // Protocol version for wire encoding
	}{
		// Latest protocol version.
		{
			MsgPong{Nonce: 123123}, // 0x1e0f3
			MsgPong{Nonce: 123123}, // 0x1e0f3
			[]byte{0xf3, 0xe0, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
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
		var msg MsgPong
		rbuf := bytes.NewReader(test.buf)
		err = msg.KaspaDecode(rbuf, test.pver)
		if err != nil {
			t.Errorf("KaspaDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(msg, test.out) {
			t.Errorf("KaspaDecode #%d\n got: %s want: %s", i,
				spew.Sdump(msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestPongWireErrors performs negative tests against wire encode and decode
// of MsgPong to confirm error paths work correctly.
func TestPongWireErrors(t *testing.T) {
	pver := ProtocolVersion

	basePong := NewMsgPong(123123) // 0x1e0f3
	basePongEncoded := []byte{
		0xf3, 0xe0, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	tests := []struct {
		in       *MsgPong // Value to encode
		buf      []byte   // Wire encoding
		pver     uint32   // Protocol version for wire encoding
		max      int      // Max size of fixed buffer to induce errors
		writeErr error    // Expected write error
		readErr  error    // Expected read error
	}{
		// Latest protocol version with intentional read/write errors.
		// Force error in nonce.
		{basePong, basePongEncoded, pver, 0, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := test.in.KaspaEncode(w, test.pver)

		// For errors which are not of type MessageError, check them for
		// equality. If the error is a MessageError, check only if it's
		// the expected type.
		if msgErr := &(MessageError{}); !errors.As(err, &msgErr) {
			if !errors.Is(err, test.writeErr) {
				t.Errorf("KaspaEncode #%d wrong error got: %v, "+
					"want: %v", i, err, test.writeErr)
				continue
			}
		} else if reflect.TypeOf(msgErr) != reflect.TypeOf(test.writeErr) {
			t.Errorf("ReadMessage #%d wrong error type got: %T, "+
				"want: %T", i, msgErr, test.writeErr)
			continue
		}

		// Decode from wire format.
		var msg MsgPong
		r := newFixedReader(test.max, test.buf)
		err = msg.KaspaDecode(r, test.pver)

		// For errors which are not of type MessageError, check them for
		// equality. If the error is a MessageError, check only if it's
		// the expected type.
		if msgErr := &(MessageError{}); !errors.As(err, &msgErr) {
			if !errors.Is(err, test.readErr) {
				t.Errorf("KaspaDecode #%d wrong error got: %v, "+
					"want: %v", i, err, test.readErr)
				continue
			}
		} else if reflect.TypeOf(msgErr) != reflect.TypeOf(test.readErr) {
			t.Errorf("ReadMessage #%d wrong error type got: %T, "+
				"want: %T", i, msgErr, test.readErr)
			continue
		}

	}
}
