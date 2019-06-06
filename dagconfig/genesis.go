// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"math"
	"time"

	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

var genesisTxIns = []*wire.TxIn{
	{
		PreviousOutpoint: wire.Outpoint{
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
}

var genesisTxPayload = []byte{
	0x51, //OP_TRUE
}

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network (version 3).
var genesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, genesisTxIns, genesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, genesisTxPayload)

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = daghash.Hash([daghash.HashSize]byte{
	0x65, 0xf3, 0x2f, 0x2d, 0x93, 0xb8, 0xff, 0x42,
	0x20, 0x1a, 0x31, 0xa7, 0xf9, 0x1c, 0x1a, 0x8c,
	0x27, 0x7e, 0xa1, 0x7b, 0xa7, 0x5a, 0x01, 0xfc,
	0x21, 0xb4, 0xbf, 0x7d, 0xc1, 0x81, 0xc0, 0x09,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{
	0x96, 0x11, 0xce, 0x08, 0x51, 0x7a, 0x34, 0x54,
	0x4a, 0xd0, 0xbe, 0xe4, 0xf3, 0x34, 0xac, 0xf5,
	0x6a, 0x86, 0x68, 0x49, 0x2e, 0x0e, 0x82, 0xdf,
	0xf7, 0xf0, 0x48, 0xd8, 0x45, 0xf7, 0xf7, 0x1c,
})

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &genesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment: &daghash.Hash{
			0x51, 0x9f, 0x81, 0xbb, 0x93, 0xfd, 0x72, 0x9d,
			0x9d, 0xe0, 0x60, 0x80, 0xfa, 0x01, 0xe6, 0x34,
			0x93, 0x22, 0xed, 0x67, 0x69, 0x14, 0x52, 0xca,
			0x95, 0xd1, 0x9b, 0x77, 0xdb, 0xb8, 0x12, 0x80,
		},
		Timestamp: time.Unix(0x5cdac4b0, 0),
		Bits:      0x207fffff,
		Nonce:     0x3,
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
	0x7b, 0xea, 0x60, 0xdc, 0xaf, 0x9f, 0xc8, 0xf1,
	0x28, 0x9d, 0xbf, 0xde, 0xf6, 0x3b, 0x1e, 0xe8,
	0x4d, 0x4c, 0xc1, 0x8c, 0xaa, 0xe7, 0xdf, 0x08,
	0xe2, 0xe7, 0xa3, 0x6a, 0x4c, 0x7c, 0x00, 0x00,
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
		UTXOCommitment: &daghash.Hash{
			0x51, 0x9f, 0x81, 0xbb, 0x93, 0xfd, 0x72, 0x9d,
			0x9d, 0xe0, 0x60, 0x80, 0xfa, 0x01, 0xe6, 0x34,
			0x93, 0x22, 0xed, 0x67, 0x69, 0x14, 0x52, 0xca,
			0x95, 0xd1, 0x9b, 0x77, 0xdb, 0xb8, 0x12, 0x80,
		},
		Timestamp: time.Unix(0x5cee6864, 0),
		Bits:      0x1e7fffff,
		Nonce:     0xe834,
	},
	Transactions: []*wire.MsgTx{devNetGenesisCoinbaseTx},
}
