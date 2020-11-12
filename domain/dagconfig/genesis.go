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
	0x37, 0xf9, 0x09, 0x98, 0x0e, 0x12, 0xf6, 0xa2,
	0xaa, 0x0d, 0x09, 0xfa, 0x41, 0xad, 0x95, 0x1a,
	0x5b, 0x38, 0x14, 0x12, 0xe9, 0x02, 0x14, 0xcd,
	0xba, 0x64, 0xf6, 0x84, 0x72, 0xc4, 0xa4, 0x3a,
}

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.DomainHash{
	0xca, 0x85, 0x56, 0x27, 0xc7, 0x6a, 0xb5, 0x7a,
	0x26, 0x1d, 0x63, 0x62, 0x1e, 0x57, 0x21, 0xf0,
	0x5e, 0x60, 0x1f, 0xee, 0x1d, 0x4d, 0xaa, 0x53,
	0x72, 0xe1, 0x16, 0xda, 0x4b, 0xb3, 0xd8, 0x0e,
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
		TimeInMilliseconds:   0x1730a81bdb4,
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
	0xd3, 0xad, 0xd6, 0xe4, 0x6b, 0xc2, 0x33, 0xa9,
	0x20, 0x03, 0x1e, 0xf3, 0xe6, 0x8a, 0xf4, 0x08,
	0x91, 0xa1, 0x25, 0xc7, 0xc1, 0xf1, 0x5b, 0x3e,
	0x74, 0x72, 0xb5, 0x8a, 0xa0, 0x10, 0x00, 0x00,
}

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.DomainHash{
	0x00, 0x94, 0xfd, 0xff, 0x4d, 0xb2, 0x4d, 0x18,
	0x95, 0x21, 0x36, 0x2a, 0x14, 0xfb, 0x19, 0x7a,
	0x99, 0x51, 0x7e, 0x3f, 0x44, 0xf6, 0x2e, 0x0b,
	0xe7, 0xb3, 0xc0, 0xbb, 0x00, 0x3b, 0x0b, 0xbd,
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
		TimeInMilliseconds:   0x17305b05694,
		Bits:                 0x1e7fffff,
		Nonce:                282366,
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
	0x50, 0x01, 0x7e, 0x84, 0x55, 0xc0, 0xab, 0x9c,
	0xca, 0xf5, 0xc1, 0x5d, 0xbe, 0x57, 0x0a, 0x80,
	0x1f, 0x93, 0x00, 0x34, 0xe6, 0xee, 0xc2, 0xee,
	0xff, 0x57, 0xc1, 0x66, 0x2a, 0x63, 0x4b, 0x23,
}

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.DomainHash{
	0x79, 0x77, 0x9c, 0xad, 0x8d, 0x5a, 0x37, 0x57,
	0x75, 0x8b, 0x2f, 0xa5, 0x82, 0x47, 0x2f, 0xb6,
	0xbe, 0x24, 0x5f, 0xcb, 0x21, 0x68, 0x21, 0x44,
	0x45, 0x39, 0x44, 0xaf, 0xab, 0x9f, 0x0f, 0xc1,
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
		TimeInMilliseconds:   0x173001df3d5,
		Bits:                 0x207fffff,
		Nonce:                1,
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
	0x29, 0x0c, 0x43, 0x42, 0x66, 0x2d, 0x1b, 0x05,
	0x0b, 0xc7, 0xb2, 0x26, 0x8d, 0x53, 0x16, 0xbf,
	0x5a, 0x2d, 0x7e, 0xff, 0xee, 0x06, 0xb7, 0x2b,
	0x41, 0x9f, 0xe8, 0xf0, 0xfa, 0xc7, 0xd1, 0x1a,
}

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.DomainHash{
	0xA0, 0xA1, 0x3D, 0xFD, 0x86, 0x41, 0x35, 0xC8,
	0xBD, 0xBB, 0xE6, 0x37, 0x35, 0xBB, 0x4C, 0x51,
	0x11, 0x7B, 0x26, 0x90, 0x15, 0x64, 0x0F, 0x42,
	0x6D, 0x2B, 0x6F, 0x37, 0x4D, 0xC1, 0xA9, 0x72,
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
		TimeInMilliseconds:   0x1730a66a9d9,
		Bits:                 0x1e7fffff,
		Nonce:                0x162ca,
	},
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
