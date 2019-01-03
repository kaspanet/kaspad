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
				TxID:  daghash.Hash{},
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
	SubNetworkID: wire.SubNetworkDAGCoin,
}

// genesisHash is the hash of the first block in the block chain for the main
// network (genesis block).
var genesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x34, 0x16, 0x17, 0x85, 0x5d, 0xf7, 0x79, 0xde,
	0xdc, 0x37, 0xf0, 0x57, 0x93, 0x9d, 0x6f, 0x22,
	0xe5, 0xff, 0xa0, 0x3e, 0xc7, 0x3e, 0x50, 0xf5,
	0xa1, 0x2a, 0x0a, 0x2e, 0x48, 0x05, 0xc2, 0x47,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x79, 0x6e, 0xd7, 0x12, 0x58, 0xea, 0x3f, 0x88,
	0xa1, 0xfa, 0x7a, 0x67, 0x2b, 0x68, 0x76, 0x6b,
	0x39, 0x6b, 0x7d, 0x15, 0x6b, 0x1a, 0xb1, 0x96,
	0xe2, 0x17, 0x2c, 0x4a, 0x6a, 0xad, 0x72, 0x96,
})

// genesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:      1,
		ParentHashes: []daghash.Hash{},
		MerkleRoot:   genesisMerkleRoot,
		Timestamp:    time.Unix(0x5c238413, 0),
		Bits:         0x207fffff,
		Nonce:        0x00000001,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x34, 0x16, 0x17, 0x85, 0x5d, 0xf7, 0x79, 0xde,
	0xdc, 0x37, 0xf0, 0x57, 0x93, 0x9d, 0x6f, 0x22,
	0xe5, 0xff, 0xa0, 0x3e, 0xc7, 0x3e, 0x50, 0xf5,
	0xa1, 0x2a, 0x0a, 0x2e, 0x48, 0x05, 0xc2, 0x47,
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
		Timestamp:    time.Unix(0x5c238413, 0),
		Bits:         0x207fffff,
		Nonce:        0x00000001,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x34, 0x16, 0x17, 0x85, 0x5d, 0xf7, 0x79, 0xde,
	0xdc, 0x37, 0xf0, 0x57, 0x93, 0x9d, 0x6f, 0x22,
	0xe5, 0xff, 0xa0, 0x3e, 0xc7, 0x3e, 0x50, 0xf5,
	0xa1, 0x2a, 0x0a, 0x2e, 0x48, 0x05, 0xc2, 0x47,
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
		Timestamp:    time.Unix(0x5c238413, 0),
		Bits:         0x207fffff,
		Nonce:        0x00000001,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = daghash.Hash([daghash.HashSize]byte{ // Make go vet happy.
	0x34, 0x16, 0x17, 0x85, 0x5d, 0xf7, 0x79, 0xde,
	0xdc, 0x37, 0xf0, 0x57, 0x93, 0x9d, 0x6f, 0x22,
	0xe5, 0xff, 0xa0, 0x3e, 0xc7, 0x3e, 0x50, 0xf5,
	0xa1, 0x2a, 0x0a, 0x2e, 0x48, 0x05, 0xc2, 0x47,
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
		Timestamp:    time.Unix(0x5c238413, 0),
		Bits:         0x207fffff,
		Nonce:        0x00000001,
	},
	Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
}
