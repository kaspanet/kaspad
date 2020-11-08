// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
)

// TestGenesisBlock tests the genesis block of the main network for validity by
// checking the encoded hash.
func TestGenesisBlock(t *testing.T) {
	// Check hash of the block against expected hash.
	hash := consensusserialization.BlockHash(MainnetParams.GenesisBlock)
	if *MainnetParams.GenesisHash != *hash {
		t.Fatalf("TestGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", hash, MainnetParams.GenesisHash)
	}
}

// TestTestnetGenesisBlock tests the genesis block of the test network for
// validity by checking the hash.
func TestTestnetGenesisBlock(t *testing.T) {
	// Check hash of the block against expected hash.
	hash := consensusserialization.BlockHash(TestnetParams.GenesisBlock)
	if *TestnetParams.GenesisHash != *hash {
		t.Fatalf("TestTestnetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", hash,
			TestnetParams.GenesisHash)
	}
}

// TestSimnetGenesisBlock tests the genesis block of the simulation test network
// for validity by checking the hash.
func TestSimnetGenesisBlock(t *testing.T) {
	// Check hash of the block against expected hash.
	hash := consensusserialization.BlockHash(SimnetParams.GenesisBlock)
	if *SimnetParams.GenesisHash != *hash {
		t.Fatalf("TestSimnetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", hash,
			SimnetParams.GenesisHash)
	}
}

// TestDevnetGenesisBlock tests the genesis block of the development network
// for validity by checking the encoded hash.
func TestDevnetGenesisBlock(t *testing.T) {
	// Check hash of the block against expected hash.
	hash := consensusserialization.BlockHash(DevnetParams.GenesisBlock)
	if *DevnetParams.GenesisHash != *hash {
		t.Fatalf("TestDevnetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", hash,
			DevnetParams.GenesisHash)
	}
}
