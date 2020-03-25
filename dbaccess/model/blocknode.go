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
	Header             wire.BlockHeader
	Status             byte
	SelectedParentHash daghash.Hash
	BlueScore          uint64
	BlueHashes         []daghash.Hash
	BluesAnticoneSizes map[daghash.Hash]dagconfig.KType
}

func SerializeBlockNode(node *BlockNode) ([]byte, error) {
	w := bytes.NewBuffer(make([]byte, 0, blockNodeSerializeSize))

	// Serialize Header
	err := node.Header.Serialize(w)
	if err != nil {
		return nil, err
	}

	// Serialize Status
	err = w.WriteByte(node.Status)
	if err != nil {
		return nil, err
	}

	// Serialize SelectedParentHash
	_, err = w.Write(node.SelectedParentHash[:])
	if err != nil {
		return nil, err
	}

	// Serialize BlueScore
	err = binaryserializer.PutUint64(w, byteOrder, node.BlueScore)
	if err != nil {
		return nil, err
	}

	// Serialize BlueHashes
	err = binaryserializer.PutUint64(w, byteOrder, uint64(len(node.BlueHashes)))
	if err != nil {
		return nil, err
	}
	for _, blueHash := range node.BlueHashes {
		_, err = w.Write(blueHash[:])
		if err != nil {
			return nil, err
		}
	}

	// Serialize BluesAnticoneSizes
	err = binaryserializer.PutUint64(w, byteOrder, uint64(len(node.BluesAnticoneSizes)))
	if err != nil {
		return nil, err
	}
	for blueHash, blueAnticoneSize := range node.BluesAnticoneSizes {
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

	// Deserialize Header
	var header wire.BlockHeader
	err := header.Deserialize(buffer)
	if err != nil {
		return nil, err
	}
	node.Header = header

	// Deserialize Status
	status, err := buffer.ReadByte()
	if err != nil {
		return nil, err
	}
	node.Status = status

	// Deserialize SelectedParentHash
	selectedParentHash := daghash.Hash{}
	_, err = io.ReadFull(buffer, selectedParentHash[:])
	if err != nil {
		return nil, err
	}
	node.SelectedParentHash = selectedParentHash

	// Deserialize BlueScore
	blueScore, err := binaryserializer.Uint64(buffer, byteOrder)
	if err != nil {
		return nil, err
	}
	node.BlueScore = blueScore

	// Deserialize BlueHashes
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
	node.BlueHashes = blueHashes

	// Deserialize BluesAnticoneSizes
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
	node.BluesAnticoneSizes = bluesAnticoneSizes

	return node, nil
}
