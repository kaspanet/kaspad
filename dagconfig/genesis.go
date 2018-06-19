// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
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
				0x62, 0x61, 0x6e, 0x6b, 0x73,                   /* |banks| */
			},
			Sequence: 0xffffffff,
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
				0x1d, 0x5f, 0xac,                               /* |._.| */
			},
		},
	},
	LockTime: 0,
}

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x82, 0xdc, 0xbd, 0xe6, 0x88, 0x37, 0x74, 0x5b,
	0x78, 0x6b, 0x03, 0x1d, 0xa3, 0x48, 0x3c, 0x45,
	0x3f, 0xc3, 0x2e, 0xd4, 0x53, 0x5b, 0x6f, 0x26,
	0x26, 0xb0, 0x48, 0x4f, 0x09, 0x00, 0x00, 0x00,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2,
	0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
	0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
	0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a,
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
		Nonce:         0xc0192550,               // 2148484547
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x7c, 0x03, 0x27, 0x80, 0x78, 0xfc, 0xe8, 0xed,
	0xcd, 0xfc, 0x3d, 0xe1, 0x63, 0x45, 0x4c, 0x03,
	0x0f, 0xef, 0x38, 0x16, 0xec, 0x54, 0x61, 0x6f,
	0xca, 0xc9, 0x58, 0x12, 0xb4, 0x6a, 0x15, 0x08,
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
		Nonce:         1,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x91, 0x1c, 0xe0, 0x47, 0x77, 0xad, 0x5e, 0xc5,
	0x60, 0x10, 0x21, 0x32, 0x51, 0x1b, 0x39, 0x06,
	0x24, 0xb3, 0xbf, 0x08, 0x8e, 0x04, 0x8c, 0xd3,
	0x80, 0xb4, 0x83, 0x83, 0xed, 0x00, 0x00, 0x00,
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
	0xc1, 0x5b, 0x71, 0xfe, 0x20, 0x70, 0x0f, 0xd0,
	0x08, 0x49, 0x88, 0x1b, 0x32, 0xb5, 0xbd, 0x13,
	0x17, 0xbe, 0x75, 0xe7, 0x29, 0x46, 0xdd, 0x03,
	0x01, 0x92, 0x90, 0xf1, 0xca, 0x8a, 0x88, 0x11,
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
		Timestamp:     time.Unix(0x5b28c7ec, 0), // 2018-06-19 09:07:56 +0000 UTC
		Bits:          0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:         0x9ffffffb,               // 2684354555
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}
