// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

var genesisTxOuts = []*appmessage.TxOut{}

var genesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
}

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network.
var genesisCoinbaseTx = appmessage.NewSubnetworkMsgTx(1, []*appmessage.TxIn{}, genesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, genesisTxPayload)

// genesisHash is the hash of the first block in the block DAG for the main
// network (genesis block).
var genesisHash = daghash.Hash{
	0xfa, 0x00, 0xbd, 0xcb, 0x46, 0x74, 0xc5, 0xdb,
	0xf7, 0x63, 0xcb, 0x78, 0x7a, 0x94, 0xc5, 0xbf,
	0xd4, 0x81, 0xd3, 0x52, 0x2d, 0x79, 0xac, 0x57,
	0x73, 0xe6, 0x14, 0x7e, 0x15, 0xef, 0x85, 0x27,
}

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = daghash.Hash{
	0xca, 0x85, 0x56, 0x27, 0xc7, 0x6a, 0xb5, 0x7a,
	0x26, 0x1d, 0x63, 0x62, 0x1e, 0x57, 0x21, 0xf0,
	0x5e, 0x60, 0x1f, 0xee, 0x1d, 0x4d, 0xaa, 0x53,
	0x72, 0xe1, 0x16, 0xda, 0x4b, 0xb3, 0xd8, 0x0e,
}

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = appmessage.MsgBlock{
	Header: appmessage.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &genesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            mstime.UnixMilliseconds(0x1730a81bdb4),
		Bits:                 0x207fffff,
		Nonce:                0x1,
	},
	Transactions: []*appmessage.MsgTx{genesisCoinbaseTx},
}

var devnetGenesisTxOuts = []*appmessage.TxOut{}

var devnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x64, 0x65, 0x76, 0x6e, 0x65, 0x74, // kaspa-devnet
}

// devnetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the development network.
var devnetGenesisCoinbaseTx = appmessage.NewSubnetworkMsgTx(1, []*appmessage.TxIn{}, devnetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, devnetGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devnetGenesisHash = daghash.Hash{
	0x2e, 0x03, 0x7d, 0x31, 0x09, 0x56, 0x82, 0x72,
	0x1d, 0x49, 0x39, 0xf3, 0x7d, 0xd5, 0xc8, 0xf4,
	0xef, 0x4f, 0xcd, 0xeb, 0x1d, 0x95, 0xad, 0x6e,
	0x02, 0x4f, 0x52, 0xf2, 0xd6, 0x66, 0x00, 0x00,
}

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = daghash.Hash{
	0x68, 0x60, 0xe7, 0x77, 0x47, 0x74, 0x7f, 0xd5,
	0x55, 0x58, 0x8a, 0xb5, 0xc2, 0x29, 0x0c, 0xa6,
	0x65, 0x44, 0xb4, 0x4f, 0xfa, 0x31, 0x7a, 0xfa,
	0x55, 0xe0, 0xcf, 0xac, 0x9c, 0x86, 0x30, 0x2a,
}

// devnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devnetGenesisBlock = appmessage.MsgBlock{
	Header: appmessage.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &devnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            mstime.UnixMilliseconds(0x17305b05694),
		Bits:                 0x1e7fffff,
		Nonce:                0x10bb,
	},
	Transactions: []*appmessage.MsgTx{devnetGenesisCoinbaseTx},
}

var simnetGenesisTxOuts = []*appmessage.TxOut{}

var simnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x73, 0x69, 0x6d, 0x6e, 0x65, 0x74, // kaspa-simnet
}

// simnetGenesisCoinbaseTx is the coinbase transaction for the simnet genesis block.
var simnetGenesisCoinbaseTx = appmessage.NewSubnetworkMsgTx(1, []*appmessage.TxIn{}, simnetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, simnetGenesisTxPayload)

// simnetGenesisHash is the hash of the first block in the block DAG for
// the simnet (genesis block).
var simnetGenesisHash = daghash.Hash{
	0x86, 0x27, 0xdc, 0x5e, 0xa9, 0x38, 0xc7, 0xa5,
	0x7a, 0x18, 0xcd, 0xe7, 0xda, 0xed, 0x13, 0xe0,
	0x24, 0x1b, 0xab, 0xfe, 0xbd, 0xe6, 0x6f, 0xd3,
	0x95, 0x34, 0x81, 0x1c, 0x57, 0xd1, 0xc4, 0x3f,
}

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = daghash.Hash{
	0x47, 0x52, 0xc7, 0x23, 0x70, 0x4d, 0x89, 0x17,
	0xbd, 0x44, 0x26, 0xfa, 0x82, 0x7e, 0x1b, 0xa9,
	0xc6, 0x46, 0x1a, 0x37, 0x5a, 0x73, 0x88, 0x09,
	0xe8, 0x17, 0xff, 0xb1, 0xdb, 0x1a, 0xb3, 0x3f,
}

// simnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simnetGenesisBlock = appmessage.MsgBlock{
	Header: appmessage.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &simnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.Hash{},
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            mstime.UnixMilliseconds(0x173001df3d5),
		Bits:                 0x207fffff,
		Nonce:                0x0,
	},
	Transactions: []*appmessage.MsgTx{simnetGenesisCoinbaseTx},
}

var testnetGenesisTxOuts = []*appmessage.TxOut{}

var testnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x01,                                                                         // Varint
	0x00,                                                                         // OP-FALSE
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x74, 0x65, 0x73, 0x74, 0x6e, 0x65, 0x74, // kaspa-testnet
}

// testnetGenesisCoinbaseTx is the coinbase transaction for the testnet genesis block.
var testnetGenesisCoinbaseTx = appmessage.NewSubnetworkMsgTx(1, []*appmessage.TxIn{}, testnetGenesisTxOuts, subnetworkid.SubnetworkIDCoinbase, 0, testnetGenesisTxPayload)

// testnetGenesisHash is the hash of the first block in the block DAG for the test
// network (genesis block).
var testnetGenesisHash = daghash.Hash{
	0x34, 0x8c, 0x71, 0x99, 0x70, 0x13, 0x00, 0xe5,
	0xf5, 0x35, 0x98, 0x45, 0x89, 0xc7, 0xa2, 0xab,
	0xd0, 0x8f, 0x26, 0x00, 0x9c, 0xc6, 0x6b, 0xa3,
	0x20, 0x88, 0x86, 0x55, 0x3f, 0x61, 0x00, 0x00,
}

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = daghash.Hash{
	0xA0, 0xA1, 0x3D, 0xFD, 0x86, 0x41, 0x35, 0xC8,
	0xBD, 0xBB, 0xE6, 0x37, 0x35, 0xBB, 0x4C, 0x51,
	0x11, 0x7B, 0x26, 0x90, 0x15, 0x64, 0x0F, 0x42,
	0x6D, 0x2B, 0x6F, 0x37, 0x4D, 0xC1, 0xA9, 0x72,
}

// testnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for testnet.
var testnetGenesisBlock = appmessage.MsgBlock{
	Header: appmessage.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{},
		HashMerkleRoot:       &testnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: &daghash.ZeroHash,
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            mstime.UnixMilliseconds(0x1730a66a9d9),
		Bits:                 0x1e7fffff,
		Nonce:                0x162ca,
	},
	Transactions: []*appmessage.MsgTx{testnetGenesisCoinbaseTx},
}
