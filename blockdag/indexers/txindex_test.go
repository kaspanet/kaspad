package indexers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

func tempDb() (database.DB, func(), error) {
	dbPath, err := ioutil.TempDir("", "ffldb")
	if err != nil {
		return nil, nil, err
	}
	db, err := database.Create("ffldb", dbPath, wire.MainNet)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating db: %v", err)
	}
	teardown := func() {
		db.Close()
		os.RemoveAll(dbPath)
	}
	return db, teardown, nil
}

func TestTxIndexConnectBlock(t *testing.T) {
	db, teardown, err := tempDb()
	if teardown != nil {
		defer teardown()
	}
	if err != nil {
		t.Fatalf("TestTxIndexConnectBlock: %v", err)
	}
	err = db.Update(func(dbTx database.Tx) error {
		txIndex := NewTxIndex(db)
		err := txIndex.Create(dbTx)
		if err != nil {
			t.Errorf("TestTxIndexConnectBlock: Couldn't create txIndex: %v", err)
			return nil
		}
		msgBlock1 := wire.NewMsgBlock(wire.NewBlockHeader(1,
			[]daghash.Hash{{1}}, &daghash.Hash{}, 1, 1))

		dummyPrevOutHash, err := daghash.NewHashFromStr("01")
		if err != nil {
			t.Errorf("TestTxIndexConnectBlock: NewShaHashFromStr: unexpected error: %v", err)
			return nil
		}
		dummyPrevOut1 := wire.OutPoint{Hash: *dummyPrevOutHash, Index: 0}
		dummySigScript := bytes.Repeat([]byte{0x00}, 65)
		dummyTxOut := &wire.TxOut{
			Value:    5000000000,
			PkScript: bytes.Repeat([]byte{0x00}, 65),
		}

		tx1 := wire.NewMsgTx(wire.TxVersion)
		tx1.AddTxIn(wire.NewTxIn(&dummyPrevOut1, dummySigScript))
		tx1.AddTxOut(dummyTxOut)

		dummyPrevOut2 := wire.OutPoint{Hash: *dummyPrevOutHash, Index: 1}

		tx2 := wire.NewMsgTx(wire.TxVersion)
		tx2.AddTxIn(wire.NewTxIn(&dummyPrevOut2, dummySigScript))
		tx2.AddTxOut(dummyTxOut)

		msgBlock1.AddTransaction(tx1)
		msgBlock1.AddTransaction(tx2)
		block1 := util.NewBlock(msgBlock1)
		err = txIndex.ConnectBlock(dbTx, block1, &blockdag.BlockDAG{}, []*blockdag.TxWithBlockHash{
			{
				Tx:      util.NewTx(tx1),
				InBlock: block1.Hash(),
			},
			{
				Tx:      util.NewTx(tx2),
				InBlock: block1.Hash(),
			},
		})
		if err != nil {
			t.Errorf("TestTxIndexConnectBlock: Couldn't connect block 1 to txindex")
			return nil
		}

		tx1Hash := tx1.TxHash()

		tx1Blocks, err := dbFetchTxBlocks(dbTx, &tx1Hash)
		if err != nil {
			t.Errorf("TestTxIndexConnectBlock: dbFetchTxBlocks: %v", err)
			return nil
		}
		expectedTx1Blocks := []daghash.Hash{
			*block1.Hash(),
		}
		if !daghash.AreEqual(tx1Blocks, expectedTx1Blocks) {
			t.Errorf("TestTxIndexConnectBlock: tx1Blocks expected to be %v but got %v", expectedTx1Blocks, tx1Blocks)
			return nil
		}

		block1IDBytes := make([]byte, 4)
		byteOrder.PutUint32(block1IDBytes, uint32(1))
		regionTx1, err := dbFetchFirstTxRegion(dbTx, &tx1Hash)
		if err != nil {
			t.Errorf("TestTxIndexConnectBlock: no block region was found for tx1")
			return nil
		}

		block1Bytes, err := block1.Bytes()
		if err != nil {
			t.Errorf("TestTxIndexConnectBlock: Couldn't serialize block 1 to bytes")
			return nil
		}
		tx1InBlock1 := block1Bytes[regionTx1.Offset : regionTx1.Offset+regionTx1.Len]

		wTx1 := bytes.NewBuffer(make([]byte, 0, tx1.SerializeSize()))
		tx1.BtcEncode(wTx1, 0)
		tx1Bytes := wTx1.Bytes()

		if !reflect.DeepEqual(tx1Bytes, tx1InBlock1) {
			t.Errorf("TestTxIndexConnectBlock: the block region that was in the bucket doesn't match tx1")
		}

		tx1AcceptingBlocksBucket := dbTx.Metadata().Bucket(acceptingBlocksIndexKey).Bucket(tx1Hash[:])
		if tx1AcceptingBlocksBucket == nil {
			t.Errorf("TestTxIndexConnectBlock: No accepting blocks bucket was found for tx1")
			return nil
		}

		block1Tx1AcceptingEntry := tx1AcceptingBlocksBucket.Get(block1IDBytes)
		tx1IncludingBlockID := byteOrder.Uint32(block1Tx1AcceptingEntry)
		if tx1IncludingBlockID != 1 {
			t.Errorf("TestTxIndexConnectBlock: tx1 should've been included in block 1, but got %v", tx1IncludingBlockID)
			return nil
		}

		msgBlock2 := wire.NewMsgBlock(wire.NewBlockHeader(1,
			[]daghash.Hash{{2}}, &daghash.Hash{}, 1, 1))
		msgBlock2.AddTransaction(tx2)

		dummyPrevOut3 := wire.OutPoint{Hash: *dummyPrevOutHash, Index: 2}
		tx3 := wire.NewMsgTx(wire.TxVersion)
		tx3.AddTxIn(wire.NewTxIn(&dummyPrevOut3, dummySigScript))
		tx3.AddTxOut(dummyTxOut)
		msgBlock2.AddTransaction(tx3)

		block2 := util.NewBlock(msgBlock2)
		err = txIndex.ConnectBlock(dbTx, block2, &blockdag.BlockDAG{}, []*blockdag.TxWithBlockHash{
			{
				Tx:      util.NewTx(tx1),
				InBlock: block1.Hash(),
			},
			{
				Tx:      util.NewTx(tx2),
				InBlock: block2.Hash(),
			},
			{
				Tx:      util.NewTx(tx3),
				InBlock: block2.Hash(),
			},
		})
		if err != nil {
			t.Errorf("TestTxIndexConnectBlock: Couldn't connect block 2 to txindex")
			return nil
		}

		tx2Hash := tx2.TxHash()
		tx2Blocks, err := dbFetchTxBlocks(dbTx, &tx2Hash)
		if err != nil {
			t.Errorf("TestTxIndexConnectBlock: dbFetchTxBlocks: %v", err)
			return nil
		}
		daghash.Sort(tx2Blocks)
		expectedTx2Blocks := []daghash.Hash{
			*block1.Hash(),
			*block2.Hash(),
		}
		daghash.Sort(expectedTx2Blocks)
		if !daghash.AreEqual(tx2Blocks, expectedTx2Blocks) {
			t.Errorf("TestTxIndexConnectBlock: tx2Blocks expected to be %v but got %v", expectedTx2Blocks, tx2Blocks)
			return nil
		}

		tx3Hash := tx3.TxHash()

		tx3Blocks, err := dbFetchTxBlocks(dbTx, &tx3Hash)
		expectedTx3Blocks := []daghash.Hash{
			*block2.Hash(),
		}
		if !daghash.AreEqual(tx3Blocks, expectedTx3Blocks) {
			t.Errorf("TestTxIndexConnectBlock: tx1Blocks expected to be %v but got %v", expectedTx3Blocks, tx3Blocks)
			return nil
		}

		block2IDBytes := make([]byte, 4)
		byteOrder.PutUint32(block2IDBytes, uint32(2))

		tx3AcceptingBlocksBucket := dbTx.Metadata().Bucket(acceptingBlocksIndexKey).Bucket(tx3Hash[:])
		if tx3AcceptingBlocksBucket == nil {
			t.Errorf("TestTxIndexConnectBlock: No accepting blocks bucket was found for tx3")
			return nil
		}

		block2Tx3AcceptingEntry := tx3AcceptingBlocksBucket.Get(block2IDBytes)
		tx3IncludingBlockID := byteOrder.Uint32(block2Tx3AcceptingEntry)
		if tx3IncludingBlockID != 2 {
			t.Errorf("TestTxIndexConnectBlock: tx3 should've been included in block 2, but got %v", tx1IncludingBlockID)
			return nil
		}

		block2Tx1AcceptingEntry := tx1AcceptingBlocksBucket.Get(block2IDBytes)
		tx1Block2IncludingBlockID := byteOrder.Uint32(block2Tx1AcceptingEntry)
		if tx1Block2IncludingBlockID != 1 {
			t.Errorf("TestTxIndexConnectBlock: tx3 should've been included in block 1, but got %v", tx1Block2IncludingBlockID)
			return nil
		}

		return nil
	})
	if err != nil {
		t.Fatalf("TestTxIndexConnectBlock: %v", err)
	}
}
