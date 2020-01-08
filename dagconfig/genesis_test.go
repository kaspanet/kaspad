// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"bytes"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// TestGenesisBlock tests the genesis block of the main network for validity by
// checking the encoded bytes and hashes.
func TestGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := MainNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), genesisBlockBytes) {
		t.Fatalf("TestGenesisBlock: Genesis block does not appear valid - "+
			"got %v, want %v", spew.Sdump(buf.Bytes()),
			spew.Sdump(genesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := MainNetParams.GenesisBlock.BlockHash()
	if !MainNetParams.GenesisHash.IsEqual(hash) {
		t.Fatalf("TestGenesisBlock: Genesis block hash does not "+
			"appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(MainNetParams.GenesisHash))
	}
}

// TestRegTestGenesisBlock tests the genesis block of the regression test
// network for validity by checking the encoded bytes and hashes.
func TestRegTestGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := RegressionNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestRegTestGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), regTestGenesisBlockBytes) {
		t.Fatalf("TestRegTestGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(regTestGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := RegressionNetParams.GenesisBlock.BlockHash()
	if !RegressionNetParams.GenesisHash.IsEqual(hash) {
		t.Fatalf("TestRegTestGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(RegressionNetParams.GenesisHash))
	}
}

// TestTestNetGenesisBlock tests the genesis block of the test network for
// validity by checking the encoded bytes and hashes.
func TestTestNetGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := TestNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestTestNetGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), testNetGenesisBlockBytes) {
		t.Fatalf("TestTestNetGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(testNetGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := TestNetParams.GenesisBlock.BlockHash()
	if !TestNetParams.GenesisHash.IsEqual(hash) {
		t.Fatalf("TestTestNetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(TestNetParams.GenesisHash))
	}
}

// TestSimNetGenesisBlock tests the genesis block of the simulation test network
// for validity by checking the encoded bytes and hashes.
func TestSimNetGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := SimNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestSimNetGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), simNetGenesisBlockBytes) {
		t.Fatalf("TestSimNetGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(simNetGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := SimNetParams.GenesisBlock.BlockHash()
	if !SimNetParams.GenesisHash.IsEqual(hash) {
		t.Fatalf("TestSimNetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(SimNetParams.GenesisHash))
	}
}

// genesisBlockBytes are the wire encoded bytes for the genesis block of the
// main network as of protocol version 1.
var genesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x72, 0x10, 0x35, 0x85, 0xdd, 0xac, 0x82, 0x5c, 0x49, 0x13, 0x9f,
	0xc0, 0x0e, 0x37, 0xc0, 0x45, 0x71, 0xdf, 0xd9, 0xf6, 0x36, 0xdf, 0x4c, 0x42, 0x72, 0x7b, 0x9e,
	0x86, 0xdd, 0x37, 0xd2, 0xbd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0xb0, 0xc4, 0xda, 0x5c, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7f,
	0x20, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xff, 0xff, 0x0e, 0x00, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48, 0x2f, 0x62, 0x74, 0x63,
	0x64, 0x2f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd2,
	0xea, 0x82, 0x4e, 0xb8, 0x87, 0x42, 0xd0, 0x6d, 0x1f, 0x8d, 0xc3, 0xad, 0x9f, 0x43, 0x9e, 0xed,
	0x6f, 0x43, 0x3c, 0x02, 0x71, 0x71, 0x69, 0xfb, 0xbc, 0x91, 0x44, 0xac, 0xf1, 0x93, 0xd3, 0x18,
	0x17, 0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, 0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71,
	0xc7, 0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
}

// regTestGenesisBlockBytes are the wire encoded bytes for the genesis block of
// the regression test network as of protocol version 1.
var regTestGenesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x3a, 0x9f, 0x62,
	0xc9, 0x2b, 0x16, 0x17, 0xb3, 0x41, 0x6d, 0x9e,
	0x2d, 0x87, 0x93, 0xfd, 0x72, 0x77, 0x4d, 0x1d,
	0x6f, 0x6d, 0x38, 0x5b, 0xf1, 0x24, 0x1b, 0xdc,
	0x96, 0xce, 0xbf, 0xa1, 0x09, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0xd8, 0xe2, 0x15,
	0x5e, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7f,
	0x1e, 0xa6, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xff, 0xff, 0x0e, 0x00, 0x00, 0x0b, 0x2f,
	0x50, 0x32, 0x53, 0x48, 0x2f, 0x62, 0x74, 0x63,
	0x64, 0x2f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xed,
	0x32, 0xec, 0xb4, 0xf8, 0x3c, 0x7a, 0x32, 0x0f,
	0xd2, 0xe5, 0x24, 0x77, 0x89, 0x43, 0x3a, 0x78,
	0x0a, 0xda, 0x68, 0x2d, 0xf6, 0xaa, 0xb1, 0x19,
	0xdd, 0xd8, 0x97, 0x15, 0x4b, 0xcb, 0x42, 0x25,
	0x17, 0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5,
	0x49, 0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71,
	0xc7, 0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x72, 0x65,
	0x67, 0x74, 0x65, 0x73, 0x74,
}

// testNetGenesisBlockBytes are the wire encoded bytes for the genesis block of
// the test network as of protocol version 1.
var testNetGenesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x88, 0x05, 0xd0,
	0xe7, 0x8f, 0x41, 0x77, 0x39, 0x2c, 0xb6, 0xbb,
	0xb4, 0x19, 0xa8, 0x48, 0x4a, 0xdf, 0x77, 0xb0,
	0x82, 0xd6, 0x70, 0xd8, 0x24, 0x6a, 0x36, 0x05,
	0xaa, 0xbd, 0x7a, 0xd1, 0x62, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0xfe, 0xad, 0x15,
	0x5e, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7f,
	0x1e, 0xa1, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xff, 0xff, 0x0e, 0x00, 0x00, 0x0b, 0x2f,
	0x50, 0x32, 0x53, 0x48, 0x2f, 0x62, 0x74, 0x63,
	0x64, 0x2f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xcc,
	0x72, 0xe6, 0x7e, 0x37, 0xa1, 0x34, 0x89, 0x23,
	0x24, 0xaf, 0xae, 0x99, 0x1f, 0x89, 0x09, 0x41,
	0x1a, 0x4d, 0x58, 0xfe, 0x5a, 0x04, 0xb0, 0x3e,
	0xeb, 0x1b, 0x5b, 0xb8, 0x65, 0xa8, 0x65, 0x0f,
	0x01, 0x00, 0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d,
	0x74, 0x65, 0x73, 0x74, 0x6e, 0x65, 0x74,
}

