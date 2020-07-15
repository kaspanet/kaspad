// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"encoding/binary"
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"io"
	"math"
	"net"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/util/daghash"
)

// makeHeader is a convenience function to make a message header in the form of
// a byte slice. It is used to force errors when reading messages.
func makeHeader(kaspaNet KaspaNet, command MessageCommand,
	payloadLen uint32, checksum uint32) []byte {

	// The length of a kaspa message header is 13 bytes.
	// 4 byte magic number of the kaspa network + 4 bytes command + 4 byte
	// payload length + 4 byte checksum.
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint32(buf, uint32(kaspaNet))
	binary.LittleEndian.PutUint32(buf[4:], uint32(command))
	binary.LittleEndian.PutUint32(buf[8:], payloadLen)
	binary.LittleEndian.PutUint32(buf[12:], checksum)
	return buf
}

// TestMessage tests the Read/WriteMessage and Read/WriteMessageN API.
func TestMessage(t *testing.T) {
	pver := ProtocolVersion

	// Create the various types of messages to test.

	// MsgVersion.
	addrYou := &net.TCPAddr{IP: net.ParseIP("192.168.0.1"), Port: 16111}
	you := NewNetAddress(addrYou, SFNodeNetwork)
	you.Timestamp = mstime.Time{} // Version message has zero value timestamp.
	addrMe := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 16111}
	me := NewNetAddress(addrMe, SFNodeNetwork)
	me.Timestamp = mstime.Time{} // Version message has zero value timestamp.
	idMeBytes := make([]byte, id.IDLength)
	idMeBytes[0] = 0xff
	idMe := id.FromBytes(idMeBytes)
	msgVersion := NewMsgVersion(me, idMe, &daghash.ZeroHash, nil)

	msgVerack := NewMsgVerAck()
	msgGetAddr := NewMsgGetAddr(false, nil)
	msgAddr := NewMsgAddr(false, nil)
	msgGetBlockInvs := NewMsgGetBlockInvs(&daghash.Hash{}, &daghash.Hash{})
	msgBlock := &blockOne
	msgInv := NewMsgInv()
	msgGetData := NewMsgGetData()
	msgNotFound := NewMsgNotFound()
	msgTx := NewNativeMsgTx(1, nil, nil)
	msgPing := NewMsgPing(123123)
	msgPong := NewMsgPong(123123)
	msgGetBlockLocator := NewMsgGetBlockLocator(&daghash.ZeroHash, &daghash.ZeroHash)
	msgBlockLocator := NewMsgBlockLocator()
	msgFeeFilter := NewMsgFeeFilter(123456)
	msgFilterAdd := NewMsgFilterAdd([]byte{0x01})
	msgFilterClear := NewMsgFilterClear()
	msgFilterLoad := NewMsgFilterLoad([]byte{0x01}, 10, 0, BloomUpdateNone)
	bh := NewBlockHeader(1, []*daghash.Hash{mainnetGenesisHash, simnetGenesisHash}, &daghash.Hash{}, &daghash.Hash{}, &daghash.Hash{}, 0, 0)
	msgMerkleBlock := NewMsgMerkleBlock(bh)
	msgReject := NewMsgReject(CmdBlock, RejectDuplicate, "duplicate block")

	tests := []struct {
		in       Message  // Value to encode
		out      Message  // Expected decoded value
		pver     uint32   // Protocol version for wire encoding
		kaspaNet KaspaNet // Network to use for wire encoding
		bytes    int      // Expected num bytes read/written
	}{
		{msgVersion, msgVersion, pver, Mainnet, 128},
		{msgVerack, msgVerack, pver, Mainnet, 16},
		{msgGetAddr, msgGetAddr, pver, Mainnet, 18},
		{msgAddr, msgAddr, pver, Mainnet, 19},
		{msgGetBlockInvs, msgGetBlockInvs, pver, Mainnet, 80},
		{msgBlock, msgBlock, pver, Mainnet, 364},
		{msgInv, msgInv, pver, Mainnet, 17},
		{msgGetData, msgGetData, pver, Mainnet, 17},
		{msgNotFound, msgNotFound, pver, Mainnet, 17},
		{msgTx, msgTx, pver, Mainnet, 50},
		{msgPing, msgPing, pver, Mainnet, 24},
		{msgPong, msgPong, pver, Mainnet, 24},
		{msgGetBlockLocator, msgGetBlockLocator, pver, Mainnet, 80},
		{msgBlockLocator, msgBlockLocator, pver, Mainnet, 17},
		{msgFeeFilter, msgFeeFilter, pver, Mainnet, 24},
		{msgFilterAdd, msgFilterAdd, pver, Mainnet, 18},
		{msgFilterClear, msgFilterClear, pver, Mainnet, 16},
		{msgFilterLoad, msgFilterLoad, pver, Mainnet, 27},
		{msgMerkleBlock, msgMerkleBlock, pver, Mainnet, 207},
		{msgReject, msgReject, pver, Mainnet, 69},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		var buf bytes.Buffer
		nw, err := WriteMessageN(&buf, test.in, test.pver, test.kaspaNet)
		if err != nil {
			t.Errorf("WriteMessage #%d error %v", i, err)
			continue
		}

		// Ensure the number of bytes written match the expected value.
		if nw != test.bytes {
			t.Errorf("WriteMessage #%d unexpected num bytes "+
				"written - got %d, want %d", i, nw, test.bytes)
		}

		// Decode from wire format.
		rbuf := bytes.NewReader(buf.Bytes())
		nr, msg, _, err := ReadMessageN(rbuf, test.pver, test.kaspaNet)
		if err != nil {
			t.Errorf("ReadMessage #%d error %v, msg %+v", i, err,
				spew.Sdump(msg))
			continue
		}
		if !reflect.DeepEqual(msg, test.out) {
			t.Errorf("ReadMessage #%d\n got: %v want: %v", i,
				spew.Sdump(msg), spew.Sdump(test.out))
			continue
		}

		// Ensure the number of bytes read match the expected value.
		if nr != test.bytes {
			t.Errorf("ReadMessage #%d unexpected num bytes read - "+
				"got %d, want %d", i, nr, test.bytes)
		}
	}

	// Do the same thing for Read/WriteMessage, but ignore the bytes since
	// they don't return them.
	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		var buf bytes.Buffer
		err := WriteMessage(&buf, test.in, test.pver, test.kaspaNet)
		if err != nil {
			t.Errorf("WriteMessage #%d error %v", i, err)
			continue
		}

		// Decode from wire format.
		rbuf := bytes.NewReader(buf.Bytes())
		msg, _, err := ReadMessage(rbuf, test.pver, test.kaspaNet)
		if err != nil {
			t.Errorf("ReadMessage #%d error %v, msg %v", i, err,
				spew.Sdump(msg))
			continue
		}
		if !reflect.DeepEqual(msg, test.out) {
			t.Errorf("ReadMessage #%d\n got: %v want: %v", i,
				spew.Sdump(msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestReadMessageWireErrors performs negative tests against wire decoding into
// concrete messages to confirm error paths work correctly.
func TestReadMessageWireErrors(t *testing.T) {
	pver := ProtocolVersion
	kaspaNet := Mainnet

	// Ensure message errors are as expected with no function specified.
	wantErr := "something bad happened"
	testErr := MessageError{Description: wantErr}
	if testErr.Error() != wantErr {
		t.Errorf("MessageError: wrong error - got %v, want %v",
			testErr.Error(), wantErr)
	}

	// Ensure message errors are as expected with a function specified.
	wantFunc := "foo"
	testErr = MessageError{Func: wantFunc, Description: wantErr}
	if testErr.Error() != wantFunc+": "+wantErr {
		t.Errorf("MessageError: wrong error - got %v, want %v",
			testErr.Error(), wantErr)
	}

	bogusCommand := MessageCommand(math.MaxUint8)

	// Wire encoded bytes for main and testnet networks magic identifiers.
	testnetBytes := makeHeader(Testnet, bogusCommand, 0, 0)

	// Wire encoded bytes for a message that exceeds max overall message
	// length.
	mpl := uint32(MaxMessagePayload)
	exceedMaxPayloadBytes := makeHeader(kaspaNet, CmdAddr, mpl+1, 0)

	// Wire encoded bytes for a command which is invalid utf-8.
	badCommandBytes := makeHeader(kaspaNet, bogusCommand, 0, 0)
	badCommandBytes[4] = 0x81

	// Wire encoded bytes for a command which is valid, but not supported.
	unsupportedCommandBytes := makeHeader(kaspaNet, bogusCommand, 0, 0)

	// Wire encoded bytes for a message which exceeds the max payload for
	// a specific message type.
	exceedTypePayloadBytes := makeHeader(kaspaNet, CmdGetAddr, 23, 0)

	// Wire encoded bytes for a message which does not deliver the full
	// payload according to the header length.
	shortPayloadBytes := makeHeader(kaspaNet, CmdVersion, 115, 0)

	// Wire encoded bytes for a message with a bad checksum.
	badChecksumBytes := makeHeader(kaspaNet, CmdVersion, 2, 0xbeef)
	badChecksumBytes = append(badChecksumBytes, []byte{0x0, 0x0}...)

	// Wire encoded bytes for a message which has a valid header, but is
	// the wrong format. An addr starts with a varint of the number of
	// contained in the message. Claim there is two, but don't provide
	// them. At the same time, forge the header fields so the message is
	// otherwise accurate.
	badMessageBytes := makeHeader(kaspaNet, CmdAddr, 1, 0xeaadc31c)
	badMessageBytes = append(badMessageBytes, 0x2)

	// Wire encoded bytes for a message which the header claims has 15k
	// bytes of data to discard.
	discardBytes := makeHeader(kaspaNet, bogusCommand, 15*1024, 0)

	tests := []struct {
		buf      []byte   // Wire encoding
		pver     uint32   // Protocol version for wire encoding
		kaspaNet KaspaNet // Kaspa network for wire encoding
		max      int      // Max size of fixed buffer to induce errors
		readErr  error    // Expected read error
		bytes    int      // Expected num bytes read
	}{
		// Latest protocol version with intentional read errors.

		// Short header.
		{
			[]byte{},
			pver,
			kaspaNet,
			0,
			io.EOF,
			0,
		},

		// Wrong network. Want Mainnet, but giving Testnet.
		{
			testnetBytes,
			pver,
			kaspaNet,
			len(testnetBytes),
			&MessageError{},
			16,
		},

		// Exceed max overall message payload length.
		{
			exceedMaxPayloadBytes,
			pver,
			kaspaNet,
			len(exceedMaxPayloadBytes),
			&MessageError{},
			16,
		},

		// Invalid UTF-8 command.
		{
			badCommandBytes,
			pver,
			kaspaNet,
			len(badCommandBytes),
			&MessageError{},
			16,
		},

		// Valid, but unsupported command.
		{
			unsupportedCommandBytes,
			pver,
			kaspaNet,
			len(unsupportedCommandBytes),
			&MessageError{},
			16,
		},

		// Exceed max allowed payload for a message of a specific type.
		{
			exceedTypePayloadBytes,
			pver,
			kaspaNet,
			len(exceedTypePayloadBytes),
			&MessageError{},
			16,
		},

		// Message with a payload shorter than the header indicates.
		{
			shortPayloadBytes,
			pver,
			kaspaNet,
			len(shortPayloadBytes),
			io.EOF,
			16,
		},

		// Message with a bad checksum.
		{
			badChecksumBytes,
			pver,
			kaspaNet,
			len(badChecksumBytes),
			&MessageError{},
			18,
		},

		// Message with a valid header, but wrong format.
		{
			badMessageBytes,
			pver,
			kaspaNet,
			len(badMessageBytes),
			io.EOF,
			17,
		},

		// 15k bytes of data to discard.
		{
			discardBytes,
			pver,
			kaspaNet,
			len(discardBytes),
			&MessageError{},
			16,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Decode from wire format.
		r := newFixedReader(test.max, test.buf)
		nr, _, _, err := ReadMessageN(r, test.pver, test.kaspaNet)

		// Ensure the number of bytes written match the expected value.
		if nr != test.bytes {
			t.Errorf("ReadMessage #%d unexpected num bytes read - "+
				"got %d, want %d", i, nr, test.bytes)
		}

		// For errors which are not of type MessageError, check them for
		// equality. If the error is a MessageError, check only if it's
		// the expected type.
		if msgErr := &(MessageError{}); !errors.As(err, &msgErr) {
			if !errors.Is(err, test.readErr) {
				t.Errorf("ReadMessage #%d wrong error got: %v <%T>, "+
					"want: %v <%T>", i, err, err,
					test.readErr, test.readErr)
				continue
			}
		} else if reflect.TypeOf(msgErr) != reflect.TypeOf(test.readErr) {
			t.Errorf("ReadMessage #%d wrong error type got: %T, "+
				"want: %T", i, msgErr, test.readErr)
			continue
		}
	}
}

// TestWriteMessageWireErrors performs negative tests against wire encoding from
// concrete messages to confirm error paths work correctly.
func TestWriteMessageWireErrors(t *testing.T) {
	pver := ProtocolVersion
	kaspaNet := Mainnet
	wireErr := &MessageError{}

	// Fake message with a problem during encoding
	encodeErrMsg := &fakeMessage{forceEncodeErr: true}

	// Fake message that has payload which exceeds max overall message size.
	exceedOverallPayload := make([]byte, MaxMessagePayload+1)
	exceedOverallPayloadErrMsg := &fakeMessage{payload: exceedOverallPayload}

	// Fake message that has payload which exceeds max allowed per message.
	exceedPayload := make([]byte, 1)
	exceedPayloadErrMsg := &fakeMessage{payload: exceedPayload, forceLenErr: true}

	// Fake message that is used to force errors in the header and payload
	// writes.
	bogusPayload := []byte{0x01, 0x02, 0x03, 0x04}
	bogusMsg := &fakeMessage{command: MessageCommand(math.MaxUint8), payload: bogusPayload}

	tests := []struct {
		msg      Message  // Message to encode
		pver     uint32   // Protocol version for wire encoding
		kaspaNet KaspaNet // Kaspa network for wire encoding
		max      int      // Max size of fixed buffer to induce errors
		err      error    // Expected error
		bytes    int      // Expected num bytes written
	}{
		// Force error in payload encode.
		{encodeErrMsg, pver, kaspaNet, 0, wireErr, 0},
		// Force error due to exceeding max overall message payload size.
		{exceedOverallPayloadErrMsg, pver, kaspaNet, 0, wireErr, 0},
		// Force error due to exceeding max payload for message type.
		{exceedPayloadErrMsg, pver, kaspaNet, 0, wireErr, 0},
		// Force error in header write.
		{bogusMsg, pver, kaspaNet, 0, io.ErrShortWrite, 0},
		// Force error in payload write.
		{bogusMsg, pver, kaspaNet, 16, io.ErrShortWrite, 16},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode wire format.
		w := newFixedWriter(test.max)
		nw, err := WriteMessageN(w, test.msg, test.pver, test.kaspaNet)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("WriteMessage #%d wrong error got: %v <%T>, "+
				"want: %T", i, err, err, test.err)
			continue
		}

		// Ensure the number of bytes written match the expected value.
		if nw != test.bytes {
			t.Errorf("WriteMessage #%d unexpected num bytes "+
				"written - got %d, want %d", i, nw, test.bytes)
		}

		// For errors which are not of type MessageError, check them for
		// equality. If the error is a MessageError, check only if it's
		// the expected type.
		if msgErr := &(MessageError{}); !errors.As(err, &msgErr) {
			if err != test.err {
				t.Errorf("ReadMessage #%d wrong error got: %v <%T>, "+
					"want: %v <%T>", i, err, err,
					test.err, test.err)
				continue
			}
		}
	}
}
