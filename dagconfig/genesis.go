// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"math"
	"time"

	"github.com/daglabs/btcd/util/daghash"
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
var genesisHash = daghash.Hash([daghash.HashSize]byte{
	0x95, 0x77, 0xb3, 0xed, 0xa7, 0x10, 0x0a, 0x5a,
	0x9d, 0x70, 0xb0, 0x17, 0x70, 0xb5, 0x5e, 0x58,
	0x44, 0x59, 0xb9, 0xfa, 0x71, 0x52, 0x03, 0xdd,
	0x88, 0xfd, 0xdb, 0x7b, 0x20, 0xd0, 0x04, 0x58,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{
	0xd4, 0xdc, 0x8b, 0xb8, 0x76, 0x57, 0x9d, 0x7d,
	0xe9, 0x9d, 0xae, 0xdb, 0xf8, 0x22, 0xd2, 0x0d,
	0xa2, 0xe0, 0xbb, 0xbe, 0xed, 0xb0, 0xdb, 0xba,
	0xeb, 0x18, 0x4d, 0x42, 0x01, 0xff, 0xed, 0x9d,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &genesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.Hash{},
		Timestamp:            time.Unix(0x5cd83da0, 0),
		Bits:                 0x207fffff,
		Nonce:                0x1,
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
var devNetGenesisHash = daghash.Hash([daghash.HashSize]byte{
	0x53, 0x68, 0x95, 0xd4, 0xf7, 0xc7, 0x3b, 0x94,
	0x64, 0xfa, 0x98, 0xc7, 0xcb, 0x92, 0x53, 0x71,
	0x5e, 0x14, 0xd1, 0x83, 0x01, 0xd4, 0x1e, 0x17,
	0xd4, 0xc4, 0xd3, 0x50, 0xa7, 0x64, 0x00, 0x00,
})

// devNetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devNetGenesisMerkleRoot = genesisMerkleRoot

// devNetGenesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the development network.
var devNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &devNetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.Hash{},
		Timestamp:            time.Unix(0x5cd83da0, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0xfe2a,
	},
	Transactions: []*wire.MsgTx{devNetGenesisCoinbaseTx},
}
