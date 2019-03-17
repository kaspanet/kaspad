package indexers

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/mining"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

func createTransaction(value uint64, originTx *wire.MsgTx, outputIndex uint32) *wire.MsgTx {
	txIn := &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			TxID:  originTx.TxID(),
			Index: outputIndex,
		},
		Sequence: wire.MaxTxInSequenceNum,
	}
	txOut := wire.NewTxOut(value, blockdag.OpTrueScript)
	tx := wire.NewMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, nil, 0, nil)

	return tx
}

func TestTxIndexConnectBlock(t *testing.T) {
	blocks := make(map[daghash.Hash]*util.Block)

	txIndex := NewTxIndex()
	indexManager := NewManager([]Indexer{txIndex})

	params := dagconfig.SimNetParams
	params.BlockRewardMaturity = 1
	params.K = 1

	config := blockdag.Config{
		IndexManager: indexManager,
		DAGParams:    &params,
		SubnetworkID: subnetworkid.SubnetworkIDSupportsAll,
	}

	dag, teardown, err := blockdag.DAGSetup("TestTxIndexConnectBlock", config)
	if err != nil {
		t.Fatalf("TestTxIndexConnectBlock: Failed to setup DAG instance: %v", err)
	}
	if teardown != nil {
		defer teardown()
	}

	prepareAndProcessBlock := func(parentHashes []daghash.Hash, transactions []*wire.MsgTx, blockName string) *wire.MsgBlock {
		block, err := mining.PrepareBlockForTest(dag, &params, parentHashes, transactions, false, 1)
		if err != nil {
			t.Fatalf("TestTxIndexConnectBlock: block %v got unexpected error from PrepareBlockForTest: %v", blockName, err)
		}
		utilBlock := util.NewBlock(block)
		blocks[block.BlockHash()] = utilBlock
		isOrphan, err := dag.ProcessBlock(utilBlock, blockdag.BFNoPoWCheck)
		if err != nil {
			t.Fatalf("TestTxIndexConnectBlock: dag.ProcessBlock got unexpected error for block %v: %v", blockName, err)
		}
		if isOrphan {
			t.Fatalf("TestTxIndexConnectBlock: block %v was unexpectedly orphan", blockName)
		}
		return block
	}

	block1 := prepareAndProcessBlock([]daghash.Hash{*params.GenesisHash}, nil, "1")
	block2Tx := createTransaction(block1.Transactions[0].TxOut[0].Value, block1.Transactions[0], 0)
	block2 := prepareAndProcessBlock([]daghash.Hash{block1.BlockHash()}, []*wire.MsgTx{block2Tx}, "2")
	block3Tx := createTransaction(block2.Transactions[0].TxOut[0].Value, block2.Transactions[0], 0)
	block3 := prepareAndProcessBlock([]daghash.Hash{block2.BlockHash()}, []*wire.MsgTx{block3Tx}, "3")

	block3TxID := block3Tx.TxID()
	block3TxNewAcceptedBlock, err := txIndex.BlockThatAcceptedTx(dag, &block3TxID)
	if err != nil {
		t.Errorf("TestTxIndexConnectBlock: TxAcceptedInBlock: %v", err)
	}
	block3Hash := block3.BlockHash()
	if !block3TxNewAcceptedBlock.IsEqual(&block3Hash) {
		t.Errorf("TestTxIndexConnectBlock: block3Tx should've "+
			"been accepted in block %v but instead got accepted in block %v", block3Hash, block3TxNewAcceptedBlock)
	}

	block3A := prepareAndProcessBlock([]daghash.Hash{block2.BlockHash()}, []*wire.MsgTx{block3Tx}, "3A")
	block4 := prepareAndProcessBlock([]daghash.Hash{block3.BlockHash()}, nil, "4")
	prepareAndProcessBlock([]daghash.Hash{block3A.BlockHash(), block4.BlockHash()}, nil, "5")

	block3TxAcceptedBlock, err := txIndex.BlockThatAcceptedTx(dag, &block3TxID)
	if err != nil {
		t.Errorf("TestTxIndexConnectBlock: TxAcceptedInBlock: %v", err)
	}
	block3AHash := block3A.BlockHash()
	if !block3TxAcceptedBlock.IsEqual(&block3AHash) {
		t.Errorf("TestTxIndexConnectBlock: block3Tx should've "+
			"been accepted in block %v but instead got accepted in block %v", block3AHash, block3TxAcceptedBlock)
	}

	region, err := txIndex.TxFirstBlockRegion(&block3TxID)
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
