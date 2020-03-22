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

// TestBlockLocator tests the MsgBlockLocator API.
func TestBlockLocator(t *testing.T) {
	pver := ProtocolVersion

	hashStr := "000000000002e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	locatorHash, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	msg := NewMsgBlockLocator()

	// Ensure the command is expected value.
	wantCmd := "locator"
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgBlockLocator: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	// Num hashes (varInt) + max block locator
	// hashes.
	wantPayload := uint32(16009)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Ensure block locator hashes are added properly.
	err = msg.AddBlockLocatorHash(locatorHash)
	if err != nil {
		t.Errorf("AddBlockLocatorHash: %v", err)
	}
	if msg.BlockLocatorHashes[0] != locatorHash {
		t.Errorf("AddBlockLocatorHash: wrong block locator added - "+
			"got %v, want %v",
			spew.Sprint(msg.BlockLocatorHashes[0]),
			spew.Sprint(locatorHash))
	}

	// Ensure adding more than the max allowed block locator hashes per
	// message returns an error.
	for i := 0; i < MaxBlockLocatorsPerMsg; i++ {
		err = msg.AddBlockLocatorHash(locatorHash)
	}
	if err == nil {
		t.Errorf("AddBlockLocatorHash: expected error on too many " +
			"block locator hashes not received")
	}
}

// TestBlockLocatorWire tests the MsgBlockLocator wire encode and decode for various
// numbers of block locator hashes.
func TestBlockLocatorWire(t *testing.T) {
	hashStr := "2710f40c87ec93d010a6fd95f42c59a2cbacc60b18cf6b7957535"
	hashLocator, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	hashStr = "2e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	hashLocator2, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// MsgBlockLocator message with no block locators.
	noLocators := NewMsgBlockLocator()
	noLocatorsEncoded := []byte{
		0x00, // Varint for number of block locator hashes
	}

	// MsgBlockLocator message with multiple block locators.
	multiLocators := NewMsgBlockLocator()
	multiLocators.AddBlockLocatorHash(hashLocator2)
	multiLocators.AddBlockLocatorHash(hashLocator)
	multiLocatorsEncoded := []byte{
		0x02, // Varint for number of block locator hashes
		0xe0, 0xde, 0x06, 0x44, 0x68, 0x13, 0x2c, 0x63,
		0xd2, 0x20, 0xcc, 0x69, 0x12, 0x83, 0xcb, 0x65,
		0xbc, 0xaa, 0xe4, 0x79, 0x94, 0xef, 0x9e, 0x7b,
		0xad, 0xe7, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // first hash
		0x35, 0x75, 0x95, 0xb7, 0xf6, 0x8c, 0xb1, 0x60,
		0xcc, 0xba, 0x2c, 0x9a, 0xc5, 0x42, 0x5f, 0xd9,
		0x6f, 0x0a, 0x01, 0x3d, 0xc9, 0x7e, 0xc8, 0x40,
		0x0f, 0x71, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // second hash
	}

	tests := []struct {
		in   *MsgBlockLocator // Message to encode
		out  *MsgBlockLocator // Expected decoded message
		buf  []byte           // Wire encoding
		pver uint32           // Protocol version for wire encoding
	}{
		// Latest protocol version with no block locators.
		{
			noLocators,
			noLocators,
			noLocatorsEncoded,
			ProtocolVersion,
		},

		// Latest protocol version with multiple block locators.
		{
			multiLocators,
			multiLocators,
			multiLocatorsEncoded,
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
		var msg MsgBlockLocator
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

// TestBlockLocatorWireErrors performs negative tests against wire encode and
// decode of MsgBlockLocator to confirm error paths work correctly.
func TestBlockLocatorWireErrors(t *testing.T) {
	// Set protocol inside locator message. Use protocol version 1
	// specifically here instead of the latest because the test data is
	// using bytes encoded with that protocol version.
	pver := uint32(1)
	wireErr := &MessageError{}

	hashStr := "2710f40c87ec93d010a6fd95f42c59a2cbacc60b18cf6b7957535"
	hashLocator, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	hashStr = "2e7ad7b9eef9479e4aabc65cb831269cc20d2632c13684406dee0"
	hashLocator2, err := daghash.NewHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// MsgBlockLocator message with multiple block locators and a low hash.
	baseGetBlocks := NewMsgBlockLocator()
	baseGetBlocks.AddBlockLocatorHash(hashLocator2)
	baseGetBlocks.AddBlockLocatorHash(hashLocator)
	baseGetBlocksEncoded := []byte{
		0x02, // Varint for number of block locator hashes
		0xe0, 0xde, 0x06, 0x44, 0x68, 0x13, 0x2c, 0x63,
		0xd2, 0x20, 0xcc, 0x69, 0x12, 0x83, 0xcb, 0x65,
		0xbc, 0xaa, 0xe4, 0x79, 0x94, 0xef, 0x9e, 0x7b,
		0xad, 0xe7, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // Block 99500 hash
		0x35, 0x75, 0x95, 0xb7, 0xf6, 0x8c, 0xb1, 0x60,
		0xcc, 0xba, 0x2c, 0x9a, 0xc5, 0x42, 0x5f, 0xd9,
		0x6f, 0x0a, 0x01, 0x3d, 0xc9, 0x7e, 0xc8, 0x40,
		0x0f, 0x71, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, // Block 99499 hash
	}

	// Message that forces an error by having more than the max allowed
	// block locator hashes.
	maxGetBlocks := NewMsgBlockLocator()
	for i := 0; i < MaxBlockLocatorsPerMsg; i++ {
		maxGetBlocks.AddBlockLocatorHash(mainnetGenesisHash)
	}
	maxGetBlocks.BlockLocatorHashes = append(maxGetBlocks.BlockLocatorHashes,
		mainnetGenesisHash)
	maxGetBlocksEncoded := []byte{
		0xfd, 0xf5, 0x01, // Varint for number of block loc hashes (501)
	}

	tests := []struct {
		in       *MsgBlockLocator // Value to encode
		buf      []byte           // Wire encoding
		pver     uint32           // Protocol version for wire encoding
		max      int              // Max size of fixed buffer to induce errors
		writeErr error            // Expected write error
		readErr  error            // Expected read error
	}{
		// Force error in block locator hash count.
		{baseGetBlocks, baseGetBlocksEncoded, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in block locator hashes.
		{baseGetBlocks, baseGetBlocksEncoded, pver, 1, io.ErrShortWrite, io.EOF},
		// Force error with greater than max block locator hashes.
		{maxGetBlocks, maxGetBlocksEncoded, pver, 3, wireErr, wireErr},
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
		var msg MsgBlockLocator
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
