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
	LockTime:     0,
	SubnetworkID: wire.SubnetworkDAGCoin,
}

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xca, 0xd9, 0x5f, 0x65, 0x44, 0xd4, 0x2f, 0x08,
	0x23, 0x22, 0x93, 0x4c, 0x07, 0xd9, 0xa4, 0xc0,
	0x1a, 0x51, 0x77, 0xf6, 0x13, 0x7c, 0x06, 0x8b,
	0xd2, 0x6d, 0xe1, 0x38, 0xea, 0x12, 0xcd, 0x4a,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x5a, 0x7c, 0x26, 0x54, 0x0d, 0x46, 0x6b, 0x5f,
	0x4a, 0x2c, 0x29, 0xff, 0x2f, 0xa5, 0xa0, 0xad,
	0x9c, 0xd2, 0x4f, 0x6e, 0xa9, 0x16, 0xfa, 0xb5,
	0x52, 0x5e, 0x40, 0xa2, 0x69, 0xb6, 0x8d, 0x5c,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:      1,
		ParentHashes: []daghash.Hash{},
		MerkleRoot:   genesisMerkleRoot,
		Timestamp:    time.Unix(0x5c34b3ba, 0),
		Bits:         0x207fffff,
		Nonce:        0xdffffffffffffff9,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xca, 0xd9, 0x5f, 0x65, 0x44, 0xd4, 0x2f, 0x08,
	0x23, 0x22, 0x93, 0x4c, 0x07, 0xd9, 0xa4, 0xc0,
	0x1a, 0x51, 0x77, 0xf6, 0x13, 0x7c, 0x06, 0x8b,
	0xd2, 0x6d, 0xe1, 0x38, 0xea, 0x12, 0xcd, 0x4a,
})

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the regression test network.  It is the same as the merkle root for
// the main network.
var regTestGenesisMerkleRoot = genesisMerkleRoot

// regTestGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the regression test network.
var regTestGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:      1,
		ParentHashes: []daghash.Hash{},
		MerkleRoot:   genesisMerkleRoot,
		Timestamp:    time.Unix(0x5c34b3ba, 0),
		Bits:         0x207fffff,
		Nonce:        0xdffffffffffffff9,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xca, 0xd9, 0x5f, 0x65, 0x44, 0xd4, 0x2f, 0x08,
	0x23, 0x22, 0x93, 0x4c, 0x07, 0xd9, 0xa4, 0xc0,
	0x1a, 0x51, 0x77, 0xf6, 0x13, 0x7c, 0x06, 0x8b,
	0xd2, 0x6d, 0xe1, 0x38, 0xea, 0x12, 0xcd, 0x4a,
})

// testNet3GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 3).  It is the same as the merkle root
// for the main network.
var testNet3GenesisMerkleRoot = genesisMerkleRoot

// testNet3GenesisBlock defines the genesis block of the block chain which
// serves as the public transaction ledger for the test network (version 3).
var testNet3GenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:      1,
		ParentHashes: []daghash.Hash{},
		MerkleRoot:   genesisMerkleRoot,
		Timestamp:    time.Unix(0x5c34b3ba, 0),
		Bits:         0x207fffff,
		Nonce:        0xdffffffffffffff9,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0xca, 0xd9, 0x5f, 0x65, 0x44, 0xd4, 0x2f, 0x08,
	0x23, 0x22, 0x93, 0x4c, 0x07, 0xd9, 0xa4, 0xc0,
	0x1a, 0x51, 0x77, 0xf6, 0x13, 0x7c, 0x06, 0x8b,
	0xd2, 0x6d, 0xe1, 0x38, 0xea, 0x12, 0xcd, 0x4a,
})

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the simulation test network.  It is the same as the merkle root for
// the main network.
var simNetGenesisMerkleRoot = genesisMerkleRoot

// simNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the simulation test network.
var simNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:      1,
		ParentHashes: []daghash.Hash{},
		MerkleRoot:   genesisMerkleRoot,
		Timestamp:    time.Unix(0x5c34b3ba, 0),
		Bits:         0x207fffff,
		Nonce:        0xdffffffffffffff9,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}
