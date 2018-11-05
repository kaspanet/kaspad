package indexers

import (
	"bytes"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

func TestTxIndexConnectBlock(t *testing.T) {
	blocks := make(map[daghash.Hash]*util.Block)
	processBlock := func(t *testing.T, dag *blockdag.BlockDAG, msgBlock *wire.MsgBlock, blockName string) {
		block := util.NewBlock(msgBlock)
		blocks[*block.Hash()] = block
		isOrphan, err := dag.ProcessBlock(block, blockdag.BFNone)
		if err != nil {
			t.Fatalf("TestTxIndexConnectBlock: dag.ProcessBlock got unexpected error for block %v: %v", blockName, err)
		}
		if isOrphan {
			t.Fatalf("TestTxIndexConnectBlock: block %v was unexpectedly orphan", blockName)
		}
	}

	txIndex := NewTxIndex()
	indexManager := NewManager([]Indexer{txIndex})

	params := dagconfig.SimNetParams
	params.CoinbaseMaturity = 1
	params.K = 1

	config := blockdag.Config{
		IndexManager: indexManager,
		DAGParams:    &params,
	}

	dag, teardown, err := blockdag.DAGSetup("TestTxIndexConnectBlock", config)
	if err != nil {
		t.Fatalf("TestTxIndexConnectBlock: Failed to setup DAG instance: %v", err)
	}
	if teardown != nil {
		defer teardown()
	}

	processBlock(t, dag, &block1, "1")
	processBlock(t, dag, &block2, "2")
	processBlock(t, dag, &block3, "3")

	block3TxHash := block3Tx.TxHash()
	block3TxNewAcceptedBlock, err := txIndex.BlockThatAcceptedTx(dag, &block3TxHash)
	if err != nil {
		t.Errorf("TestTxIndexConnectBlock: TxAcceptedInBlock: %v", err)
	}
	block3Hash := block3.Header.BlockHash()
	if !block3TxNewAcceptedBlock.IsEqual(&block3Hash) {
		t.Errorf("TestTxIndexConnectBlock: block3Tx should've "+
			"been accepted in block %v but instead got accepted in block %v", block3Hash, block3TxNewAcceptedBlock)
	}

	processBlock(t, dag, &block3A, "3A")
	processBlock(t, dag, &block4, "4")
	processBlock(t, dag, &block5, "5")

	block3TxAcceptedBlock, err := txIndex.BlockThatAcceptedTx(dag, &block3TxHash)
	if err != nil {
		t.Errorf("TestTxIndexConnectBlock: TxAcceptedInBlock: %v", err)
	}
	block3AHash := block3A.Header.BlockHash()
	if !block3TxAcceptedBlock.IsEqual(&block3AHash) {
		t.Errorf("TestTxIndexConnectBlock: block3Tx should've "+
			"been accepted in block %v but instead got accepted in block %v", block3AHash, block3TxAcceptedBlock)
	}

	region, err := txIndex.TxFirstBlockRegion(&block3TxHash)
	if err != nil {
		t.Fatalf("TestTxIndexConnectBlock: no block region was found for block3Tx")
	}
	regionBlock, ok := blocks[*region.Hash]
	if !ok {
		t.Fatalf("TestTxIndexConnectBlock: couldn't find block with hash %v", region.Hash)
	}

	regionBlockBytes, err := regionBlock.Bytes()
	if err != nil {
		t.Fatalf("TestTxIndexConnectBlock: Couldn't serialize block to bytes")
	}
	block3TxInBlock := regionBlockBytes[region.Offset : region.Offset+region.Len]

	block3TxBuf := bytes.NewBuffer(make([]byte, 0, block3Tx.SerializeSize()))
	block3Tx.BtcEncode(block3TxBuf, 0)
	blockTxBytes := block3TxBuf.Bytes()

	if !reflect.DeepEqual(blockTxBytes, block3TxInBlock) {
		t.Errorf("TestTxIndexConnectBlock: the block region that was in the bucket doesn't match block3Tx")
	}

}

var block1 = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 1,
		PrevBlocks: []daghash.Hash{
			[32]byte{ // Make go vet happy.
				0x4a, 0xc1, 0x82, 0x2e, 0x43, 0x05, 0xea, 0x0c,
				0x4f, 0xcc, 0x77, 0x87, 0xae, 0x26, 0x48, 0x87,
				0x50, 0x13, 0xee, 0x2f, 0x55, 0xa7, 0x18, 0xa7,
				0x1e, 0xf2, 0xd8, 0x7c, 0xc1, 0x13, 0xac, 0x22,
			},
		},
		MerkleRoot: daghash.Hash([32]byte{ // Make go vet happy.
			0xec, 0x37, 0x81, 0x75, 0x51, 0x79, 0x41, 0x34,
			0x3a, 0xae, 0x05, 0x48, 0x67, 0xfa, 0xdf, 0x84,
			0xef, 0x06, 0x5b, 0x93, 0x07, 0xa8, 0xc2, 0xb7,
			0x2a, 0x94, 0x07, 0x3b, 0x5f, 0xee, 0xb8, 0x6a,
		}),
		Timestamp: time.Unix(0x5bd58c4a, 0),
		Bits:      0x207fffff,
		Nonce:     0xdffffffffffffffa,
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  daghash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x51, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
						0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
					},
					Sequence: math.MaxUint64,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 5000000000,
					PkScript: []byte{
						0x76, 0xa9, 0x14, 0x3d, 0xee, 0x47, 0x71, 0x6e,
						0x3c, 0xfa, 0x57, 0xdf, 0x45, 0x11, 0x34, 0x73,
						0xa6, 0x31, 0x2e, 0xbe, 0xae, 0xf3, 0x11, 0x88,
						0xac,
					},
				},
			},
			LockTime: 0,
		},
	},
}

