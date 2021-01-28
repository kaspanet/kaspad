package utxodiffstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"math/rand"
	"testing"
)

func BenchmarkUTXODiffSerialization(b *testing.B) {
	utxoDiffStore := New(0).(*utxoDiffStore)

	testUTXODiff := buildTestUTXODiff(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := utxoDiffStore.serializeUTXODiff(testUTXODiff)
		if err != nil {
			b.Fatalf("Could not serialize UTXO diff: %s", err)
		}
	}
}

func BenchmarkUTXODiffDeserialization(b *testing.B) {
	utxoDiffStore := New(0).(*utxoDiffStore)

	testUTXODiff := buildTestUTXODiff(b)
	serializedUTXODiff, err := utxoDiffStore.serializeUTXODiff(testUTXODiff)
	if err != nil {
		b.Fatalf("Could not serialize UTXO diff: %s", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = utxoDiffStore.deserializeUTXODiff(serializedUTXODiff)
		if err != nil {
			b.Fatalf("Could not deserialize UTXO diff: %s", err)
		}
	}
}

func BenchmarkUTXODiffSerializationAndDeserialization(b *testing.B) {
	utxoDiffStore := New(0).(*utxoDiffStore)

	testUTXODiff := buildTestUTXODiff(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serializedUTXODiff, err := utxoDiffStore.serializeUTXODiff(testUTXODiff)
		if err != nil {
			b.Fatalf("Could not serialize UTXO diff: %s", err)
		}
		_, err = utxoDiffStore.deserializeUTXODiff(serializedUTXODiff)
		if err != nil {
			b.Fatalf("Could not deserialize UTXO diff: %s", err)
		}
	}
}

func buildTestUTXODiff(b *testing.B) model.UTXODiff {
	toAdd := buildTestUTXOCollection()
	toRemove := buildTestUTXOCollection()

	utxoDiff, err := utxo.NewUTXODiffFromCollections(toAdd, toRemove)
	if err != nil {
		b.Fatalf("Could not create UTXODiff from toAdd and toRemove collections: %s", err)
	}
	return utxoDiff
}

func buildTestUTXOCollection() model.UTXOCollection {
	utxoMap := make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry)

	for i := 0; i < 100_000; i++ {
		var outpointTransactionIDBytes [32]byte
		rand.Read(outpointTransactionIDBytes[:])
		outpointTransactionID := externalapi.NewDomainTransactionIDFromByteArray(&outpointTransactionIDBytes)
		outpointIndex := rand.Uint32()
		outpoint := externalapi.NewDomainOutpoint(outpointTransactionID, outpointIndex)

		utxoEntryAmount := rand.Uint64()
		var utxoEntryScriptPublicKeyScript [256]byte
		rand.Read(utxoEntryScriptPublicKeyScript[:])
		utxoEntryScriptPublicKeyVersion := uint16(rand.Uint32())
		utxoEntryScriptPublicKey := &externalapi.ScriptPublicKey{
			Script:  utxoEntryScriptPublicKeyScript[:],
			Version: utxoEntryScriptPublicKeyVersion,
		}
		utxoEntryIsCoinbase := rand.Float32() > 0.5
		utxoEntryBlockBlueScore := rand.Uint64()
		utxoEntry := utxo.NewUTXOEntry(utxoEntryAmount, utxoEntryScriptPublicKey, utxoEntryIsCoinbase, utxoEntryBlockBlueScore)

		utxoMap[*outpoint] = utxoEntry
	}

	return utxo.NewUTXOCollection(utxoMap)
}
