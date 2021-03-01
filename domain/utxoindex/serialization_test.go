package utxoindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"testing"
)

func Test_serializeHashes(t *testing.T) {
	r := rand.New(rand.NewSource(0))

	for length := 0; length < 32; length++ {
		hashes := make([]*externalapi.DomainHash, length)
		for i := range hashes {
			var hashBytes [32]byte
			r.Read(hashBytes[:])
			hashes[i] = externalapi.NewDomainHashFromByteArray(&hashBytes)
		}
		result, err := deserializeHashes(serializeHashes(hashes))
		if err != nil {
			t.Fatalf("Failed deserializing hashes: %v", err)
		}
		if !externalapi.HashesEqual(hashes, result) {
			t.Fatalf("Expected \n %s \n==\n %s\n", hashes, result)
		}
	}
}

func Test_deserializeHashesFailure(t *testing.T) {
	hashes := []*externalapi.DomainHash{
		externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
		externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
		externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
	}
	serialized := serializeHashes(hashes)
	binary.LittleEndian.PutUint64(serialized[:8], uint64(len(hashes)+1))
	_, err := deserializeHashes(serialized)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Expected error to be EOF, instead got: %v", err)
	}
}
