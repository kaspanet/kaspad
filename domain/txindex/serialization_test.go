package txindex

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func Test_serializeTxIndexData(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	serializedtxIndex := make([]byte, 68) // 32 bytes including block hash 32 bytes accepting blockhash and 4 bytes uint32
	r.Read(serializedtxIndex[:])
	includingBlockHash, err := externalapi.NewDomainHashFromByteSlice(serializedtxIndex[:32])
	if err != nil {
		t.Fatalf(err.Error())
	}
	acceptingBlockHash, err := externalapi.NewDomainHashFromByteSlice(serializedtxIndex[32:64])
	if err != nil {
		t.Fatalf(err.Error())
	}
	includingIndex := binary.BigEndian.Uint32(serializedtxIndex[64:68])

	testdeserializedtxIndex := &TxData{
		IncludingBlockHash: includingBlockHash,
		AcceptingBlockHash: acceptingBlockHash,
		IncludingIndex:     includingIndex,
	}

	result, err := deserializeTxIndexData(serializeTxIndexData(testdeserializedtxIndex))
	if err != nil {
		t.Fatalf("Failed deserializing txIndexData: %v", err)
	}
	if !testdeserializedtxIndex.IncludingBlockHash.Equal(result.IncludingBlockHash) {
		t.Fatalf("Expected including block hash: \n %s \n Got: \n %s\n", testdeserializedtxIndex.IncludingBlockHash.String(), result.IncludingBlockHash.String())
	} else if !testdeserializedtxIndex.AcceptingBlockHash.Equal(result.AcceptingBlockHash) {
		t.Fatalf("Expected accepting block hash \n %s \n Got: \n %s\n", testdeserializedtxIndex.AcceptingBlockHash.String(), result.AcceptingBlockHash.String())
	} else if testdeserializedtxIndex.IncludingIndex != result.IncludingIndex {
		t.Fatalf("Expected including index \n %d \n Got: \n %d\n", testdeserializedtxIndex.IncludingIndex, result.IncludingIndex)
	}
}
