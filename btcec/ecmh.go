package btcec

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"

	"github.com/daglabs/btcd/util/daghash"
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
}

// NewMultiset returns an empty multiset. The hash of an empty set
// is the 32 byte value of zero.
func NewMultiset(curve *KoblitzCurve) *Multiset {
	return &Multiset{curve: curve, x: big.NewInt(0), y: big.NewInt(0)}
}

// NewMultisetFromPoint initializes a new multiset with the given x, y
// coordinate.
func NewMultisetFromPoint(curve *KoblitzCurve, x, y *big.Int) *Multiset {
	var copyX, copyY big.Int
	if x != nil {
		copyX.Set(x)
	}
	if y != nil {
		copyY.Set(y)
	}
	return &Multiset{curve: curve, x: &copyX, y: &copyY}
}

// NewMultisetFromDataSlice gets a curve and a slice of byte
// slices, creates an empty multiset, hashes each data and
// add it to the multiset, and return the resulting multiset.
func NewMultisetFromDataSlice(curve *KoblitzCurve, datas [][]byte) *Multiset {
	ms := NewMultiset(curve)
	for _, data := range datas {
		x, y := hashToPoint(curve, data)
		ms.addPoint(x, y)
	}
	return ms
}

// Clone returns a clone of this multiset.
func (ms *Multiset) Clone() *Multiset {
	return NewMultisetFromPoint(ms.curve, ms.x, ms.y)
}

// Add hashes the data onto the curve and returns
// a multiset with the new resulting point.
func (ms *Multiset) Add(data []byte) *Multiset {
	newMs := ms.Clone()
	x, y := hashToPoint(ms.curve, data)
	newMs.addPoint(x, y)
	return newMs
}

func (ms *Multiset) addPoint(x, y *big.Int) {
	ms.x, ms.y = ms.curve.Add(ms.x, ms.y, x, y)
}

// Remove hashes the data onto the curve, subtracts
// the point from the existing multiset, and returns
// a multiset with the new point. This function
// will execute regardless of whether or not the passed
// data was previously added to the set. Hence if you
// remove an element that was never added and also remove
// all the elements that were added, you will not get
// back to the point at infinity (empty set).
func (ms *Multiset) Remove(data []byte) *Multiset {
	newMs := ms.Clone()
	x, y := hashToPoint(ms.curve, data)
	newMs.removePoint(x, y)
	return newMs
}

func (ms *Multiset) removePoint(x, y *big.Int) {
	y.Neg(y).Mod(y, ms.curve.P)
	ms.x, ms.y = ms.curve.Add(ms.x, ms.y, x, y)
}

// Union will add the point of the passed multiset instance to the point
// of this multiset and will return a multiset with the resulting point.
func (ms *Multiset) Union(otherMultiset *Multiset) *Multiset {
	newMs := ms.Clone()
	otherMsCopy := otherMultiset.Clone()
	newMs.addPoint(otherMsCopy.x, otherMsCopy.y)
	return newMs
}

// Subtract will remove the point of the passed multiset instance from the point
// of this multiset and will return a multiset with the resulting point.
func (ms *Multiset) Subtract(otherMultiset *Multiset) *Multiset {
	newMs := ms.Clone()
	otherMsCopy := otherMultiset.Clone()
	newMs.removePoint(otherMsCopy.x, otherMsCopy.y)
	return newMs
}

// Hash serializes and returns the hash of the multiset. The hash of an empty
// set is the 32 byte value of zero. The hash of a non-empty multiset is the
// sha256 hash of the 32 byte x value concatenated with the 32 byte y value.
func (ms *Multiset) Hash() daghash.Hash {
	if ms.x.Sign() == 0 && ms.y.Sign() == 0 {
		return daghash.Hash{}
	}

	hash := sha256.Sum256(append(ms.x.Bytes(), ms.y.Bytes()...))
	return daghash.Hash(hash)
}

// Point returns a copy of the x and y coordinates of the current multiset state.
func (ms *Multiset) Point() (x *big.Int, y *big.Int) {
	var copyX, copyY big.Int
	copyX.Set(ms.x)
	copyY.Set(ms.y)
	return &copyX, &copyY
}

// hashToPoint hashes the passed data into a point on the curve. The x value
// is sha256(n, sha256(data)) where n starts at zero. If the resulting x value
// is not in the field or x^3+7 is not quadratic residue then n is incremented
// and we try again. There is a 50% chance of success for any given iteration.
func hashToPoint(curve *KoblitzCurve, data []byte) (x *big.Int, y *big.Int) {
	i := uint64(0)
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
