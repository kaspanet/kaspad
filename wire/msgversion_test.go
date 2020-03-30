// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/random"
)

// TestVersion tests the MsgVersion API.
func TestVersion(t *testing.T) {
	pver := ProtocolVersion

	// Create version message data.
	selectedTipHash := &daghash.Hash{12, 34}
	tcpAddrMe := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 16111}
	me := NewNetAddress(tcpAddrMe, SFNodeNetwork)
	tcpAddrYou := &net.TCPAddr{IP: net.ParseIP("192.168.0.1"), Port: 16111}
	you := NewNetAddress(tcpAddrYou, SFNodeNetwork)
	nonce, err := random.Uint64()
	if err != nil {
		t.Errorf("random.Uint64: error generating nonce: %v", err)
	}

	// Ensure we get the correct data back out.
	msg := NewMsgVersion(me, you, nonce, selectedTipHash, nil)
	if msg.ProtocolVersion != int32(pver) {
		t.Errorf("NewMsgVersion: wrong protocol version - got %v, want %v",
			msg.ProtocolVersion, pver)
	}
	if !reflect.DeepEqual(&msg.AddrMe, me) {
		t.Errorf("NewMsgVersion: wrong me address - got %v, want %v",
			spew.Sdump(&msg.AddrMe), spew.Sdump(me))
	}
	if !reflect.DeepEqual(&msg.AddrYou, you) {
		t.Errorf("NewMsgVersion: wrong you address - got %v, want %v",
			spew.Sdump(&msg.AddrYou), spew.Sdump(you))
	}
	if msg.Nonce != nonce {
		t.Errorf("NewMsgVersion: wrong nonce - got %v, want %v",
			msg.Nonce, nonce)
	}
	if msg.UserAgent != DefaultUserAgent {
		t.Errorf("NewMsgVersion: wrong user agent - got %v, want %v",
			msg.UserAgent, DefaultUserAgent)
	}
	if !msg.SelectedTipHash.IsEqual(selectedTipHash) {
		t.Errorf("NewMsgVersion: wrong selected tip hash - got %s, want %s",
			msg.SelectedTipHash, selectedTipHash)
	}
	if msg.DisableRelayTx {
		t.Errorf("NewMsgVersion: disable relay tx is not false by "+
			"default - got %v, want %v", msg.DisableRelayTx, false)
	}

	msg.AddUserAgent("myclient", "1.2.3", "optional", "comments")
	customUserAgent := DefaultUserAgent + "myclient:1.2.3(optional; comments)/"
	if msg.UserAgent != customUserAgent {
		t.Errorf("AddUserAgent: wrong user agent - got %s, want %s",
			msg.UserAgent, customUserAgent)
	}

	msg.AddUserAgent("mygui", "3.4.5")
	customUserAgent += "mygui:3.4.5/"
	if msg.UserAgent != customUserAgent {
		t.Errorf("AddUserAgent: wrong user agent - got %s, want %s",
			msg.UserAgent, customUserAgent)
	}

	// accounting for ":", "/"
	err = msg.AddUserAgent(strings.Repeat("t",
		MaxUserAgentLen-len(customUserAgent)-2+1), "")
	if msgErr := &(MessageError{}); !errors.As(err, &msgErr) {
		t.Errorf("AddUserAgent: expected error not received "+
			"- got %v, want %T", err, MessageError{})

	}

	// Version message should not have any services set by default.
	if msg.Services != 0 {
		t.Errorf("NewMsgVersion: wrong default services - got %v, want %v",
			msg.Services, 0)

	}
	if msg.HasService(SFNodeNetwork) {
		t.Errorf("HasService: SFNodeNetwork service is set")
	}

	// Ensure the command is expected value.
	wantCmd := "version"
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgVersion: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value.
	// Protocol version 4 bytes + services 8 bytes + timestamp 16 bytes +
	// remote and local net addresses + nonce 8 bytes + length of user
	// agent (varInt) + max allowed useragent length + selected tip hash length +
	// relay transactions flag 1 byte.
	wantPayload := uint32(394)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Ensure adding the full service node flag works.
	msg.AddService(SFNodeNetwork)
	if msg.Services != SFNodeNetwork {
		t.Errorf("AddService: wrong services - got %v, want %v",
			msg.Services, SFNodeNetwork)
	}
	if !msg.HasService(SFNodeNetwork) {
		t.Errorf("HasService: SFNodeNetwork service not set")
	}
}

