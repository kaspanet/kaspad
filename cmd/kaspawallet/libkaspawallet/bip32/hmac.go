package bip32

import (
	"crypto/hmac"
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
