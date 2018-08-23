// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bloom_test

import (
	"bytes"
	"testing"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
	"github.com/daglabs/btcd/util"
	"github.com/davecgh/go-spew/spew"
	"github.com/daglabs/btcd/util/bloom"
)

func TestMerkleBlock3(t *testing.T) {
	blockBytes := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x01,                                                             //NumPrevBlocks
		0x79, 0xCD, 0xA8, 0x56, 0xB1, 0x43, 0xD9, 0xDB, 0x2C, 0x1C, 0xAF, //HashPrevBlocks
		0xF0, 0x1D, 0x1A, 0xEC, 0xC8, 0x63, 0x0D, 0x30, 0x62, 0x5D, 0x10,
		0xE8, 0xB4, 0xB8, 0xB0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xB5, 0x0C, 0xC0, 0x69, 0xD6, 0xA3, 0xE3, 0x3E, 0x3F, 0xF8, 0x4A, // HashMerkleRoot
		0x5C, 0x41, 0xD9, 0xD3, 0xFE, 0xBE, 0x7C, 0x77, 0x0F, 0xDC, 0xC9,
		0x6B, 0x2C, 0x3F, 0xF6, 0x0A, 0xBE, 0x18, 0x4F, 0x19, 0x63,
		0x67, 0x29, 0x1B, 0x4D, 0x00, 0x00, 0x00, 0x00, //Time
		0x4C, 0x86, 0x04, 0x1B, // Bits
		0x8F, 0xA4, 0x5D, 0x63, // Nonce
		0x01,                                                             // NumTxs
		0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Tx
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0x08, 0x04, 0x4C,
		0x86, 0x04, 0x1B, 0x02, 0x0A, 0x02, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01,
		0x00, 0xF2, 0x05, 0x2A, 0x01, 0x00, 0x00, 0x00, 0x43, 0x41, 0x04,
		0xEC, 0xD3, 0x22, 0x9B, 0x05, 0x71, 0xC3, 0xBE, 0x87, 0x6F, 0xEA,
		0xAC, 0x04, 0x42, 0xA9, 0xF1, 0x3C, 0x5A, 0x57, 0x27, 0x42, 0x92,
		0x7A, 0xF1, 0xDC, 0x62, 0x33, 0x53, 0xEC, 0xF8, 0xC2, 0x02, 0x22,
		0x5F, 0x64, 0x86, 0x81, 0x37, 0xA1, 0x8C, 0xDD, 0x85, 0xCB, 0xBB,
		0x4C, 0x74, 0xFB, 0xCC, 0xFD, 0x4F, 0x49, 0x63, 0x9C, 0xF1, 0xBD,
		0xC9, 0x4A, 0x56, 0x72, 0xBB, 0x15, 0xAD, 0x5D, 0x4C, 0xAC, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	blk, err := util.NewBlockFromBytes(blockBytes)
	if err != nil {
		t.Errorf("TestMerkleBlock3 NewBlockFromBytes failed: %v", err)
		return
	}

	f := bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)

	inputStr := "63194f18be0af63f2c6bc9dc0f777cbefed3d9415c4af83f3ee3a3d669c00cb5"
	hash, err := daghash.NewHashFromStr(inputStr)
	if err != nil {
		t.Errorf("TestMerkleBlock3 NewHashFromStr failed: %v", err)
		return
	}

	f.AddHash(hash)

	mBlock, _ := bloom.NewMerkleBlock(blk, f)

	want := []byte{
		0x01, 0x00, 0x00, 0x00, 0x01, 0x79, 0xcd, 0xa8, 0x56, 0xb1, 0x43, 0xd9, 0xdb, 0x2c, 0x1c, 0xaf,
		0xf0, 0x1d, 0x1a, 0xec, 0xc8, 0x63, 0x0d, 0x30, 0x62, 0x5d, 0x10, 0xe8, 0xb4, 0xb8, 0xb0, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xb5, 0x0c, 0xc0, 0x69, 0xd6, 0xa3, 0xe3, 0x3e, 0x3f, 0xf8, 0x4a,
		0x5c, 0x41, 0xd9, 0xd3, 0xfe, 0xbe, 0x7c, 0x77, 0x0f, 0xdc, 0xc9, 0x6b, 0x2c, 0x3f, 0xf6, 0x0a,
		0xbe, 0x18, 0x4f, 0x19, 0x63, 0x67, 0x29, 0x1b, 0x4d, 0x00, 0x00, 0x00, 0x00, 0x4c, 0x86, 0x04,
		0x1b, 0x8f, 0xa4, 0x5d, 0x63, 0x01, 0x00, 0x00, 0x00, 0x01, 0x70, 0x2b, 0x1e, 0x73, 0x5a, 0x15,
		0xbc, 0xd7, 0xdb, 0xc9, 0x6f, 0xe5, 0x0f, 0x86, 0xe9, 0xd1, 0x9b, 0xdf, 0x75, 0x95, 0x30, 0xa6,
		0x95, 0x06, 0x6e, 0xe5, 0x1e, 0xb0, 0xb9, 0x94, 0xe8, 0x01, 0x01, 0x00,
	}
	t.Log(spew.Sdump(want))
	if err != nil {
		t.Errorf("TestMerkleBlock3 DecodeString failed: %v", err)
		return
	}

	got := bytes.NewBuffer(nil)
	err = mBlock.BtcEncode(got, wire.ProtocolVersion)
	if err != nil {
		t.Errorf("TestMerkleBlock3 BtcEncode failed: %v", err)
		return
	}

	if !bytes.Equal(want, got.Bytes()) {
		t.Errorf("TestMerkleBlock3 failed merkle block comparison: "+
			"got:\n %v want:\n %v", spew.Sdump(got.Bytes()), spew.Sdump(want))
		return
	}
}