var block2 = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 1,
		PrevBlocks: []daghash.Hash{
			[32]byte{ // Make go vet happy.
				0x42, 0xb9, 0x2c, 0xee, 0x3e, 0x3e, 0x35, 0x02,
				0xf5, 0x8d, 0xd2, 0xc8, 0xff, 0x61, 0xe3, 0x44,
				0x59, 0xb2, 0x5d, 0x72, 0x10, 0x29, 0x62, 0x58,
				0x3f, 0xc9, 0x41, 0xe2, 0xcd, 0xa9, 0x05, 0x11,
			},
		},
		MerkleRoot: daghash.Hash([32]byte{ // Make go vet happy.
			0x3e, 0x89, 0x5f, 0xb4, 0xa8, 0x2f, 0x64, 0xb9,
			0xe7, 0x1d, 0x5d, 0xce, 0x41, 0x4a, 0xb0, 0x36,
			0x4e, 0xd0, 0x4b, 0xfc, 0x0c, 0xe1, 0x82, 0xfc,
			0x51, 0x0d, 0x03, 0x7b, 0x8c, 0xdd, 0x3e, 0x49,
		}),
		Timestamp: time.Unix(0x5bd58c4b, 0),
		Bits:      0x207fffff,
		Nonce:     0x9ffffffffffffffb,
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  daghash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x52, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
						0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
					},
					Sequence: math.MaxUint64,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 5000000000,
					PkScript: []byte{
						0x76, 0xa9, 0x14, 0x3d, 0xee, 0x47, 0x71, 0x6e,
						0x3c, 0xfa, 0x57, 0xdf, 0x45, 0x11, 0x34, 0x73,
						0xa6, 0x31, 0x2e, 0xbe, 0xae, 0xf3, 0x11, 0x88,
						0xac,
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: daghash.Hash{
							0xec, 0x37, 0x81, 0x75, 0x51, 0x79, 0x41, 0x34,
							0x3a, 0xae, 0x05, 0x48, 0x67, 0xfa, 0xdf, 0x84,
							0xef, 0x06, 0x5b, 0x93, 0x07, 0xa8, 0xc2, 0xb7,
							0x2a, 0x94, 0x07, 0x3b, 0x5f, 0xee, 0xb8, 0x6a,
						},
						Index: 0,
					},
					SignatureScript: []byte{
						0x47, 0x30, 0x44, 0x02, 0x20, 0x5d, 0xca, 0x41,
						0xb0, 0x73, 0x9e, 0xba, 0x0c, 0xba, 0x59, 0xdd,
						0xb5, 0x6a, 0x6e, 0xd2, 0xd2, 0x36, 0x61, 0xa5,
						0xa0, 0x5c, 0xb5, 0x2b, 0xee, 0x5f, 0x30, 0x62,
						0x72, 0xb3, 0x26, 0xa2, 0xdb, 0x02, 0x20, 0x0d,
						0xc5, 0x22, 0xd8, 0x88, 0x5a, 0xf7, 0xef, 0x60,
						0xa6, 0xd9, 0x5c, 0x7a, 0x44, 0x96, 0xfc, 0x14,
						0x66, 0x74, 0xda, 0x2b, 0x6c, 0x99, 0x2c, 0x56,
						0x34, 0x3d, 0x64, 0xdf, 0xc2, 0x36, 0xe8, 0x01,
						0x21, 0x02, 0xa6, 0x73, 0x63, 0x8c, 0xb9, 0x58,
						0x7c, 0xb6, 0x8e, 0xa0, 0x8d, 0xbe, 0xf6, 0x85,
						0xc6, 0xf2, 0xd2, 0xa7, 0x51, 0xa8, 0xb3, 0xc6,
						0xf2, 0xa7, 0xe9, 0xa4, 0x99, 0x9e, 0x6e, 0x4b,
						0xfa, 0xf5,
					},
					Sequence: math.MaxUint64,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 5000000000,
					PkScript: []byte{
						0x76, 0xa9, 0x14, 0x3d, 0xee, 0x47, 0x71, 0x6e,
						0x3c, 0xfa, 0x57, 0xdf, 0x45, 0x11, 0x34, 0x73,
						0xa6, 0x31, 0x2e, 0xbe, 0xae, 0xf3, 0x11, 0x88,
						0xac,
					},
				},
			},
			LockTime: 0,
		},
	},
}

