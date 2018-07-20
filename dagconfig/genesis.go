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
	0x32, 0x93, 0xa9, 0x6f, 0x69, 0xf3, 0x80, 0x4b,
	0x50, 0x35, 0x29, 0x16, 0xcd, 0x68, 0x4c, 0x36,
	0x99, 0xf7, 0x05, 0x08, 0xea, 0xbd, 0x8a, 0x78,
	0xc0, 0x39, 0xfb, 0x39, 0xf5, 0x84, 0x04, 0x2f,
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
		Version:       1,
		NumPrevBlocks: 0,
		PrevBlocks:    []daghash.Hash{},
		MerkleRoot:    genesisMerkleRoot,        // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:     time.Unix(0x5b28c4c8, 0), // 2018-06-19 08:54:32 +0000 UTC
		Bits:          0x1e00ffff,               // 503382015 [000000ffff000000000000000000000000000000000000000000000000000000]
		Nonce:         0x400a4c8f,               // 1074416783
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xeb, 0x98, 0xea, 0xaf, 0xf3, 0xf4, 0x7b, 0x3f, 0x4b, 0x61, 0x57, 0xf8, 0x75, 0xfb, 0xc6, 0x9f, 0xde, 0x88, 0x68, 0xbe, 0x5d, 0xc8, 0xab, 0xc6, 0xb4, 0x1f, 0x85, 0xd8, 0x77, 0x03, 0xbf, 0x0f,
})

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the regression test network.  It is the same as the merkle root for
// the main network.
var regTestGenesisMerkleRoot = genesisMerkleRoot

// regTestGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the regression test network.
var regTestGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 0,
		PrevBlocks:    []daghash.Hash{},
		MerkleRoot:    regTestGenesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:     time.Unix(0x5b28c636, 0), // 2018-06-19 09:00:38 +0000 UTC
		Bits:          0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:         0xdffffff9,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x25, 0x62, 0xec, 0x05, 0x42, 0x4b, 0x74, 0x37, 0x5a, 0x67, 0xd2, 0x6e, 0x24, 0x6c, 0xe8, 0x96, 0xdf, 0xd6, 0x71, 0x88, 0xc8, 0xbb, 0x89, 0xd6, 0xd9, 0x23, 0x84, 0x2b, 0xd0, 0x69, 0x39, 0xe7,
})

// testNet3GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 3).  It is the same as the merkle root
// for the main network.
var testNet3GenesisMerkleRoot = genesisMerkleRoot

// testNet3GenesisBlock defines the genesis block of the block chain which
// serves as the public transaction ledger for the test network (version 3).
var testNet3GenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 0,
		PrevBlocks:    []daghash.Hash{},
		MerkleRoot:    testNet3GenesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:     time.Unix(0x5b28c706, 0),  // 2018-06-19 09:04:06 +0000 UTC
		Bits:          0x1e00ffff,                // 503382015 [000000ffff000000000000000000000000000000000000000000000000000000]
		Nonce:         0x802f1b3b,                // 2150570811
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x93, 0x0f, 0x30, 0xd1, 0x0b, 0xaf, 0x9d, 0x5b,
	0x58, 0xdc, 0xad, 0x78, 0xee, 0x16, 0xd0, 0x12,
	0x10, 0xac, 0x2c, 0xa3, 0x08, 0xc4, 0x83, 0x33,
	0x57, 0xb2, 0xaf, 0x5a, 0x22, 0xa2, 0xf9, 0x20,
})

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the simulation test network.  It is the same as the merkle root for
// the main network.
var simNetGenesisMerkleRoot = genesisMerkleRoot

// simNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the simulation test network.
var simNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 0,
		PrevBlocks:    []daghash.Hash{},
		MerkleRoot:    simNetGenesisMerkleRoot,  // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
		Timestamp:     time.Unix(0x5b50a002, 0), // 2018-06-19 09:07:56 +0000 UTC
		Bits:          0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:         0x5ffffffd,               // 2684354555
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}
