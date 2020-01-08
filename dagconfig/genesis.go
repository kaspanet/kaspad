// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"math"
	"time"

	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
)

var genesisTxIns = []*wire.TxIn{
	{
		PreviousOutpoint: wire.Outpoint{
			TxID:  daghash.TxID{},
			Index: math.MaxUint32,
		},
		SignatureScript: []byte{
			0x00, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
			0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
		},
		Sequence: math.MaxUint64,
	},
}
var genesisTxOuts = []*wire.TxOut{}

var genesisTxPayload = []byte{
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
}

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network.
var genesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, genesisTxIns, genesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, genesisTxPayload)

// genesisHash is the hash of the first block in the block DAG for the main
// network (genesis block).
var genesisHash = daghash.Hash([daghash.HashSize]byte{
	0x9b, 0x22, 0x59, 0x44, 0x66, 0xf0, 0xbe, 0x50,
	0x7c, 0x1c, 0x8a, 0xf6, 0x06, 0x27, 0xe6, 0x33,
	0x38, 0x7e, 0xd1, 0xd5, 0x8c, 0x42, 0x59, 0x1a,
	0x31, 0xac, 0x9a, 0xa6, 0x2e, 0xd5, 0x2b, 0x0f,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash([daghash.HashSize]byte{
	0x72, 0x10, 0x35, 0x85, 0xdd, 0xac, 0x82, 0x5c,
	0x49, 0x13, 0x9f, 0xc0, 0x0e, 0x37, 0xc0, 0x45,
	0x71, 0xdf, 0xd9, 0xf6, 0x36, 0xdf, 0x4c, 0x42,
	0x72, 0x7b, 0x9e, 0x86, 0xdd, 0x37, 0xd2, 0xbd,
})

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &genesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5cdac4b0, 0),
		Bits:                 0x207fffff,
		Nonce:                0x1,
	},
	Transactions: []*wire.MsgTx{genesisCoinbaseTx},
}

// devNetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network.
var devNetGenesisCoinbaseTx = genesisCoinbaseTx

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devNetGenesisHash = daghash.Hash([daghash.HashSize]byte{
	0xf4, 0xd6, 0x37, 0x0b, 0xc6, 0x67, 0x41, 0x90,
	0x06, 0x57, 0xef, 0x65, 0x45, 0x07, 0x3a, 0x50,
	0xb7, 0x85, 0x21, 0xb9, 0xa2, 0xe3, 0xf5, 0x9e,
	0xe0, 0x45, 0x9b, 0xb0, 0xab, 0x33, 0x00, 0x00,
})

// devNetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devNetGenesisMerkleRoot = genesisMerkleRoot

// devNetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &devNetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5d021d3e, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x5221,
	},
	Transactions: []*wire.MsgTx{devNetGenesisCoinbaseTx},
}

// devNetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network.
var regTestGenesisCoinbaseTx = genesisCoinbaseTx

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var regTestGenesisHash = daghash.Hash([daghash.HashSize]byte{
	0xf4, 0xd6, 0x37, 0x0b, 0xc6, 0x67, 0x41, 0x90,
	0x06, 0x57, 0xef, 0x65, 0x45, 0x07, 0x3a, 0x50,
	0xb7, 0x85, 0x21, 0xb9, 0xa2, 0xe3, 0xf5, 0x9e,
	0xe0, 0x45, 0x9b, 0xb0, 0xab, 0x33, 0x00, 0x00,
})

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var regTestGenesisMerkleRoot = genesisMerkleRoot

// regTestGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var regTestGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &regTestGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5d021d3e, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x5221,
	},
	Transactions: []*wire.MsgTx{regTestGenesisCoinbaseTx},
}

// simNetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network.
var simNetGenesisCoinbaseTx = genesisCoinbaseTx

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var simNetGenesisHash = daghash.Hash([daghash.HashSize]byte{
	0xf4, 0xd6, 0x37, 0x0b, 0xc6, 0x67, 0x41, 0x90,
	0x06, 0x57, 0xef, 0x65, 0x45, 0x07, 0x3a, 0x50,
	0xb7, 0x85, 0x21, 0xb9, 0xa2, 0xe3, 0xf5, 0x9e,
	0xe0, 0x45, 0x9b, 0xb0, 0xab, 0x33, 0x00, 0x00,
})

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simNetGenesisMerkleRoot = genesisMerkleRoot

// simNetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &simNetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5d021d3e, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x5221,
	},
	Transactions: []*wire.MsgTx{simNetGenesisCoinbaseTx},
}

var testnetGenesisTxIns = []*wire.TxIn{
	{
		PreviousOutpoint: wire.Outpoint{
			TxID:  daghash.TxID{},
			Index: math.MaxUint32,
		},
		SignatureScript: []byte{
			0x00, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
			0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
		},
		Sequence: math.MaxUint64,
	},
}
var testnetGenesisTxOuts = []*wire.TxOut{}

var testnetGenesisTxPayload = []byte{
	0x01,                                                                         // Varint
	0x00,                                                                         // OP-FALSE
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x74, 0x65, 0x73, 0x74, 0x6e, 0x65, 0x74, // kaspa-testnet
}

// testNetGenesisCoinbaseTx is the coinbase transaction for the testnet genesis block.
var testNetGenesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, testnetGenesisTxIns, testnetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, testnetGenesisTxPayload)

// testNetGenesisHash is the hash of the first block in the block DAG for the test
// network (genesis block).
var testNetGenesisHash = daghash.Hash{
	0x22, 0x15, 0x34, 0xa9, 0xff, 0x10, 0xdd, 0x47,
	0xcd, 0x21, 0x11, 0x25, 0xc5, 0x6d, 0x85, 0x9a,
	0x97, 0xc8, 0x63, 0x63, 0x79, 0x40, 0x80, 0x04,
	0x74, 0xe6, 0x29, 0x7b, 0xbc, 0x08, 0x00, 0x00,
}

// testNetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testNetGenesisMerkleRoot = daghash.Hash{
	0x88, 0x05, 0xd0, 0xe7, 0x8f, 0x41, 0x77, 0x39,
	0x2c, 0xb6, 0xbb, 0xb4, 0x19, 0xa8, 0x48, 0x4a,
	0xdf, 0x77, 0xb0, 0x82, 0xd6, 0x70, 0xd8, 0x24,
	0x6a, 0x36, 0x05, 0xaa, 0xbd, 0x7a, 0xd1, 0x62,
}

// testNetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for testnet.
var testNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &testNetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.ZeroHash,
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5e15adfe, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x20a1,
	},
	Transactions: []*wire.MsgTx{testNetGenesisCoinbaseTx},
}
