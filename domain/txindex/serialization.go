package txindex

import (
	"encoding/binary"
	"io"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

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

const hashesLengthSize = 8

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

func deserializeTxIndexData(serializedTxIndexData []byte) (*TxData, error) {
	var err error

	deserializedTxIndexData := &TxData{}
	deserializedTxIndexData.IncludingBlockHash, err = externalapi.NewDomainHashFromByteSlice(serializedTxIndexData[:32])
	if err != nil {
		return nil, err
	}
	deserializedTxIndexData.AcceptingBlockHash, err = externalapi.NewDomainHashFromByteSlice(serializedTxIndexData[32:64])
	if err != nil {
		return nil, err
	}
	deserializedTxIndexData.IncludingIndex = binary.BigEndian.Uint32(serializedTxIndexData[64:68])

	return deserializedTxIndexData, nil
}

func serializeTxIndexData(blockTxIndexData *TxData) []byte {
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, blockTxIndexData.IncludingIndex)
	serializedTxIndexData := append(
		append(
			blockTxIndexData.IncludingBlockHash.ByteSlice(),
			blockTxIndexData.AcceptingBlockHash.ByteSlice()...,
		),
		indexBytes...,
	)
	return serializedTxIndexData
}
