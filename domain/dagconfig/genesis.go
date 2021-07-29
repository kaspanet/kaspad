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
)

var genesisTxOuts = []*externalapi.DomainTransactionOutput{}

var genesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
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
	0x3b, 0x5e, 0xae, 0x38, 0x47, 0xc5, 0x9e, 0x06,
	0x8b, 0xc2, 0x9d, 0xe3, 0x26, 0xa2, 0x75, 0x10,
	0x35, 0xdd, 0xf8, 0x58, 0xe5, 0xb8, 0xcc, 0x83,
	0xd8, 0x9e, 0x07, 0xe0, 0x40, 0xdf, 0xb1, 0x96,
})

// genesisMerkleRoot is the hash of the first transaction in the genesis block
// for the main network.
var genesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xc5, 0x04, 0xfd, 0x49, 0x6a, 0xa5, 0x3f, 0x44,
	0x51, 0x2c, 0x5a, 0x73, 0x9c, 0x28, 0xd5, 0x30,
	0xfa, 0x54, 0xbb, 0x8e, 0x88, 0x82, 0xbb, 0x9a,
	0xdb, 0xd6, 0x7f, 0x09, 0xd0, 0x2c, 0x6c, 0x47,
})

// genesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the main network.
var genesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]*externalapi.DomainHash{},
		genesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		0x177bfd44981,
		0x207fffff,
		0x0,
	),
	Transactions: []*externalapi.DomainTransaction{genesisCoinbaseTx},
}

var devnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var devnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
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
	0x31, 0x4d, 0x29, 0xb4, 0x1e, 0x53, 0x50, 0x6a,
	0xbb, 0x7c, 0xfa, 0x59, 0x84, 0xf8, 0xe4, 0x84,
	0x66, 0x3e, 0x97, 0xe3, 0x4d, 0x56, 0xd0, 0x86,
	0xbc, 0x18, 0x69, 0x6b, 0xb2, 0x7b, 0x9c, 0x02,
})

// devnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var devnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0xfc, 0x35, 0x93, 0x85, 0x4d, 0x0a, 0x24, 0xe3,
	0xc4, 0x52, 0xdd, 0x4d, 0xe5, 0xf1, 0x4d, 0xf1,
	0x5e, 0xff, 0xcd, 0x40, 0x81, 0x63, 0x53, 0x4e,
	0xec, 0x86, 0x62, 0x99, 0x91, 0x28, 0x45, 0xa4,
})

// devnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var devnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]*externalapi.DomainHash{},
		devnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		0x11e9db49828,
		0x1e7fffff,
		0x10eff,
	),
	Transactions: []*externalapi.DomainTransaction{devnetGenesisCoinbaseTx},
}

var simnetGenesisTxOuts = []*externalapi.DomainTransactionOutput{}

var simnetGenesisTxPayload = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Blue score
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
	0x5b, 0x2e, 0xbf, 0x45, 0x96, 0x12, 0x84, 0xf7,
	0x73, 0xb0, 0x88, 0x45, 0x5e, 0x9b, 0xb3, 0x80,
	0x9a, 0xcc, 0x92, 0x4c, 0x99, 0xcd, 0xdc, 0xd0,
	0x5f, 0xe2, 0x24, 0xa8, 0x9d, 0xd6, 0x59, 0x25,
})

// simnetGenesisMerkleRoot is the hash of the first transaction in the genesis block
// for the devopment network.
var simnetGenesisMerkleRoot = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	0x28, 0x5d, 0x64, 0x1f, 0x27, 0xdd, 0x7e, 0x56,
	0x46, 0x75, 0xeb, 0xe5, 0x05, 0x58, 0x5b, 0x7a,
	0xd4, 0x0c, 0x65, 0x07, 0xb8, 0x66, 0x0e, 0xd8,
	0x17, 0xfc, 0xc7, 0x53, 0x77, 0x48, 0xd2, 0x3a,
})

// simnetGenesisBlock defines the genesis block of the block DAG which serves as the
// public transaction ledger for the development network.
var simnetGenesisBlock = externalapi.DomainBlock{
	Header: blockheader.NewImmutableBlockHeader(
		0,
		[]*externalapi.DomainHash{},
		simnetGenesisMerkleRoot,
		&externalapi.DomainHash{},
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		0x177bfd44a10,
		0x207fffff,
		0x0,
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
	0xa0, 0xef, 0xc3, 0x0b, 0x5e, 0x19, 0x71, 0x1b,
	0xe2, 0x62, 0xa3, 0x98, 0x0f, 0xdf, 0xb8, 0x99,
	0x28, 0xb6, 0x86, 0x88, 0xff, 0xde, 0x78, 0x69,
	0xe8, 0xdd, 0x75, 0x2a, 0xbf, 0xc1, 0x70, 0xda,
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
		externalapi.NewDomainHashFromByteArray(muhash.EmptyMuHashHash.AsArray()),
		0x17a56662e7d,
		0x1e7fffff,
		0x4c933,
	),
	Transactions: []*externalapi.DomainTransaction{testnetGenesisCoinbaseTx},
}
