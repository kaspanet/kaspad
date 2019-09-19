package indexers

import (
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"reflect"
	"testing"
)

func TestSerializationAnDeserialization(t *testing.T) {
	txsAcceptanceData := blockdag.MultiBlockTxsAcceptanceData{}

	// Create test data
	hash, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	txIn1 := &wire.TxIn{SignatureScript: []byte{1}, PreviousOutpoint: wire.Outpoint{Index: 1}, Sequence: 0}
	txIn2 := &wire.TxIn{SignatureScript: []byte{2}, PreviousOutpoint: wire.Outpoint{Index: 2}, Sequence: 0}
	txOut1 := &wire.TxOut{ScriptPubKey: []byte{1}, Value: 10}
	txOut2 := &wire.TxOut{ScriptPubKey: []byte{2}, Value: 20}
	blockTxsAcceptanceData := blockdag.BlockTxsAcceptanceData{
		{
			Tx:         util.NewTx(wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn1}, []*wire.TxOut{txOut1})),
			IsAccepted: true,
		},
		{
			Tx:         util.NewTx(wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn2}, []*wire.TxOut{txOut2})),
			IsAccepted: false,
		},
	}
	txsAcceptanceData[*hash] = blockTxsAcceptanceData

	// Serialize
	serializedTxsAcceptanceData, err := serializeMultiBlockTxsAcceptanceData(txsAcceptanceData)
	if err != nil {
		t.Fatalf("TestSerializationAnDeserialization: serialization failed: %s", err)
	}

	// Deserialize
	deserializedTxsAcceptanceData, err := deserializeMultiBlockTxsAcceptanceData(serializedTxsAcceptanceData)
	if err != nil {
		t.Fatalf("TestSerializationAnDeserialization: deserialization failed: %s", err)
	}

	// Check that they're the same
	if !reflect.DeepEqual(txsAcceptanceData, deserializedTxsAcceptanceData) {
		t.Fatalf("TestSerializationAnDeserialization: original data and deseralize data aren't equal")
	}
}
