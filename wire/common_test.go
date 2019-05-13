// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/davecgh/go-spew/spew"
)

// mainNetGenesisHash is the hash of the first block in the block chain for the
// main network (genesis block).
var mainNetGenesisHash = &daghash.Hash{
	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
	0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = &daghash.Hash{
	0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a,
	0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
	0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
	0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
}

// mainNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the main network.
var mainNetGenesisMerkleRoot = &daghash.Hash{
	0x4a, 0x5e, 0x1e, 0x4b, 0xaa, 0xb8, 0x9f, 0x3a,
	0x32, 0x51, 0x8a, 0x88, 0xc3, 0x1b, 0xc8, 0x7f,
	0x61, 0x8f, 0x76, 0x67, 0x3e, 0x2c, 0xc7, 0x7a,
	0xb2, 0x12, 0x7b, 0x7a, 0xfd, 0xed, 0xa3, 0x3b,
}

var exampleIDMerkleRoot = &daghash.Hash{
	0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
	0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
	0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C,
	0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
}

var exampleAcceptedIDMerkleRoot = &daghash.Hash{
	0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C,
	0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
	0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
	0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
}

var exampleUTXOCommitment = &daghash.Hash{
	0x10, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C,
	0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
	0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
	0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
}

// TestElementWire tests wire encode and decode for various element types.  This
// is mainly to test the "fast" paths in readElement and writeElement which use
// type assertions to avoid reflection when possible.
func TestElementWire(t *testing.T) {
	type writeElementReflect int32

	tests := []struct {
		in  interface{} // Value to encode
		buf []byte      // Wire encoding
	}{
		{int32(1), []byte{0x01, 0x00, 0x00, 0x00}},
		{uint32(256), []byte{0x00, 0x01, 0x00, 0x00}},
		{
			int64(65536),
			[]byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			uint64(4294967296),
			[]byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			true,
			[]byte{0x01},
		},
		{
			false,
			[]byte{0x00},
		},
		{
			[4]byte{0x01, 0x02, 0x03, 0x04},
			[]byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			[CommandSize]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c,
			},
			[]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c,
			},
		},
		{
			[16]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
			},
			[]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
			},
		},
		{
			(*daghash.Hash)(&[daghash.HashSize]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
				0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
				0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
			}),
			[]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
				0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
				0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
			},
		},
		{
			ServiceFlag(SFNodeNetwork),
			[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			InvType(InvTypeTx),
			[]byte{0x01, 0x00, 0x00, 0x00},
		},
		{
			BitcoinNet(MainNet),
			[]byte{0xf9, 0xbe, 0xb4, 0xd9},
		},
		// Type not supported by the "fast" path and requires reflection.
		{
			writeElementReflect(1),
			[]byte{0x01, 0x00, 0x00, 0x00},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Write to wire format.
		var buf bytes.Buffer
		err := WriteElement(&buf, test.in)
		if err != nil {
			t.Errorf("writeElement #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("writeElement #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Read from wire format.
		rbuf := bytes.NewReader(test.buf)
		val := test.in
		if reflect.ValueOf(test.in).Kind() != reflect.Ptr {
			val = reflect.New(reflect.TypeOf(test.in)).Interface()
		}
		err = ReadElement(rbuf, val)
		if err != nil {
			t.Errorf("readElement #%d error %v", i, err)
			continue
		}
		ival := val
		if reflect.ValueOf(test.in).Kind() != reflect.Ptr {
			ival = reflect.Indirect(reflect.ValueOf(val)).Interface()
		}
		if !reflect.DeepEqual(ival, test.in) {
			t.Errorf("readElement #%d\n got: %s want: %s", i,
				spew.Sdump(ival), spew.Sdump(test.in))
			continue
		}
	}
}

// TestElementWireErrors performs negative tests against wire encode and decode
// of various element types to confirm error paths work correctly.
func TestElementWireErrors(t *testing.T) {
	tests := []struct {
		in       interface{} // Value to encode
		max      int         // Max size of fixed buffer to induce errors
		writeErr error       // Expected write error
		readErr  error       // Expected read error
	}{
		{int32(1), 0, io.ErrShortWrite, io.EOF},
		{uint32(256), 0, io.ErrShortWrite, io.EOF},
		{int64(65536), 0, io.ErrShortWrite, io.EOF},
		{true, 0, io.ErrShortWrite, io.EOF},
		{[4]byte{0x01, 0x02, 0x03, 0x04}, 0, io.ErrShortWrite, io.EOF},
		{
			[CommandSize]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c,
			},
			0, io.ErrShortWrite, io.EOF,
		},
		{
			[16]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
			},
			0, io.ErrShortWrite, io.EOF,
		},
		{
			(*daghash.Hash)(&[daghash.HashSize]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
				0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
				0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
			}),
			0, io.ErrShortWrite, io.EOF,
		},
		{ServiceFlag(SFNodeNetwork), 0, io.ErrShortWrite, io.EOF},
		{InvType(InvTypeTx), 0, io.ErrShortWrite, io.EOF},
		{BitcoinNet(MainNet), 0, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := WriteElement(w, test.in)
		if err != test.writeErr {
			t.Errorf("writeElement #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Decode from wire format.
		r := newFixedReader(test.max, nil)
		val := test.in
		if reflect.ValueOf(test.in).Kind() != reflect.Ptr {
			val = reflect.New(reflect.TypeOf(test.in)).Interface()
		}
		err = ReadElement(r, val)
		if err != test.readErr {
			t.Errorf("readElement #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestVarIntWire tests wire encode and decode for variable length integers.
func TestVarIntWire(t *testing.T) {
	tests := []struct {
		in  uint64 // Value to encode
		out uint64 // Expected decoded value
		buf []byte // Wire encoding
	}{
		// Latest protocol version.
		// Single byte
		{0, 0, []byte{0x00}},
		// Max single byte
		{0xfc, 0xfc, []byte{0xfc}},
		// Min 2-byte
		{0xfd, 0xfd, []byte{0xfd, 0x0fd, 0x00}},
		// Max 2-byte
		{0xffff, 0xffff, []byte{0xfd, 0xff, 0xff}},
		// Min 4-byte
		{0x10000, 0x10000, []byte{0xfe, 0x00, 0x00, 0x01, 0x00}},
		// Max 4-byte
		{0xffffffff, 0xffffffff, []byte{0xfe, 0xff, 0xff, 0xff, 0xff}},
		// Min 8-byte
		{
			0x100000000, 0x100000000,
			[]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		// Max 8-byte
		{
			0xffffffffffffffff, 0xffffffffffffffff,
			[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		var buf bytes.Buffer
		err := WriteVarInt(&buf, test.in)
		if err != nil {
			t.Errorf("WriteVarInt #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("WriteVarInt #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode from wire format.
		rbuf := bytes.NewReader(test.buf)
		val, err := ReadVarInt(rbuf)
		if err != nil {
			t.Errorf("ReadVarInt #%d error %v", i, err)
			continue
		}
		if val != test.out {
			t.Errorf("ReadVarInt #%d\n got: %d want: %d", i,
				val, test.out)
			continue
		}
	}
}

// TestVarIntWireErrors performs negative tests against wire encode and decode
// of variable length integers to confirm error paths work correctly.
func TestVarIntWireErrors(t *testing.T) {
	tests := []struct {
		in       uint64 // Value to encode
		buf      []byte // Wire encoding
		max      int    // Max size of fixed buffer to induce errors
		writeErr error  // Expected write error
		readErr  error  // Expected read error
	}{
		// Force errors on discriminant.
		{0, []byte{0x00}, 0, io.ErrShortWrite, io.EOF},
		// Force errors on 2-byte read/write.
		{0xfd, []byte{0xfd}, 0, io.ErrShortWrite, io.EOF},              // error on writing length
		{0xfd, []byte{0xfd}, 2, io.ErrShortWrite, io.ErrUnexpectedEOF}, // error on writing actual data
		// Force errors on 4-byte read/write.
		{0x10000, []byte{0xfe}, 0, io.ErrShortWrite, io.EOF},              // error on writing length
		{0x10000, []byte{0xfe}, 2, io.ErrShortWrite, io.ErrUnexpectedEOF}, // error on writing actual data
		// Force errors on 8-byte read/write.
		{0x100000000, []byte{0xff}, 0, io.ErrShortWrite, io.EOF},              // error on writing length
		{0x100000000, []byte{0xff}, 2, io.ErrShortWrite, io.ErrUnexpectedEOF}, // error on writing actual data
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := WriteVarInt(w, test.in)
		if err != test.writeErr {
			t.Errorf("WriteVarInt #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Decode from wire format.
		r := newFixedReader(test.max, test.buf)
		_, err = ReadVarInt(r)
		if err != test.readErr {
			t.Errorf("ReadVarInt #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestVarIntNonCanonical ensures variable length integers that are not encoded
// canonically return the expected error.
func TestVarIntNonCanonical(t *testing.T) {
	pver := ProtocolVersion

	tests := []struct {
		name string // Test name for easier identification
		in   []byte // Value to decode
		pver uint32 // Protocol version for wire encoding
	}{
		{
			"0 encoded with 3 bytes", []byte{0xfd, 0x00, 0x00},
			pver,
		},
		{
			"max single-byte value encoded with 3 bytes",
			[]byte{0xfd, 0xfc, 0x00}, pver,
		},
		{
			"0 encoded with 5 bytes",
			[]byte{0xfe, 0x00, 0x00, 0x00, 0x00}, pver,
		},
		{
			"max three-byte value encoded with 5 bytes",
			[]byte{0xfe, 0xff, 0xff, 0x00, 0x00}, pver,
		},
		{
			"0 encoded with 9 bytes",
			[]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			pver,
		},
		{
			"max five-byte value encoded with 9 bytes",
			[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00},
			pver,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Decode from wire format.
		rbuf := bytes.NewReader(test.in)
		val, err := ReadVarInt(rbuf)
		if _, ok := err.(*MessageError); !ok {
			t.Errorf("ReadVarInt #%d (%s) unexpected error %v", i,
				test.name, err)
			continue
		}
		if val != 0 {
			t.Errorf("ReadVarInt #%d (%s)\n got: %d want: 0", i,
				test.name, val)
			continue
		}
	}
}

// TestVarIntWire tests the serialize size for variable length integers.
func TestVarIntSerializeSize(t *testing.T) {
	tests := []struct {
		val  uint64 // Value to get the serialized size for
		size int    // Expected serialized size
	}{
		// Single byte
		{0, 1},
		// Max single byte
		{0xfc, 1},
		// Min 2-byte
		{0xfd, 3},
		// Max 2-byte
		{0xffff, 3},
		// Min 4-byte
		{0x10000, 5},
		// Max 4-byte
		{0xffffffff, 5},
		// Min 8-byte
		{0x100000000, 9},
		// Max 8-byte
		{0xffffffffffffffff, 9},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		serializedSize := VarIntSerializeSize(test.val)
		if serializedSize != test.size {
			t.Errorf("VarIntSerializeSize #%d got: %d, want: %d", i,
				serializedSize, test.size)
			continue
		}
	}
}

// TestVarStringWire tests wire encode and decode for variable length strings.
func TestVarStringWire(t *testing.T) {
	pver := ProtocolVersion

	// str256 is a string that takes a 2-byte varint to encode.
	str256 := strings.Repeat("test", 64)

	tests := []struct {
		in   string // String to encode
		out  string // String to decoded value
		buf  []byte // Wire encoding
		pver uint32 // Protocol version for wire encoding
	}{
		// Latest protocol version.
		// Empty string
		{"", "", []byte{0x00}, pver},
		// Single byte varint + string
		{"Test", "Test", append([]byte{0x04}, []byte("Test")...), pver},
		// 2-byte varint + string
		{str256, str256, append([]byte{0xfd, 0x00, 0x01}, []byte(str256)...), pver},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		var buf bytes.Buffer
		err := WriteVarString(&buf, test.in)
		if err != nil {
			t.Errorf("WriteVarString #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("WriteVarString #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode from wire format.
		rbuf := bytes.NewReader(test.buf)
		val, err := ReadVarString(rbuf, test.pver)
		if err != nil {
			t.Errorf("ReadVarString #%d error %v", i, err)
			continue
		}
		if val != test.out {
			t.Errorf("ReadVarString #%d\n got: %s want: %s", i,
				val, test.out)
			continue
		}
	}
}

// TestVarStringWireErrors performs negative tests against wire encode and
// decode of variable length strings to confirm error paths work correctly.
func TestVarStringWireErrors(t *testing.T) {
	pver := ProtocolVersion

	// str256 is a string that takes a 2-byte varint to encode.
	str256 := strings.Repeat("test", 64)

	tests := []struct {
		in       string // Value to encode
		buf      []byte // Wire encoding
		pver     uint32 // Protocol version for wire encoding
		max      int    // Max size of fixed buffer to induce errors
		writeErr error  // Expected write error
		readErr  error  // Expected read error
	}{
		// Latest protocol version with intentional read/write errors.
		// Force errors on empty string.
		{"", []byte{0x00}, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error on single byte varint + string.
		{"Test", []byte{0x04}, pver, 2, io.ErrShortWrite, io.ErrUnexpectedEOF},
		// Force errors on 2-byte varint + string.
		{str256, []byte{0xfd}, pver, 2, io.ErrShortWrite, io.ErrUnexpectedEOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := WriteVarString(w, test.in)
		if err != test.writeErr {
			t.Errorf("WriteVarString #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Decode from wire format.
		r := newFixedReader(test.max, test.buf)
		_, err = ReadVarString(r, test.pver)
		if err != test.readErr {
			t.Errorf("ReadVarString #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestVarStringOverflowErrors performs tests to ensure deserializing variable
// length strings intentionally crafted to use large values for the string
// length are handled properly.  This could otherwise potentially be used as an
// attack vector.
func TestVarStringOverflowErrors(t *testing.T) {
	pver := ProtocolVersion

	tests := []struct {
		buf  []byte // Wire encoding
		pver uint32 // Protocol version for wire encoding
		err  error  // Expected error
	}{
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			pver, &MessageError{}},
		{[]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			pver, &MessageError{}},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Decode from wire format.
		rbuf := bytes.NewReader(test.buf)
		_, err := ReadVarString(rbuf, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("ReadVarString #%d wrong error got: %v, "+
				"want: %v", i, err, reflect.TypeOf(test.err))
			continue
		}
	}

}

// TestVarBytesWire tests wire encode and decode for variable length byte array.
func TestVarBytesWire(t *testing.T) {
	pver := ProtocolVersion

	// bytes256 is a byte array that takes a 2-byte varint to encode.
	bytes256 := bytes.Repeat([]byte{0x01}, 256)

	tests := []struct {
		in   []byte // Byte Array to write
		buf  []byte // Wire encoding
		pver uint32 // Protocol version for wire encoding
	}{
		// Latest protocol version.
		// Empty byte array
		{[]byte{}, []byte{0x00}, pver},
		// Single byte varint + byte array
		{[]byte{0x01}, []byte{0x01, 0x01}, pver},
		// 2-byte varint + byte array
		{bytes256, append([]byte{0xfd, 0x00, 0x01}, bytes256...), pver},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		var buf bytes.Buffer
		err := WriteVarBytes(&buf, test.pver, test.in)
		if err != nil {
			t.Errorf("WriteVarBytes #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("WriteVarBytes #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode from wire format.
		rbuf := bytes.NewReader(test.buf)
		val, err := ReadVarBytes(rbuf, test.pver, MaxMessagePayload,
			"test payload")
		if err != nil {
			t.Errorf("ReadVarBytes #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("ReadVarBytes #%d\n got: %s want: %s", i,
				val, test.buf)
			continue
		}
	}
}

// TestVarBytesWireErrors performs negative tests against wire encode and
// decode of variable length byte arrays to confirm error paths work correctly.
func TestVarBytesWireErrors(t *testing.T) {
	pver := ProtocolVersion

	// bytes256 is a byte array that takes a 2-byte varint to encode.
	bytes256 := bytes.Repeat([]byte{0x01}, 256)

	tests := []struct {
		in       []byte // Byte Array to write
		buf      []byte // Wire encoding
		pver     uint32 // Protocol version for wire encoding
		max      int    // Max size of fixed buffer to induce errors
		writeErr error  // Expected write error
		readErr  error  // Expected read error
	}{
		// Latest protocol version with intentional read/write errors.
		// Force errors on empty byte array.
		{[]byte{}, []byte{0x00}, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error on single byte varint + byte array.
		{[]byte{0x01, 0x02, 0x03}, []byte{0x04}, pver, 2, io.ErrShortWrite, io.ErrUnexpectedEOF},
		// Force errors on 2-byte varint + byte array.
		{bytes256, []byte{0xfd}, pver, 2, io.ErrShortWrite, io.ErrUnexpectedEOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := WriteVarBytes(w, test.pver, test.in)
		if err != test.writeErr {
			t.Errorf("WriteVarBytes #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Decode from wire format.
		r := newFixedReader(test.max, test.buf)
		_, err = ReadVarBytes(r, test.pver, MaxMessagePayload,
			"test payload")
		if err != test.readErr {
			t.Errorf("ReadVarBytes #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestVarBytesOverflowErrors performs tests to ensure deserializing variable
// length byte arrays intentionally crafted to use large values for the array
// length are handled properly.  This could otherwise potentially be used as an
// attack vector.
func TestVarBytesOverflowErrors(t *testing.T) {
	pver := ProtocolVersion

	tests := []struct {
		buf  []byte // Wire encoding
		pver uint32 // Protocol version for wire encoding
		err  error  // Expected error
	}{
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			pver, &MessageError{}},
		{[]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			pver, &MessageError{}},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Decode from wire format.
		rbuf := bytes.NewReader(test.buf)
		_, err := ReadVarBytes(rbuf, test.pver, MaxMessagePayload,
			"test payload")
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("ReadVarBytes #%d wrong error got: %v, "+
				"want: %v", i, err, reflect.TypeOf(test.err))
			continue
		}
	}

}
