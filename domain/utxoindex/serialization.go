package utxoindex

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/domain/consensus/database/serialization"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"google.golang.org/protobuf/proto"
)

func serializeOutpoint(outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	dbOutpoint := serialization.DomainOutpointToDbOutpoint(outpoint)
	return proto.Marshal(dbOutpoint)
}

func deserializeOutpoint(serializedOutpoint []byte) (*externalapi.DomainOutpoint, error) {
	var dbOutpoint serialization.DbOutpoint
	err := proto.Unmarshal(serializedOutpoint, &dbOutpoint)
	if err != nil {
		return nil, err
	}
	return serialization.DbOutpointToDomainOutpoint(&dbOutpoint)
}

func serializeUTXOEntry(utxoEntry externalapi.UTXOEntry) ([]byte, error) {
	dbUTXOEntry := serialization.UTXOEntryToDBUTXOEntry(utxoEntry)
	return proto.Marshal(dbUTXOEntry)
}

func deserializeUTXOEntry(serializedUTXOEntry []byte) (externalapi.UTXOEntry, error) {
	var dbUTXOEntry serialization.DbUtxoEntry
	err := proto.Unmarshal(serializedUTXOEntry, &dbUTXOEntry)
	if err != nil {
		return nil, err
	}
	return serialization.DBUTXOEntryToUTXOEntry(&dbUTXOEntry)
}

const hashesLengthSize = 8

func serializeHashes(hashes []*externalapi.DomainHash) []byte {
	serializedHashes := make([]byte, hashesLengthSize+externalapi.DomainHashSize*len(hashes))
	binary.LittleEndian.PutUint64(serializedHashes[:hashesLengthSize], uint64(len(hashes)))
	for i, hash := range hashes {
		start := hashesLengthSize + externalapi.DomainHashSize*i
		end := start + externalapi.DomainHashSize
		copy(serializedHashes[start:end], hash.ByteSlice())
	}
	return serializedHashes
}

func deserializeHashes(serializedHashes []byte) ([]*externalapi.DomainHash, error) {
	length := binary.LittleEndian.Uint64(serializedHashes[:hashesLengthSize])
	hashes := make([]*externalapi.DomainHash, length)
	for i := uint64(0); i < length; i++ {
		start := hashesLengthSize + externalapi.DomainHashSize*i
		end := start + externalapi.DomainHashSize

		if end > uint64(len(serializedHashes)) {
			return nil, errors.Wrapf(io.ErrUnexpectedEOF, "unexpected EOF while deserializing hashes")
		}

		var err error
		hashes[i], err = externalapi.NewDomainHashFromByteSlice(serializedHashes[start:end])
		if err != nil {
			return nil, err
		}
	}

	return hashes, nil
}
