package blockvalidator_test

import (
	"math"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
)

func TestChainedTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.BlockCoinbaseMaturity = 0

		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(params, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown()

		block1Hash, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block1, err := tc.GetBlock(block1Hash)
		if err != nil {
			t.Fatalf("Error getting block1: %+v", err)
		}

		tx1, err := testutils.CreateTransaction(block1.Transactions[0])
		if err != nil {
			t.Fatalf("Error creating tx1: %+v", err)
		}

		chainedTx, err := testutils.CreateTransaction(tx1)
		if err != nil {
			t.Fatalf("Error creating chainedTx: %+v", err)
		}

		// Check that a block is invalid if it contains chained transactions
		_, err = tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil,
			[]*externalapi.DomainTransaction{tx1, chainedTx})
		if !errors.Is(err, ruleerrors.ErrChainedTransactions) {
			t.Fatalf("unexpected error %+v", err)
		}

		block2Hash, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error %+v", err)
		}

		block2, err := tc.GetBlock(block2Hash)
		if err != nil {
			t.Fatalf("Error getting block2: %+v", err)
		}

		tx2, err := testutils.CreateTransaction(block2.Transactions[0])
		if err != nil {
			t.Fatalf("Error creating tx2: %+v", err)
		}

		// Check that a block is valid if it contains two non chained transactions
		_, err = tc.AddBlock([]*externalapi.DomainHash{block2Hash}, nil,
			[]*externalapi.DomainTransaction{tx1, tx2})
		if err != nil {
			t.Fatalf("unexpected error %+v", err)
		}
	})
}

// TestCheckBlockSanity tests the CheckBlockSanity function to ensure it works
// as expected.
func TestCheckBlockSanity(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		consensus, teardown, err := factory.NewTestConsensus(params, "TestCheckBlockSanity")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown()
		blockHash := consensushashing.BlockHash(&exampleValidBlock)
		if len(exampleValidBlock.Transactions) < 3 {
			t.Fatalf("Too few transactions in block, expect at least 3, got %v", len(exampleValidBlock.Transactions))
		}

		consensus.BlockStore().Stage(blockHash, &exampleValidBlock)

		err = consensus.BlockValidator().ValidateBodyInIsolation(blockHash)
		if err != nil {
			t.Fatalf("Failed validating block in isolation: %v", err)
		}

		// Test with block with wrong transactions sorting order
		blockHash = consensushashing.BlockHash(&blockWithWrongTxOrder)
		consensus.BlockStore().Stage(blockHash, &blockWithWrongTxOrder)
		err = consensus.BlockValidator().ValidateBodyInIsolation(blockHash)
		if !errors.Is(err, ruleerrors.ErrTransactionsNotSorted) {
			t.Errorf("CheckBlockSanity: Expected ErrTransactionsNotSorted error, instead got %v", err)
		}

		// Test a block with invalid parents order
		// We no longer require blocks to have ordered parents
		blockHash = consensushashing.BlockHash(&unOrderedParentsBlock)
		consensus.BlockStore().Stage(blockHash, &unOrderedParentsBlock)
		err = consensus.BlockValidator().ValidateBodyInIsolation(blockHash)
		if err != nil {
			t.Errorf("CheckBlockSanity: Expected block to be be body in isolation valid, got error instead: %v", err)
		}
	})
}

const version = 0

var unOrderedParentsBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version: 0x00000000,
		ParentHashes: []*externalapi.DomainHash{
			{
				0x4b, 0xb0, 0x75, 0x35, 0xdf, 0xd5, 0x8e, 0x0b,
				0x3c, 0xd6, 0x4f, 0xd7, 0x15, 0x52, 0x80, 0x87,
				0x2a, 0x04, 0x71, 0xbc, 0xf8, 0x30, 0x95, 0x52,
				0x6a, 0xce, 0x0e, 0x38, 0xc6, 0x00, 0x00, 0x00,
			},
			{
				0x16, 0x5e, 0x38, 0xe8, 0xb3, 0x91, 0x45, 0x95,
				0xd9, 0xc6, 0x41, 0xf3, 0xb8, 0xee, 0xc2, 0xf3,
				0x46, 0x11, 0x89, 0x6b, 0x82, 0x1a, 0x68, 0x3b,
				0x7a, 0x4e, 0xde, 0xfe, 0x2c, 0x00, 0x00, 0x00,
			},
		},
		HashMerkleRoot: externalapi.DomainHash{
			0xe7, 0xc0, 0x86, 0xcc, 0xc9, 0xc4, 0x6f, 0x1c,
			0xf2, 0x08, 0x98, 0xcc, 0x19, 0xfb, 0x45, 0x4f,
			0x87, 0xe4, 0x82, 0xa3, 0xbd, 0xc9, 0x69, 0x49,
			0xfa, 0x03, 0x3c, 0x74, 0xdc, 0xec, 0x6b, 0x44,
		},
		AcceptedIDMerkleRoot: externalapi.DomainHash{
			0x80, 0xf7, 0x00, 0xe3, 0x16, 0x3d, 0x04, 0x95,
			0x5b, 0x7e, 0xaf, 0x84, 0x7e, 0x1b, 0x6b, 0x06,
			0x4e, 0x06, 0xba, 0x64, 0xd7, 0x61, 0xda, 0x25,
			0x1a, 0x0e, 0x21, 0xd4, 0x64, 0x49, 0x02, 0xa2,
		},
		UTXOCommitment: externalapi.DomainHash{
			0x80, 0xf7, 0x00, 0xe3, 0x16, 0x3d, 0x04, 0x95,
			0x5b, 0x7e, 0xaf, 0x84, 0x7e, 0x1b, 0x6b, 0x06,
			0x4e, 0x06, 0xba, 0x64, 0xd7, 0x61, 0xda, 0x25,
			0x1a, 0x0e, 0x21, 0xd4, 0x64, 0x49, 0x02, 0xa2,
		},
		TimeInMilliseconds: 0x5cd18053000,
		Bits:               0x207fffff,
		Nonce:              0x1,
	},
	Transactions: []*externalapi.DomainTransaction{
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID{},
						Index:         0xffffffff,
					},
					SignatureScript: []byte{
						0x02, 0x10, 0x27, 0x08, 0xac, 0x29, 0x2f, 0x2f,
						0xcf, 0x70, 0xb0, 0x7e, 0x0b, 0x2f, 0x50, 0x32,
						0x53, 0x48, 0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0x12a05f200, // 5000000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x51,
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDCoinbase,
			Payload:      []byte{9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			PayloadHash: externalapi.DomainHash{
				0x5d, 0xac, 0x93, 0xd6, 0xd2, 0xa9, 0x68, 0x89,
				0x97, 0xee, 0x3b, 0x4d, 0x7e, 0x8b, 0xae, 0x3b,
				0xa7, 0x36, 0xe5, 0xad, 0xbd, 0xdc, 0xee, 0xfa,
				0xe2, 0x5c, 0x85, 0x18, 0x33, 0xe5, 0xe3, 0x6c,
			},
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID([32]byte{
							0x03, 0x2e, 0x38, 0xe9, 0xc0, 0xa8, 0x4c, 0x60,
							0x46, 0xd6, 0x87, 0xd1, 0x05, 0x56, 0xdc, 0xac,
							0xc4, 0x1d, 0x27, 0x5e, 0xc5, 0x5f, 0xc0, 0x07,
							0x79, 0xac, 0x88, 0xfd, 0xf3, 0x57, 0xa1, 0x87,
						}), // 87a157f3fd88ac7907c05fc55e271dc4acdc5605d187d646604ca8c0e9382e03
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xc3, 0x52, 0xd3,
						0xdd, 0x99, 0x3a, 0x98, 0x1b, 0xeb, 0xa4, 0xa6,
						0x3a, 0xd1, 0x5c, 0x20, 0x92, 0x75, 0xca, 0x94,
						0x70, 0xab, 0xfc, 0xd5, 0x7d, 0xa9, 0x3b, 0x58,
						0xe4, 0xeb, 0x5d, 0xce, 0x82, 0x02, 0x21, 0x00,
						0x84, 0x07, 0x92, 0xbc, 0x1f, 0x45, 0x60, 0x62,
						0x81, 0x9f, 0x15, 0xd3, 0x3e, 0xe7, 0x05, 0x5c,
						0xf7, 0xb5, 0xee, 0x1a, 0xf1, 0xeb, 0xcc, 0x60,
						0x28, 0xd9, 0xcd, 0xb1, 0xc3, 0xaf, 0x77, 0x48,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xf4, 0x6d, 0xb5, 0xe9, 0xd6, 0x1a, 0x9d,
						0xc2, 0x7b, 0x8d, 0x64, 0xad, 0x23, 0xe7, 0x38,
						0x3a, 0x4e, 0x6c, 0xa1, 0x64, 0x59, 0x3c, 0x25,
						0x27, 0xc0, 0x38, 0xc0, 0x85, 0x7e, 0xb6, 0x7e,
						0xe8, 0xe8, 0x25, 0xdc, 0xa6, 0x50, 0x46, 0xb8,
						0x2c, 0x93, 0x31, 0x58, 0x6c, 0x82, 0xe0, 0xfd,
						0x1f, 0x63, 0x3f, 0x25, 0xf8, 0x7c, 0x16, 0x1b,
						0xc6, 0xf8, 0xa6, 0x30, 0x12, 0x1d, 0xf2, 0xb3,
						0xd3, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0x2123e300, // 556000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xc3, 0x98, 0xef, 0xa9, 0xc3, 0x92, 0xba, 0x60,
						0x13, 0xc5, 0xe0, 0x4e, 0xe7, 0x29, 0x75, 0x5e,
						0xf7, 0xf5, 0x8b, 0x32,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
				{
					Value: 0x108e20f00, // 4444000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x94, 0x8c, 0x76, 0x5a, 0x69, 0x14, 0xd4, 0x3f,
						0x2a, 0x7a, 0xc1, 0x77, 0xda, 0x2c, 0x2f, 0x6b,
						0x52, 0xde, 0x3d, 0x7c,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID([32]byte{
							0xc3, 0x3e, 0xbf, 0xf2, 0xa7, 0x09, 0xf1, 0x3d,
							0x9f, 0x9a, 0x75, 0x69, 0xab, 0x16, 0xa3, 0x27,
							0x86, 0xaf, 0x7d, 0x7e, 0x2d, 0xe0, 0x92, 0x65,
							0xe4, 0x1c, 0x61, 0xd0, 0x78, 0x29, 0x4e, 0xcf,
						}), // cf4e2978d0611ce46592e02d7e7daf8627a316ab69759a9f3df109a7f2bf3ec3
						Index: 1,
					},
					SignatureScript: []byte{
						0x47, // OP_DATA_71
						0x30, 0x44, 0x02, 0x20, 0x03, 0x2d, 0x30, 0xdf,
						0x5e, 0xe6, 0xf5, 0x7f, 0xa4, 0x6c, 0xdd, 0xb5,
						0xeb, 0x8d, 0x0d, 0x9f, 0xe8, 0xde, 0x6b, 0x34,
						0x2d, 0x27, 0x94, 0x2a, 0xe9, 0x0a, 0x32, 0x31,
						0xe0, 0xba, 0x33, 0x3e, 0x02, 0x20, 0x3d, 0xee,
						0xe8, 0x06, 0x0f, 0xdc, 0x70, 0x23, 0x0a, 0x7f,
						0x5b, 0x4a, 0xd7, 0xd7, 0xbc, 0x3e, 0x62, 0x8c,
						0xbe, 0x21, 0x9a, 0x88, 0x6b, 0x84, 0x26, 0x9e,
						0xae, 0xb8, 0x1e, 0x26, 0xb4, 0xfe, 0x01,
						0x41, // OP_DATA_65
						0x04, 0xae, 0x31, 0xc3, 0x1b, 0xf9, 0x12, 0x78,
						0xd9, 0x9b, 0x83, 0x77, 0xa3, 0x5b, 0xbc, 0xe5,
						0xb2, 0x7d, 0x9f, 0xff, 0x15, 0x45, 0x68, 0x39,
						0xe9, 0x19, 0x45, 0x3f, 0xc7, 0xb3, 0xf7, 0x21,
						0xf0, 0xba, 0x40, 0x3f, 0xf9, 0x6c, 0x9d, 0xee,
						0xb6, 0x80, 0xe5, 0xfd, 0x34, 0x1c, 0x0f, 0xc3,
						0xa7, 0xb9, 0x0d, 0xa4, 0x63, 0x1e, 0xe3, 0x95,
						0x60, 0x63, 0x9d, 0xb4, 0x62, 0xe9, 0xcb, 0x85,
						0x0f, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0xf4240, // 1000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xb0, 0xdc, 0xbf, 0x97, 0xea, 0xbf, 0x44, 0x04,
						0xe3, 0x1d, 0x95, 0x24, 0x77, 0xce, 0x82, 0x2d,
						0xad, 0xbe, 0x7e, 0x10,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
				{
					Value: 0x11d260c0, // 299000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x6b, 0x12, 0x81, 0xee, 0xc2, 0x5a, 0xb4, 0xe1,
						0xe0, 0x79, 0x3f, 0xf4, 0xe0, 0x8a, 0xb1, 0xab,
						0xb3, 0x40, 0x9c, 0xd9,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID([32]byte{
							0x0b, 0x60, 0x72, 0xb3, 0x86, 0xd4, 0xa7, 0x73,
							0x23, 0x52, 0x37, 0xf6, 0x4c, 0x11, 0x26, 0xac,
							0x3b, 0x24, 0x0c, 0x84, 0xb9, 0x17, 0xa3, 0x90,
							0x9b, 0xa1, 0xc4, 0x3d, 0xed, 0x5f, 0x51, 0xf4,
						}), // f4515fed3dc4a19b90a317b9840c243bac26114cf637522373a7d486b372600b
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xbb, 0x1a, 0xd2,
						0x6d, 0xf9, 0x30, 0xa5, 0x1c, 0xce, 0x11, 0x0c,
						0xf4, 0x4f, 0x7a, 0x48, 0xc3, 0xc5, 0x61, 0xfd,
						0x97, 0x75, 0x00, 0xb1, 0xae, 0x5d, 0x6b, 0x6f,
						0xd1, 0x3d, 0x0b, 0x3f, 0x4a, 0x02, 0x21, 0x00,
						0xc5, 0xb4, 0x29, 0x51, 0xac, 0xed, 0xff, 0x14,
						0xab, 0xba, 0x27, 0x36, 0xfd, 0x57, 0x4b, 0xdb,
						0x46, 0x5f, 0x3e, 0x6f, 0x8d, 0xa1, 0x2e, 0x2c,
						0x53, 0x03, 0x95, 0x4a, 0xca, 0x7f, 0x78, 0xf3,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xa7, 0x13, 0x5b, 0xfe, 0x82, 0x4c, 0x97,
						0xec, 0xc0, 0x1e, 0xc7, 0xd7, 0xe3, 0x36, 0x18,
						0x5c, 0x81, 0xe2, 0xaa, 0x2c, 0x41, 0xab, 0x17,
						0x54, 0x07, 0xc0, 0x94, 0x84, 0xce, 0x96, 0x94,
						0xb4, 0x49, 0x53, 0xfc, 0xb7, 0x51, 0x20, 0x65,
						0x64, 0xa9, 0xc2, 0x4d, 0xd0, 0x94, 0xd4, 0x2f,
						0xdb, 0xfd, 0xd5, 0xaa, 0xd3, 0xe0, 0x63, 0xce,
						0x6a, 0xf4, 0xcf, 0xaa, 0xea, 0x4e, 0xa1, 0x4f,
						0xbb, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0xf4240, // 1000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x39, 0xaa, 0x3d, 0x56, 0x9e, 0x06, 0xa1, 0xd7,
						0x92, 0x6d, 0xc4, 0xbe, 0x11, 0x93, 0xc9, 0x9b,
						0xf2, 0xeb, 0x9e, 0xe0,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
	},
}

// exampleValidBlock defines a sample valid block
var exampleValidBlock = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version: 0x00000000,
		ParentHashes: []*externalapi.DomainHash{
			{
				0x16, 0x5e, 0x38, 0xe8, 0xb3, 0x91, 0x45, 0x95,
				0xd9, 0xc6, 0x41, 0xf3, 0xb8, 0xee, 0xc2, 0xf3,
				0x46, 0x11, 0x89, 0x6b, 0x82, 0x1a, 0x68, 0x3b,
				0x7a, 0x4e, 0xde, 0xfe, 0x2c, 0x00, 0x00, 0x00,
			},
			{
				0x4b, 0xb0, 0x75, 0x35, 0xdf, 0xd5, 0x8e, 0x0b,
				0x3c, 0xd6, 0x4f, 0xd7, 0x15, 0x52, 0x80, 0x87,
				0x2a, 0x04, 0x71, 0xbc, 0xf8, 0x30, 0x95, 0x52,
				0x6a, 0xce, 0x0e, 0x38, 0xc6, 0x00, 0x00, 0x00,
			},
		},
		HashMerkleRoot: externalapi.DomainHash{
			0x10, 0xcf, 0xf8, 0xb4, 0x14, 0x46, 0x00, 0x21,
			0xaa, 0xba, 0x4d, 0x25, 0x31, 0x0a, 0x7e, 0xb6,
			0x28, 0xc8, 0x20, 0x6d, 0x38, 0x7b, 0x70, 0x64,
			0xbf, 0x2e, 0x7c, 0x68, 0x09, 0x16, 0x41, 0x8e,
		},
		AcceptedIDMerkleRoot: externalapi.DomainHash{
			0x8a, 0xb7, 0xd6, 0x73, 0x1b, 0xe6, 0xc5, 0xd3,
			0x5d, 0x4e, 0x2c, 0xc9, 0x57, 0x88, 0x30, 0x65,
			0x81, 0xb8, 0xa0, 0x68, 0x77, 0xc4, 0x02, 0x1e,
			0x3c, 0xb1, 0x16, 0x8f, 0x5f, 0x6b, 0x45, 0x87,
		},
		UTXOCommitment:     [32]byte{},
		TimeInMilliseconds: 0x17305aa654a,
		Bits:               0x207fffff,
		Nonce:              1,
	},
	Transactions: []*externalapi.DomainTransaction{
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID{
							0x9b, 0x22, 0x59, 0x44, 0x66, 0xf0, 0xbe, 0x50,
							0x7c, 0x1c, 0x8a, 0xf6, 0x06, 0x27, 0xe6, 0x33,
							0x38, 0x7e, 0xd1, 0xd5, 0x8c, 0x42, 0x59, 0x1a,
							0x31, 0xac, 0x9a, 0xa6, 0x2e, 0xd5, 0x2b, 0x0f,
						},
						Index: 0xffffffff,
					},
					SignatureScript: nil,
					Sequence:        math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0x12a05f200, // 5000000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49,
						0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
						0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDCoinbase,
			Payload:      []byte{9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			PayloadHash: externalapi.DomainHash{
				0x5d, 0xac, 0x93, 0xd6, 0xd2, 0xa9, 0x68, 0x89,
				0x97, 0xee, 0x3b, 0x4d, 0x7e, 0x8b, 0xae, 0x3b,
				0xa7, 0x36, 0xe5, 0xad, 0xbd, 0xdc, 0xee, 0xfa,
				0xe2, 0x5c, 0x85, 0x18, 0x33, 0xe5, 0xe3, 0x6c,
			},
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID{
							0x16, 0x5e, 0x38, 0xe8, 0xb3, 0x91, 0x45, 0x95,
							0xd9, 0xc6, 0x41, 0xf3, 0xb8, 0xee, 0xc2, 0xf3,
							0x46, 0x11, 0x89, 0x6b, 0x82, 0x1a, 0x68, 0x3b,
							0x7a, 0x4e, 0xde, 0xfe, 0x2c, 0x00, 0x00, 0x00,
						},
						Index: 0xffffffff,
					},
					Sequence: math.MaxUint64,
				},
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID{
							0x4b, 0xb0, 0x75, 0x35, 0xdf, 0xd5, 0x8e, 0x0b,
							0x3c, 0xd6, 0x4f, 0xd7, 0x15, 0x52, 0x80, 0x87,
							0x2a, 0x04, 0x71, 0xbc, 0xf8, 0x30, 0x95, 0x52,
							0x6a, 0xce, 0x0e, 0x38, 0xc6, 0x00, 0x00, 0x00,
						},
						Index: 0xffffffff,
					},
					Sequence: math.MaxUint64,
				},
			},
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID([32]byte{
							0x03, 0x2e, 0x38, 0xe9, 0xc0, 0xa8, 0x4c, 0x60,
							0x46, 0xd6, 0x87, 0xd1, 0x05, 0x56, 0xdc, 0xac,
							0xc4, 0x1d, 0x27, 0x5e, 0xc5, 0x5f, 0xc0, 0x07,
							0x79, 0xac, 0x88, 0xfd, 0xf3, 0x57, 0xa1, 0x87,
						}), // 87a157f3fd88ac7907c05fc55e271dc4acdc5605d187d646604ca8c0e9382e03
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xc3, 0x52, 0xd3,
						0xdd, 0x99, 0x3a, 0x98, 0x1b, 0xeb, 0xa4, 0xa6,
						0x3a, 0xd1, 0x5c, 0x20, 0x92, 0x75, 0xca, 0x94,
						0x70, 0xab, 0xfc, 0xd5, 0x7d, 0xa9, 0x3b, 0x58,
						0xe4, 0xeb, 0x5d, 0xce, 0x82, 0x02, 0x21, 0x00,
						0x84, 0x07, 0x92, 0xbc, 0x1f, 0x45, 0x60, 0x62,
						0x81, 0x9f, 0x15, 0xd3, 0x3e, 0xe7, 0x05, 0x5c,
						0xf7, 0xb5, 0xee, 0x1a, 0xf1, 0xeb, 0xcc, 0x60,
						0x28, 0xd9, 0xcd, 0xb1, 0xc3, 0xaf, 0x77, 0x48,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xf4, 0x6d, 0xb5, 0xe9, 0xd6, 0x1a, 0x9d,
						0xc2, 0x7b, 0x8d, 0x64, 0xad, 0x23, 0xe7, 0x38,
						0x3a, 0x4e, 0x6c, 0xa1, 0x64, 0x59, 0x3c, 0x25,
						0x27, 0xc0, 0x38, 0xc0, 0x85, 0x7e, 0xb6, 0x7e,
						0xe8, 0xe8, 0x25, 0xdc, 0xa6, 0x50, 0x46, 0xb8,
						0x2c, 0x93, 0x31, 0x58, 0x6c, 0x82, 0xe0, 0xfd,
						0x1f, 0x63, 0x3f, 0x25, 0xf8, 0x7c, 0x16, 0x1b,
						0xc6, 0xf8, 0xa6, 0x30, 0x12, 0x1d, 0xf2, 0xb3,
						0xd3, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0x2123e300, // 556000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xc3, 0x98, 0xef, 0xa9, 0xc3, 0x92, 0xba, 0x60,
						0x13, 0xc5, 0xe0, 0x4e, 0xe7, 0x29, 0x75, 0x5e,
						0xf7, 0xf5, 0x8b, 0x32,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
				{
					Value: 0x108e20f00, // 4444000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x94, 0x8c, 0x76, 0x5a, 0x69, 0x14, 0xd4, 0x3f,
						0x2a, 0x7a, 0xc1, 0x77, 0xda, 0x2c, 0x2f, 0x6b,
						0x52, 0xde, 0x3d, 0x7c,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID([32]byte{
							0xc3, 0x3e, 0xbf, 0xf2, 0xa7, 0x09, 0xf1, 0x3d,
							0x9f, 0x9a, 0x75, 0x69, 0xab, 0x16, 0xa3, 0x27,
							0x86, 0xaf, 0x7d, 0x7e, 0x2d, 0xe0, 0x92, 0x65,
							0xe4, 0x1c, 0x61, 0xd0, 0x78, 0x29, 0x4e, 0xcf,
						}), // cf4e2978d0611ce46592e02d7e7daf8627a316ab69759a9f3df109a7f2bf3ec3
						Index: 1,
					},
					SignatureScript: []byte{
						0x47, // OP_DATA_71
						0x30, 0x44, 0x02, 0x20, 0x03, 0x2d, 0x30, 0xdf,
						0x5e, 0xe6, 0xf5, 0x7f, 0xa4, 0x6c, 0xdd, 0xb5,
						0xeb, 0x8d, 0x0d, 0x9f, 0xe8, 0xde, 0x6b, 0x34,
						0x2d, 0x27, 0x94, 0x2a, 0xe9, 0x0a, 0x32, 0x31,
						0xe0, 0xba, 0x33, 0x3e, 0x02, 0x20, 0x3d, 0xee,
						0xe8, 0x06, 0x0f, 0xdc, 0x70, 0x23, 0x0a, 0x7f,
						0x5b, 0x4a, 0xd7, 0xd7, 0xbc, 0x3e, 0x62, 0x8c,
						0xbe, 0x21, 0x9a, 0x88, 0x6b, 0x84, 0x26, 0x9e,
						0xae, 0xb8, 0x1e, 0x26, 0xb4, 0xfe, 0x01,
						0x41, // OP_DATA_65
						0x04, 0xae, 0x31, 0xc3, 0x1b, 0xf9, 0x12, 0x78,
						0xd9, 0x9b, 0x83, 0x77, 0xa3, 0x5b, 0xbc, 0xe5,
						0xb2, 0x7d, 0x9f, 0xff, 0x15, 0x45, 0x68, 0x39,
						0xe9, 0x19, 0x45, 0x3f, 0xc7, 0xb3, 0xf7, 0x21,
						0xf0, 0xba, 0x40, 0x3f, 0xf9, 0x6c, 0x9d, 0xee,
						0xb6, 0x80, 0xe5, 0xfd, 0x34, 0x1c, 0x0f, 0xc3,
						0xa7, 0xb9, 0x0d, 0xa4, 0x63, 0x1e, 0xe3, 0x95,
						0x60, 0x63, 0x9d, 0xb4, 0x62, 0xe9, 0xcb, 0x85,
						0x0f, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0xf4240, // 1000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xb0, 0xdc, 0xbf, 0x97, 0xea, 0xbf, 0x44, 0x04,
						0xe3, 0x1d, 0x95, 0x24, 0x77, 0xce, 0x82, 0x2d,
						0xad, 0xbe, 0x7e, 0x10,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
				{
					Value: 0x11d260c0, // 299000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x6b, 0x12, 0x81, 0xee, 0xc2, 0x5a, 0xb4, 0xe1,
						0xe0, 0x79, 0x3f, 0xf4, 0xe0, 0x8a, 0xb1, 0xab,
						0xb3, 0x40, 0x9c, 0xd9,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID{
							0x0b, 0x60, 0x72, 0xb3, 0x86, 0xd4, 0xa7, 0x73,
							0x23, 0x52, 0x37, 0xf6, 0x4c, 0x11, 0x26, 0xac,
							0x3b, 0x24, 0x0c, 0x84, 0xb9, 0x17, 0xa3, 0x90,
							0x9b, 0xa1, 0xc4, 0x3d, 0xed, 0x5f, 0x51, 0xf4,
						}, // f4515fed3dc4a19b90a317b9840c243bac26114cf637522373a7d486b372600b
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xbb, 0x1a, 0xd2,
						0x6d, 0xf9, 0x30, 0xa5, 0x1c, 0xce, 0x11, 0x0c,
						0xf4, 0x4f, 0x7a, 0x48, 0xc3, 0xc5, 0x61, 0xfd,
						0x97, 0x75, 0x00, 0xb1, 0xae, 0x5d, 0x6b, 0x6f,
						0xd1, 0x3d, 0x0b, 0x3f, 0x4a, 0x02, 0x21, 0x00,
						0xc5, 0xb4, 0x29, 0x51, 0xac, 0xed, 0xff, 0x14,
						0xab, 0xba, 0x27, 0x36, 0xfd, 0x57, 0x4b, 0xdb,
						0x46, 0x5f, 0x3e, 0x6f, 0x8d, 0xa1, 0x2e, 0x2c,
						0x53, 0x03, 0x95, 0x4a, 0xca, 0x7f, 0x78, 0xf3,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xa7, 0x13, 0x5b, 0xfe, 0x82, 0x4c, 0x97,
						0xec, 0xc0, 0x1e, 0xc7, 0xd7, 0xe3, 0x36, 0x18,
						0x5c, 0x81, 0xe2, 0xaa, 0x2c, 0x41, 0xab, 0x17,
						0x54, 0x07, 0xc0, 0x94, 0x84, 0xce, 0x96, 0x94,
						0xb4, 0x49, 0x53, 0xfc, 0xb7, 0x51, 0x20, 0x65,
						0x64, 0xa9, 0xc2, 0x4d, 0xd0, 0x94, 0xd4, 0x2f,
						0xdb, 0xfd, 0xd5, 0xaa, 0xd3, 0xe0, 0x63, 0xce,
						0x6a, 0xf4, 0xcf, 0xaa, 0xea, 0x4e, 0xa1, 0x4f,
						0xbb, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0xf4240, // 1000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x39, 0xaa, 0x3d, 0x56, 0x9e, 0x06, 0xa1, 0xd7,
						0x92, 0x6d, 0xc4, 0xbe, 0x11, 0x93, 0xc9, 0x9b,
						0xf2, 0xeb, 0x9e, 0xe0,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
	},
}