// simNetGenesisBlockBytes are the wire encoded bytes for the genesis block of
// the simulation test network as of protocol version 1.
var simNetGenesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0xb0, 0x1c, 0x3b,
	0x9e, 0x0d, 0x9a, 0xc0, 0x80, 0x0a, 0x08, 0x42,
	0x50, 0x02, 0xa3, 0xea, 0xdb, 0xed, 0xc8, 0xd0,
	0xad, 0x35, 0x03, 0xd8, 0x0e, 0x11, 0x3c, 0x7b,
	0xb2, 0xb5, 0x20, 0xe5, 0x84, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x1c, 0xd3, 0x15,
	0x5e, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7f,
	0x20, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xff, 0xff, 0x0e, 0x00, 0x00, 0x0b, 0x2f,
	0x50, 0x32, 0x53, 0x48, 0x2f, 0x62, 0x74, 0x63,
	0x64, 0x2f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x89,
	0x48, 0xd3, 0x23, 0x9c, 0xf9, 0x88, 0x2b, 0x63,
	0xc7, 0x33, 0x0f, 0xa3, 0x64, 0xf2, 0xdb, 0x39,
	0x73, 0x5f, 0x2b, 0xa8, 0xd5, 0x7b, 0x5c, 0x31,
	0x68, 0xc9, 0x63, 0x37, 0x5c, 0xe7, 0x41, 0x24,
	0x17, 0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5,
	0x49, 0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71,
	0xc7, 0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x73, 0x69,
	0x6d, 0x6e, 0x65, 0x74,
}
