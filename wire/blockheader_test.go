// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/davecgh/go-spew/spew"
)

// TestBlockHeader tests the BlockHeader API.
func TestBlockHeader(t *testing.T) {
	nonce64, err := RandomUint64()
	if err != nil {
		t.Errorf("RandomUint64: Error generating nonce: %v", err)
	}
	nonce := uint32(nonce64)

	hashes := []daghash.Hash{mainNetGenesisHash, simNetGenesisHash}

	merkleHash := mainNetGenesisMerkleRoot
	bits := uint32(0x1d00ffff)
	bh := NewBlockHeader(1, hashes, &merkleHash, bits, nonce)

	// Ensure we get the same data back out.
	if !reflect.DeepEqual(bh.PrevBlocks, hashes) {
		t.Errorf("NewBlockHeader: wrong prev hashes - got %v, want %v",
			spew.Sprint(bh.PrevBlocks), spew.Sprint(hashes))
	}
	if !bh.MerkleRoot.IsEqual(&merkleHash) {
		t.Errorf("NewBlockHeader: wrong merkle root - got %v, want %v",
			spew.Sprint(bh.MerkleRoot), spew.Sprint(merkleHash))
	}
	if bh.Bits != bits {
		t.Errorf("NewBlockHeader: wrong bits - got %v, want %v",
			bh.Bits, bits)
	}
	if bh.Nonce != nonce {
		t.Errorf("NewBlockHeader: wrong nonce - got %v, want %v",
			bh.Nonce, nonce)
	}
}