// blockWithWrongTxOrder defines invalid block 100,000 of the block DAG.
var blockWithWrongTxOrder = externalapi.DomainBlock{
	Header: &externalapi.DomainBlockHeader{
		Version: 0,
		ParentHashes: []*externalapi.DomainHash{
			{
				0x16, 0x5e, 0x38, 0xe8, 0xb3, 0x91, 0x45, 0x95,
				0xd9, 0xc6, 0x41, 0xf3, 0xb8, 0xee, 0xc2, 0xf3,
				0x46, 0x11, 0x89, 0x6b, 0x82, 0x1a, 0x68, 0x3b,
				0x7a, 0x4e, 0xde, 0xfe, 0x2c, 0x00, 0x00, 0x00,
			},
			{
				0x4b, 0xb0, 0x75, 0x35, 0xdf, 0xd5, 0x8e, 0x0b,
				0x3c, 0xd6, 0x4f, 0xd7, 0x15, 0x52, 0x80, 0x87,
				0x2a, 0x04, 0x71, 0xbc, 0xf8, 0x30, 0x95, 0x52,
				0x6a, 0xce, 0x0e, 0x38, 0xc6, 0x00, 0x00, 0x00,
			},
		},
		HashMerkleRoot: externalapi.DomainHash{
			0xac, 0xa4, 0x21, 0xe1, 0xa6, 0xc3, 0xbe, 0x5d,
			0x52, 0x66, 0xf3, 0x0b, 0x21, 0x87, 0xbc, 0xf3,
			0xf3, 0x2d, 0xd1, 0x05, 0x64, 0xb5, 0x16, 0x76,
			0xe4, 0x66, 0x7d, 0x51, 0x53, 0x18, 0x6d, 0xb1,
		},
		AcceptedIDMerkleRoot: externalapi.DomainHash{
			0xa0, 0x69, 0x2d, 0x16, 0xb5, 0xd7, 0xe4, 0xf3,
			0xcd, 0xc7, 0xc9, 0xaf, 0xfb, 0xd2, 0x1b, 0x85,
			0x0b, 0x79, 0xf5, 0x29, 0x6d, 0x1c, 0xaa, 0x90,
			0x2f, 0x01, 0xd4, 0x83, 0x9b, 0x2a, 0x04, 0x5e,
		},
		UTXOCommitment: externalapi.DomainHash{
			0x00, 0x69, 0x2d, 0x16, 0xb5, 0xd7, 0xe4, 0xf3,
			0xcd, 0xc7, 0xc9, 0xaf, 0xfb, 0xd2, 0x1b, 0x85,
			0x0b, 0x79, 0xf5, 0x29, 0x6d, 0x1c, 0xaa, 0x90,
			0x2f, 0x01, 0xd4, 0x83, 0x9b, 0x2a, 0x04, 0x5e,
		},
		TimeInMilliseconds: 0x5cd16eaa000,
		Bits:               0x207fffff,
		Nonce:              1,
	},
	Transactions: []*externalapi.DomainTransaction{
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID{
							0x9b, 0x22, 0x59, 0x44, 0x66, 0xf0, 0xbe, 0x50,
							0x7c, 0x1c, 0x8a, 0xf6, 0x06, 0x27, 0xe6, 0x33,
							0x38, 0x7e, 0xd1, 0xd5, 0x8c, 0x42, 0x59, 0x1a,
							0x31, 0xac, 0x9a, 0xa6, 0x2e, 0xd5, 0x2b, 0x0f,
						},
						Index: 0xffffffff,
					},
					SignatureScript: nil,
					Sequence:        math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0x12a05f200, // 5000000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0xa9, 0x14, 0xda, 0x17, 0x45, 0xe9, 0xb5, 0x49,
						0xbd, 0x0b, 0xfa, 0x1a, 0x56, 0x99, 0x71, 0xc7,
						0x7e, 0xba, 0x30, 0xcd, 0x5a, 0x4b, 0x87,
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDCoinbase,
			Payload:      []byte{9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			PayloadHash: externalapi.DomainHash{
				0x5d, 0xac, 0x93, 0xd6, 0xd2, 0xa9, 0x68, 0x89,
				0x97, 0xee, 0x3b, 0x4d, 0x7e, 0x8b, 0xae, 0x3b,
				0xa7, 0x36, 0xe5, 0xad, 0xbd, 0xdc, 0xee, 0xfa,
				0xe2, 0x5c, 0x85, 0x18, 0x33, 0xe5, 0xe3, 0x6c,
			},
		},

		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID{
							0x16, 0x5e, 0x38, 0xe8, 0xb3, 0x91, 0x45, 0x95,
							0xd9, 0xc6, 0x41, 0xf3, 0xb8, 0xee, 0xc2, 0xf3,
							0x46, 0x11, 0x89, 0x6b, 0x82, 0x1a, 0x68, 0x3b,
							0x7a, 0x4e, 0xde, 0xfe, 0x2c, 0x00, 0x00, 0x00,
						},
						Index: 0xffffffff,
					},
					Sequence: math.MaxUint64,
				},
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID{
							0x4b, 0xb0, 0x75, 0x35, 0xdf, 0xd5, 0x8e, 0x0b,
							0x3c, 0xd6, 0x4f, 0xd7, 0x15, 0x52, 0x80, 0x87,
							0x2a, 0x04, 0x71, 0xbc, 0xf8, 0x30, 0x95, 0x52,
							0x6a, 0xce, 0x0e, 0x38, 0xc6, 0x00, 0x00, 0x00,
						},
						Index: 0xffffffff,
					},
					Sequence: math.MaxUint64,
				},
			},
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID([32]byte{
							0x03, 0x2e, 0x38, 0xe9, 0xc0, 0xa8, 0x4c, 0x60,
							0x46, 0xd6, 0x87, 0xd1, 0x05, 0x56, 0xdc, 0xac,
							0xc4, 0x1d, 0x27, 0x5e, 0xc5, 0x5f, 0xc0, 0x07,
							0x79, 0xac, 0x88, 0xfd, 0xf3, 0x57, 0xa1, 0x87,
						}), // 87a157f3fd88ac7907c05fc55e271dc4acdc5605d187d646604ca8c0e9382e03
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xc3, 0x52, 0xd3,
						0xdd, 0x99, 0x3a, 0x98, 0x1b, 0xeb, 0xa4, 0xa6,
						0x3a, 0xd1, 0x5c, 0x20, 0x92, 0x75, 0xca, 0x94,
						0x70, 0xab, 0xfc, 0xd5, 0x7d, 0xa9, 0x3b, 0x58,
						0xe4, 0xeb, 0x5d, 0xce, 0x82, 0x02, 0x21, 0x00,
						0x84, 0x07, 0x92, 0xbc, 0x1f, 0x45, 0x60, 0x62,
						0x81, 0x9f, 0x15, 0xd3, 0x3e, 0xe7, 0x05, 0x5c,
						0xf7, 0xb5, 0xee, 0x1a, 0xf1, 0xeb, 0xcc, 0x60,
						0x28, 0xd9, 0xcd, 0xb1, 0xc3, 0xaf, 0x77, 0x48,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xf4, 0x6d, 0xb5, 0xe9, 0xd6, 0x1a, 0x9d,
						0xc2, 0x7b, 0x8d, 0x64, 0xad, 0x23, 0xe7, 0x38,
						0x3a, 0x4e, 0x6c, 0xa1, 0x64, 0x59, 0x3c, 0x25,
						0x27, 0xc0, 0x38, 0xc0, 0x85, 0x7e, 0xb6, 0x7e,
						0xe8, 0xe8, 0x25, 0xdc, 0xa6, 0x50, 0x46, 0xb8,
						0x2c, 0x93, 0x31, 0x58, 0x6c, 0x82, 0xe0, 0xfd,
						0x1f, 0x63, 0x3f, 0x25, 0xf8, 0x7c, 0x16, 0x1b,
						0xc6, 0xf8, 0xa6, 0x30, 0x12, 0x1d, 0xf2, 0xb3,
						0xd3, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0x2123e300, // 556000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xc3, 0x98, 0xef, 0xa9, 0xc3, 0x92, 0xba, 0x60,
						0x13, 0xc5, 0xe0, 0x4e, 0xe7, 0x29, 0x75, 0x5e,
						0xf7, 0xf5, 0x8b, 0x32,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
				{
					Value: 0x108e20f00, // 4444000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x94, 0x8c, 0x76, 0x5a, 0x69, 0x14, 0xd4, 0x3f,
						0x2a, 0x7a, 0xc1, 0x77, 0xda, 0x2c, 0x2f, 0x6b,
						0x52, 0xde, 0x3d, 0x7c,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: externalapi.DomainSubnetworkID{11},
			Payload:      []byte{},
			PayloadHash:  [32]byte{0xFF, 0xFF},
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID([32]byte{
							0xc3, 0x3e, 0xbf, 0xf2, 0xa7, 0x09, 0xf1, 0x3d,
							0x9f, 0x9a, 0x75, 0x69, 0xab, 0x16, 0xa3, 0x27,
							0x86, 0xaf, 0x7d, 0x7e, 0x2d, 0xe0, 0x92, 0x65,
							0xe4, 0x1c, 0x61, 0xd0, 0x78, 0x29, 0x4e, 0xcf,
						}), // cf4e2978d0611ce46592e02d7e7daf8627a316ab69759a9f3df109a7f2bf3ec3
						Index: 1,
					},
					SignatureScript: []byte{
						0x47, // OP_DATA_71
						0x30, 0x44, 0x02, 0x20, 0x03, 0x2d, 0x30, 0xdf,
						0x5e, 0xe6, 0xf5, 0x7f, 0xa4, 0x6c, 0xdd, 0xb5,
						0xeb, 0x8d, 0x0d, 0x9f, 0xe8, 0xde, 0x6b, 0x34,
						0x2d, 0x27, 0x94, 0x2a, 0xe9, 0x0a, 0x32, 0x31,
						0xe0, 0xba, 0x33, 0x3e, 0x02, 0x20, 0x3d, 0xee,
						0xe8, 0x06, 0x0f, 0xdc, 0x70, 0x23, 0x0a, 0x7f,
						0x5b, 0x4a, 0xd7, 0xd7, 0xbc, 0x3e, 0x62, 0x8c,
						0xbe, 0x21, 0x9a, 0x88, 0x6b, 0x84, 0x26, 0x9e,
						0xae, 0xb8, 0x1e, 0x26, 0xb4, 0xfe, 0x01,
						0x41, // OP_DATA_65
						0x04, 0xae, 0x31, 0xc3, 0x1b, 0xf9, 0x12, 0x78,
						0xd9, 0x9b, 0x83, 0x77, 0xa3, 0x5b, 0xbc, 0xe5,
						0xb2, 0x7d, 0x9f, 0xff, 0x15, 0x45, 0x68, 0x39,
						0xe9, 0x19, 0x45, 0x3f, 0xc7, 0xb3, 0xf7, 0x21,
						0xf0, 0xba, 0x40, 0x3f, 0xf9, 0x6c, 0x9d, 0xee,
						0xb6, 0x80, 0xe5, 0xfd, 0x34, 0x1c, 0x0f, 0xc3,
						0xa7, 0xb9, 0x0d, 0xa4, 0x63, 0x1e, 0xe3, 0x95,
						0x60, 0x63, 0x9d, 0xb4, 0x62, 0xe9, 0xcb, 0x85,
						0x0f, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0xf4240, // 1000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xb0, 0xdc, 0xbf, 0x97, 0xea, 0xbf, 0x44, 0x04,
						0xe3, 0x1d, 0x95, 0x24, 0x77, 0xce, 0x82, 0x2d,
						0xad, 0xbe, 0x7e, 0x10,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
				{
					Value: 0x11d260c0, // 299000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x6b, 0x12, 0x81, 0xee, 0xc2, 0x5a, 0xb4, 0xe1,
						0xe0, 0x79, 0x3f, 0xf4, 0xe0, 0x8a, 0xb1, 0xab,
						0xb3, 0x40, 0x9c, 0xd9,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
		{
			Version: 0,
			Inputs: []*externalapi.DomainTransactionInput{
				{
					PreviousOutpoint: externalapi.DomainOutpoint{
						TransactionID: externalapi.DomainTransactionID([32]byte{
							0x0b, 0x60, 0x72, 0xb3, 0x86, 0xd4, 0xa7, 0x73,
							0x23, 0x52, 0x37, 0xf6, 0x4c, 0x11, 0x26, 0xac,
							0x3b, 0x24, 0x0c, 0x84, 0xb9, 0x17, 0xa3, 0x90,
							0x9b, 0xa1, 0xc4, 0x3d, 0xed, 0x5f, 0x51, 0xf4,
						}), // f4515fed3dc4a19b90a317b9840c243bac26114cf637522373a7d486b372600b
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xbb, 0x1a, 0xd2,
						0x6d, 0xf9, 0x30, 0xa5, 0x1c, 0xce, 0x11, 0x0c,
						0xf4, 0x4f, 0x7a, 0x48, 0xc3, 0xc5, 0x61, 0xfd,
						0x97, 0x75, 0x00, 0xb1, 0xae, 0x5d, 0x6b, 0x6f,
						0xd1, 0x3d, 0x0b, 0x3f, 0x4a, 0x02, 0x21, 0x00,
						0xc5, 0xb4, 0x29, 0x51, 0xac, 0xed, 0xff, 0x14,
						0xab, 0xba, 0x27, 0x36, 0xfd, 0x57, 0x4b, 0xdb,
						0x46, 0x5f, 0x3e, 0x6f, 0x8d, 0xa1, 0x2e, 0x2c,
						0x53, 0x03, 0x95, 0x4a, 0xca, 0x7f, 0x78, 0xf3,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xa7, 0x13, 0x5b, 0xfe, 0x82, 0x4c, 0x97,
						0xec, 0xc0, 0x1e, 0xc7, 0xd7, 0xe3, 0x36, 0x18,
						0x5c, 0x81, 0xe2, 0xaa, 0x2c, 0x41, 0xab, 0x17,
						0x54, 0x07, 0xc0, 0x94, 0x84, 0xce, 0x96, 0x94,
						0xb4, 0x49, 0x53, 0xfc, 0xb7, 0x51, 0x20, 0x65,
						0x64, 0xa9, 0xc2, 0x4d, 0xd0, 0x94, 0xd4, 0x2f,
						0xdb, 0xfd, 0xd5, 0xaa, 0xd3, 0xe0, 0x63, 0xce,
						0x6a, 0xf4, 0xcf, 0xaa, 0xea, 0x4e, 0xa1, 0x4f,
						0xbb, // 65-byte pubkey
					},
					Sequence: math.MaxUint64,
				},
			},
			Outputs: []*externalapi.DomainTransactionOutput{
				{
					Value: 0xf4240, // 1000000
					ScriptPublicKey: &externalapi.ScriptPublicKey{[]byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x39, 0xaa, 0x3d, 0x56, 0x9e, 0x06, 0xa1, 0xd7,
						0x92, 0x6d, 0xc4, 0xbe, 0x11, 0x93, 0xc9, 0x9b,
						0xf2, 0xeb, 0x9e, 0xe0,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					}, version},
				},
			},
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
		},
	},
}