// TestVersionWire tests the MsgVersion wire encode and decode for various
// protocol versions.
func TestVersionWire(t *testing.T) {
	// verRelayTxFalse and verRelayTxFalseEncoded is a version message with the transaction relay disabled.
	baseVersionWithRelayTxCopy := *baseVersionWithRelayTx
	verRelayTxFalse := &baseVersionWithRelayTxCopy
	verRelayTxFalse.DisableRelayTx = true
	verRelayTxFalseEncoded := make([]byte, len(baseVersionWithRelayTxEncoded))
	copy(verRelayTxFalseEncoded, baseVersionWithRelayTxEncoded)
	verRelayTxFalseEncoded[len(verRelayTxFalseEncoded)-1] = 0

	tests := []struct {
		in   *MsgVersion // Message to encode
		out  *MsgVersion // Expected decoded message
		buf  []byte      // Wire encoding
		pver uint32      // Protocol version for wire encoding
	}{
		// Latest protocol version.
		{
			baseVersionWithRelayTx,
			baseVersionWithRelayTx,
			baseVersionWithRelayTxEncoded,
			ProtocolVersion,
		},
		{
			verRelayTxFalse,
			verRelayTxFalse,
			verRelayTxFalseEncoded,
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
		var msg MsgVersion
		rbuf := bytes.NewBuffer(test.buf)
		err = msg.KaspaDecode(rbuf, test.pver)
		if err != nil {
			t.Errorf("KaspaDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&msg, test.out) {
			t.Errorf("KaspaDecode #%d\n got: %s want: %s", i,
				spew.Sdump(msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestVersionWireErrors performs negative tests against wire encode and
// decode of MsgGetHeaders to confirm error paths work correctly.
func TestVersionWireErrors(t *testing.T) {
	pver := ProtocolVersion
	wireErr := &MessageError{}

	// Ensure calling MsgVersion.KaspaDecode with a non *bytes.Buffer returns
	// error.
	fr := newFixedReader(0, []byte{})
	if err := baseVersion.KaspaDecode(fr, pver); err == nil {
		t.Errorf("Did not received error when calling " +
			"MsgVersion.KaspaDecode with non *bytes.Buffer")
	}

	// Copy the base version and change the user agent to exceed max limits.
	bvc := *baseVersion
	exceedUAVer := &bvc
	newUA := "/" + strings.Repeat("t", MaxUserAgentLen-8+1) + ":0.0.1/"
	exceedUAVer.UserAgent = newUA

	// Encode the new UA length as a varint.
	var newUAVarIntBuf bytes.Buffer
	err := WriteVarInt(&newUAVarIntBuf, uint64(len(newUA)))
	if err != nil {
		t.Errorf("WriteVarInt: error %v", err)
	}

	// Make a new buffer big enough to hold the base version plus the new
	// bytes for the bigger varint to hold the new size of the user agent
	// and the new user agent string. Then stitch it all together.
	newLen := len(baseVersionEncoded) - len(baseVersion.UserAgent)
	newLen = newLen + len(newUAVarIntBuf.Bytes()) - 1 + len(newUA)
	exceedUAVerEncoded := make([]byte, newLen)
	copy(exceedUAVerEncoded, baseVersionEncoded[0:81])
	copy(exceedUAVerEncoded[81:], newUAVarIntBuf.Bytes())
	copy(exceedUAVerEncoded[84:], []byte(newUA))
	copy(exceedUAVerEncoded[84+len(newUA):], baseVersionEncoded[98:101])

	tests := []struct {
		in       *MsgVersion // Value to encode
		buf      []byte      // Wire encoding
		pver     uint32      // Protocol version for wire encoding
		max      int         // Max size of fixed buffer to induce errors
		writeErr error       // Expected write error
		readErr  error       // Expected read error
	}{
		// Force error in protocol version.
		{baseVersion, baseVersionEncoded, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in services.
		{baseVersion, baseVersionEncoded, pver, 4, io.ErrShortWrite, io.EOF},
		// Force error in timestamp.
		{baseVersion, baseVersionEncoded, pver, 12, io.ErrShortWrite, io.EOF},
		// Force error in subnetworkID.
		{baseVersion, baseVersionEncoded, pver, 20, io.ErrShortWrite, io.EOF},
		// Force error in remote address.
		{baseVersion, baseVersionEncoded, pver, 21, io.ErrShortWrite, io.EOF},
		// Force error in local address.
		{baseVersion, baseVersionEncoded, pver, 48, io.ErrShortWrite, io.ErrUnexpectedEOF},
		// Force error in nonce.
		{baseVersion, baseVersionEncoded, pver, 74, io.ErrShortWrite, io.ErrUnexpectedEOF},
		// Force error in user agent length.
		{baseVersion, baseVersionEncoded, pver, 82, io.ErrShortWrite, io.EOF},
		// Force error in user agent.
		{baseVersion, baseVersionEncoded, pver, 83, io.ErrShortWrite, io.ErrUnexpectedEOF},
		// Force error in last block.
		{baseVersion, baseVersionEncoded, pver, 99, io.ErrShortWrite, io.ErrUnexpectedEOF},
		// Force error due to user agent too big
		{exceedUAVer, exceedUAVerEncoded, pver, newLen, wireErr, wireErr},
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
		var msg MsgVersion
		buf := bytes.NewBuffer(test.buf[0:test.max])
		err = msg.KaspaDecode(buf, test.pver)

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

// baseVersion is used in the various tests as a baseline MsgVersion.
var baseVersion = &MsgVersion{
	ProtocolVersion: 60002,
	Services:        SFNodeNetwork,
	Timestamp:       time.Unix(0x495fab29, 0), // 2009-01-03 12:15:05 -0600 CST)
	AddrYou: NetAddress{
		Timestamp: time.Time{}, // Zero value -- no timestamp in version
		Services:  SFNodeNetwork,
		IP:        net.ParseIP("192.168.0.1"),
		Port:      16111,
	},
	AddrMe: NetAddress{
		Timestamp: time.Time{}, // Zero value -- no timestamp in version
		Services:  SFNodeNetwork,
		IP:        net.ParseIP("127.0.0.1"),
		Port:      16111,
	},
	Nonce:           123123, // 0x1e0f3
	UserAgent:       "/kaspadtest:0.0.1/",
	SelectedTipHash: &daghash.Hash{0x12, 0x34},
}

// baseVersionEncoded is the wire encoded bytes for baseVersion using protocol
// version 60002 and is used in the various tests.
var baseVersionEncoded = []byte{
	0x62, 0xea, 0x00, 0x00, // Protocol version 60002
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
	0x29, 0xab, 0x5f, 0x49, 0x00, 0x00, 0x00, 0x00, // 64-bit Timestamp
	0x01, // is full node
	// AddrYou -- No timestamp for NetAddress in version message
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xc0, 0xa8, 0x00, 0x01, // IP 192.168.0.1
	0x3e, 0xef, // Port 16111 in big-endian
	// AddrMe -- No timestamp for NetAddress in version message
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x01, // IP 127.0.0.1
	0x3e, 0xef, // Port 16111 in big-endian
	0xf3, 0xe0, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, // Fake Nonce. TODO: (Ori) Replace to a real nonce
	0x12, // Varint for user agent length
	0x2f, 0x6b, 0x61, 0x73, 0x70, 0x61, 0x64, 0x74,
	0x65, 0x73, 0x74, 0x3a, 0x30, 0x2e, 0x30, 0x2e, // User agent
	0x31, 0x2f, 0x12, 0x34, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Selected Tip
	0x00, 0x00,
}

// baseVersionWithRelayTx is used in the various tests as a baseline MsgVersion
var baseVersionWithRelayTx = &MsgVersion{
	ProtocolVersion: 70001,
	Services:        SFNodeNetwork,
	Timestamp:       time.Unix(0x495fab29, 0), // 2009-01-03 12:15:05 -0600 CST)
	AddrYou: NetAddress{
		Timestamp: time.Time{}, // Zero value -- no timestamp in version
		Services:  SFNodeNetwork,
		IP:        net.ParseIP("192.168.0.1"),
		Port:      16111,
	},
	AddrMe: NetAddress{
		Timestamp: time.Time{}, // Zero value -- no timestamp in version
		Services:  SFNodeNetwork,
		IP:        net.ParseIP("127.0.0.1"),
		Port:      16111,
	},
	Nonce:           123123, // 0x1e0f3
	UserAgent:       "/kaspadtest:0.0.1/",
	SelectedTipHash: &daghash.Hash{0x12, 0x34},
}

// baseVersionWithRelayTxEncoded is the wire encoded bytes for
// baseVersionWithRelayTx and is used in the various tests.
var baseVersionWithRelayTxEncoded = []byte{
	0x71, 0x11, 0x01, 0x00, // Protocol version 70001
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
	0x29, 0xab, 0x5f, 0x49, 0x00, 0x00, 0x00, 0x00, // 64-bit Timestamp
	0x01, // is full node
	// AddrYou -- No timestamp for NetAddress in version message
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xc0, 0xa8, 0x00, 0x01, // IP 192.168.0.1
	0x3e, 0xef, // Port 16111 in big-endian
	// AddrMe -- No timestamp for NetAddress in version message
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x01, // IP 127.0.0.1
	0x3e, 0xef, // Port 16111 in big-endian
	0xf3, 0xe0, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, // Nonce
	0x12, // Varint for user agent length
	0x2f, 0x6b, 0x61, 0x73, 0x70, 0x61, 0x64, 0x74,
	0x65, 0x73, 0x74, 0x3a, 0x30, 0x2e, 0x30, 0x2e, // User agent
	0x31, 0x2f, 0x12, 0x34, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Selected Tip
	0x00, 0x00,
	0x01, // Relay tx
}
