package utxo

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
)

func Benchmark_serializeUTXO(b *testing.B) {
	script, err := hex.DecodeString("76a914ad06dd6ddee55cbca9a9e3713bd7587509a3056488ac")
	if err != nil {
		b.Fatalf("Error decoding scriptPublicKey string: %s", err)
	}
	scriptPublicKey := &externalapi.ScriptPublicKey{script, 0}
	entry := NewUTXOEntry(5000000000, scriptPublicKey, false, 1432432)
	outpoint := &externalapi.DomainOutpoint{
		TransactionID: *externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x16, 0x5e, 0x38, 0xe8, 0xb3, 0x91, 0x45, 0x95,
			0xd9, 0xc6, 0x41, 0xf3, 0xb8, 0xee, 0xc2, 0xf3,
			0x46, 0x11, 0x89, 0x6b, 0x82, 0x1a, 0x68, 0x3b,
			0x7a, 0x4e, 0xde, 0xfe, 0x2c, 0x00, 0x00, 0x00,
		}),
		Index: 0xffffffff,
	}

	for i := 0; i < b.N; i++ {
		_, err := SerializeUTXO(entry, outpoint)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Test_serializeUTXO(t *testing.T) {
	script, err := hex.DecodeString("76a914ad06dd6ddee55cbca9a9e3713bd7587509a3056488ac")
	if err != nil {
		t.Fatalf("Error decoding scriptPublicKey script string: %s", err)
	}
	scriptPublicKey := &externalapi.ScriptPublicKey{Script: script, Version: 0}
	entry := NewUTXOEntry(5000000000, scriptPublicKey, false, 1432432)
	outpoint := &externalapi.DomainOutpoint{
		TransactionID: *externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x16, 0x5e, 0x38, 0xe8, 0xb3, 0x91, 0x45, 0x95,
			0xd9, 0xc6, 0x41, 0xf3, 0xb8, 0xee, 0xc2, 0xf3,
			0x46, 0x11, 0x89, 0x6b, 0x82, 0x1a, 0x68, 0x3b,
			0x7a, 0x4e, 0xde, 0xfe, 0x2c, 0x00, 0x00, 0x00,
		}),
		Index: 0xffffffff,
	}

	serialized, err := SerializeUTXO(entry, outpoint)
	if err != nil {
		t.Fatalf("SerializeUTXO: %+v", err)
	}

	deserializedEntry, deserializedOutpoint, err := DeserializeUTXO(serialized)
	if err != nil {
		t.Fatalf("DeserializeUTXO: %+v", err)
	}

	if !reflect.DeepEqual(deserializedEntry, entry) {
		t.Fatalf("deserialized entry is not equal to the original")
	}

	if !reflect.DeepEqual(deserializedOutpoint, outpoint) {
		t.Fatalf("deserialized outpoint is not equal to the original")
	}
}
