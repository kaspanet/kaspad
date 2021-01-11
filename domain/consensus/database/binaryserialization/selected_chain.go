package binaryserialization

import "encoding/binary"

// SerializeChainBlockIndex serializes chain block index
func SerializeChainBlockIndex(index uint64) []byte {
	var keyBytes [8]byte
	binary.LittleEndian.PutUint64(keyBytes[:], index)
	return keyBytes[:]
}

// DeserializeChainBlockIndex deserializes chain block index to uint64
func DeserializeChainBlockIndex(indexBytes []byte) uint64 {
	return binary.LittleEndian.Uint64(indexBytes)
}
