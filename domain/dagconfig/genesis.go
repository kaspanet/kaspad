// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

var genesisTxOuts = []*externalapi.DomainTransactionOutput{}

var genesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x00, 0x00, //script version
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
}

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network.
var genesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(1, []*externalapi.DomainTransactionInput{}, genesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, genesisTxPayload)

// genesisHash is the hash of the first block in the block DAG for the main
// network (genesis block).
var genesisHash = externalapi.DomainHash{
	0x6e, 0x1e, 0x36, 0xdf, 0x8d, 0xd4, 0x43, 0xa7,
	0x55, 0x79, 0xc7, 0x40, 0xc8, 0x2f, 0x8d, 0xfa,
	0x4d, 0x32, 0x47, 0xcb, 0x5e, 0xea, 0x34, 0xf9,
	0x22, 0xe5, 0xf3, 0x1c, 0xe8, 0x38, 0x5a, 0x15,
}

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.DomainHash{
	0x2b, 0x6c, 0xbf, 0x58, 0xfc, 0x0a, 0x07, 0x25,
	0xff, 0x31, 0xf7, 0x1a, 0x78, 0x3c, 0xd4, 0x25,
	0xfe, 0xb6, 0x9f, 0x9e, 0xc5, 0xf2, 0x2e, 0x4b,
	0xcc, 0x7b, 0xf5, 0xe9, 0x25, 0xb5, 0x34, 0xec,
}

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version:              0,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       genesisMerkleRoot,
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   0x176850d2445,
		Bits:                 0x207fffff,
		Nonce:                0x5,
	},
	Transactions: []*externalapi.DomainTransaction{genesisCoinbaseTx},
}

var devnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var devnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x17,       // Varint
	0x00, 0x00, // Script version
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x64, 0x65, 0x76, 0x6e, 0x65, 0x74, // kaspa-devnet
}

// devnetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the development network.
var devnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(1,
	[]*externalapi.DomainTransactionInput{}, devnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, devnetGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devnetGenesisHash = externalapi.DomainHash{
	0x93, 0x71, 0x75, 0xac, 0x87, 0x55, 0xd3, 0x18,
	0xbb, 0xb9, 0x0f, 0xaf, 0x05, 0x59, 0x88, 0xf4,
	0x9c, 0x55, 0x06, 0x0d, 0xcd, 0xc2, 0x10, 0x84,
	0x66, 0x98, 0xb6, 0xc9, 0xa8, 0x4d, 0x00, 0x00,
}

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.DomainHash{
	0x3d, 0x69, 0x66, 0x1d, 0xf3, 0x7d, 0x3f, 0x75,
	0x44, 0xf8, 0x10, 0x25, 0x7d, 0x81, 0x91, 0x9b,
	0x0b, 0x01, 0xdd, 0x3f, 0x27, 0x0b, 0x23, 0x25,
	0xba, 0x30, 0x23, 0xbe, 0x04, 0x2b, 0x6c, 0xc5,
}

// devnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devnetGenesisBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version:              0,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       devnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   0x176850d2445,
		Bits:                 0x1e7fffff,
		Nonce:                0xdfb2,
	},
	Transactions: []*externalapi.DomainTransaction{devnetGenesisCoinbaseTx},
}

var simnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var simnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x00, 0x00, // Script version
	0x17,                                           // Varint
	0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49, // OP-TRUE p2sh
	0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
	0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x73, 0x69, 0x6d, 0x6e, 0x65, 0x74, // kaspa-simnet
}

// simnetGenesisCoinbaseTx is the coinbase transaction for the simnet genesis block.
var simnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(1,
	[]*externalapi.DomainTransactionInput{}, simnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, simnetGenesisTxPayload)

// simnetGenesisHash is the hash of the first block in the block DAG for
// the simnet (genesis block).
var simnetGenesisHash = externalapi.DomainHash{
	0x9f, 0x94, 0xd9, 0x2f, 0xc3, 0x93, 0xd5, 0xda,
	0x85, 0x43, 0x72, 0xc6, 0x80, 0xc8, 0x7b, 0x33,
	0xe4, 0xb4, 0x60, 0xd1, 0xed, 0xa8, 0xd4, 0xd5,
	0x7f, 0x5e, 0x8e, 0x96, 0x01, 0xfb, 0xed, 0x4f,
}

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.DomainHash{
	0xfd, 0x7d, 0xc3, 0x95, 0x9f, 0xc7, 0x0c, 0x51,
	0xba, 0x24, 0x6b, 0xfb, 0x96, 0x93, 0x21, 0xff,
	0x8d, 0xf4, 0x88, 0xfa, 0x76, 0x5e, 0x68, 0xd4,
	0x48, 0x4c, 0x27, 0xb0, 0x9b, 0x3e, 0xbc, 0x90,
}

// simnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simnetGenesisBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version:              0,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       simnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   0x176850d2494,
		Bits:                 0x207fffff,
		Nonce:                0x0,
	},
	Transactions: []*externalapi.DomainTransaction{simnetGenesisCoinbaseTx},
}

var testnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var testnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x00, 0x00, // Script version
	0x01,                                                                         // Varint
	0x00,                                                                         // OP-FALSE
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x74, 0x65, 0x73, 0x74, 0x6e, 0x65, 0x74, // kaspa-testnet
}

// testnetGenesisCoinbaseTx is the coinbase transaction for the testnet genesis block.
var testnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(1,
	[]*externalapi.DomainTransactionInput{}, testnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, testnetGenesisTxPayload)

// testnetGenesisHash is the hash of the first block in the block DAG for the test
// network (genesis block).
var testnetGenesisHash = externalapi.DomainHash{
	0x4a, 0xed, 0x01, 0x2d, 0x13, 0x16, 0x33, 0x6c,
	0x6a, 0xf8, 0xcd, 0x8f, 0xd3, 0xbf, 0x28, 0x00,
	0xa7, 0x1f, 0x4a, 0xe7, 0x89, 0x65, 0x21, 0x2e,
	0x3a, 0xd1, 0x65, 0x97, 0xb7, 0x07, 0x00, 0x00,
}

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.DomainHash{
	0xa3, 0x5a, 0xcd, 0xf2, 0x71, 0xf4, 0xd9, 0x7c,
	0x84, 0x7c, 0x05, 0x0d, 0xa1, 0x94, 0x19, 0x73,
	0x07, 0xb8, 0xf8, 0xd4, 0x4c, 0x82, 0x83, 0x92,
	0x86, 0xf6, 0x5f, 0x96, 0x38, 0x02, 0x0b, 0x05,
}

// testnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for testnet.
var testnetGenesisBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version:              0,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       testnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   0x176850d2494,
		Bits:                 0x1e7fffff,
		Nonce:                0x19f3,
	},
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
