// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// TestGetAddr tests the MsgGetAddr API.
func TestGetAddr(t *testing.T) {
	pver := ProtocolVersion

	// Ensure the command is expected value.
	wantCmd := "getaddr"
	msg := NewMsgGetAddr(true, false, nil)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgGetAddr: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	// Num addresses (varInt) + max allowed addresses.
	wantPayload := uint32(23)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}
}

// TestGetAddrWire tests the MsgGetAddr wire encode and decode for various
// protocol versions.
func TestGetAddrWire(t *testing.T) {
	// With all subnetworks
	msgGetAddr := NewMsgGetAddr(true, false, nil)
	msgGetAddrEncoded := []byte{
		0x01, // Need addresses
		0x00, // All subnetworks
		0x01, // Get full nodes
	}

	// With specific subnetwork
	msgGetAddrSubnet := NewMsgGetAddr(true, false, subnetworkid.SubnetworkIDNative)
	msgGetAddrSubnetEncoded := []byte{
		0x01,                                           // Need addresses
		0x00,                                           // Is all subnetworks
		0x00,                                           // Is full node
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Subnetwork ID
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	msgGetAddrNoAddressesNeeded := NewMsgGetAddr(false, false, nil)
	msgGetAddrNoAddressesNeededEncoded := []byte{
		0x00, // Need addresses
		0x00, // All subnetworks
		0x01, // Get full nodes
	}

	tests := []struct {
		in   *MsgGetAddr // Message to encode
		out  *MsgGetAddr // Expected decoded message
		buf  []byte      // Wire encoding
		pver uint32      // Protocol version for wire encoding
	}{
		// Latest protocol version. All subnetworks
		{
			msgGetAddr,
			msgGetAddr,
			msgGetAddrEncoded,
			ProtocolVersion,
		},
		// Latest protocol version. Specific subnetwork
		{
			msgGetAddrSubnet,
			msgGetAddrSubnet,
			msgGetAddrSubnetEncoded,
			ProtocolVersion,
		},
		// Latest protocol version. No addresses needed
		{
			msgGetAddrNoAddressesNeeded,
			msgGetAddrNoAddressesNeeded,
			msgGetAddrNoAddressesNeededEncoded,
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
		var msg MsgGetAddr
		rbuf := bytes.NewReader(test.buf)
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
