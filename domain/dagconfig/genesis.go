// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
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
var genesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x92, 0x29, 0x3c, 0xbd, 0x65, 0xa8, 0x6d, 0x9c,
	0xc1, 0xb2, 0x8f, 0x63, 0xc9, 0x2a, 0x50, 0x90,
	0x28, 0xe7, 0x45, 0x57, 0x1d, 0xdc, 0xc2, 0xcd,
	0xdd, 0x9b, 0x99, 0x4c, 0x22, 0xc6, 0x21, 0x89,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x0f, 0xa7, 0x42, 0x5e, 0xa9, 0xec, 0xd7, 0x1f,
	0x40, 0x53, 0x31, 0xe4, 0x88, 0x22, 0x31, 0x9a,
	0xfb, 0xa7, 0xf4, 0x66, 0x9b, 0xa4, 0x37, 0xe0,
	0x86, 0x54, 0x21, 0xaa, 0x6d, 0x4e, 0x87, 0xe6,
})

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]*externalapi.DomainHash{},
		genesisMerkleRoot,
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0x176eb9ddaf4,
		0x207fffff,
		0x0,
	),
	Transactions: []*externalapi.DomainTransaction{genesisCoinbaseTx},
}

var devnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var devnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
	0x00, 0x00, // Script version
	0x17,                                           // Varint
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
var devnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x86, 0xd8, 0x94, 0xee, 0x2f, 0x73, 0x62, 0xbd,
	0x8d, 0xa9, 0x5a, 0x2b, 0x45, 0x91, 0xb9, 0x65,
	0xcc, 0x7f, 0x0d, 0xf9, 0x5d, 0x20, 0x3f, 0xf4,
	0x92, 0x37, 0xc2, 0x15, 0xf3, 0x9c, 0x8c, 0x2d,
})

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x62, 0xf4, 0xfa, 0xfc, 0xb2, 0x28, 0xfc, 0x33,
	0x1b, 0xae, 0xaf, 0x4a, 0xdc, 0xa9, 0xc8, 0xc6,
	0xfb, 0xc5, 0xfc, 0xc7, 0x2c, 0x86, 0x44, 0x33,
	0xbd, 0x75, 0xf7, 0x93, 0x2c, 0x11, 0xa8, 0x2a,
})

// devnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]*externalapi.DomainHash{},
		devnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0x11e9db49828,
		0x1e7fffff,
		0x20a4f,
	),
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
var simnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x2e, 0x37, 0x4f, 0xac, 0xa9, 0xfb, 0x88, 0xea,
	0x0e, 0xb7, 0x8f, 0xb2, 0x1e, 0xbe, 0xb6, 0xe5,
	0xbf, 0x59, 0xde, 0x29, 0x98, 0x55, 0x9e, 0x21,
	0xf2, 0x3b, 0x55, 0xcc, 0x41, 0xb8, 0xd9, 0x55,
})

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x0a, 0x84, 0xe5, 0xf0, 0xae, 0x6d, 0x26, 0xd2,
	0xf6, 0xdb, 0x94, 0x00, 0xfc, 0xcd, 0xea, 0x4b,
	0x61, 0x17, 0x1b, 0xa4, 0x32, 0xae, 0xde, 0x27,
	0xfb, 0x3e, 0x1d, 0x46, 0x17, 0xf2, 0xb8, 0xac,
})

// simnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]*externalapi.DomainHash{},
		simnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0x176c86a5c26,
		0x207fffff,
		0x1,
	),
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
var testnetGenesisHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x49, 0xa6, 0x19, 0xe4, 0x25, 0xa0, 0x8d, 0xae,
	0xd9, 0x67, 0x76, 0x82, 0xd8, 0x8e, 0x93, 0xa3,
	0xf5, 0x42, 0x5a, 0x02, 0x4d, 0x55, 0xef, 0x5e,
	0x39, 0x61, 0x9f, 0x2d, 0xd9, 0x51, 0xe4, 0x55,
})

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xc5, 0xef, 0xd2, 0xc7, 0xa6, 0x18, 0xe0, 0xd0,
	0xd1, 0x47, 0x3c, 0x44, 0x58, 0xaa, 0xdb, 0xfb,
	0x82, 0xfc, 0x9f, 0x88, 0x73, 0x93, 0xb1, 0x91,
	0x32, 0xec, 0xf9, 0x20, 0xd1, 0x6c, 0xce, 0xe9,
})

// testnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for testnet.
var testnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]*externalapi.DomainHash{},
		testnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		0x177e83f864a,
		0x1e7fffff,
		0x4d8e0,
	),
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