var block3Tx *wire.MsgTx = &wire.MsgTx{
	Version: 1,
	TxIn: []*wire.TxIn{
		{
			PreviousOutPoint: wire.OutPoint{
				Hash: daghash.Hash{
					0x69, 0x11, 0xbd, 0x7e, 0x46, 0x5e, 0xe8, 0xf7,
					0xbe, 0x80, 0xb0, 0x21, 0x6a, 0xc8, 0xb4, 0xea,
					0xef, 0xfa, 0x6a, 0x34, 0x75, 0x6e, 0xb5, 0x96,
					0xd9, 0x3b, 0xe2, 0x6a, 0xd6, 0x49, 0xac, 0x6e,
				},
				Index: 0,
			},
			SignatureScript: []byte{
				0x48, 0x30, 0x45, 0x02, 0x21, 0x00, 0xea, 0xa8,
				0xa5, 0x8b, 0x2d, 0xeb, 0x15, 0xc1, 0x18, 0x79,
				0xa4, 0xad, 0xc3, 0xde, 0x57, 0x09, 0xac, 0xdb,
				0x16, 0x16, 0x9f, 0x07, 0xe8, 0x7d, 0xbe, 0xf1,
				0x4b, 0xaa, 0xd3, 0x76, 0xb4, 0x87, 0x02, 0x20,
				0x03, 0xb3, 0xee, 0xc8, 0x9f, 0x87, 0x18, 0xee,
				0xf3, 0xc3, 0x29, 0x29, 0x57, 0xb9, 0x93, 0x95,
				0x4a, 0xe9, 0x49, 0x74, 0x90, 0xa1, 0x5b, 0xae,
				0x49, 0x16, 0xa9, 0x3e, 0xb8, 0xf0, 0xf9, 0x6b,
				0x01, 0x21, 0x02, 0xa6, 0x73, 0x63, 0x8c, 0xb9,
				0x58, 0x7c, 0xb6, 0x8e, 0xa0, 0x8d, 0xbe, 0xf6,
				0x85, 0xc6, 0xf2, 0xd2, 0xa7, 0x51, 0xa8, 0xb3,
				0xc6, 0xf2, 0xa7, 0xe9, 0xa4, 0x99, 0x9e, 0x6e,
				0x4b, 0xfa, 0xf5,
			},
			Sequence: math.MaxUint64,
		},
	},
	TxOut: []*wire.TxOut{
		{
			Value: 5000000000,
			PkScript: []byte{
				0x76, 0xa9, 0x14, 0x3d, 0xee, 0x47, 0x71, 0x6e,
				0x3c, 0xfa, 0x57, 0xdf, 0x45, 0x11, 0x34, 0x73,
				0xa6, 0x31, 0x2e, 0xbe, 0xae, 0xf3, 0x11, 0x88,
				0xac,
			},
		},
	},
	LockTime: 0,
}

