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
			t.Fatalf("TestTxIndexConnectBlock: Couldn't create txIndex: %v", err)
		}
		msgBlock1 := wire.NewMsgBlock(wire.NewBlockHeader(1,
			[]daghash.Hash{
				daghash.Hash{1},
			}, &daghash.Hash{}, 1, 1))

		dummyPrevOutHash, err := daghash.NewHashFromStr("01")
		if err != nil {
			t.Fatalf("TestTxIndexConnectBlock: NewShaHashFromStr: unexpected error: %v", err)
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
		msgBlock1.AddTransaction(tx1)
		block1 := util.NewBlock(msgBlock1)
		err = txIndex.ConnectBlock(dbTx, block1, &blockdag.BlockDAG{}, []*blockdag.AcceptedTxData{
			&blockdag.AcceptedTxData{
				Tx:      util.NewTx(tx1),
				InBlock: block1.Hash(),
			},
		})
		if err != nil {
			return err
		}

		tx1Hash := tx1.TxHash()
		block1IDBytes := make([]byte, 4)
		byteOrder.PutUint32(block1IDBytes, uint32(1))
		tx1IncludingBucket := dbTx.Metadata().Bucket(includingBlocksIndexKey).Bucket(tx1Hash[:])
		if tx1IncludingBucket == nil {
			t.Fatalf("TestTxIndexConnectBlock: No including blocks bucket was found for tx1")
		}
		block1Tx1includingBlocksIndexEntry := tx1IncludingBucket.Get(block1IDBytes)
		if len(block1Tx1includingBlocksIndexEntry) == 0 {
			t.Fatalf("TestTxIndexConnectBlock: there was no entry for block1 in tx1's including blocks bucket")
		}

		tx1Offset := byteOrder.Uint32(block1Tx1includingBlocksIndexEntry[:4])
		tx1Len := byteOrder.Uint32(block1Tx1includingBlocksIndexEntry[4:])

		block1Bytes, err := block1.Bytes()
		if err != nil {
			t.Fatalf("TestTxIndexConnectBlock: Couldn't serialize block 1 to bytes")
		}
		tx1InBlock1 := block1Bytes[tx1Offset : tx1Offset+tx1Len]

		wTx1 := bytes.NewBuffer(make([]byte, 0, tx1.SerializeSize()))
		tx1.BtcEncode(wTx1, 0)
		tx1Bytes := wTx1.Bytes()

		if !reflect.DeepEqual(tx1Bytes, tx1InBlock1) {
			t.Errorf("TestTxIndexConnectBlock: the block region that was in the bucket doesn't match tx1")
		}

		tx1AcceptingBlocksBucket := dbTx.Metadata().Bucket(acceptingBlocksIndexKey).Bucket(tx1Hash[:])
		if tx1AcceptingBlocksBucket == nil {
			t.Fatalf("TestTxIndexConnectBlock: No accepting blocks bucket was found for tx1")
		}

		block1Tx1AcceptingEntry := tx1AcceptingBlocksBucket.Get(block1IDBytes)
		tx1IncludingBlockID := byteOrder.Uint32(block1Tx1AcceptingEntry)
		if tx1IncludingBlockID != 1 {
			t.Fatalf("TestTxIndexConnectBlock: tx1 should've been included in block 1, but got %v", tx1IncludingBlockID)
		}

		msgBlock2 := wire.NewMsgBlock(wire.NewBlockHeader(1,
			[]daghash.Hash{
				daghash.Hash{2},
			}, &daghash.Hash{}, 1, 1))
		dummyPrevOut2 := wire.OutPoint{Hash: *dummyPrevOutHash, Index: 1}
		tx2 := wire.NewMsgTx(wire.TxVersion)
		tx2.AddTxIn(wire.NewTxIn(&dummyPrevOut2, dummySigScript))
		tx2.AddTxOut(dummyTxOut)
		msgBlock2.AddTransaction(tx2)
		block2 := util.NewBlock(msgBlock2)
		err = txIndex.ConnectBlock(dbTx, block2, &blockdag.BlockDAG{}, []*blockdag.AcceptedTxData{
			&blockdag.AcceptedTxData{
				Tx:      util.NewTx(tx1),
				InBlock: block1.Hash(),
			},
			&blockdag.AcceptedTxData{
				Tx:      util.NewTx(tx2),
				InBlock: block2.Hash(),
			},
		})
		if err != nil {
			t.Fatalf("TestTxIndexConnectBlock: Couldn't connect block 1 to txindex")
		}

		tx2Hash := tx2.TxHash()
		block2IDBytes := make([]byte, 4)
		byteOrder.PutUint32(block2IDBytes, uint32(2))

		tx2IncludingBlocksBucket := dbTx.Metadata().Bucket(includingBlocksIndexKey).Bucket(tx2Hash[:])
		if tx2IncludingBlocksBucket == nil {
			t.Fatalf("TestTxIndexConnectBlock: No including blocks bucket was found for tx2")
		}

		block2Tx2includingBlocksIndexEntry := tx2IncludingBlocksBucket.Get(block2IDBytes)
		if len(block2Tx2includingBlocksIndexEntry) == 0 {
			t.Fatalf("TestTxIndexConnectBlock: there was no entry for block2 in tx2's including blocks bucket")
		}

		tx2AcceptingBlocksBucket := dbTx.Metadata().Bucket(acceptingBlocksIndexKey).Bucket(tx2Hash[:])
		if tx2AcceptingBlocksBucket == nil {
			t.Fatalf("TestTxIndexConnectBlock: No accepting blocks bucket was found for tx2")
		}

		block2Tx2AcceptingEntry := tx2AcceptingBlocksBucket.Get(block2IDBytes)
		tx2IncludingBlockID := byteOrder.Uint32(block2Tx2AcceptingEntry)
		if tx2IncludingBlockID != 2 {
			t.Fatalf("TestTxIndexConnectBlock: tx2 should've been included in block 2, but got %v", tx1IncludingBlockID)
		}

		block2Tx1AcceptingEntry := tx1AcceptingBlocksBucket.Get(block2IDBytes)
		tx1Block2IncludingBlockID := byteOrder.Uint32(block2Tx1AcceptingEntry)
		if tx1Block2IncludingBlockID != 1 {
			t.Fatalf("TestTxIndexConnectBlock: tx2 should've been included in block 1, but got %v", tx1Block2IncludingBlockID)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("TestTxIndexConnectBlock: %v", err)
	}
}
