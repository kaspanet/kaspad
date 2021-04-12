package bip32

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"github.com/pkg/errors"
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
