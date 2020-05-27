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
// the main network.
var genesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, genesisTxIns, genesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, genesisTxPayload)

// genesisHash is the hash of the first block in the block DAG for the main
// network (genesis block).
var genesisHash = daghash.Hash{
	0x9c, 0xf9, 0x7d, 0xd6, 0xbc, 0x25, 0xb2, 0xb8,
	0x6c, 0xd0, 0xe1, 0x9e, 0x3a, 0x2f, 0xab, 0x3d,
	0x3e, 0x3f, 0x4d, 0x95, 0x09, 0x85, 0x8f, 0x99,
	0xc8, 0xe4, 0xc2, 0x15, 0x78, 0xac, 0x79, 0x6a,
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
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &genesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5ece5ba4, 0),
		Bits:                 0x207fffff,
		Nonce:                0,
	},
	Transactions: []*wire.MsgTx{genesisCoinbaseTx},
}

var devnetGenesisTxIns = []*wire.TxIn{
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
var devnetGenesisTxOuts = []*wire.TxOut{}

var devnetGenesisTxPayload = []byte{
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x64, 0x65, 0x76, 0x6e, 0x65, 0x74, // kaspa-devnet
}

// devnetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the development network.
var devnetGenesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, devnetGenesisTxIns, devnetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, devnetGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devnetGenesisHash = daghash.Hash{
	0xd3, 0xc0, 0xf4, 0xa7, 0x91, 0xa2, 0x2e, 0x27,
	0x90, 0x38, 0x6d, 0x47, 0x7b, 0x26, 0x15, 0xaf,
	0xaf, 0xa6, 0x3a, 0xad, 0xd5, 0xfa, 0x37, 0xf3,
	0x5e, 0x70, 0xfb, 0xfc, 0x07, 0x31, 0x00, 0x00,
}

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = daghash.Hash{
	0x16, 0x0a, 0xc6, 0x8b, 0x77, 0x08, 0xf4, 0x96,
	0xa3, 0x07, 0x05, 0xbc, 0x92, 0xda, 0xee, 0x73,
	0x26, 0x5e, 0xd0, 0x85, 0x78, 0xa2, 0x5d, 0x02,
	0x49, 0x8a, 0x2a, 0x22, 0xef, 0x41, 0xc9, 0xc3,
}

// devnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devnetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &devnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5ece5ba4, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x227e6,
	},
	Transactions: []*wire.MsgTx{devnetGenesisCoinbaseTx},
}

var regtestGenesisTxIns = []*wire.TxIn{
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
var regtestGenesisTxOuts = []*wire.TxOut{}

var regtestGenesisTxPayload = []byte{
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x72, 0x65, 0x67, 0x74, 0x65, 0x73, 0x74, // kaspa-regtest
}

// regtestGenesisCoinbaseTx is the coinbase transaction for
// the genesis blocks for the regtest network.
var regtestGenesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, regtestGenesisTxIns, regtestGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, regtestGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var regtestGenesisHash = daghash.Hash{
	0xc7, 0x7f, 0x3f, 0xb1, 0xe8, 0xf8, 0xcf, 0xa4,
	0xf5, 0x6e, 0xeb, 0x9a, 0x35, 0xd4, 0x58, 0x10,
	0xc8, 0xd6, 0x6d, 0x07, 0x76, 0x53, 0x75, 0xa2,
	0x73, 0xc0, 0x4e, 0xeb, 0xed, 0x61, 0x00, 0x00,
}

// regtestGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the regtest.
var regtestGenesisMerkleRoot = daghash.Hash{
	0x3a, 0x9f, 0x62, 0xc9, 0x2b, 0x16, 0x17, 0xb3,
	0x41, 0x6d, 0x9e, 0x2d, 0x87, 0x93, 0xfd, 0x72,
	0x77, 0x4d, 0x1d, 0x6f, 0x6d, 0x38, 0x5b, 0xf1,
	0x24, 0x1b, 0xdc, 0x96, 0xce, 0xbf, 0xa1, 0x09,
}

// regtestGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var regtestGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &regtestGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5ece5ba4, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x31516,
	},
	Transactions: []*wire.MsgTx{regtestGenesisCoinbaseTx},
}

var simnetGenesisTxIns = []*wire.TxIn{
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
var simnetGenesisTxOuts = []*wire.TxOut{}

var simnetGenesisTxPayload = []byte{
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x73, 0x69, 0x6d, 0x6e, 0x65, 0x74, // kaspa-simnet
}

// simnetGenesisCoinbaseTx is the coinbase transaction for the simnet genesis block.
var simnetGenesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, simnetGenesisTxIns, simnetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, simnetGenesisTxPayload)

// simnetGenesisHash is the hash of the first block in the block DAG for
// the simnet (genesis block).
var simnetGenesisHash = daghash.Hash{
	0x2b, 0x7b, 0x81, 0x60, 0x79, 0x74, 0x83, 0x0a,
	0x33, 0x71, 0x88, 0x2d, 0x67, 0x7e, 0x06, 0x7b,
	0x58, 0x87, 0xa3, 0x2b, 0xed, 0xa7, 0x65, 0xb9,
	0x13, 0x1b, 0xce, 0x49, 0xa5, 0x56, 0xe4, 0x44,
}

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = daghash.Hash{
	0xb0, 0x1c, 0x3b, 0x9e, 0x0d, 0x9a, 0xc0, 0x80,
	0x0a, 0x08, 0x42, 0x50, 0x02, 0xa3, 0xea, 0xdb,
	0xed, 0xc8, 0xd0, 0xad, 0x35, 0x03, 0xd8, 0x0e,
	0x11, 0x3c, 0x7b, 0xb2, 0xb5, 0x20, 0xe5, 0x84,
}

// simnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simnetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &simnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5ece5ba5, 0),
		Bits:                 0x207fffff,
		Nonce:                0x0,
	},
	Transactions: []*wire.MsgTx{simnetGenesisCoinbaseTx},
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

// testnetGenesisCoinbaseTx is the coinbase transaction for the testnet genesis block.
var testnetGenesisCoinbaseTx = wire.NewSubnetworkMsgTx(1, testnetGenesisTxIns, testnetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, testnetGenesisTxPayload)

// testnetGenesisHash is the hash of the first block in the block DAG for the test
// network (genesis block).
var testnetGenesisHash = daghash.Hash{
	0x6b, 0xac, 0xe2, 0xfc, 0x1d, 0x1c, 0xaf, 0x38,
	0x72, 0x0b, 0x9d, 0xf5, 0xcc, 0x2b, 0xf4, 0x6d,
	0xf4, 0x2c, 0x05, 0xf9, 0x3d, 0x94, 0xb1, 0xc6,
	0x6a, 0xea, 0x1b, 0x81, 0x4c, 0x22, 0x00, 0x00,
}

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = daghash.Hash{
	0x88, 0x05, 0xd0, 0xe7, 0x8f, 0x41, 0x77, 0x39,
	0x2c, 0xb6, 0xbb, 0xb4, 0x19, 0xa8, 0x48, 0x4a,
	0xdf, 0x77, 0xb0, 0x82, 0xd6, 0x70, 0xd8, 0x24,
	0x6a, 0x36, 0x05, 0xaa, 0xbd, 0x7a, 0xd1, 0x62,
}

// testnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for testnet.
var testnetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &testnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.ZeroHash,
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            time.Unix(0x5ece5ba4, 0),
		Bits:                 0x1e7fffff,
		Nonce:                0x6d249,
	},
	Transactions: []*wire.MsgTx{testnetGenesisCoinbaseTx},
}
