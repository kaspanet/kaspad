// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcec

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
)

// PrivateKey wraps an ecdsa.PrivateKey as a convenience mainly for signing
// things with the the private key without having to directly import the ecdsa
// package.
type PrivateKey ecdsa.PrivateKey

// PrivKeyFromBytes returns a private and public key for `curve' based on the
// private key passed as an argument as a byte slice.
func PrivKeyFromBytes(pk []byte) (*PrivateKey,
	*PublicKey) {
	curve := S256()
	x, y := curve.ScalarBaseMult(pk)

	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(pk),
	}

	return (*PrivateKey)(priv), (*PublicKey)(&priv.PublicKey)
}

// NewPrivateKey is a wrapper for ecdsa.GenerateKey that returns a PrivateKey
// instead of the normal ecdsa.PrivateKey.
func NewPrivateKey() (*PrivateKey, error) {
	key, err := ecdsa.GenerateKey(S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return (*PrivateKey)(key), nil
}

// PubKey returns the PublicKey corresponding to this private key.
func (p *PrivateKey) PubKey() *PublicKey {
	return (*PublicKey)(&p.PublicKey)
}

// ToECDSA returns the private key as a *ecdsa.PrivateKey.
func (p *PrivateKey) ToECDSA() *ecdsa.PrivateKey {
	return (*ecdsa.PrivateKey)(p)
}

// Sign generates a Schnorr signature for the provided hash (which should be the result
// of hashing a larger message) using the private key. Produced signature
// is deterministic (same message and same key yield the same signature).
func (p *PrivateKey) Sign(hash []byte) (*Signature, error) {
	return sign(p, hash)
}

// PrivKeyBytesLen defines the length in bytes of a serialized private key.
const PrivKeyBytesLen = 32

// Serialize returns the private key number d as a big-endian binary-encoded
// number, padded to a length of 32 bytes.
func (p *PrivateKey) Serialize() []byte {
	b := make([]byte, 0, PrivKeyBytesLen)
	return paddedAppend(PrivKeyBytesLen, b, p.ToECDSA().D.Bytes())
}
