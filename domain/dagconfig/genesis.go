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
var genesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0, []*externalapi.DomainTransactionInput{}, genesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, genesisTxPayload)

// genesisHash is the hash of the first block in the block DAG for the main
// network (genesis block).
var genesisHash = externalapi.DomainHash{
	0xf2, 0xb8, 0x52, 0x1f, 0xb8, 0xa7, 0x6a, 0x30,
	0xd1, 0x23, 0x14, 0x6e, 0x3c, 0x87, 0x0f, 0x52,
	0xb0, 0xc9, 0xa6, 0xc6, 0x83, 0x0b, 0x50, 0xfe,
	0xb3, 0x9e, 0x20, 0xa1, 0xc0, 0x9a, 0x84, 0x34,
}

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.DomainHash{
	0xc8, 0xa9, 0xfe, 0x04, 0xcc, 0x59, 0x97, 0xf4,
	0x60, 0xdc, 0xe2, 0xec, 0xd8, 0x47, 0x15, 0x6d,
	0x04, 0xca, 0x2c, 0x79, 0xff, 0xd6, 0x8e, 0x18,
	0x61, 0xe7, 0x6a, 0x59, 0x46, 0x44, 0xbc, 0x7e,
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
		TimeInMilliseconds:   0x176ad4afe4e,
		Bits:                 0x207fffff,
		Nonce:                0x0,
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
var devnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0,
	[]*externalapi.DomainTransactionInput{}, devnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, devnetGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devnetGenesisHash = externalapi.DomainHash{
	0x29, 0xbc, 0x24, 0x32, 0x00, 0x05, 0xd9, 0x45,
	0x8b, 0x97, 0xbe, 0xb5, 0x88, 0xb2, 0x70, 0x4e,
	0x97, 0x4b, 0x41, 0x78, 0x74, 0x7b, 0x9f, 0xbe,
	0x08, 0x76, 0x45, 0xab, 0x24, 0x89, 0xf2, 0x7f,
}

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.DomainHash{
	0xd4, 0x6b, 0x41, 0xb0, 0xb6, 0x02, 0x34, 0xbd,
	0x0f, 0xbf, 0x6b, 0x0d, 0x67, 0x94, 0x0a, 0x2c,
	0x5b, 0xaf, 0x2b, 0x16, 0x20, 0x86, 0x90, 0x56,
	0x61, 0xc1, 0x68, 0xf9, 0xec, 0x48, 0x82, 0x11,
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
		TimeInMilliseconds:   0x176ad4afe4e,
		Bits:                 0x1e7fffff,
		Nonce:                0xd304,
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
var simnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0,
	[]*externalapi.DomainTransactionInput{}, simnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, simnetGenesisTxPayload)

// simnetGenesisHash is the hash of the first block in the block DAG for
// the simnet (genesis block).
var simnetGenesisHash = externalapi.DomainHash{
	0x00, 0x02, 0xf0, 0xd8, 0x2d, 0xa1, 0x73, 0x75,
	0x21, 0xd2, 0xa2, 0x50, 0xc9, 0xc9, 0xfb, 0x06,
	0x9a, 0x0b, 0x5a, 0x6d, 0xd5, 0x73, 0x03, 0x3f,
	0xe3, 0xb5, 0xd5, 0x7c, 0xcc, 0x25, 0x49, 0x4f,
}

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.DomainHash{
	0x66, 0xa3, 0xa9, 0x01, 0x25, 0xf7, 0xe5, 0x40,
	0x51, 0xb0, 0x57, 0x25, 0x9b, 0x8d, 0xa8, 0x5c,
	0xb2, 0x4f, 0xb6, 0xb5, 0x6e, 0xad, 0x00, 0xb4,
	0xcc, 0x1a, 0xb2, 0x98, 0x5d, 0xf4, 0x4a, 0x8b,
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
		TimeInMilliseconds:   0x176ad4afedf,
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
var testnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0,
	[]*externalapi.DomainTransactionInput{}, testnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, testnetGenesisTxPayload)

// testnetGenesisHash is the hash of the first block in the block DAG for the test
// network (genesis block).
var testnetGenesisHash = externalapi.DomainHash{
	0x82, 0x98, 0x3a, 0xb9, 0xc5, 0x27, 0xe6, 0x0d,
	0xb0, 0xe3, 0x63, 0x49, 0xa3, 0x5b, 0xe8, 0x02,
	0xd0, 0x0f, 0x12, 0x21, 0xf6, 0x8b, 0x9e, 0xb0,
	0xd7, 0xe1, 0xce, 0x72, 0x88, 0xda, 0x8f, 0x06,
}

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.DomainHash{
	0xf2, 0x9c, 0xb7, 0xb5, 0x90, 0x64, 0x0b, 0x2c,
	0xea, 0x47, 0x86, 0x19, 0x9b, 0x1f, 0x16, 0x19,
	0xf7, 0xfb, 0x61, 0x9b, 0x38, 0x87, 0x78, 0x54,
	0x52, 0x75, 0x39, 0x94, 0xe6, 0xb4, 0x8f, 0x6b,
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
		TimeInMilliseconds:   0x176ad4afedf,
		Bits:                 0x1e7fffff,
		Nonce:                0x85777,
	},
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