// TestBlockHeaderWire tests the BlockHeader wire encode and decode for various
// protocol versions.
func TestBlockHeaderWire(t *testing.T) {
	nonce := uint32(123123) // 0x1e0f3
	pver := uint32(70001)

	// baseBlockHdr is used in the various tests as a baseline BlockHeader.
	bits := uint32(0x1d00ffff)
	baseBlockHdr := &BlockHeader{
		Version:       1,
		NumPrevBlocks: 2,
		PrevBlocks:    []daghash.Hash{mainNetGenesisHash, simNetGenesisHash},
		MerkleRoot:    mainNetGenesisMerkleRoot,
		Timestamp:     time.Unix(0x495fab29, 0), // 2009-01-03 12:15:05 -0600 CST
		Bits:          bits,
		Nonce:         nonce,
	}

	// baseBlockHdrEncoded is the wire encoded bytes of baseBlockHdr.
	baseBlockHdrEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version 1
		0x02,                                           // NumPrevBlocks
		0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, // PrevBlock mainNetGenesisHash
		0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
		0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
		0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a, // PrevBlock simNetGenesisHash
		0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
		0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
		0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
		0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2, // MerkleRoot
		0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
		0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
		0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a,
		0x29, 0xab, 0x5f, 0x49, // Timestamp
		0xff, 0xff, 0x00, 0x1d, // Bits
		0xf3, 0xe0, 0x01, 0x00, // Nonce
	}

	tests := []struct {
		in   *BlockHeader // Data to encode
		out  *BlockHeader // Expected decoded data
		buf  []byte       // Wire encoding
		pver uint32       // Protocol version for wire encoding
	}{
		// Latest protocol version.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			ProtocolVersion,
		},

		// Protocol version BIP0035Version.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			BIP0035Version,
		},

		// Protocol version BIP0031Version.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			BIP0031Version,
		},

		// Protocol version NetAddressTimeVersion.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			NetAddressTimeVersion,
		},

		// Protocol version MultipleAddressVersion.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			MultipleAddressVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		var buf bytes.Buffer
		err := writeBlockHeader(&buf, test.pver, test.in)
		if err != nil {
			t.Errorf("writeBlockHeader #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("writeBlockHeader #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		buf.Reset()
		err = test.in.BtcEncode(&buf, pver)
		if err != nil {
			t.Errorf("BtcEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("BtcEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the block header from wire format.
		var bh BlockHeader
		rbuf := bytes.NewReader(test.buf)
		err = readBlockHeader(rbuf, test.pver, &bh)
		if err != nil {
			t.Errorf("readBlockHeader #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&bh, test.out) {
			t.Errorf("readBlockHeader #%d\n got: %s want: %s", i,
				spew.Sdump(&bh), spew.Sdump(test.out))
			continue
		}

		rbuf = bytes.NewReader(test.buf)
		err = bh.BtcDecode(rbuf, pver)
		if err != nil {
			t.Errorf("BtcDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&bh, test.out) {
			t.Errorf("BtcDecode #%d\n got: %s want: %s", i,
				spew.Sdump(&bh), spew.Sdump(test.out))
			continue
		}
	}
}

// TestBlockHeaderSerialize tests BlockHeader serialize and deserialize.
func TestBlockHeaderSerialize(t *testing.T) {
	nonce := uint32(123123) // 0x1e0f3

	// baseBlockHdr is used in the various tests as a baseline BlockHeader.
	bits := uint32(0x1d00ffff)
	baseBlockHdr := &BlockHeader{
		Version:       1,
		NumPrevBlocks: 2,
		PrevBlocks:    []daghash.Hash{mainNetGenesisHash, simNetGenesisHash},
		MerkleRoot:    mainNetGenesisMerkleRoot,
		Timestamp:     time.Unix(0x495fab29, 0), // 2009-01-03 12:15:05 -0600 CST
		Bits:          bits,
		Nonce:         nonce,
	}

	// baseBlockHdrEncoded is the wire encoded bytes of baseBlockHdr.
	baseBlockHdrEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version 1
		0x02,                                           // NumPrevBlocks
		0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, // PrevBlock mainNetGenesisHash
		0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
		0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
		0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a, // PrevBlock simNetGenesisHash
		0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
		0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
		0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
		0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2, // MerkleRoot
		0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
		0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
		0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a,
		0x29, 0xab, 0x5f, 0x49, // Timestamp
		0xff, 0xff, 0x00, 0x1d, // Bits
		0xf3, 0xe0, 0x01, 0x00, // Nonce
	}

	tests := []struct {
		in  *BlockHeader // Data to encode
		out *BlockHeader // Expected decoded data
		buf []byte       // Serialized data
	}{
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Serialize the block header.
		var buf bytes.Buffer
		err := test.in.Serialize(&buf)
		if err != nil {
			t.Errorf("Serialize #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("Serialize #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Deserialize the block header.
		var bh BlockHeader
		rbuf := bytes.NewReader(test.buf)
		err = bh.Deserialize(rbuf)
		if err != nil {
			t.Errorf("Deserialize #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&bh, test.out) {
			t.Errorf("Deserialize #%d\n got: %s want: %s", i,
				spew.Sdump(&bh), spew.Sdump(test.out))
			continue
		}
	}
}

// TestBlockHeaderSerializeSize performs tests to ensure the serialize size for
// various block headers is accurate.
func TestBlockHeaderSerializeSize(t *testing.T) {
	nonce := uint32(123123) // 0x1e0f3
	bits := uint32(0x1d00ffff)
	timestamp := time.Unix(0x495fab29, 0) // 2009-01-03 12:15:05 -0600 CST
	baseBlockHdr := &BlockHeader{
		Version:       1,
		NumPrevBlocks: 2,
		PrevBlocks:    []daghash.Hash{mainNetGenesisHash, simNetGenesisHash},
		MerkleRoot:    mainNetGenesisMerkleRoot,
		Timestamp:     timestamp,
		Bits:          bits,
		Nonce:         nonce,
	}

	genesisBlockHdr := &BlockHeader{
		Version:       1,
		NumPrevBlocks: 0,
		PrevBlocks:    []daghash.Hash{},
		MerkleRoot:    mainNetGenesisMerkleRoot,
		Timestamp:     timestamp,
		Bits:          bits,
		Nonce:         nonce,
	}
	tests := []struct {
		in   *BlockHeader // Block header to encode
		size int          // Expected serialized size
	}{
		// Block with no transactions.
		{genesisBlockHdr, 49},

		// First block in the mainnet block chain.
		{baseBlockHdr, 113},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		serializedSize := test.in.SerializeSize()
		if serializedSize != test.size {
			t.Errorf("BlockHeader.SerializeSize: #%d got: %d, want: "+
				"%d", i, serializedSize, test.size)

			continue
		}
	}
}

func TestIsGenesis(t *testing.T) {
	nonce := uint32(123123) // 0x1e0f3
	bits := uint32(0x1d00ffff)
	timestamp := time.Unix(0x495fab29, 0) // 2009-01-03 12:15:05 -0600 CST

	baseBlockHdr := &BlockHeader{
		Version:       1,
		NumPrevBlocks: 2,
		PrevBlocks:    []daghash.Hash{mainNetGenesisHash, simNetGenesisHash},
		MerkleRoot:    mainNetGenesisMerkleRoot,
		Timestamp:     timestamp,
		Bits:          bits,
		Nonce:         nonce,
	}
	genesisBlockHdr := &BlockHeader{
		Version:       1,
		NumPrevBlocks: 0,
		PrevBlocks:    []daghash.Hash{},
		MerkleRoot:    mainNetGenesisMerkleRoot,
		Timestamp:     timestamp,
		Bits:          bits,
		Nonce:         nonce,
	}

	tests := []struct {
		in        *BlockHeader // Block header to encode
		isGenesis bool         // Expected result for call of .IsGenesis
	}{
		{genesisBlockHdr, true},
		{baseBlockHdr, false},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		isGenesis := test.in.IsGenesis()
		if isGenesis != test.isGenesis {
			t.Errorf("BlockHeader.IsGenesis: #%d got: %t, want: %t",
				i, isGenesis, test.isGenesis)
		}
	}
}
