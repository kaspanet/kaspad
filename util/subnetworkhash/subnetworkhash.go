// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package subnetworkhash

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// HashSize of array used to store hashes.  See Hash.
const HashSize = 20

// MaxHashStringSize is the maximum length of a Hash hash string.
const MaxHashStringSize = HashSize * 2

// ErrHashStrSize describes an error that indicates the caller specified a hash
// string that has too many characters.
var ErrHashStrSize = fmt.Errorf("max hash string length is %v bytes", MaxHashStringSize)

// SubNetworkHash is used in several of the bitcoin messages and common structures.  It
// typically represents the double sha256 of data.
type SubNetworkHash [HashSize]byte

// String returns the Hash as the hexadecimal string of the byte-reversed
// hash.
func (hash SubNetworkHash) String() string {
	for i := 0; i < HashSize/2; i++ {
		hash[i], hash[HashSize-1-i] = hash[HashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}

// Strings returns a slice of strings representing the hashes in the given slice of hashes
func Strings(hashes []SubNetworkHash) []string {
	strings := make([]string, len(hashes))
	for i, hash := range hashes {
		strings[i] = hash.String()
	}

	return strings
}

// CloneBytes returns a copy of the bytes which represent the hash as a byte
// slice.
//
// NOTE: It is generally cheaper to just slice the hash directly thereby reusing
// the same bytes rather than calling this method.
func (hash *SubNetworkHash) CloneBytes() []byte {
	newHash := make([]byte, HashSize)
	copy(newHash, hash[:])

	return newHash
}

// SetBytes sets the bytes which represent the hash.  An error is returned if
// the number of bytes passed in is not HashSize.
func (hash *SubNetworkHash) SetBytes(newHash []byte) error {
	nhlen := len(newHash)
	if nhlen != HashSize {
		return fmt.Errorf("invalid hash length of %v, want %v", nhlen,
			HashSize)
	}
	copy(hash[:], newHash)

	return nil
}

// IsEqual returns true if target is the same as hash.
func (hash *SubNetworkHash) IsEqual(target *SubNetworkHash) bool {
	if hash == nil && target == nil {
		return true
	}
	if hash == nil || target == nil {
		return false
	}
	return *hash == *target
}

// AreEqual returns true if both slices contain the same hashes.
// Either slice must not contain duplicates.
func AreEqual(first []SubNetworkHash, second []SubNetworkHash) bool {
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

// NewHash returns a new Hash from a byte slice.  An error is returned if
// the number of bytes passed in is not HashSize.
func NewHash(newHash []byte) (*SubNetworkHash, error) {
	var sh SubNetworkHash
	err := sh.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &sh, err
}

// NewHashFromStr creates a Hash from a hash string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the Hash.
func NewHashFromStr(hash string) (*SubNetworkHash, error) {
	ret := new(SubNetworkHash)
	err := Decode(ret, hash)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Decode decodes the byte-reversed hexadecimal string encoding of a Hash to a
// destination.
func Decode(dst *SubNetworkHash, src string) error {
	// Return error if hash string is too long.
	if len(src) > MaxHashStringSize {
		return ErrHashStrSize
	}

	// Hex decoder expects the hash to be a multiple of two.  When not, pad
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
	var reversedHash SubNetworkHash
	_, err := hex.Decode(reversedHash[HashSize-hex.DecodedLen(len(srcBytes)):], srcBytes)
	if err != nil {
		return err
	}

	// Reverse copy from the temporary hash to destination.  Because the
	// temporary was zeroed, the written result will be correctly padded.
	for i, b := range reversedHash[:HashSize/2] {
		dst[i], dst[HashSize-1-i] = reversedHash[HashSize-1-i], b
	}

	return nil
}

// HashToBig converts a daghash.Hash into a big.Int that can be used to
// perform math comparisons.
func HashToBig(hash *SubNetworkHash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	buf := *hash
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

// Cmp compares hash and target and returns:
//
//   -1 if hash <  target
//    0 if hash == target
//   +1 if hash >  target
//
func (hash *SubNetworkHash) Cmp(target *SubNetworkHash) int {
	return HashToBig(hash).Cmp(HashToBig(target))
}

//Less returns true iff hash a is less than hash b
func Less(a *SubNetworkHash, b *SubNetworkHash) bool {
	return a.Cmp(b) < 0
}

//JoinHashesStrings joins all the stringified hashes separated by a separator
func JoinHashesStrings(hashes []SubNetworkHash, separator string) string {
	return strings.Join(Strings(hashes), separator)
}

func Sort(hashes []SubNetworkHash) {
	sort.Slice(hashes, func(i, j int) bool {
		return Less(&hashes[i], &hashes[j])
	})
}
