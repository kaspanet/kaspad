package model

import (
	"bytes"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"io"
)

const blockNodeSerializeSize = wire.MaxBlockHeaderPayload

type BlockNode struct {
	header             wire.BlockHeader
	status             byte
	selectedParentHash daghash.Hash
	blueScore          uint64
	blueHashes         []daghash.Hash
	bluesAnticoneSizes map[daghash.Hash]dagconfig.KType
}

func SerializeBlockNode(node *BlockNode) ([]byte, error) {
	w := bytes.NewBuffer(make([]byte, 0, blockNodeSerializeSize))

	// Serialize the header
	err := node.header.Serialize(w)
	if err != nil {
		return nil, err
	}

	// Serialize the status
	err = w.WriteByte(node.status)
	if err != nil {
		return nil, err
	}

	// Serialize the selectedParentHash
	_, err = w.Write(node.selectedParentHash[:])
	if err != nil {
		return nil, err
	}

	// Serialize the blueScore
	err = binaryserializer.PutUint64(w, byteOrder, node.blueScore)
	if err != nil {
		return nil, err
	}

	// Serialize the blueHashes
	err = binaryserializer.PutUint64(w, byteOrder, uint64(len(node.blueHashes)))
	if err != nil {
		return nil, err
	}
	for _, blueHash := range node.blueHashes {
		_, err = w.Write(blueHash[:])
		if err != nil {
			return nil, err
		}
	}

	// Serialize the bluesAnticoneSizes
	err = binaryserializer.PutUint64(w, byteOrder, uint64(len(node.bluesAnticoneSizes)))
	if err != nil {
		return nil, err
	}
	for blueHash, blueAnticoneSize := range node.bluesAnticoneSizes {
		_, err = w.Write(blueHash[:])
		if err != nil {
			return nil, err
		}
		err = binaryserializer.PutUint8(w, uint8(blueAnticoneSize))
		if err != nil {
			return nil, err
		}
	}

	return w.Bytes(), nil
}

func DeserializeBlockNode(serializedBlockNode []byte) (*BlockNode, error) {
	buffer := bytes.NewReader(serializedBlockNode)
	node := &BlockNode{}

	// Deserialize the header
	var header wire.BlockHeader
	err := header.Deserialize(buffer)
	if err != nil {
		return nil, err
	}
	node.header = header

	// Deserialize the status
	status, err := buffer.ReadByte()
	if err != nil {
		return nil, err
	}
	node.status = status

	// Deserialize the selectedParentHash
	selectedParentHash := daghash.Hash{}
	_, err = io.ReadFull(buffer, selectedParentHash[:])
	if err != nil {
		return nil, err
	}
	node.selectedParentHash = selectedParentHash

	// Deserialize the blueScore
	blueScore, err := binaryserializer.Uint64(buffer, byteOrder)
	if err != nil {
		return nil, err
	}
	node.blueScore = blueScore

	// Deserialize the blueHashes
	blueHashesCount, err := binaryserializer.Uint64(buffer, byteOrder)
	if err != nil {
		return nil, err
	}
	blueHashes := make([]daghash.Hash, blueHashesCount)
	for i := uint64(0); i < blueHashesCount; i++ {
		blueHash := daghash.Hash{}
		_, err := io.ReadFull(buffer, blueHash[:])
		if err != nil {
			return nil, err
		}
		blueHashes[i] = blueHash
	}
	node.blueHashes = blueHashes

	// Deserialize the bluesAnticoneSizes
	bluesAnticoneSizesCount, err := binaryserializer.Uint64(buffer, byteOrder)
	if err != nil {
		return nil, err
	}
	bluesAnticoneSizes := make(map[daghash.Hash]dagconfig.KType, bluesAnticoneSizesCount)
	for i := uint64(0); i < bluesAnticoneSizesCount; i++ {
		blueHash := daghash.Hash{}
		_, err := io.ReadFull(buffer, blueHash[:])
		if err != nil {
			return nil, err
		}
		bluesAnticoneSize, err := binaryserializer.Uint8(buffer)
		if err != nil {
			return nil, err
		}
		bluesAnticoneSizes[blueHash] = dagconfig.KType(bluesAnticoneSize)
	}
	node.bluesAnticoneSizes = bluesAnticoneSizes

	return node, nil
}