var block3 = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 1,
		PrevBlocks: []daghash.Hash{
			[32]byte{ // Make go vet happy.
				0x8b, 0xdf, 0xd1, 0x48, 0xef, 0xf5, 0x2b, 0x5e,
				0xfe, 0x26, 0xba, 0x37, 0xcb, 0x23, 0x0d, 0x41,
				0x24, 0x80, 0xfe, 0x9a, 0x38, 0x90, 0xb9, 0xd3,
				0x07, 0x30, 0xcc, 0xa0, 0x4f, 0x4e, 0xf1, 0x02,
			},
		},
		MerkleRoot: daghash.Hash([32]byte{ // Make go vet happy.
			0x24, 0x0f, 0x21, 0x89, 0x94, 0xd1, 0x77, 0x32,
			0xff, 0x5d, 0xb4, 0xe9, 0x11, 0xd2, 0x74, 0xc9,
			0x0f, 0x0c, 0xb7, 0xe5, 0x16, 0xf6, 0xca, 0x63,
			0xac, 0xaa, 0x6c, 0x23, 0x42, 0xe9, 0xd5, 0x58,
		}),
		Timestamp: time.Unix(0x5bd58c4c, 0),
		Bits:      0x207fffff,
		Nonce:     0x7ffffffffffffffc,
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  daghash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x53, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
						0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
					},
					Sequence: math.MaxUint64,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 5000000000,
					PkScript: []byte{
						0x76, 0xa9, 0x14, 0x3d, 0xee, 0x47, 0x71, 0x6e,
						0x3c, 0xfa, 0x57, 0xdf, 0x45, 0x11, 0x34, 0x73,
						0xa6, 0x31, 0x2e, 0xbe, 0xae, 0xf3, 0x11, 0x88,
						0xac,
					},
				},
			},
			LockTime: 0,
		},
		block3Tx,
	},
}

var block3A = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 1,
		PrevBlocks: []daghash.Hash{
			[32]byte{ // Make go vet happy.
				0x8b, 0xdf, 0xd1, 0x48, 0xef, 0xf5, 0x2b, 0x5e,
				0xfe, 0x26, 0xba, 0x37, 0xcb, 0x23, 0x0d, 0x41,
				0x24, 0x80, 0xfe, 0x9a, 0x38, 0x90, 0xb9, 0xd3,
				0x07, 0x30, 0xcc, 0xa0, 0x4f, 0x4e, 0xf1, 0x02,
			},
		},
		MerkleRoot: daghash.Hash([32]byte{ // Make go vet happy.
			0x4b, 0xd6, 0xbf, 0x21, 0xa0, 0x62, 0x77, 0xb5,
			0xc0, 0xd3, 0x3b, 0x31, 0x9d, 0x30, 0x9b, 0x89,
			0x93, 0x75, 0x50, 0xdb, 0x3b, 0x87, 0x23, 0x67,
			0x2f, 0xeb, 0xf9, 0xf2, 0x1b, 0x63, 0x5f, 0x1c,
		}),
		Timestamp: time.Unix(0x5bd58c4c, 0),
		Bits:      0x207fffff,
		Nonce:     0xdffffffffffffff9,
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  daghash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x53, 0x51, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
						0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
					},
					Sequence: math.MaxUint64,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 5000000000,
					PkScript: []byte{
						0x76, 0xa9, 0x14, 0x3d, 0xee, 0x47, 0x71, 0x6e,
						0x3c, 0xfa, 0x57, 0xdf, 0x45, 0x11, 0x34, 0x73,
						0xa6, 0x31, 0x2e, 0xbe, 0xae, 0xf3, 0x11, 0x88,
						0xac,
					},
				},
			},
			LockTime: 0,
		},
		block3Tx,
	},
}

