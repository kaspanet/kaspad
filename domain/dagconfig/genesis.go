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
	0xbb, 0xc1, 0x88, 0xdd, 0x56, 0x9d, 0x46, 0xbd,
	0x36, 0xb0, 0x31, 0x52, 0x49, 0x93, 0xac, 0x70,
	0x1d, 0x36, 0xf1, 0xb3, 0xd2, 0x2f, 0xe5, 0x51,
	0x7c, 0x8b, 0x1a, 0xaf, 0x3c, 0x82, 0x6f, 0x18,
}

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.DomainHash{
	0x2f, 0x55, 0x37, 0x11, 0x8a, 0x30, 0xcd, 0xcd,
	0xa7, 0xdb, 0xe5, 0xe6, 0x42, 0xf9, 0x1b, 0xf3,
	0xb4, 0x62, 0xd6, 0xb6, 0xed, 0xc0, 0x5c, 0xe1,
	0x6e, 0xee, 0x0f, 0x3c, 0xdc, 0xf6, 0x01, 0x15,
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
		TimeInMilliseconds:   0x175bc9e305a,
		Bits:                 0x207fffff,
		Nonce:                0x0,
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
	0x92, 0xb5, 0x28, 0xd3, 0xaa, 0x6d, 0x8b, 0x30,
	0x49, 0x19, 0x53, 0x6f, 0x62, 0xce, 0x9a, 0x82,
	0x2f, 0x91, 0xd4, 0x33, 0x24, 0xbc, 0x39, 0xe6,
	0xad, 0x53, 0xe3, 0x97, 0x5f, 0x03, 0x00, 0x00,
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
		TimeInMilliseconds:   0x175bca27b7f,
		Bits:                 0x1e7fffff,
		Nonce:                0x1a9ba,
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
	0x84, 0x96, 0x38, 0xb6, 0x5c, 0x44, 0xc0, 0xb9,
	0x3c, 0x48, 0x03, 0x7c, 0x2e, 0xee, 0x0a, 0xbf,
	0xfb, 0x54, 0xc8, 0x5f, 0x99, 0xd6, 0x21, 0x3d,
	0x3f, 0xdd, 0xac, 0xb1, 0xe7, 0x30, 0x7e, 0x05,
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
		TimeInMilliseconds:   0x175bca27c39,
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
	0xa9, 0xbe, 0xa7, 0xd9, 0x0f, 0xd2, 0xbd, 0xfb,
	0xd8, 0x09, 0x4d, 0x6a, 0x49, 0xa7, 0x59, 0x93,
	0xd1, 0x35, 0xce, 0x61, 0x18, 0x07, 0x0b, 0xe6,
	0xb9, 0xec, 0xad, 0x68, 0xe4, 0x2d, 0x00, 0x00,
}

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.DomainHash{
	0x7c, 0x5a, 0x9a, 0xb4, 0xa6, 0xd5, 0x03, 0xf3,
	0x19, 0x3c, 0x26, 0x82, 0xf5, 0x45, 0xdf, 0xe3,
	0x08, 0x6d, 0x94, 0xfc, 0x7d, 0xb9, 0x42, 0x4b,
	0x2c, 0x38, 0xf2, 0x5c, 0x64, 0xbc, 0xb4, 0x98,
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
		TimeInMilliseconds:   0x175bcac06ec,
		Bits:                 0x1e7fffff,
		Nonce:                0x568f,
	},
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
