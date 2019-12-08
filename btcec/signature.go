// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcec

import (
	"bytes"
	"crypto/sha256"
	"github.com/pkg/errors"
	"math/big"
)

// Errors returned by canonicalPadding.
var (
	errNegativeValue          = errors.New("value may be interpreted as negative")
	errExcessivelyPaddedValue = errors.New("value is excessively padded")
)

// Signature is a type representing a Schnorr signature.
type Signature struct {
	R *big.Int
	S *big.Int
}

var (
	zero = big.NewInt(0)
	// Group order in secp256k1
	n = S256().N
)

// Serialize returns a serialized signature (R and S concatenated).
func (sig *Signature) Serialize() []byte {
	return append(bigIntTo32Bytes(sig.R), bigIntTo32Bytes(sig.S)...)
}

// Verify verifies digital signatures. It returns true if the signature
// is valid, false otherwise.
func (sig *Signature) Verify(hash []byte, pubKey *PublicKey) bool {
	return verifySchnorr(pubKey, hash, sig.R, sig.S)
}

// verifySchnorr verifies the schnorr signature of the hash using the pubkey key.
// It returns true if the signature is valid, false otherwise.
func verifySchnorr(pubKey *PublicKey, hash []byte, r *big.Int, s *big.Int) bool {
	// This schnorr specification is specific to the secp256k1 curve so if the
	// provided curve is not a KoblitizCurve then we'll just return false.
	curve, ok := pubKey.Curve.(*KoblitzCurve)
	if !ok {
		return false
	}

	// Signature is invalid if s >= order or r >= p.
	if s.Cmp(curve.Params().N) >= 0 || r.Cmp(curve.Params().P) >= 0 {
		return false
	}

	// Compute scalar e = Hash(r || compressed(P) || m) mod N
	eBytes := sha256.Sum256(append(append(bigIntTo32Bytes(r), pubKey.SerializeCompressed()...), hash...))
	e := new(big.Int).SetBytes(eBytes[:])
	e.Mod(e, curve.Params().N)

	// Negate e
	e.Neg(e).Mod(e, curve.Params().N)

	// Compute point R = s * G - e * P.
	sgx, sgy, sgz := curve.scalarBaseMultJacobian(s.Bytes())
	epx, epy, epz := curve.scalarMultJacobian(pubKey.X, pubKey.Y, e.Bytes())
	rx, ry, rz := new(fieldVal), new(fieldVal), new(fieldVal)
	curve.addJacobian(sgx, sgy, sgz, epx, epy, epz, rx, ry, rz)

	// Check that R is not infinity
	if rz.Equals(new(fieldVal).SetInt(0)) {
		return false
	}

	// Check if R.y is quadratic residue
	yz := ry.Mul(rz).Normalize()
	b := yz.Bytes()
	if big.Jacobi(new(big.Int).SetBytes(b[:]), curve.P) != 1 {
		return false
	}

	// Check R values match
	// rx ≠ rz^2 * r mod p
	fieldR := new(fieldVal).SetByteSlice(r.Bytes())
	return rx.Normalize().Equals(rz.Square().Mul(fieldR).Normalize())
}

// IsEqual compares this Signature instance to the one passed, returning true
// if both Signatures are equivalent. A signature is equivalent to another, if
// they both have the same scalar value for R and S.
func (sig *Signature) IsEqual(otherSig *Signature) bool {
	return sig.R.Cmp(otherSig.R) == 0 &&
		sig.S.Cmp(otherSig.S) == 0
}

// ParseSignature parses a 64 byte schnorr signature into a Signature type.
func ParseSignature(sigStr []byte) (*Signature, error) {
	if len(sigStr) != 64 {
		return nil, errors.New("malformed schnorr signature: not 64 bytes")
	}
	bigR := new(big.Int).SetBytes(sigStr[:32])
	bigS := new(big.Int).SetBytes(sigStr[32:64])
	return &Signature{
		R: bigR,
		S: bigS,
	}, nil
}

// bigIntTo32Bytes pads a big int bytes with leading zeros if they
// are missing to get the length up to 32 bytes.
func bigIntTo32Bytes(val *big.Int) []byte {
	b := val.Bytes()
	pad := bytes.Repeat([]byte{0x00}, 32-len(b))
	return append(pad, b...)
}

// sign signs the hash using the schnorr signature algorithm.
func sign(privateKey *PrivateKey, hash []byte) (*Signature, error) {

	kBytes := sha256.Sum256(append(privateKey.Serialize(), hash...))
	k := new(big.Int).SetBytes(kBytes[:])
	// The modulo is fine because n is close enough to 2^256 that this is very very rare
	k.Mod(k, n)
	if k.Cmp(zero) == 0 {
		return nil, errors.New("Something bad happend with sha256. got 0")
	}
	// Compute point R = k * G
	rx, ry := privateKey.Curve.ScalarBaseMult(k.Bytes())

	//  Negate nonce if R.y is not a quadratic residue.
	if big.Jacobi(ry, privateKey.Params().P) != 1 {
		k = k.Neg(k)
	}

	// Compute scalar e = Hash(R.x || compressed(P) || m) mod N
	eBytes := sha256.Sum256(append(append(bigIntTo32Bytes(rx), privateKey.PubKey().SerializeCompressed()...), hash...))
	e := new(big.Int).SetBytes(eBytes[:])
	// The modulo is fine because n is close enough to 2^256 that this is very very rare
	e.Mod(e, n)

	// Compute scalar s = (k + e * x) mod N
	s := e.Mul(e, privateKey.D)
	s.Add(s, k)
	s.Mod(s, n)
	return &Signature{
		R: rx,
		S: s,
	}, nil
}
