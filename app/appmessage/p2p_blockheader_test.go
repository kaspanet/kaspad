// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/util/random"
)

// TestBlockHeader tests the BlockHeader API.
func TestBlockHeader(t *testing.T) {
	nonce, err := random.Uint64()
	if err != nil {
		t.Errorf("random.Uint64: Error generating nonce: %v", err)
	}

	hashes := []*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash}

	merkleHash := mainnetGenesisMerkleRoot
	acceptedIDMerkleRoot := exampleAcceptedIDMerkleRoot
	bits := uint32(0x1d00ffff)
	bh := NewBlockHeader(1, hashes, merkleHash, acceptedIDMerkleRoot, exampleUTXOCommitment, bits, nonce)

	// Ensure we get the same data back out.
	if !reflect.DeepEqual(bh.ParentHashes, hashes) {
		t.Errorf("NewBlockHeader: wrong prev hashes - got %v, want %v",
			spew.Sprint(bh.ParentHashes), spew.Sprint(hashes))
	}
	if bh.HashMerkleRoot != merkleHash {
		t.Errorf("NewBlockHeader: wrong merkle root - got %v, want %v",
			spew.Sprint(bh.HashMerkleRoot), spew.Sprint(merkleHash))
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

// TestBlockHeaderEncoding tests the BlockHeader appmessage encode and decode for various
// protocol versions.
func TestBlockHeaderEncoding(t *testing.T) {
	nonce := uint64(123123) // 0x000000000001e0f3
	pver := ProtocolVersion

	// baseBlockHdr is used in the various tests as a baseline BlockHeader.
	bits := uint32(0x1d00ffff)
	baseBlockHdr := &BlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash},
		HashMerkleRoot:       mainnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: exampleAcceptedIDMerkleRoot,
		UTXOCommitment:       exampleUTXOCommitment,
		Timestamp:            mstime.UnixMilliseconds(0x17315ed0f99),
		Bits:                 bits,
		Nonce:                nonce,
	}

	// baseBlockHdrEncoded is the appmessage encoded bytes of baseBlockHdr.
	baseBlockHdrEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version 1
		0x02,                                           // NumParentBlocks
		0xdc, 0x5f, 0x5b, 0x5b, 0x1d, 0xc2, 0xa7, 0x25, // mainnetGenesisHash
		0x49, 0xd5, 0x1d, 0x4d, 0xee, 0xd7, 0xa4, 0x8b,
		0xaf, 0xd3, 0x14, 0x4b, 0x56, 0x78, 0x98, 0xb1,
		0x8c, 0xfd, 0x9f, 0x69, 0xdd, 0xcf, 0xbb, 0x63,
		0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a, // simnetGenesisHash
		0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
		0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
		0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
		0x4a, 0x5e, 0x1e, 0x4b, 0xaa, 0xb8, 0x9f, 0x3a, // HashMerkleRoot
		0x32, 0x51, 0x8a, 0x88, 0xc3, 0x1b, 0xc8, 0x7f,
		0x61, 0x8f, 0x76, 0x67, 0x3e, 0x2c, 0xc7, 0x7a,
		0xb2, 0x12, 0x7b, 0x7a, 0xfd, 0xed, 0xa3, 0x3b,
		0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // AcceptedIDMerkleRoot
		0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
		0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
		0x10, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // UTXOCommitment
		0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
		0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
		0x99, 0x0f, 0xed, 0x15, 0x73, 0x01, 0x00, 0x00, // Timestamp
		0xff, 0xff, 0x00, 0x1d, // Bits
		0xf3, 0xe0, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, // Fake Nonce
	}

	tests := []struct {
		in   *BlockHeader // Data to encode
		out  *BlockHeader // Expected decoded data
		buf  []byte       // Encoded data
		pver uint32       // Protocol version for appmessage encoding
	}{
		// Latest protocol version.
		{
			baseBlockHdr,
			baseBlockHdr,
			baseBlockHdrEncoded,
			ProtocolVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to appmessage format.
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
		err = test.in.KaspaEncode(&buf, pver)
		if err != nil {
			t.Errorf("KaspaEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("KaspaEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the block header from appmessage format.
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
		err = bh.KaspaDecode(rbuf, pver)
		if err != nil {
			t.Errorf("KaspaDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&bh, test.out) {
			t.Errorf("KaspaDecode #%d\n got: %s want: %s", i,
				spew.Sdump(&bh), spew.Sdump(test.out))
			continue
		}
	}
}

// TestBlockHeaderSerialize tests BlockHeader serialize and deserialize.
func TestBlockHeaderSerialize(t *testing.T) {
	nonce := uint64(123123) // 0x01e0f3

	// baseBlockHdr is used in the various tests as a baseline BlockHeader.
	bits := uint32(0x1d00ffff)
	baseBlockHdr := &BlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash},
		HashMerkleRoot:       mainnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: exampleAcceptedIDMerkleRoot,
		UTXOCommitment:       exampleUTXOCommitment,
		Timestamp:            mstime.UnixMilliseconds(0x17315ed0f99),
		Bits:                 bits,
		Nonce:                nonce,
	}

	// baseBlockHdrEncoded is the appmessage encoded bytes of baseBlockHdr.
	baseBlockHdrEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version 1
		0x02,                                           // NumParentBlocks
		0xdc, 0x5f, 0x5b, 0x5b, 0x1d, 0xc2, 0xa7, 0x25, // mainnetGenesisHash
		0x49, 0xd5, 0x1d, 0x4d, 0xee, 0xd7, 0xa4, 0x8b,
		0xaf, 0xd3, 0x14, 0x4b, 0x56, 0x78, 0x98, 0xb1,
		0x8c, 0xfd, 0x9f, 0x69, 0xdd, 0xcf, 0xbb, 0x63,
		0xf6, 0x7a, 0xd7, 0x69, 0x5d, 0x9b, 0x66, 0x2a, // simnetGenesisHash
		0x72, 0xff, 0x3d, 0x8e, 0xdb, 0xbb, 0x2d, 0xe0,
		0xbf, 0xa6, 0x7b, 0x13, 0x97, 0x4b, 0xb9, 0x91,
		0x0d, 0x11, 0x6d, 0x5c, 0xbd, 0x86, 0x3e, 0x68,
		0x4a, 0x5e, 0x1e, 0x4b, 0xaa, 0xb8, 0x9f, 0x3a, // HashMerkleRoot
		0x32, 0x51, 0x8a, 0x88, 0xc3, 0x1b, 0xc8, 0x7f,
		0x61, 0x8f, 0x76, 0x67, 0x3e, 0x2c, 0xc7, 0x7a,
		0xb2, 0x12, 0x7b, 0x7a, 0xfd, 0xed, 0xa3, 0x3b,
		0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // AcceptedIDMerkleRoot
		0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
		0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
		0x10, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // UTXOCommitment
		0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
		0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
		0x99, 0x0f, 0xed, 0x15, 0x73, 0x01, 0x00, 0x00, // Timestamp
		0xff, 0xff, 0x00, 0x1d, // Bits
		0xf3, 0xe0, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, // Fake Nonce
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
	nonce := uint64(123123) // 0x1e0f3
	bits := uint32(0x1d00ffff)
	timestamp := mstime.UnixMilliseconds(0x495fab29000)
	baseBlockHdr := &BlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash},
		HashMerkleRoot:       mainnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &externalapi.DomainHash{},
		UTXOCommitment:       &externalapi.DomainHash{},
		Timestamp:            timestamp,
		Bits:                 bits,
		Nonce:                nonce,
	}

	genesisBlockHdr := &BlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       mainnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &externalapi.DomainHash{},
		UTXOCommitment:       &externalapi.DomainHash{},
		Timestamp:            timestamp,
		Bits:                 bits,
		Nonce:                nonce,
	}
	tests := []struct {
		in   *BlockHeader // Block header to encode
		size int          // Expected serialized size
	}{
		// Block with no transactions.
		{genesisBlockHdr, 121},

		// First block in the mainnet block DAG.
		{baseBlockHdr, 185},
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
	nonce := uint64(123123) // 0x1e0f3
	bits := uint32(0x1d00ffff)
	timestamp := mstime.UnixMilliseconds(0x495fab29000)

	baseBlockHdr := &BlockHeader{
		Version:        1,
		ParentHashes:   []*externalapi.DomainHash{mainnetGenesisHash, simnetGenesisHash},
		HashMerkleRoot: mainnetGenesisMerkleRoot,
		Timestamp:      timestamp,
		Bits:           bits,
		Nonce:          nonce,
	}
	genesisBlockHdr := &BlockHeader{
		Version:        1,
		ParentHashes:   []*externalapi.DomainHash{},
		HashMerkleRoot: mainnetGenesisMerkleRoot,
		Timestamp:      timestamp,
		Bits:           bits,
		Nonce:          nonce,
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
