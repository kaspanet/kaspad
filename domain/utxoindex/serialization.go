package utxoindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"io"
)

func (uis *utxoIndexStore) serializeOutpoint(outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	dbOutpoint := serialization.DomainOutpointToDbOutpoint(outpoint)
	return proto.Marshal(dbOutpoint)
}

func (uis *utxoIndexStore) deserializeOutpoint(serializedOutpoint []byte) (*externalapi.DomainOutpoint, error) {
	var dbOutpoint serialization.DbOutpoint
	err := proto.Unmarshal(serializedOutpoint, &dbOutpoint)
	if err != nil {
		return nil, err
	}
	return serialization.DbOutpointToDomainOutpoint(&dbOutpoint)
}

func (uis *utxoIndexStore) serializeUTXOEntry(utxoEntry externalapi.UTXOEntry) ([]byte, error) {
	dbUTXOEntry := serialization.UTXOEntryToDBUTXOEntry(utxoEntry)
	return proto.Marshal(dbUTXOEntry)
}

func (uis *utxoIndexStore) deserializeUTXOEntry(serializedUTXOEntry []byte) (externalapi.UTXOEntry, error) {
	var dbUTXOEntry serialization.DbUtxoEntry
	err := proto.Unmarshal(serializedUTXOEntry, &dbUTXOEntry)
	if err != nil {
		return nil, err
	}
	return serialization.DBUTXOEntryToUTXOEntry(&dbUTXOEntry)
}

const hashesLengthSize = 8

func (uis *utxoIndexStore) serializeHashes(hashes []*externalapi.DomainHash) []byte {
	serializedHashes := make([]byte, hashesLengthSize+externalapi.DomainHashSize*len(hashes))
	binary.LittleEndian.PutUint64(serializedHashes[:hashesLengthSize], uint64(len(hashes)))
	for i, hash := range hashes {
		start := hashesLengthSize + externalapi.DomainHashSize*i
		end := start + externalapi.DomainHashSize
		copy(serializedHashes[start:end],
			hash.ByteSlice())
	}
	return serializedHashes
}

func (uis *utxoIndexStore) deserializeHashes(serializedHashes []byte) ([]*externalapi.DomainHash, error) {
	length := binary.LittleEndian.Uint64(serializedHashes[:hashesLengthSize])
	hashes := make([]*externalapi.DomainHash, length)
	for i := uint64(0); i < length; i++ {
		start := hashesLengthSize + externalapi.DomainHashSize*i
		end := start + externalapi.DomainHashSize

		if end > uint64(len(serializedHashes)) {
			return nil, errors.Wrapf(io.ErrUnexpectedEOF, "unexpected EOF while deserializing hashes")
		}

		var err error
		hashes[i], err = externalapi.
			NewDomainHashFromByteSlice(serializedHashes[start:end])
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
