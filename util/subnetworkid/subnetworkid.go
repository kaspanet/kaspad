// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package subnetworkid

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// IDLength of array used to store the sub-network ID.  See SubNetworkID.
const IDLength = 20

// MaxStringSize is the maximum length of a SubNetworkID string.
const MaxStringSize = IDLength * 2

// ErrIDStrSize describes an error that indicates the caller specified an ID
// string that has too many characters.
var ErrIDStrSize = fmt.Errorf("max ID string length is %v bytes", MaxStringSize)

// SubNetworkID is used in several of the bitcoin messages and common structures.  It
// typically represents ripmed160(sha256(data)).
type SubNetworkID [IDLength]byte

// String returns the SubNetworkID as the hexadecimal string of the byte-reversed
// hash.
func (id SubNetworkID) String() string {
	for i := 0; i < IDLength/2; i++ {
		id[i], id[IDLength-1-i] = id[IDLength-1-i], id[i]
	}
	return hex.EncodeToString(id[:])
}

// Strings returns a slice of strings representing the IDs in the given slice of IDs
func Strings(ids []SubNetworkID) []string {
	strings := make([]string, len(ids))
	for i, id := range ids {
		strings[i] = id.String()
	}

	return strings
}

// CloneBytes returns a copy of the bytes which represent the ID as a byte
// slice.
//
// NOTE: It is generally cheaper to just slice the ID directly thereby reusing
// the same bytes rather than calling this method.
func (id *SubNetworkID) CloneBytes() []byte {
	newID := make([]byte, IDLength)
	copy(newID, id[:])

	return newID
}

// SetBytes sets the bytes which represent the ID.  An error is returned if
// the number of bytes passed in is not IDLength.
func (id *SubNetworkID) SetBytes(newID []byte) error {
	nhlen := len(newID)
	if nhlen != IDLength {
		return fmt.Errorf("invalid ID length of %v, want %v", nhlen,
			IDLength)
	}
	copy(id[:], newID)

	return nil
}

// IsEqual returns true if target is the same as ID.
func (id *SubNetworkID) IsEqual(target *SubNetworkID) bool {
	if id == nil && target == nil {
		return true
	}
	if id == nil || target == nil {
		return false
	}
	return *id == *target
}

// AreEqual returns true if both slices contain the same IDs.
// Either slice must not contain duplicates.
func AreEqual(first []SubNetworkID, second []SubNetworkID) bool {
	if len(first) != len(second) {
		return false
	}

	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}

	return true
}

// New returns a new ID from a byte slice.  An error is returned if
// the number of bytes passed in is not IDLength.
func New(newID []byte) (*SubNetworkID, error) {
	var sh SubNetworkID
	err := sh.SetBytes(newID)
	if err != nil {
		return nil, err
	}
	return &sh, err
}

// NewFromStr creates a SubNetworkID from a string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the SubNetworkID.
func NewFromStr(id string) (*SubNetworkID, error) {
	ret := new(SubNetworkID)
	err := Decode(ret, id)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Decode decodes the byte-reversed hexadecimal string encoding of a SubNetworkID to a
// destination.
func Decode(dst *SubNetworkID, src string) error {
	// Return error if ID string is too long.
	if len(src) > MaxStringSize {
		return ErrIDStrSize
	}

	// Hex decoder expects the ID to be a multiple of two.  When not, pad
	// with a leading zero.
	var srcBytes []byte
	if len(src)%2 == 0 {
		srcBytes = []byte(src)
	} else {
		srcBytes = make([]byte, 1+len(src))
		srcBytes[0] = '0'
		copy(srcBytes[1:], src)
	}

	// Hex decode the source bytes to a temporary destination.
	var reversedHash SubNetworkID
	_, err := hex.Decode(reversedHash[IDLength-hex.DecodedLen(len(srcBytes)):], srcBytes)
	if err != nil {
		return err
	}

	// Reverse copy from the temporary hash to destination.  Because the
	// temporary was zeroed, the written result will be correctly padded.
	for i, b := range reversedHash[:IDLength/2] {
		dst[i], dst[IDLength-1-i] = reversedHash[IDLength-1-i], b
	}

	return nil
}

// ToBig converts a SubNetworkID into a big.Int that can be used to
// perform math comparisons.
func ToBig(id *SubNetworkID) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	buf := *id
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

// Cmp compares id and target and returns:
//
//   -1 if id <  target
//    0 if id == target
//   +1 if id >  target
//
func (id *SubNetworkID) Cmp(target *SubNetworkID) int {
	return ToBig(id).Cmp(ToBig(target))
}

// Less returns true iff id a is less than id b
func Less(a *SubNetworkID, b *SubNetworkID) bool {
	return a.Cmp(b) < 0
}

// JoinIDsStrings joins all the stringified IDs separated by a separator
func JoinIDsStrings(ids []SubNetworkID, separator string) string {
	return strings.Join(Strings(ids), separator)
}

// Sort sorts a slice of ids
func Sort(ids []SubNetworkID) {
	sort.Slice(ids, func(i, j int) bool {
		return Less(&ids[i], &ids[j])
	})
}
