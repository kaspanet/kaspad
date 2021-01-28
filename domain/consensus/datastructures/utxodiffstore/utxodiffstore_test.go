package utxodiffstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"math/rand"
	"testing"
)

func TestUTXODiffSerializationAndDeserialization(t *testing.T) {
	utxoDiffStore := New(0).(*utxoDiffStore)

	testUTXODiff, err := buildTestUTXODiff()
	if err != nil {
		t.Fatalf("Could not create UTXODiff from toAdd and toRemove collections: %s", err)
	}

	serializedUTXODiff, err := utxoDiffStore.serializeUTXODiff(testUTXODiff)
	if err != nil {
		t.Fatalf("Could not serialize UTXO diff: %s", err)
	}
	deserializedUTXODiff, err := utxoDiffStore.deserializeUTXODiff(serializedUTXODiff)
	if err != nil {
		t.Fatalf("Could not deserialize UTXO diff: %s", err)
	}

	if testUTXODiff.ToAdd().Len() != deserializedUTXODiff.ToAdd().Len() {
		t.Fatalf("Unexpected toAdd length in deserialized utxoDiff. Want: %d, got: %d",
			testUTXODiff.ToAdd().Len(), deserializedUTXODiff.ToAdd().Len())
	}
	if testUTXODiff.ToRemove().Len() != deserializedUTXODiff.ToRemove().Len() {
		t.Fatalf("Unexpected toRemove length in deserialized utxoDiff. Want: %d, got: %d",
			testUTXODiff.ToRemove().Len(), deserializedUTXODiff.ToRemove().Len())
	}

	testToAddIterator := testUTXODiff.ToAdd().Iterator()
	for ok := testToAddIterator.First(); ok; ok = testToAddIterator.Next() {
		testOutpoint, testUTXOEntry, err := testToAddIterator.Get()
		if err != nil {
			t.Fatalf("Could not get an outpoint-utxoEntry pair out of the toAdd iterator: %s", err)
		}
		deserializedUTXOEntry, ok := deserializedUTXODiff.ToAdd().Get(testOutpoint)
		if !ok {
			t.Fatalf("Outpoint %s:%d not found in the deserialized toAdd collection",
				testOutpoint.TransactionID, testOutpoint.Index)
		}
		if !testUTXOEntry.Equal(deserializedUTXOEntry) {
			t.Fatalf("Deserialized UTXO entry is not equal to the original UTXO entry for outpoint %s:%d "+
				"in the toAdd collection", testOutpoint.TransactionID, testOutpoint.Index)
		}
	}

	testToRemoveIterator := testUTXODiff.ToRemove().Iterator()
	for ok := testToRemoveIterator.First(); ok; ok = testToRemoveIterator.Next() {
		testOutpoint, testUTXOEntry, err := testToRemoveIterator.Get()
		if err != nil {
			t.Fatalf("Could not get an outpoint-utxoEntry pair out of the toRemove iterator: %s", err)
		}
		deserializedUTXOEntry, ok := deserializedUTXODiff.ToRemove().Get(testOutpoint)
		if !ok {
			t.Fatalf("Outpoint %s:%d not found in the deserialized toRemove collection",
				testOutpoint.TransactionID, testOutpoint.Index)
		}
		if !testUTXOEntry.Equal(deserializedUTXOEntry) {
			t.Fatalf("Deserialized UTXO entry is not equal to the original UTXO entry for outpoint %s:%d "+
				"in the toRemove collection", testOutpoint.TransactionID, testOutpoint.Index)
		}
	}
}

func BenchmarkUTXODiffSerialization(b *testing.B) {
	utxoDiffStore := New(0).(*utxoDiffStore)

	testUTXODiff, err := buildTestUTXODiff()
	if err != nil {
		b.Fatalf("Could not create UTXODiff from toAdd and toRemove collections: %s", err)
	}

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

	testUTXODiff, err := buildTestUTXODiff()
	if err != nil {
		b.Fatalf("Could not create UTXODiff from toAdd and toRemove collections: %s", err)
	}
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

	testUTXODiff, err := buildTestUTXODiff()
	if err != nil {
		b.Fatalf("Could not create UTXODiff from toAdd and toRemove collections: %s", err)
	}

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

func buildTestUTXODiff() (model.UTXODiff, error) {
	toAdd := buildTestUTXOCollection()
	toRemove := buildTestUTXOCollection()

	utxoDiff, err := utxo.NewUTXODiffFromCollections(toAdd, toRemove)
	if err != nil {
		return nil, err
	}
	return utxoDiff, nil
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
