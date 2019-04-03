// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"math"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
)

var genesisTxIns = []*wire.TxIn{
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
}
var genesisTxOuts = []*wire.TxOut{
	{
		Value: 0x12a05f200,
		PkScript: []byte{
			0x51,
		},
	},
}

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network (version 3).
var genesisCoinbaseTx = wire.NewNativeMsgTx(1, genesisTxIns, genesisTxOuts)

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x08, 0xcc, 0x73, 0xd0, 0xe0, 0x70, 0x52, 0xfa,
	0x00, 0x1f, 0x2b, 0x02, 0x67, 0xe3, 0x4f, 0xda,
	0x37, 0x75, 0x1c, 0x67, 0xa4, 0x6d, 0x55, 0x2e,
	0x90, 0xac, 0x0f, 0x19, 0x3f, 0x86, 0x77, 0x64,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xd4, 0xdc, 0x8b, 0xb8, 0x76, 0x57, 0x9d, 0x7d,
	0xe9, 0x9d, 0xae, 0xdb, 0xf8, 0x22, 0xd2, 0x0d,
	0xa2, 0xe0, 0xbb, 0xbe, 0xed, 0xb0, 0xdb, 0xba,
	0xeb, 0x18, 0x4d, 0x42, 0x01, 0xff, 0xed, 0x9d,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:        1,
		ParentHashes:   []*daghash.Hash{},
		HashMerkleRoot: &genesisMerkleRoot,
		IDMerkleRoot:   &genesisMerkleRoot,
		Timestamp:      time.Unix(0x5ca09ba1, 0),
		Bits:           0x207fffff,
		Nonce:          0x6,
	},
	Transactions: []*wire.MsgTx{genesisCoinbaseTx},
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
	0xe8, 0xf0, 0x59, 0x4b, 0x54, 0xe7, 0xa1, 0xba,
	0x1b, 0xe0, 0x6c, 0x72, 0xf8, 0x3d, 0x75, 0x5d,
	0xa3, 0x3c, 0xc5, 0x71, 0x43, 0x94, 0x81, 0xbf,
	0x32, 0x82, 0x5e, 0xfa, 0x6a, 0x0d, 0x00, 0x00,
})

// devNetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devNetGenesisMerkleRoot = genesisMerkleRoot

// devNetGenesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the development network.
var devNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:        1,
		ParentHashes:   []*daghash.Hash{},
		HashMerkleRoot: &devNetGenesisMerkleRoot,
		IDMerkleRoot:   &devNetGenesisMerkleRoot,
		Timestamp:      time.Unix(0x5ca09ba1, 0),
		Bits:           0x1e7fffff,
		Nonce:          0x155d9,
	},
	Transactions: []*wire.MsgTx{devNetGenesisCoinbaseTx},
}
