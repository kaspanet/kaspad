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
	0x8c, 0x74, 0x62, 0xc9, 0xb6, 0xa8, 0xb2, 0x7c,
	0x8d, 0x03, 0xa3, 0x7e, 0x45, 0x73, 0x31, 0x77,
	0xc7, 0xe1, 0x00, 0xa8, 0xc7, 0x75, 0xe9, 0xaa,
	0x31, 0x02, 0xa9, 0x82, 0x9f, 0xad, 0x34, 0xc8,
}

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.DomainHash{
	0x32, 0xea, 0x93, 0x9a, 0x1f, 0x00, 0x50, 0xc3,
	0x97, 0x2c, 0x3d, 0xdf, 0x28, 0xb4, 0x8f, 0x1d,
	0x75, 0x9f, 0xb1, 0x82, 0x99, 0x79, 0x7a, 0x48,
	0xc9, 0xf6, 0x05, 0xc6, 0xae, 0x30, 0x49, 0xf7,
}

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       genesisMerkleRoot,
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   0x1763db5c4a9,
		Bits:                 0x207fffff,
		Nonce:                0x1,
	},
	Transactions: []*externalapi.DomainTransaction{genesisCoinbaseTx},
}

var devnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

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
var devnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(1,
	[]*externalapi.DomainTransactionInput{}, devnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, devnetGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devnetGenesisHash = externalapi.DomainHash{
	0xee, 0xce, 0x68, 0x63, 0x61, 0xb4, 0xa8, 0x09,
	0x5d, 0xa3, 0x91, 0x6c, 0x12, 0x20, 0x27, 0xdd,
	0xf8, 0x16, 0x74, 0x8e, 0xd8, 0x7a, 0xfe, 0x2c,
	0xb7, 0x98, 0xe6, 0x9d, 0x47, 0x07, 0x02, 0xc5,
}

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.DomainHash{
	0xdf, 0x52, 0x65, 0x3a, 0x5a, 0xd4, 0x07, 0x4e,
	0xad, 0xac, 0xb3, 0xd7, 0xd6, 0x9a, 0xf5, 0xd3,
	0x68, 0x05, 0x4d, 0xef, 0xd9, 0x41, 0x28, 0x84,
	0xa9, 0x56, 0xdd, 0x68, 0x60, 0x1b, 0x8d, 0x2c,
}

// devnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devnetGenesisBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       devnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   0x1763db5c4a9,
		Bits:                 0x1e7fffff,
		Nonce:                0xb6c8,
	},
	Transactions: []*externalapi.DomainTransaction{devnetGenesisCoinbaseTx},
}

var simnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var simnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
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
	0xe3, 0xa4, 0x4a, 0xe5, 0xdc, 0x3d, 0x39, 0x6a,
	0xc8, 0x5b, 0x1b, 0x95, 0x30, 0x05, 0x7d, 0xb9,
	0xd4, 0xfa, 0x30, 0x9a, 0x20, 0x7a, 0x42, 0x54,
	0xf8, 0x10, 0x73, 0xc0, 0x15, 0x31, 0xf5, 0x1a,
}

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.DomainHash{
	0x16, 0x07, 0x15, 0x0f, 0x1b, 0xc0, 0x26, 0x27,
	0x42, 0xc5, 0x84, 0x77, 0xdb, 0x58, 0xf7, 0x87,
	0xa8, 0xe9, 0x9f, 0x21, 0x73, 0xa0, 0x9d, 0x96,
	0x6a, 0x99, 0x55, 0x46, 0x7b, 0xb2, 0x1b, 0x99,
}

// simnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simnetGenesisBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       simnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   0x1763db5c4a9,
		Bits:                 0x207fffff,
		Nonce:                0x0,
	},
	Transactions: []*externalapi.DomainTransaction{simnetGenesisCoinbaseTx},
}

var testnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var testnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
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
	0x17, 0xb3, 0x16, 0xd3, 0x4f, 0xb5, 0x2c, 0xc1,
	0x22, 0x53, 0x1a, 0xc9, 0xde, 0x79, 0xc3, 0x03,
	0x53, 0xa2, 0x1a, 0x0d, 0x00, 0x40, 0x7d, 0x49,
	0x66, 0x0c, 0x76, 0xf2, 0x61, 0xe4, 0x9a, 0x23,
}

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.DomainHash{
	0xd7, 0x16, 0x4a, 0x38, 0x3b, 0x8a, 0x67, 0xc2,
	0x3b, 0x89, 0x12, 0x1c, 0xcb, 0x97, 0x89, 0xe1,
	0x12, 0x82, 0x12, 0xc2, 0x69, 0x95, 0x7f, 0x03,
	0x29, 0xd1, 0x4f, 0xdd, 0xf1, 0x93, 0xd8, 0x47,
}

// testnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for testnet.
var testnetGenesisBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version:              1,
		ParentHashes:         []*externalapi.DomainHash{},
		HashMerkleRoot:       testnetGenesisMerkleRoot,
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   0x1763db5c4a9,
		Bits:                 0x1e7fffff,
		Nonce:                0x493d,
	},
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
