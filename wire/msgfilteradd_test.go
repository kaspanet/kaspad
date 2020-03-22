// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"reflect"
	"testing"
)

// TestFilterAddLatest tests the MsgFilterAdd API against the latest protocol
// version.
func TestFilterAddLatest(t *testing.T) {
	pver := ProtocolVersion

	data := []byte{0x01, 0x02}
	msg := NewMsgFilterAdd(data)

	// Ensure the command is expected value.
	wantCmd := "filteradd"
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgFilterAdd: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	wantPayload := uint32(523)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Test encode with latest protocol version.
	var buf bytes.Buffer
	err := msg.KaspaEncode(&buf, pver)
	if err != nil {
		t.Errorf("encode of MsgFilterAdd failed %v err <%v>", msg, err)
	}

	// Test decode with latest protocol version.
	var readmsg MsgFilterAdd
	err = readmsg.KaspaDecode(&buf, pver)
	if err != nil {
		t.Errorf("decode of MsgFilterAdd failed [%v] err <%v>", buf, err)
	}
}

// TestFilterAddCrossProtocol tests the MsgFilterAdd API.
func TestFilterAddCrossProtocol(t *testing.T) {
	data := []byte{0x01, 0x02}
	msg := NewMsgFilterAdd(data)
	if !bytes.Equal(msg.Data, data) {
		t.Errorf("should get same data back out")
	}

	// Encode with latest protocol version.
	var buf bytes.Buffer
	err := msg.KaspaEncode(&buf, ProtocolVersion)
	if err != nil {
		t.Errorf("encode of MsgFilterAdd failed %v err <%v>", msg, err)
	}

}

// TestFilterAddMaxDataSize tests the MsgFilterAdd API maximum data size.
func TestFilterAddMaxDataSize(t *testing.T) {
	data := bytes.Repeat([]byte{0xff}, 521)
	msg := NewMsgFilterAdd(data)

	// Encode with latest protocol version.
	var buf bytes.Buffer
	err := msg.KaspaEncode(&buf, ProtocolVersion)
	if err == nil {
		t.Errorf("encode of MsgFilterAdd succeeded when it shouldn't "+
			"have %v", msg)
	}

	// Decode with latest protocol version.
	readbuf := bytes.NewReader(data)
	err = msg.KaspaDecode(readbuf, ProtocolVersion)
	if err == nil {
		t.Errorf("decode of MsgFilterAdd succeeded when it shouldn't "+
			"have %v", msg)
	}
}

// TestFilterAddWireErrors performs negative tests against wire encode and decode
// of MsgFilterAdd to confirm error paths work correctly.
func TestFilterAddWireErrors(t *testing.T) {
	pver := ProtocolVersion

	baseData := []byte{0x01, 0x02, 0x03, 0x04}
	baseFilterAdd := NewMsgFilterAdd(baseData)
	baseFilterAddEncoded := append([]byte{0x04}, baseData...)

	tests := []struct {
		in       *MsgFilterAdd // Value to encode
		buf      []byte        // Wire encoding
		pver     uint32        // Protocol version for wire encoding
		max      int           // Max size of fixed buffer to induce errors
		writeErr error         // Expected write error
		readErr  error         // Expected read error
	}{
		// Latest protocol version with intentional read/write errors.
		// Force error in data size.
		{baseFilterAdd, baseFilterAddEncoded, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in data.
		{baseFilterAdd, baseFilterAddEncoded, pver, 1, io.ErrShortWrite, io.EOF},
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
		var msg MsgFilterAdd
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
