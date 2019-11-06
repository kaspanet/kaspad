// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcec

import (
	"bytes"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"github.com/pkg/errors"
	"hash"
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
	// Used in RFC6979 implementation when testing the nonce for correctness
	one = big.NewInt(1)

	// oneInitializer is used to fill a byte slice with byte 0x01.  It is provided
	// here to avoid the need to create it multiple times.
	oneInitializer = []byte{0x01}
)

// Serialize returns a serialized signature (R and S concatenated).
func (sig *Signature) Serialize() []byte {
	return append(intTo32Bytes(sig.R), intTo32Bytes(sig.S)...)
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
	eBytes := sha256.Sum256(append(append(intTo32Bytes(r), pubKey.SerializeCompressed()...), hash...))
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

// intTo32Bytes pads a big int bytes with leading zeros if they
// are missing to get the length up to 32 bytes.
func intTo32Bytes(val *big.Int) []byte {
	b := val.Bytes()
	pad := bytes.Repeat([]byte{0x00}, 32-len(b))
	return append(pad, b...)
}

// hashToInt converts a hash value to an integer. There is some disagreement
// about how this is done. [NSA] suggests that this is done in the obvious
// manner, but [SECG] truncates the hash to the bit-length of the curve order
// first. We follow [SECG] because that's what OpenSSL does. Additionally,
// OpenSSL right shifts excess bits from the number if the hash is too large
// and we mirror that too.
// This is borrowed from crypto/ecdsa.
func hashToInt(hash []byte, c elliptic.Curve) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

// sign signs the hash using the schnorr signature algorithm.
func sign(privateKey *PrivateKey, hash []byte) (*Signature, error) {
	// The rfc6979 nonce derivation function accepts additional entropy.
	// See https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/2019-05-15-schnorr.md#recommended-practices-for-secure-signature-generation
	additionalData := []byte{'S', 'c', 'h', 'n', 'o', 'r', 'r', '+', 'S', 'H', 'A', '2', '5', '6', ' ', ' '}
	k := nonceRFC6979(privateKey.D, hash, additionalData)
	// Compute point R = k * G
	rx, ry := privateKey.Curve.ScalarBaseMult(k.Bytes())

	//  Negate nonce if R.y is not a quadratic residue.
	if big.Jacobi(ry, privateKey.Params().P) != 1 {
		k = k.Neg(k)
	}

	// Compute scalar e = Hash(R.x || compressed(P) || m) mod N
	eBytes := sha256.Sum256(append(append(intTo32Bytes(rx), privateKey.PubKey().SerializeCompressed()...), hash...))
	e := new(big.Int).SetBytes(eBytes[:])
	e.Mod(e, privateKey.Params().N)

	// Compute scalar s = (k + e * x) mod N
	x := new(big.Int).SetBytes(privateKey.Serialize())
	s := e.Mul(e, x)
	s.Add(s, k)
	s.Mod(s, privateKey.Params().N)
	return &Signature{
		R: rx,
		S: s,
	}, nil
}

// nonceRFC6979 generates a nonce (`k`) deterministically according to RFC 6979.
// It takes a 32-byte hash as an input and returns 32-byte nonce to be used in the digital signature algorithm.
func nonceRFC6979(privkey *big.Int, hash []byte, additionalData []byte) *big.Int {

	// Step A
	curve := S256()
	q := curve.Params().N
	x := privkey
	alg := sha256.New

	qlen := q.BitLen()
	holen := alg().Size()
	rolen := (qlen + 7) >> 3
	bx := append(int2octets(x, rolen), bits2octets(hash, curve, rolen)...)

	// Step B
	v := bytes.Repeat(oneInitializer, holen)

	// Step C (Go zeroes the all allocated memory)
	k := make([]byte, holen)

	// Step D
	k = mac(alg, k, append(append(append(v, 0x00), bx...), additionalData...))

	// Step E
	v = mac(alg, k, v)

	// Step F
	k = mac(alg, k, append(append(append(v, 0x01), bx...), additionalData...))

	// Step G
	v = mac(alg, k, v)

	// Step H
	for {
		// Step H1
		var t []byte

		// Step H2
		for len(t)*8 < qlen {
			v = mac(alg, k, v)
			t = append(t, v...)
		}

		// Step H3
		secret := hashToInt(t, curve)
		if secret.Cmp(one) >= 0 && secret.Cmp(q) < 0 {
			return secret
		}
		k = mac(alg, k, append(v, 0x00))
		v = mac(alg, k, v)
	}
}

// mac returns an HMAC of the given key and message.
func mac(alg func() hash.Hash, k, m []byte) []byte {
	h := hmac.New(alg, k)
	h.Write(m)
	return h.Sum(nil)
}

// https://tools.ietf.org/html/rfc6979#section-2.3.3
func int2octets(v *big.Int, rolen int) []byte {
	out := v.Bytes()

	// left pad with zeros if it's too short
	if len(out) < rolen {
		out2 := make([]byte, rolen)
		copy(out2[rolen-len(out):], out)
		return out2
	}

	// drop most significant bytes if it's too long
	if len(out) > rolen {
		out2 := make([]byte, rolen)
		copy(out2, out[len(out)-rolen:])
		return out2
	}

	return out
}

func bits2octets(in []byte, curve elliptic.Curve, rolen int) []byte {
	// https://tools.ietf.org/html/rfc6979#section-2.3.4
	z1 := hashToInt(in, curve)
	z2 := new(big.Int).Sub(z1, curve.Params().N)
	if z2.Sign() < 0 {
		return int2octets(z1, rolen)
	}
	return int2octets(z2, rolen)
}