var block4 = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 1,
		PrevBlocks: []daghash.Hash{
			[32]byte{ // Make go vet happy.
				0xde, 0xe3, 0x62, 0x5f, 0x0c, 0x98, 0x26, 0x5f,
				0x9b, 0x3e, 0xb1, 0xd9, 0x32, 0x0a, 0x84, 0xb3,
				0xe1, 0xbe, 0xe2, 0xb7, 0x8e, 0x4a, 0xfb, 0x97,
				0x7a, 0x53, 0x32, 0xff, 0x32, 0x17, 0xfc, 0x57,
			},
		},
		MerkleRoot: daghash.Hash([32]byte{ // Make go vet happy.
			0xe1, 0x13, 0x4a, 0xd8, 0xd5, 0x43, 0x33, 0x95,
			0x55, 0x19, 0x00, 0xaf, 0x13, 0x3f, 0xd6, 0x3a,
			0x63, 0x98, 0x50, 0x61, 0xfc, 0x02, 0x2c, 0x44,
			0x1b, 0x0e, 0x74, 0x7d, 0x5c, 0x19, 0x58, 0xb4,
		}),
		Timestamp: time.Unix(0x5bd58c4d, 0),
		Bits:      0x207fffff,
		Nonce:     0xdffffffffffffffa,
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  daghash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x54, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
						0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
					},
					Sequence: math.MaxUint64,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 5000000000,
					PkScript: []byte{
						0x76, 0xa9, 0x14, 0x3d, 0xee, 0x47, 0x71, 0x6e,
						0x3c, 0xfa, 0x57, 0xdf, 0x45, 0x11, 0x34, 0x73,
						0xa6, 0x31, 0x2e, 0xbe, 0xae, 0xf3, 0x11, 0x88,
						0xac,
					},
				},
			},
			LockTime: 0,
		},
	},
}

var block5 = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:       1,
		NumPrevBlocks: 2,
		PrevBlocks: []daghash.Hash{
			[32]byte{ // Make go vet happy.
				0xfd, 0x96, 0x3c, 0xfb, 0xed, 0x5a, 0xeb, 0xdb, 0x3d, 0x8e, 0xe9, 0x53, 0xf1, 0xe6, 0xad, 0x12, 0x21, 0x02, 0x55, 0x62, 0xbc, 0x2e, 0x52, 0xee, 0xb9, 0xd0, 0x60, 0xda, 0xd6, 0x4a, 0x20, 0x5a},
			[32]byte{ // Make go vet happy.
				0xec, 0x42, 0x2c, 0x0c, 0x8c, 0x94, 0x50, 0x17, 0x85, 0xbb, 0x8c, 0xaf, 0x72, 0xd9, 0x39, 0x28, 0x26, 0xaa, 0x42, 0x8d, 0xd5, 0x09, 0xa2, 0xb6, 0xa6, 0x8c, 0x4e, 0x85, 0x72, 0x44, 0xd5, 0x70},
		},
		MerkleRoot: daghash.Hash([32]byte{ // Make go vet happy.
			0x77, 0xc7, 0x09, 0x46, 0x0f, 0x81, 0x37, 0xca,
			0xf5, 0xec, 0xa5, 0xae, 0x4c, 0xad, 0x65, 0xc5,
			0xdd, 0x73, 0x4f, 0xb5, 0xcf, 0x04, 0x20, 0x38,
			0x29, 0x10, 0x5b, 0x66, 0xfe, 0x15, 0x8a, 0xfb,
		}),
		Timestamp: time.Unix(0x5bd58c4e, 0),
		Bits:      0x207fffff,
		Nonce:     4,
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  daghash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x55, 0x00, 0x0b, 0x2f, 0x50, 0x32, 0x53, 0x48,
						0x2f, 0x62, 0x74, 0x63, 0x64, 0x2f,
					},
					Sequence: math.MaxUint64,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 5000000000,
					PkScript: []byte{
						0x76, 0xa9, 0x14, 0x3d, 0xee, 0x47, 0x71, 0x6e,
						0x3c, 0xfa, 0x57, 0xdf, 0x45, 0x11, 0x34, 0x73,
						0xa6, 0x31, 0x2e, 0xbe, 0xae, 0xf3, 0x11, 0x88,
						0xac,
					},
				},
			},
			LockTime: 0,
		},
	},
}
