package bip32

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ripemd160"
	"hash"
)

func newHMACWriter(key []byte) hmacWriter {
	return hmacWriter{
		Hash: hmac.New(sha512.New, key),
	}
}

type hmacWriter struct {
	hash.Hash
}

func (hw hmacWriter) InfallibleWrite(p []byte) {
	_, err := hw.Write(p)
	if err != nil {
		panic(errors.Wrap(err, "writing to hmac should never fail"))
	}
}

func calcChecksum(data []byte) []byte {
	return doubleSha256(data)[:checkSumLen]
}

func doubleSha256(data []byte) []byte {
	inner := sha256.Sum256(data)
	outer := sha256.Sum256(inner[:])
	return outer[:]
}

// validateChecksum validates that the last checkSumLen bytes of the
// given data are its valid checksum.
func validateChecksum(data []byte) error {
	checksum := data[len(data)-checkSumLen:]
	expectedChecksum := calcChecksum(data[:len(data)-checkSumLen])
	if !bytes.Equal(expectedChecksum, checksum) {
		return errors.Errorf("expected checksum %x but got %x", expectedChecksum, checksum)
	}

	return nil
}

func hash160(data []byte) []byte {
	sha := sha256.New()
	ripe := ripemd160.New()
	sha.Write(data)
	ripe.Write(sha.Sum(nil))
	return ripe.Sum(nil)
}
