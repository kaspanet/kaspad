// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// TestAddr tests the MsgAddr API.
func TestAddr(t *testing.T) {
	pver := ProtocolVersion

	// Ensure the command is expected value.
	wantCmd := "addr"
	msg := NewMsgAddr(false, nil)
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgAddr: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	// Num addresses (varInt) + max allowed addresses.
	wantPayload := uint32(34031)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Ensure NetAddresses are added properly.
	tcpAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 16111}
	na := NewNetAddress(tcpAddr, SFNodeNetwork)
	err := msg.AddAddress(na)
	if err != nil {
		t.Errorf("AddAddress: %v", err)
	}
	if msg.AddrList[0] != na {
		t.Errorf("AddAddress: wrong address added - got %v, want %v",
			spew.Sprint(msg.AddrList[0]), spew.Sprint(na))
	}

	// Ensure the address list is cleared properly.
	msg.ClearAddresses()
	if len(msg.AddrList) != 0 {
		t.Errorf("ClearAddresses: address list is not empty - "+
			"got %v [%v], want %v", len(msg.AddrList),
			spew.Sprint(msg.AddrList[0]), 0)
	}

	// Ensure adding more than the max allowed addresses per message returns
	// error.
	for i := 0; i < MaxAddrPerMsg+1; i++ {
		err = msg.AddAddress(na)
	}
	if err == nil {
		t.Errorf("AddAddress: expected error on too many addresses " +
			"not received")
	}
	err = msg.AddAddresses(na)
	if err == nil {
		t.Errorf("AddAddresses: expected error on too many addresses " +
			"not received")
	}
}

