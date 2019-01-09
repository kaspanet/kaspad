package blockdag

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/wire"
)

func TestSerializeSubNetworkRegistryTxs(t *testing.T) {
	payload1 := make([]byte, 8)
	binary.LittleEndian.PutUint64(payload1, uint64(100))
	tx1 := wire.MsgTx{
		Version:      1,
		SubNetworkID: wire.SubNetworkRegistry,
		Payload:      payload1,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{
					Hash:  *newHashFromStr("0437cd7f8525ceed2324359c2d0ba26006d92d856a9c20fa0241106ee5a597c9"),
					Index: 0,
				},
				SignatureScript: hexToBytes("47304402204e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd410220181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d0901"),
				Sequence:        math.MaxUint64,
			},
		},
		TxOut: []*wire.TxOut{{
			Value:    1000000000,
			PkScript: hexToBytes("4104ae1a62fe09c5f51b13905f07f06b99a2f7159b2225f374cd378d71302fa28414e7aab37397f554a7df5f142c21c1b7303b8a0626f1baded5c72a704f7e6cd84cac"),
		}, {
			Value:    4000000000,
			PkScript: hexToBytes("410411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3ac"),
		}},
	}

	payload2 := make([]byte, 8)
	binary.LittleEndian.PutUint64(payload2, uint64(200))
	tx2 := wire.MsgTx{
		Version:      1,
		SubNetworkID: wire.SubNetworkRegistry,
		Payload:      payload2,
		TxIn: []*wire.TxIn{{
			PreviousOutPoint: wire.OutPoint{
				Hash:  *newHashFromStr("0437cd7f8525ceed2324359c2d0ba26006d92d856a9c20fa0241106ee5a597c9"),
				Index: 0,
			},
			SignatureScript: hexToBytes("47304402204e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd410220181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d0901"),
			Sequence:        math.MaxUint64,
		}},
		TxOut: []*wire.TxOut{{
			Value:    5000000,
			PkScript: hexToBytes("76a914f419b8db4ba65f3b6fcc233acb762ca6f51c23d488ac"),
		}, {
			Value:    34400000000,
			PkScript: hexToBytes("76a914cadf4fc336ab3c6a4610b75f31ba0676b7f663d288ac"),
		}},
		LockTime: 0,
	}

	tests := []struct {
		name string
		txs  []*wire.MsgTx
	}{
		{
			name: "empty slice",
			txs:  []*wire.MsgTx{},
		},
		{
			name: "one transaction",
			txs:  []*wire.MsgTx{&tx1},
		},
		{
			name: "two transactions",
			txs:  []*wire.MsgTx{&tx2, &tx1},
		},
	}

	for _, test := range tests {
		serializedTxs, err := serializeSubNetworkRegistryTxs(test.txs)
		if err != nil {
			t.Errorf("serialization in test '%s' unexpectedly failed: %s", test.name, err)
			continue
		}

		deserializedTxs, err := deserializeSubNetworkRegistryTxs(serializedTxs)
		if err != nil {
			t.Errorf("deserialization in test '%s' unexpectedly failed: %s", test.name, err)
			continue
		}

		if !reflect.DeepEqual(test.txs, deserializedTxs) {
			t.Errorf("original txs and deserialized txs are not equal in test '%s'", test.name)
		}
	}
}

func TestSerializeSubNetwork(t *testing.T) {
	sNet := &subNetwork{
		txHash:   *newHashFromStr("0437cd7f8525ceed2324359c2d0ba26006d92d856a9c20fa0241106ee5a597c9"),
		gasLimit: 1000,
	}

	serializedSNet, err := serializeSubNetwork(sNet)
	if err != nil {
		t.Fatalf("sub-network serialization unexpectedly failed: %s", err)
	}

	deserializedSNet, err := deserializeSubNetwork(serializedSNet)
	if err != nil {
		t.Fatalf("sub-network deserialization unexpectedly failed: %s", err)
	}

	if !reflect.DeepEqual(sNet, deserializedSNet) {
		t.Errorf("original sub-network and deserialized sub-network are not equal")
	}
}
