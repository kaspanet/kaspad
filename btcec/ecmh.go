package btcec

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"sync"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

// Multiset tracks the state of a multiset as used to calculate the ECMH
// (elliptic curve multiset hash) hash of an unordered set. The state is
// a point on the curve. New elements are hashed onto a point on the curve
// and then added to the current state. Hence elements can be added in any
// order and we can also remove elements to return to a prior hash.
type Multiset struct {
	curve *KoblitzCurve
	x     *big.Int
	y     *big.Int
	mtx   sync.RWMutex
}

// NewMultiset returns an empty multiset. The hash of an empty set
// is the 32 byte value of zero.
func NewMultiset(curve *KoblitzCurve) *Multiset {
	return &Multiset{curve: curve, x: big.NewInt(0), y: big.NewInt(0), mtx: sync.RWMutex{}}
}

// NewMultisetFromPoint initializes a new multiset with the given x, y
// coordinate.
func NewMultisetFromPoint(curve *KoblitzCurve, x, y *big.Int) *Multiset {
	var copyX, copyY big.Int
	if x != nil {
		copyX = *x
	}
	if y != nil {
		copyY = *y
	}
	return &Multiset{curve: curve, x: &copyX, y: &copyY, mtx: sync.RWMutex{}}
}

// Add hashes the data onto the curve and updates the state
// of the multiset.
func (ms *Multiset) Add(data []byte) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	x, y := hashToPoint(ms.curve, data)
	ms.x, ms.y = ms.curve.Add(ms.x, ms.y, x, y)
}

// Remove hashes the data onto the curve and subtracts the value
// from the state. This function will execute regardless of whether
// or not the passed data was previously added to the set. Hence if
// you remove an element that was never added and also remove all the
// elements that were added, you will not get back to the point at
// infinity (empty set).
func (ms *Multiset) Remove(data []byte) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	x, y := hashToPoint(ms.curve, data)
	y = y.Neg(y).Mod(y, ms.curve.P)
	ms.x, ms.y = ms.curve.Add(ms.x, ms.y, x, y)
}

// Merge will add the point of the passed in multiset instance to the point
// of this multiset and save the new point in this instance.
func (ms *Multiset) Merge(otherMultiset *Multiset) {
	ms.x, ms.y = ms.curve.Add(ms.x, ms.y, otherMultiset.x, otherMultiset.y)
}

// Hash serializes and returns the hash of the multiset. The hash of an empty
// set is the 32 byte value of zero. The hash of a non-empty multiset is the
// sha256 hash of the 32 byte x value concatenated with the 32 byte y value.
func (ms *Multiset) Hash() daghash.Hash {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()

	if ms.x.Sign() == 0 && ms.y.Sign() == 0 {
		return daghash.Hash{}
	}

	h := sha256.Sum256(append(ms.x.Bytes(), ms.y.Bytes()...))
	var reversed [32]byte
	for i, b := range h {
		reversed[len(h)-i-1] = b
	}
	return daghash.Hash(reversed)
}

// Point returns a copy of the x and y coordinates of the current multiset state.
func (ms *Multiset) Point() (x *big.Int, y *big.Int) {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()

	copyX, copyY := *ms.x, *ms.y
	return &copyX, &copyY
}

// hashToPoint hashes the passed data into a point on the curve. The x value
// is sha256(n, sha256(data)) where n starts at zero. If the resulting x value
// is not in the field or x^3+7 is not quadratic residue then n is incremented
// and we try again. There is a 50% chance of success for any given iteration.
func hashToPoint(curve *KoblitzCurve, data []byte) (*big.Int, *big.Int) {
	i := uint64(0)
	var x, y *big.Int
	var err error
	h := sha256.Sum256(data)
	n := make([]byte, 8)
	for {
		binary.LittleEndian.PutUint64(n, i)
		h2 := sha256.Sum256(append(n, h[:]...))

		x = new(big.Int).SetBytes(h2[:])

		y, err = decompressPoint(curve, x, false)
		if err == nil && x.Cmp(curve.N) < 0 {
			break
		}
		i++
	}
	return x, y
}
