// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network (version 3).
var genesisCoinbaseTx = wire.MsgTx{
	Version: 1,
	TxIn: []*wire.TxIn{
		{
			PreviousOutPoint: wire.OutPoint{
				TxID:  daghash.TxID{},
				Index: 0xffffffff,
			},
			SignatureScript: []byte{
				0x00, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
				0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
			},
			Sequence: math.MaxUint64,
		},
	},
	TxOut: []*wire.TxOut{
		{
			Value: 0x12a05f200,
			PkScript: []byte{
				0x51,
			},
		},
	},
	LockTime:     0,
	SubnetworkID: *subnetworkid.SubnetworkIDNative,
}

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xb8, 0xb0, 0x7e, 0x33, 0x56, 0x47, 0xb9, 0xe2,
	0x19, 0x9f, 0xff, 0x44, 0xa8, 0xba, 0x3c, 0x62,
	0xc7, 0xd5, 0x68, 0xe9, 0x3b, 0x2a, 0xd9, 0x3d,
	0xab, 0xf0, 0x98, 0x7b, 0x49, 0xbe, 0x0f, 0x4f,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x76, 0x2b, 0x33, 0xa9, 0x4c, 0xd4, 0x36, 0x13,
	0x29, 0x5e, 0x9b, 0x68, 0xb7, 0xad, 0x2b, 0x16,
	0x7c, 0x63, 0x89, 0xc3, 0x54, 0xc9, 0xa7, 0x06,
	0x8c, 0x23, 0x24, 0x3c, 0x53, 0x6d, 0x56, 0x23,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:        1,
		ParentHashes:   []daghash.Hash{},
		HashMerkleRoot: genesisMerkleRoot,
		IDMerkleRoot:   genesisMerkleRoot,
		Timestamp:      time.Unix(0x5c3cafec, 0),
		Bits:           0x207fffff,
		Nonce:          0,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = genesisHash

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the regression test network.  It is the same as the merkle root for
// the main network.
var regTestGenesisMerkleRoot = genesisMerkleRoot

// regTestGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the regression test network.
var regTestGenesisBlock = genesisBlock

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = genesisHash

// testNet3GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 3).  It is the same as the merkle root
// for the main network.
var testNet3GenesisMerkleRoot = genesisMerkleRoot

// testNet3GenesisBlock defines the genesis block of the block chain which
// serves as the public transaction ledger for the test network (version 3).
var testNet3GenesisBlock = genesisBlock

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = genesisHash

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the simulation test network.  It is the same as the merkle root for
// the main network.
var simNetGenesisMerkleRoot = genesisMerkleRoot

// simNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the simulation test network.
var simNetGenesisBlock = genesisBlock

// devNetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network (version 3).
var devNetGenesisCoinbaseTx = genesisCoinbaseTx

// devGenesisHash is the hash of the first block in the block chain for the development
// network (genesis block).
var devNetGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x4d, 0x6a, 0xc5, 0x8c, 0xfd, 0x73, 0xff, 0x60,
	0x5e, 0x0b, 0x03, 0x4f, 0x05, 0xcf, 0x8b, 0xa2,
	0x21, 0x50, 0x05, 0xf4, 0x16, 0xd2, 0xa6, 0x75,
	0x11, 0x36, 0xa9, 0xa3, 0x21, 0x3f, 0x00, 0x00,
})

// devNetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devNetGenesisMerkleRoot = genesisMerkleRoot

// devNetGenesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the development network.
var devNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:        1,
		ParentHashes:   []daghash.Hash{},
		HashMerkleRoot: devNetGenesisMerkleRoot,
		IDMerkleRoot:   devNetGenesisMerkleRoot,
		Timestamp:      time.Unix(0x5c922d07, 0),
		Bits:           0x1e7fffff,
		Nonce:          0x2633,
	},
	Transactions: []*wire.MsgTx{&devNetGenesisCoinbaseTx},
}

// SolveGenesisBlock attempts to find some combination of a nonce and
// current timestamp which makes the passed block hash to a value less than the
// target difficulty.
func SolveGenesisBlock() {
	// Create some convenience variables.
	header := &devNetGenesisBlock.Header
	targetDifficulty := util.CompactToBig(header.Bits)

	// set POW limit to 2^239 - 1.
	bigOne := big.NewInt(1)
	header.Bits = util.BigToCompact(new(big.Int).Sub(new(big.Int).Lsh(bigOne, 239), bigOne))

	// Search through the entire nonce range for a solution while
	// periodically checking for early quit and stale block
	// conditions along with updates to the speed monitor.
	maxNonce := ^uint64(0) // 2^64 - 1
	for {
		header.Timestamp = time.Unix(time.Now().Unix(), 0)
		for i := uint64(0); i <= maxNonce; i++ {
			// Update the nonce and hash the block header.  Each
			// hash is actually a double sha256 (two hashes), so
			// increment the number of hashes completed for each
			// attempt accordingly.
			header.Nonce = i
			hash := header.BlockHash()

			// The block is solved when the new block hash is less
			// than the target difficulty.  Yay!
			if daghash.HashToBig(&hash).Cmp(targetDifficulty) <= 0 {
				fmt.Printf("\n\nGenesis block solved:\n")
				fmt.Printf("timestamp: 0x%x\n", header.Timestamp.Unix())
				fmt.Printf("bits (difficulty): 0x%x\n", header.Bits)
				fmt.Printf("nonce: 0x%x\n", header.Nonce)
				fmt.Printf("hash: %v\n\n\n", hex.EncodeToString(hash[:]))
				return
			}
		}
	}
}
