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
var genesisHash = daghash.Hash{
	0x9b, 0x22, 0x59, 0x44, 0x66, 0xf0, 0xbe, 0x50,
	0x7c, 0x1c, 0x8a, 0xf6, 0x06, 0x27, 0xe6, 0x33,
	0x38, 0x7e, 0xd1, 0xd5, 0x8c, 0x42, 0x59, 0x1a,
	0x31, 0xac, 0x9a, 0xa6, 0x2e, 0xd5, 0x2b, 0x0f,
}

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash{
	0x72, 0x10, 0x35, 0x85, 0xdd, 0xac, 0x82, 0x5c,
	0x49, 0x13, 0x9f, 0xc0, 0x0e, 0x37, 0xc0, 0x45,
	0x71, 0xdf, 0xd9, 0xf6, 0x36, 0xdf, 0x4c, 0x42,
	0x72, 0x7b, 0x9e, 0x86, 0xdd, 0x37, 0xd2, 0xbd,
}

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

var devNetGenesisTxIns = []*wire.TxIn{
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
var devNetGenesisTxOuts = []*wire.TxOut{}

var devNetGenesisTxPayload = []byte{
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x64, 0x65, 0x76, 0x6e, 0x65, 0x74, // kaspa-devnet
}

// devNetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network, regression test network, and test network.
var devNetGenesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, devNetGenesisTxIns, devNetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, devNetGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devNetGenesisHash = daghash.Hash{
	0x17, 0x59, 0x5c, 0x09, 0xdd, 0x1a, 0x51, 0x65,
	0x14, 0xbc, 0x19, 0xff, 0x29, 0xea, 0xf3, 0xcb,
	0xe2, 0x76, 0xf0, 0xc7, 0x86, 0xf8, 0x0c, 0x53,
	0x59, 0xbe, 0xee, 0x0c, 0x2b, 0x5d, 0x00, 0x00,
}

// devNetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devNetGenesisMerkleRoot = daghash.Hash{
	0x16, 0x0a, 0xc6, 0x8b, 0x77, 0x08, 0xf4, 0x96,
	0xa3, 0x07, 0x05, 0xbc, 0x92, 0xda, 0xee, 0x73,
	0x26, 0x5e, 0xd0, 0x85, 0x78, 0xa2, 0x5d, 0x02,
	0x49, 0x8a, 0x2a, 0x22, 0xef, 0x41, 0xc9, 0xc3,
}

// devNetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &devNetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5e15e758, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x282ac,
	},
	Transactions: []*wire.MsgTx{devNetGenesisCoinbaseTx},
}

var regNetGenesisTxIns = []*wire.TxIn{
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
var regNetGenesisTxOuts = []*wire.TxOut{}

var regNetGenesisTxPayload = []byte{
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x72, 0x65, 0x67, 0x74, 0x65, 0x73, 0x74, // kaspa-regtest
}

// regTestGenesisCoinbaseTx is the coinbase transaction for
// the genesis blocks for the regTest network.
var regTestGenesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, regNetGenesisTxIns, regNetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, regNetGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var regTestGenesisHash = daghash.Hash{
	0xfc, 0x02, 0x19, 0x6f, 0x79, 0x7a, 0xed, 0x2d,
	0x0f, 0x31, 0xa5, 0xbd, 0x32, 0x13, 0x29, 0xc7,
	0x7c, 0x0c, 0x5c, 0x1a, 0x5b, 0x7c, 0x20, 0x68,
	0xb7, 0xc9, 0x9f, 0x61, 0x13, 0x11, 0x00, 0x00,
}

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the regTest.
var regTestGenesisMerkleRoot = daghash.Hash{
	0x3a, 0x9f, 0x62, 0xc9, 0x2b, 0x16, 0x17, 0xb3,
	0x41, 0x6d, 0x9e, 0x2d, 0x87, 0x93, 0xfd, 0x72,
	0x77, 0x4d, 0x1d, 0x6f, 0x6d, 0x38, 0x5b, 0xf1,
	0x24, 0x1b, 0xdc, 0x96, 0xce, 0xbf, 0xa1, 0x09,
}

// regTestGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var regTestGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &regTestGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5e15e2d8, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x15a6,
	},
	Transactions: []*wire.MsgTx{regTestGenesisCoinbaseTx},
}

var simNetGenesisTxIns = []*wire.TxIn{
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
var simNetGenesisTxOuts = []*wire.TxOut{}

var simNetGenesisTxPayload = []byte{
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x73, 0x69, 0x6d, 0x6e, 0x65, 0x74, // kaspa-simnet
}

// simNetGenesisCoinbaseTx is the coinbase transaction for the simnet genesis block.
var simNetGenesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, simNetGenesisTxIns, simNetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, simNetGenesisTxPayload)

// simNetGenesisHash is the hash of the first block in the block DAG for
// the simnet (genesis block).
var simNetGenesisHash = daghash.Hash{
	0xff, 0x69, 0xcc, 0x45, 0x45, 0x74, 0x5b, 0xf9,
	0xd5, 0x4e, 0x43, 0x56, 0x4f, 0x1b, 0xdf, 0x31,
	0x09, 0xb7, 0x76, 0xaa, 0x2a, 0x33, 0x35, 0xc9,
	0xa1, 0x80, 0xe0, 0x92, 0xbb, 0xae, 0xcd, 0x49,
}

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simNetGenesisMerkleRoot = daghash.Hash{
	0xb0, 0x1c, 0x3b, 0x9e, 0x0d, 0x9a, 0xc0, 0x80,
	0x0a, 0x08, 0x42, 0x50, 0x02, 0xa3, 0xea, 0xdb,
	0xed, 0xc8, 0xd0, 0xad, 0x35, 0x03, 0xd8, 0x0e,
	0x11, 0x3c, 0x7b, 0xb2, 0xb5, 0x20, 0xe5, 0x84,
}

// simNetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              1,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &simNetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5e15d31c, 0),
		Bits:                 0x207fffff,
		Nonce:                0x3,
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
