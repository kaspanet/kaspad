// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package daghash

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// HashSize of array used to store hashes.  See Hash.
const HashSize = 32

// MaxHashStringSize is the maximum length of a Hash hash string.
const MaxHashStringSize = HashSize * 2

// ErrHashStrSize describes an error that indicates the caller specified a hash
// string that has too many characters.
var ErrHashStrSize = fmt.Errorf("max hash string length is %v bytes", MaxHashStringSize)

// Hash is used in several of the bitcoin messages and common structures.  It
// typically represents the double sha256 of data.
type Hash [HashSize]byte

// TxID is transaction hash not including payload and signature.
type TxID Hash

// String returns the Hash as the hexadecimal string of the byte-reversed
// hash.
func (hash Hash) String() string {
	for i := 0; i < HashSize/2; i++ {
		hash[i], hash[HashSize-1-i] = hash[HashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}

// String returns the TxId as the hexadecimal string of the byte-reversed
// hash.
func (txID TxID) String() string {
	return Hash(txID).String()
}

// Strings returns a slice of strings representing the hashes in the given slice of hashes
func Strings(hashes []Hash) []string {
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
func (hash *Hash) CloneBytes() []byte {
	newHash := make([]byte, HashSize)
	copy(newHash, hash[:])

	return newHash
}

// CloneBytes returns a copy of the bytes which represent the TxID as a byte
// slice.
//
// NOTE: It is generally cheaper to just slice the hash directly thereby reusing
// the same bytes rather than calling this method.
func (txID *TxID) CloneBytes() []byte {
	return (*Hash)(txID).CloneBytes()
}

// SetBytes sets the bytes which represent the hash.  An error is returned if
// the number of bytes passed in is not HashSize.
func (hash *Hash) SetBytes(newHash []byte) error {
	nhlen := len(newHash)
	if nhlen != HashSize {
		return fmt.Errorf("invalid hash length of %v, want %v", nhlen,
			HashSize)
	}
	copy(hash[:], newHash)

	return nil
}

// SetBytes sets the bytes which represent the TxID.  An error is returned if
// the number of bytes passed in is not HashSize.
func (txID *TxID) SetBytes(newID []byte) error {
	return (*Hash)(txID).SetBytes(newID)
}

// IsEqual returns true if target is the same as hash.
func (hash *Hash) IsEqual(target *Hash) bool {
	if hash == nil && target == nil {
		return true
	}
	if hash == nil || target == nil {
		return false
	}
	return *hash == *target
}

// IsEqual returns true if target is the same as TxID.
func (txID *TxID) IsEqual(target *TxID) bool {
	return (*Hash)(txID).IsEqual((*Hash)(target))
}

// AreEqual returns true if both slices contain the same hashes.
// Either slice must not contain duplicates.
func AreEqual(first []Hash, second []Hash) bool {
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
func NewHash(newHash []byte) (*Hash, error) {
	var sh Hash
	err := sh.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &sh, err
}

// NewTxID returns a new TxID from a byte slice.  An error is returned if
// the number of bytes passed in is not HashSize.
func NewTxID(newTxID []byte) (*TxID, error) {
	hash, err := NewHash(newTxID)
	return (*TxID)(hash), err
}

// NewHashFromStr creates a Hash from a hash string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the Hash.
func NewHashFromStr(hash string) (*Hash, error) {
	ret := new(Hash)
	err := Decode(ret, hash)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// NewTxIDFromStr creates a TxID from a hash string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the Hash.
func NewTxIDFromStr(idStr string) (*TxID, error) {
	hash, err := NewHashFromStr(idStr)
	return (*TxID)(hash), err
}

// Decode decodes the byte-reversed hexadecimal string encoding of a Hash to a
// destination.
func Decode(dst *Hash, src string) error {
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
	var reversedHash Hash
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
func HashToBig(hash *Hash) *big.Int {
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
func (hash *Hash) Cmp(target *Hash) int {
	return HashToBig(hash).Cmp(HashToBig(target))
}

//Less returns true iff hash a is less than hash b
func Less(a *Hash, b *Hash) bool {
	return a.Cmp(b) < 0
}

//JoinHashesStrings joins all the stringified hashes separated by a separator
func JoinHashesStrings(hashes []Hash, separator string) string {
	return strings.Join(Strings(hashes), separator)
}

// Sort sorts a slice of hashes
func Sort(hashes []Hash) {
	sort.Slice(hashes, func(i, j int) bool {
		return Less(&hashes[i], &hashes[j])
	})
}

// ZeroHash is the Hash value of all zero bytes, defined here for
// convenience.
var ZeroHash Hash

// ZeroTxID is the Hash value of all zero bytes, defined here for
// convenience.
var ZeroTxID TxID
