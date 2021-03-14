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
	0xf2, 0xfa, 0x21, 0x85, 0xbd, 0x7c, 0x01, 0x22,
	0x23, 0xf7, 0x58, 0x8e, 0xe9, 0xb4, 0x01, 0xf3,
	0xfe, 0xf9, 0xb9, 0x70, 0xff, 0xec, 0xdd, 0x19,
	0xc1, 0xdd, 0x6d, 0x8d, 0x8b, 0x93, 0x65, 0xc9,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xca, 0x21, 0x52, 0xc9, 0x86, 0x99, 0x26, 0xf5,
	0x66, 0x23, 0xb9, 0xbf, 0x22, 0x01, 0x83, 0x2e,
	0x7f, 0xf4, 0xda, 0xfb, 0x6f, 0x9d, 0x69, 0x94,
	0xcf, 0x48, 0x8b, 0x8a, 0x16, 0x8b, 0x33, 0x37,
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
		0x17772f4a238,
		0x207fffff,
		0x5,
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
	0xac, 0x97, 0xf4, 0x05, 0x65, 0x60, 0x1c, 0x89,
	0x93, 0x86, 0x86, 0x9d, 0xa8, 0x25, 0x75, 0xd0,
	0x73, 0x18, 0x3f, 0xfc, 0x7d, 0x48, 0xc4, 0x51,
	0x6e, 0x86, 0x23, 0x0c, 0x29, 0x8b, 0x83, 0x4c,
})

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x07, 0xf1, 0x02, 0xe1, 0xca, 0xdc, 0x1c, 0xac,
	0xb3, 0xc1, 0x5e, 0x3e, 0x41, 0x1b, 0x55, 0x78,
	0x71, 0xab, 0xed, 0xc6, 0xaf, 0xe0, 0xbe, 0x40,
	0x7e, 0xe0, 0x7e, 0x31, 0x8a, 0xf7, 0x23, 0xa4,
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
		0x5a8d,
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
	0xde, 0x29, 0x53, 0xb7, 0xa3, 0xac, 0x9c, 0xc1,
	0xf6, 0xca, 0xa7, 0xc8, 0x25, 0x12, 0x80, 0xf7,
	0xb1, 0xe4, 0x42, 0xbc, 0x58, 0x89, 0x3f, 0x36,
	0xd9, 0x70, 0x6b, 0xbd, 0xae, 0x8b, 0xae, 0xc4,
})

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xb7, 0xcd, 0xe4, 0x96, 0x5b, 0xbc, 0x6b, 0x5c,
	0xca, 0x40, 0x0a, 0x8f, 0x13, 0xf8, 0x15, 0x1b,
	0x0e, 0xd2, 0xdc, 0x2c, 0x6e, 0x8c, 0x25, 0x7a,
	0x89, 0x6a, 0x69, 0xcd, 0x58, 0x7a, 0x30, 0xd0,
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
		0x17772f7a1f8,
		0x207fffff,
		0x5,
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
	0x30, 0x66, 0xaa, 0x51, 0xc0, 0x1f, 0x8a, 0x8d,
	0x6e, 0x09, 0x65, 0x7e, 0x24, 0x3b, 0x9e, 0xb0,
	0x0d, 0x89, 0xa7, 0x21, 0x64, 0x67, 0xbd, 0x4d,
	0x76, 0x74, 0x67, 0xe4, 0xdf, 0x16, 0x9a, 0x93,
})

// testnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for testnet.
var testnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x16, 0x02, 0x73, 0xc3, 0x15, 0x78, 0x7e, 0x7b,
	0xc9, 0x7f, 0x6f, 0x54, 0xd0, 0x4b, 0x39, 0xed,
	0xa7, 0xea, 0xa0, 0x63, 0xc9, 0xbc, 0x23, 0x4b,
	0xaa, 0x24, 0xaf, 0x04, 0x74, 0x2d, 0x95, 0x2b,
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
		0x17772f4a25e,
		0x1e7fffff,
		0x11447,
	),
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
