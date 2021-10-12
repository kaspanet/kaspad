// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"github.com/kaspanet/go-muhash"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"math/big"
)

var genesisTxOuts = []*externalapi.DomainTransactionOutput{}

var genesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Subsidy
	0x00, 0x00, //script version
	0x01, // Varint
	0x00, // OP-FALSE
}

// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the main network.
var genesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0, []*externalapi.DomainTransactionInput{}, genesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, genesisTxPayload)

// genesisHash is the hash of the first block in the block DAG for the main
// network (genesis block).
var genesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xf6, 0xa5, 0xb7, 0x71, 0x27, 0x74, 0x9d, 0x30,
	0x2c, 0x86, 0xa0, 0xc4, 0xeb, 0x1d, 0x53, 0x8e,
	0xcb, 0x3d, 0xb1, 0x2f, 0x16, 0x63, 0x14, 0x9b,
	0xb3, 0x0a, 0xcb, 0x38, 0x06, 0x22, 0xa9, 0x8d,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x3e, 0x23, 0x94, 0x28, 0x80, 0xd8, 0x20, 0x30,
	0x69, 0x3f, 0xec, 0x30, 0xb8, 0x7a, 0x7f, 0x82,
	0x37, 0x83, 0x1b, 0xb5, 0x8c, 0xcd, 0x9f, 0xef,
	0xb1, 0x24, 0xbc, 0xdd, 0x4b, 0xfd, 0xbf, 0x5d,
})

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]externalapi.BlockLevelParents{},
		genesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		0x17bda5ea2ba,
		0x207fffff,
		0x0,
		0,
		0,
		big.NewInt(0),
		&externalapi.DomainHash{},
	),
	Transactions: []*externalapi.DomainTransaction{genesisCoinbaseTx},
}

var devnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var devnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Subsidy
	0x00, 0x00, // Script version
	0x01,                                                                   // Varint
	0x00,                                                                   // OP-FALSE
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x64, 0x65, 0x76, 0x6e, 0x65, 0x74, // kaspa-devnet
}

// devnetGenesisCoinbaseTx is the coinbase transaction for the genesis blocks for
// the development network.
var devnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0,
	[]*externalapi.DomainTransactionInput{}, devnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, devnetGenesisTxPayload)

// devGenesisHash is the hash of the first block in the block DAG for the development
// network (genesis block).
var devnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xce, 0xd1, 0x89, 0xfb, 0x39, 0x89, 0xd6, 0x18,
	0x7e, 0x80, 0xf0, 0x9c, 0x49, 0x61, 0x8f, 0xe1,
	0x4b, 0x1f, 0x9e, 0x7c, 0x59, 0x21, 0x65, 0xe7,
	0x1c, 0x45, 0xfa, 0xcf, 0x95, 0xb3, 0x4c, 0xa4,
})

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x4a, 0x24, 0x2c, 0xe8, 0x81, 0xed, 0x81, 0xf7,
	0x5e, 0xab, 0xde, 0xfa, 0x8e, 0xd0, 0xd7, 0x01,
	0xaa, 0x0a, 0x64, 0x3f, 0x0f, 0xf6, 0xf5, 0x0b,
	0xca, 0x28, 0x43, 0x79, 0xa6, 0x2f, 0x5c, 0xe1,
})

// devnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]externalapi.BlockLevelParents{},
		devnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		0x11e9db49828,
		0x1e7fffff,
		0x1a226,
		0,
		0,
		big.NewInt(0),
		&externalapi.DomainHash{},
	),
	Transactions: []*externalapi.DomainTransaction{devnetGenesisCoinbaseTx},
}

var simnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var simnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Subsidy
	0x00, 0x00, // Script version
	0x01,                                                                   // Varint
	0x00,                                                                   // OP-FALSE
	0x6b, 0x61, 0x73, 0x70, 0x61, 0x2d, 0x73, 0x69, 0x6d, 0x6e, 0x65, 0x74, // kaspa-simnet
}

// simnetGenesisCoinbaseTx is the coinbase transaction for the simnet genesis block.
var simnetGenesisCoinbaseTx = transactionhelper.NewSubnetworkTransaction(0,
	[]*externalapi.DomainTransactionInput{}, simnetGenesisTxOuts,
	&subnetworks.SubnetworkIDCoinbase, 0, simnetGenesisTxPayload)

// simnetGenesisHash is the hash of the first block in the block DAG for
// the simnet (genesis block).
var simnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x74, 0xb5, 0x85, 0x99, 0xb7, 0x29, 0x68, 0x00,
	0x88, 0x29, 0xe2, 0x07, 0xef, 0x1e, 0x9a, 0x79,
	0xf7, 0xab, 0xeb, 0xa1, 0x8a, 0xa8, 0xec, 0x8d,
	0xc9, 0xa0, 0x94, 0x5f, 0x89, 0xd7, 0x54, 0x2b,
})

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xd0, 0x61, 0xde, 0xac, 0x16, 0xf0, 0x2a, 0x70,
	0x64, 0x6d, 0x3b, 0x74, 0x3e, 0xfb, 0xce, 0xc2,
	0x8d, 0x2e, 0x8c, 0x4b, 0x52, 0x9f, 0x99, 0xdf,
	0x24, 0xe6, 0x99, 0xd8, 0xe2, 0xa9, 0xfa, 0x51,
})

// simnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]externalapi.BlockLevelParents{},
		simnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		0x17bda60426a,
		0x207fffff,
		0x1,
		0,
		0,
		big.NewInt(0),
		&externalapi.DomainHash{},
	),
	Transactions: []*externalapi.DomainTransaction{simnetGenesisCoinbaseTx},
}

var testnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var testnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Subsidy
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
var testnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xd8, 0x9e, 0xd0, 0xdc, 0x49, 0x17, 0xa0, 0xc7,
	0x11, 0x6a, 0x90, 0xd6, 0x45, 0xf6, 0x02, 0xa4,
	0xcf, 0xda, 0x2e, 0x4b, 0x4e, 0x04, 0x03, 0x76,
	0x6a, 0xf7, 0x35, 0x24, 0xd8, 0x94, 0xd0, 0xd6,
})

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x9e, 0xd2, 0x41, 0xe1, 0x56, 0x38, 0xaf, 0x52,
	0x92, 0x84, 0xb2, 0x41, 0x94, 0xd4, 0xdc, 0x38,
	0x23, 0xb7, 0x82, 0x64, 0xbe, 0x5f, 0x0a, 0x44,
	0x52, 0x66, 0xae, 0xd6, 0xe8, 0xf2, 0x3b, 0xef,
})

// testnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for testnet.
var testnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]externalapi.BlockLevelParents{},
		testnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		0x17bda60426a,
		0x1e7fffff,
		0x278c9,
		0,
		0,
		big.NewInt(0),
		&externalapi.DomainHash{},
	),
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
