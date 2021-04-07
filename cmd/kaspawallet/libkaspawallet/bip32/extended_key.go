package bip32

import (
	"bytes"
	"crypto/sha256"
	"github.com/pkg/errors"
)

type extendedKey struct {
	Version     [4]byte
	Depth       uint8
	Fingerprint [4]byte
	ChildNumber uint32
	ChainCode   [32]byte
}

const (
	versionSerializationLen     = 4
	depthSerializationLen       = 1
	fingerprintSerializationLen = 4
	childNumberSerializationLen = 4
	chainCodeSerializationLen   = 32
	keySerializationLen         = 33
	checkSumLen                 = 4
)

const extendedKeySerializationLen = versionSerializationLen +
	depthSerializationLen +
	fingerprintSerializationLen +
	childNumberSerializationLen +
	chainCodeSerializationLen +
	keySerializationLen +
	checkSumLen

func calcChecksum(data []byte) []byte {
	return doubleSha256(data)[:checkSumLen]
}

func doubleSha256(data []byte) []byte {
	sha1 := sha256.New()
	sha2 := sha256.New()
	sha1.Write(data)
	sha2.Write(sha1.Sum(nil))
	return sha2.Sum(nil)
}

func validateChecksum(data []byte) error {
	checksum := data[len(data)-checkSumLen:]
	expectedChecksum := calcChecksum(data[:len(data)-checkSumLen])
	if !bytes.Equal(expectedChecksum, checksum) {
		return errors.Errorf("expected checksum %x but got %x", expectedChecksum, checksum)
	}

	return nil
}