// TestAddrWire tests the MsgAddr wire encode and decode for various numbers
// of addresses and protocol versions.
func TestAddrWire(t *testing.T) {
	// A couple of NetAddresses to use for testing.
	na := &NetAddress{
		Timestamp: mstime.UnixMilliseconds(0x17315ed0f99),
		Services:  SFNodeNetwork,
		IP:        net.ParseIP("127.0.0.1"),
		Port:      16111,
	}
	na2 := &NetAddress{
		Timestamp: mstime.UnixMilliseconds(0x17315ed0f99),
		Services:  SFNodeNetwork,
		IP:        net.ParseIP("192.168.0.1"),
		Port:      16110,
	}

	// Empty address message.
	noAddr := NewMsgAddr(false, nil)
	noAddrEncoded := []byte{
		0x00, // All subnetworks
		0x01, // Is full node
		0x00, // Varint for number of addresses
	}

	// Address message with multiple addresses.
	multiAddr := NewMsgAddr(true, nil)
	multiAddr.AddAddresses(na, na2)
	multiAddrEncoded := []byte{
		0x01,                                           // All subnetworks
		0x02,                                           // Varint for number of addresses
		0x99, 0x0f, 0xed, 0x15, 0x73, 0x01, 0x00, 0x00, // Timestamp
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x01, // IP 127.0.0.1
		0x3e, 0xef, // Port 16111 in big-endian
		0x99, 0x0f, 0xed, 0x15, 0x73, 0x01, 0x00, 0x00, // Timestamp
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xff, 0xff, 0xc0, 0xa8, 0x00, 0x01, // IP 192.168.0.1
		0x3e, 0xee, // Port 16110 in big-endian
	}

	// Address message with multiple addresses and subnetworkID.
	multiAddrSubnet := NewMsgAddr(false, subnetworkid.SubnetworkIDNative)
	multiAddrSubnet.AddAddresses(na, na2)
	multiAddrSubnetEncoded := []byte{
		0x00,                                           // All subnetworks
		0x00,                                           // Is full node
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Subnetwork ID
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x02,                                           // Varint for number of addresses
		0x99, 0x0f, 0xed, 0x15, 0x73, 0x01, 0x00, 0x00, // Timestamp
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x01, // IP 127.0.0.1
		0x3e, 0xef, // Port 16111 in big-endian
		0x99, 0x0f, 0xed, 0x15, 0x73, 0x01, 0x00, 0x00, // Timestamp
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xff, 0xff, 0xc0, 0xa8, 0x00, 0x01, // IP 192.168.0.1
		0x3e, 0xee, // Port 16110 in big-endian
	}

	tests := []struct {
		in   *MsgAddr // Message to encode
		out  *MsgAddr // Expected decoded message
		buf  []byte   // Wire encoding
		pver uint32   // Protocol version for wire encoding
	}{
		// Latest protocol version with no addresses.
		{
			noAddr,
			noAddr,
			noAddrEncoded,
			ProtocolVersion,
		},

		// Latest protocol version with multiple addresses.
		{
			multiAddr,
			multiAddr,
			multiAddrEncoded,
			ProtocolVersion,
		},

		// Latest protocol version with multiple addresses and subnetwork.
		{
			multiAddrSubnet,
			multiAddrSubnet,
			multiAddrSubnetEncoded,
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
		var msg MsgAddr
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

// TestAddrWireErrors performs negative tests against wire encode and decode
// of MsgAddr to confirm error paths work correctly.
func TestAddrWireErrors(t *testing.T) {
	pver := ProtocolVersion
	wireErr := &MessageError{}

	// A couple of NetAddresses to use for testing.
	na := &NetAddress{
		Timestamp: mstime.UnixMilliseconds(0x495fab29000),
		Services:  SFNodeNetwork,
		IP:        net.ParseIP("127.0.0.1"),
		Port:      16111,
	}
	na2 := &NetAddress{
		Timestamp: mstime.UnixMilliseconds(0x495fab29000),
		Services:  SFNodeNetwork,
		IP:        net.ParseIP("192.168.0.1"),
		Port:      16110,
	}

	// Address message with multiple addresses.
	baseAddr := NewMsgAddr(false, nil)
	baseAddr.AddAddresses(na, na2)
	baseAddrEncoded := []byte{
		0x01,                                           // All subnetworks
		0x02,                                           // Varint for number of addresses
		0x29, 0xab, 0x5f, 0x49, 0x00, 0x00, 0x00, 0x00, // Timestamp
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x01, // IP 127.0.0.1
		0x3e, 0xef, // Port 16111 in big-endian
		0x29, 0xab, 0x5f, 0x49, 0x00, 0x00, 0x00, 0x00, // Timestamp
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // SFNodeNetwork
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xff, 0xff, 0xc0, 0xa8, 0x00, 0x01, // IP 192.168.0.1
		0x3e, 0xee, // Port 16110 in big-endian
	}

	// Message that forces an error by having more than the max allowed
	// addresses.
	maxAddr := NewMsgAddr(false, nil)
	for i := 0; i < MaxAddrPerMsg; i++ {
		maxAddr.AddAddress(na)
	}
	maxAddr.AddrList = append(maxAddr.AddrList, na)
	maxAddrEncoded := []byte{
		0x01,             // All subnetworks
		0xfd, 0x03, 0xe9, // Varint for number of addresses (1001)
	}

	tests := []struct {
		in       *MsgAddr // Value to encode
		buf      []byte   // Wire encoding
		pver     uint32   // Protocol version for wire encoding
		max      int      // Max size of fixed buffer to induce errors
		writeErr error    // Expected write error
		readErr  error    // Expected read error
	}{
		// Latest protocol version with intentional read/write errors.
		// Force error in addresses count
		{baseAddr, baseAddrEncoded, pver, 1, io.ErrShortWrite, io.EOF},
		// Force error in address list.
		{baseAddr, baseAddrEncoded, pver, 2, io.ErrShortWrite, io.EOF},
		// Force error with greater than max inventory vectors.
		{maxAddr, maxAddrEncoded, pver, 4, wireErr, wireErr},
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
		var msg MsgAddr
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
