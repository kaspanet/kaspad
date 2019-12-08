// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bloom_test

import (
	"bytes"
	"testing"

	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/bloom"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/davecgh/go-spew/spew"
)

func TestMerkleBlock3(t *testing.T) {
	blockBytes := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x01,                                                             // NumParentBlocks
		0x79, 0xCD, 0xA8, 0x56, 0xB1, 0x43, 0xD9, 0xDB, 0x2C, 0x1C, 0xAF, // ParentHashes
		0xF0, 0x1D, 0x1A, 0xEC, 0xC8, 0x63, 0x0D, 0x30, 0x62, 0x5D, 0x10,
		0xE8, 0xB4, 0xB8, 0xB0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xB5, 0x0C, 0xC0, 0x69, 0xD6, 0xA3, 0xE3, 0x3E, 0x3F, 0xF8, 0x4A, // HashMerkleRoot
		0x5C, 0x41, 0xD9, 0xD3, 0xFE, 0xBE, 0x7C, 0x77, 0x0F, 0xDC, 0xC9,
		0x6B, 0x2C, 0x3F, 0xF6, 0x0A, 0xBE, 0x18, 0x4F, 0x19, 0x63,
		0x3C, 0xE3, 0x70, 0xD9, 0x5F, 0x09, 0x3B, 0xC7, 0xE3, 0x67, 0x11, // AcceptedIDMerkleRoot
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63, 0x65, 0x9C, 0x79,
		0x7B, 0x3C, 0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x10, 0x3B, 0xC7, 0xE3, 0x67, 0x11, 0x7B, 0x3C, // UTXOCommitment
		0x30, 0xC1, 0xF8, 0xFD, 0xD0, 0xD9, 0x72, 0x87,
		0x7F, 0x16, 0xC5, 0x96, 0x2E, 0x8B, 0xD9, 0x63,
		0x65, 0x9C, 0x79, 0x3C, 0xE3, 0x70, 0xD9, 0x5F,
		0x67, 0x29, 0x1B, 0x4D, 0x00, 0x00, 0x00, 0x00, //Time
		0x4C, 0x86, 0x04, 0x1B, // Bits
		0x8F, 0xA4, 0x5D, 0x63, 0x00, 0x00, 0x00, 0x00, // Fake Nonce. TODO: (Ori) Replace to a real nonce
		0x01, // NumTxs
		0x01, 0x00, 0x00, 0x00, 0x01, 0x9b, 0x22, 0x59,
		0x44, 0x66, 0xf0, 0xbe, 0x50, 0x7c, 0x1c, 0x8a, // Tx
		0xf6, 0x06, 0x27, 0xe6, 0x33, 0x38, 0x7e, 0xd1,
		0xd5, 0x8c, 0x42, 0x59, 0x1a, 0x31, 0xac, 0x9a,
		0xa6, 0x2e, 0xd5, 0x2b, 0x0f, 0xff, 0xff, 0xff,
		0xff, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0x01, 0x00, 0xf2, 0x05, 0x2a, 0x01,
		0x00, 0x00, 0x00, 0x17, 0xa9, 0x14, 0xda, 0x17,
		0x45, 0xe9, 0xb5, 0x49, 0xbd, 0x0b, 0xfa, 0x1a,
		0x56, 0x99, 0x71, 0xc7, 0x7e, 0xba, 0x30, 0xcd,
		0x5a, 0x4b, 0x87, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14,
		0x06, 0xe0, 0x58, 0x81, 0xe2, 0x99, 0x36, 0x77,
		0x66, 0xd3, 0x13, 0xe2, 0x6c, 0x05, 0x56, 0x4e,
		0xc9, 0x1b, 0xf7, 0x21, 0xd3, 0x17, 0x26, 0xbd,
		0x6e, 0x46, 0xe6, 0x06, 0x89, 0x53, 0x9a, 0x01,
		0x00,
	}

	blk, err := util.NewBlockFromBytes(blockBytes)
	if err != nil {
		t.Errorf("TestMerkleBlock3 NewBlockFromBytes failed: %v", err)
		return
	}

	f := bloom.NewFilter(10, 0, 0.000001, wire.BloomUpdateAll)

	inputStr := "4ee77df1e2c3126a4a3469e7b1ee3c73093f7f79fef726690fde230c47a02dc6"
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
		0xbe, 0x18, 0x4f, 0x19, 0x63, 0x3c, 0xe3, 0x70, 0xd9, 0x5f, 0x09, 0x3b, 0xc7, 0xe3, 0x67, 0x11,
		0x7f, 0x16, 0xc5, 0x96, 0x2e, 0x8b, 0xd9, 0x63, 0x65, 0x9c, 0x79, 0x7b, 0x3c, 0x30, 0xc1, 0xf8,
		0xfd, 0xd0, 0xd9, 0x72, 0x87, 0x10, 0x3b, 0xc7, 0xe3, 0x67, 0x11, 0x7b, 0x3c, 0x30, 0xc1, 0xf8,
		0xfd, 0xd0, 0xd9, 0x72, 0x87, 0x7f, 0x16, 0xc5, 0x96, 0x2e, 0x8b, 0xd9, 0x63, 0x65, 0x9c, 0x79,
		0x3c, 0xe3, 0x70, 0xd9, 0x5f, 0x67, 0x29, 0x1b, 0x4d, 0x00, 0x00, 0x00, 0x00, 0x4c, 0x86, 0x04,
		0x1b, 0x8f, 0xa4, 0x5d, 0x63, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x27, 0x8a,
		0xff, 0xcc, 0xbf, 0xd9, 0x15, 0x12, 0xf8, 0x91, 0xbd, 0xd3, 0x46, 0x91, 0x3d, 0xe6, 0x13, 0xdc,
		0x9c, 0xb9, 0x7e, 0xa3, 0xc5, 0x8f, 0xab, 0x6b, 0xd8, 0xb7, 0x2b, 0x97, 0x39, 0x03, 0x01, 0x00,
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
