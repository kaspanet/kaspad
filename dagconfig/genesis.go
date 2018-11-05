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

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network (version 3).
var genesisCoinbaseTx = wire.MsgTx{
	Version: 1,
	TxIn: []*wire.TxIn{
		{
			PreviousOutPoint: wire.OutPoint{
				Hash:  daghash.Hash{},
				Index: 0xffffffff,
			},
			SignatureScript: []byte{
				0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04, 0x45, /* |.......E| */
				0x54, 0x68, 0x65, 0x20, 0x54, 0x69, 0x6d, 0x65, /* |The Time| */
				0x73, 0x20, 0x30, 0x33, 0x2f, 0x4a, 0x61, 0x6e, /* |s 03/Jan| */
				0x2f, 0x32, 0x30, 0x30, 0x39, 0x20, 0x43, 0x68, /* |/2009 Ch| */
				0x61, 0x6e, 0x63, 0x65, 0x6c, 0x6c, 0x6f, 0x72, /* |ancellor| */
				0x20, 0x6f, 0x6e, 0x20, 0x62, 0x72, 0x69, 0x6e, /* | on brin| */
				0x6b, 0x20, 0x6f, 0x66, 0x20, 0x73, 0x65, 0x63, /* |k of sec|*/
				0x6f, 0x6e, 0x64, 0x20, 0x62, 0x61, 0x69, 0x6c, /* |ond bail| */
				0x6f, 0x75, 0x74, 0x20, 0x66, 0x6f, 0x72, 0x20, /* |out for |*/
				0x62, 0x61, 0x6e, 0x6b, 0x73, /* |banks| */
			},
			Sequence: math.MaxUint64,
		},
	},
	TxOut: []*wire.TxOut{
		{
			Value: 0x12a05f200,
			PkScript: []byte{
				0x41, 0x04, 0x67, 0x8a, 0xfd, 0xb0, 0xfe, 0x55, /* |A.g....U| */
				0x48, 0x27, 0x19, 0x67, 0xf1, 0xa6, 0x71, 0x30, /* |H'.g..q0| */
				0xb7, 0x10, 0x5c, 0xd6, 0xa8, 0x28, 0xe0, 0x39, /* |..\..(.9| */
				0x09, 0xa6, 0x79, 0x62, 0xe0, 0xea, 0x1f, 0x61, /* |..yb...a| */
				0xde, 0xb6, 0x49, 0xf6, 0xbc, 0x3f, 0x4c, 0xef, /* |..I..?L.| */
				0x38, 0xc4, 0xf3, 0x55, 0x04, 0xe5, 0x1e, 0xc1, /* |8..U....| */
				0x12, 0xde, 0x5c, 0x38, 0x4d, 0xf7, 0xba, 0x0b, /* |..\8M...| */
				0x8d, 0x57, 0x8a, 0x4c, 0x70, 0x2b, 0x6b, 0xf1, /* |.W.Lp+k.| */
				0x1d, 0x5f, 0xac, /* |._.| */
			},
		},
	},
	LockTime: 0,
}

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xc3, 0xc9, 0xa8, 0xba, 0x10, 0x1e, 0x5e, 0xf2,
	0x03, 0xfd, 0x54, 0x15, 0x34, 0xae, 0x4e, 0x74,
	0xc5, 0x5f, 0x48, 0x2e, 0xc9, 0xef, 0xba, 0x58,
	0x92, 0xce, 0x27, 0xfa, 0xf9, 0x00, 0x00, 0x00,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x69, 0x1f, 0x2e, 0x9c, 0x49, 0x12, 0x54, 0xc5,
	0xde, 0x1d, 0x8f, 0xd0, 0x95, 0x40, 0x07, 0xf0,
	0xda, 0x1b, 0x4c, 0x84, 0x3c, 0x07, 0x94, 0x87,
	0x6a, 0xd3, 0x0e, 0xca, 0x87, 0xcc, 0x4b, 0xb9,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:         1,
		NumParentBlocks: 0,
		ParentHashes:    []daghash.Hash{},
		MerkleRoot:      genesisMerkleRoot,        // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:       time.Unix(0x5bbe076c, 0), // 2018-10-10 14:06:36 +0000 UTC
		Bits:            0x1e00ffff,               // 503382015 [000000ffff000000000000000000000000000000000000000000000000000000]
		Nonce:           0x80000000000d8796,       // 9223372036855662486
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xe7, 0xbe, 0x90, 0x9a, 0x85, 0xf4, 0xea, 0x42,
	0xe0, 0xd6, 0x68, 0x09, 0xc5, 0xa9, 0xd6, 0x94,
	0xa2, 0x34, 0x52, 0x4c, 0x5f, 0x1a, 0x77, 0xb1,
	0xcf, 0x26, 0x12, 0x0a, 0x1b, 0x26, 0x92, 0x13,
})

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the regression test network.  It is the same as the merkle root for
// the main network.
var regTestGenesisMerkleRoot = genesisMerkleRoot

// regTestGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the regression test network.
var regTestGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:         1,
		NumParentBlocks: 0,
		ParentHashes:    []daghash.Hash{},
		MerkleRoot:      regTestGenesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:       time.Unix(0x5bbe0d4b, 0), // 2018-06-19 09:00:38 +0000 UTC
		Bits:            0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:           0x00000000,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xcb, 0xcb, 0x39, 0xdb, 0xc7, 0x18, 0xe3, 0xd3,
	0x59, 0xb2, 0xf4, 0xe3, 0x9c, 0x05, 0xa3, 0x1d,
	0x60, 0xc6, 0xf7, 0x06, 0xb5, 0x93, 0xdf, 0x11,
	0x71, 0x1b, 0x82, 0xb1, 0xa3, 0x00, 0x00, 0x00,
})

// testNet3GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 3).  It is the same as the merkle root
// for the main network.
var testNet3GenesisMerkleRoot = genesisMerkleRoot

// testNet3GenesisBlock defines the genesis block of the block chain which
// serves as the public transaction ledger for the test network (version 3).
var testNet3GenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:         1,
		NumParentBlocks: 0,
		ParentHashes:    []daghash.Hash{},
		MerkleRoot:      testNet3GenesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:       time.Unix(0x5bbe0e49, 0),  // 2018-06-19 09:04:06 +0000 UTC
		Bits:            0x1e00ffff,                // 503382015 [000000ffff000000000000000000000000000000000000000000000000000000]
		Nonce:           0xc00000000032560b,        // 2150570811
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x4a, 0xc1, 0x82, 0x2e, 0x43, 0x05, 0xea, 0x0c,
	0x4f, 0xcc, 0x77, 0x87, 0xae, 0x26, 0x48, 0x87,
	0x50, 0x13, 0xee, 0x2f, 0x55, 0xa7, 0x18, 0xa7,
	0x1e, 0xf2, 0xd8, 0x7c, 0xc1, 0x13, 0xac, 0x22,
})

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the simulation test network.  It is the same as the merkle root for
// the main network.
var simNetGenesisMerkleRoot = genesisMerkleRoot

// simNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the simulation test network.
var simNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:         1,
		NumParentBlocks: 0,
		ParentHashes:    []daghash.Hash{},
		MerkleRoot:      simNetGenesisMerkleRoot,  // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:       time.Unix(0x5bbe00fe, 0), // 2018-10-10 13:39:10 +0000 UTC
		Bits:            0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:           0xdffffffffffffffc,       // 1610612733
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}
