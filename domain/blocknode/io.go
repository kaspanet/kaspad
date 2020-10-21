package blocknode

import (
	"bytes"
	"encoding/binary"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/kaspanet/kaspad/util/daghash"
)

var (
	// byteOrder is the preferred byte order used for serializing numeric
	// fields for storage in the database.
	byteOrder = binary.LittleEndian
)

// SerializeNode serializes the given node
func SerializeNode(node *Node) ([]byte, error) {
	w := bytes.NewBuffer(make([]byte, 0, appmessage.MaxBlockHeaderPayload+1))
	header := node.Header()
	err := header.Serialize(w)
	if err != nil {
		return nil, err
	}

	err = w.WriteByte(byte(node.Status))
	if err != nil {
		return nil, err
	}

	// Because genesis doesn't have selected parent, it's serialized as zero hash
	selectedParentHash := &daghash.ZeroHash
	if node.SelectedParent != nil {
		selectedParentHash = node.SelectedParent.Hash
	}
	_, err = w.Write(selectedParentHash[:])
	if err != nil {
		return nil, err
	}

	err = binaryserializer.PutUint64(w, byteOrder, node.BlueScore)
	if err != nil {
		return nil, err
	}

	err = appmessage.WriteVarInt(w, uint64(len(node.Blues)))
	if err != nil {
		return nil, err
	}

	for _, blue := range node.Blues {
		_, err = w.Write(blue.Hash[:])
		if err != nil {
			return nil, err
		}
	}

	err = appmessage.WriteVarInt(w, uint64(len(node.Reds)))
	if err != nil {
		return nil, err
	}

	for _, red := range node.Reds {
		_, err = w.Write(red.Hash[:])
		if err != nil {
			return nil, err
		}
	}

	err = appmessage.WriteVarInt(w, uint64(len(node.BluesAnticoneSizes)))
	if err != nil {
		return nil, err
	}
	for blue, blueAnticoneSize := range node.BluesAnticoneSizes {
		_, err = w.Write(blue.Hash[:])
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
