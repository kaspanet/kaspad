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

// TestGetAddr tests the MsgGetAddresses API.
func TestGetAddr(t *testing.T) {
	pver := ProtocolVersion

	// Ensure the command is expected value.
	wantCmd := "getaddr"
	msg := NewMsgGetAddresses(false, nil)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgGetAddresses: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	// Num addresses (varInt) + max allowed addresses.
	wantPayload := uint32(22)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}
}

// TestGetAddrWire tests the MsgGetAddresses wire encode and decode for various
// protocol versions.
func TestGetAddrWire(t *testing.T) {
	// With all subnetworks
	msgGetAddresses := NewMsgGetAddresses(false, nil)
	msgGetAddrEncoded := []byte{
		0x00, // All subnetworks
		0x01, // Get full nodes
	}

	// With specific subnetwork
	msgGetAddressesSubnetwork := NewMsgGetAddresses(false, subnetworkid.SubnetworkIDNative)
	msgGetAddressesSubnetworkEncoded := []byte{
		0x00,                                           // Is all subnetworks
		0x00,                                           // Is full node
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Subnetwork ID
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	tests := []struct {
		in   *MsgGetAddresses // Message to encode
		out  *MsgGetAddresses // Expected decoded message
		buf  []byte           // Wire encoding
		pver uint32           // Protocol version for wire encoding
	}{
		// Latest protocol version. All subnetworks
		{
			msgGetAddresses,
			msgGetAddresses,
			msgGetAddrEncoded,
			ProtocolVersion,
		},
		// Latest protocol version. Specific subnetwork
		{
			msgGetAddressesSubnetwork,
			msgGetAddressesSubnetwork,
			msgGetAddressesSubnetworkEncoded,
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
		var msg MsgGetAddresses
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
