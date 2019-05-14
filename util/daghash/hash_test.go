// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package daghash

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"reflect"
	"testing"
)

// mainNetGenesisHash is the hash of the first block in the block chain for the
// main network (genesis block).
var mainNetGenesisHash = Hash([HashSize]byte{
	0xdc, 0x5f, 0x5b, 0x5b, 0x1d, 0xc2, 0xa7, 0x25,
	0x49, 0xd5, 0x1d, 0x4d, 0xee, 0xd7, 0xa4, 0x8b,
	0xaf, 0xd3, 0x14, 0x4b, 0x56, 0x78, 0x98, 0xb1,
	0x8c, 0xfd, 0x9f, 0x69, 0xdd, 0xcf, 0xbb, 0x63,
})

// TestHash tests the Hash API.
func TestHash(t *testing.T) {
	// Hash of block 234439.
	blockHashStr := "14a0810ac680a3eb3f82edc878cea25ec41d6b790744e5daeef"
	blockHash, err := NewHashFromStr(blockHashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// Hash of block 234440 as byte slice.
	buf := []byte{
		0x79, 0xa6, 0x1a, 0xdb, 0xc6, 0xe5, 0xa2, 0xe1,
		0x39, 0xd2, 0x71, 0x3a, 0x54, 0x6e, 0xc7, 0xc8,
		0x75, 0x63, 0x2e, 0x75, 0xf1, 0xdf, 0x9c, 0x3f,
		0xa6, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	hash, err := NewHash(buf)
	if err != nil {
		t.Errorf("NewHash: unexpected error %v", err)
	}

	// Ensure proper size.
	if len(hash) != HashSize {
		t.Errorf("NewHash: hash length mismatch - got: %v, want: %v",
			len(hash), HashSize)
	}

	// Ensure contents match.
	if !bytes.Equal(hash[:], buf) {
		t.Errorf("NewHash: hash contents mismatch - got: %v, want: %v",
			hash[:], buf)
	}

	// Ensure contents of hash of block 234440 don't match 234439.
	if hash.IsEqual(blockHash) {
		t.Errorf("IsEqual: hash contents should not match - got: %v, want: %v",
			hash, blockHash)
	}

	// Set hash from byte slice and ensure contents match.
	err = hash.SetBytes(blockHash.CloneBytes())
	if err != nil {
		t.Errorf("SetBytes: %v", err)
	}
	if !hash.IsEqual(blockHash) {
		t.Errorf("IsEqual: hash contents mismatch - got: %v, want: %v",
			hash, blockHash)
	}

	// Ensure nil hashes are handled properly.
	if !(*Hash)(nil).IsEqual(nil) {
		t.Error("IsEqual: nil hashes should match")
	}
	if hash.IsEqual(nil) {
		t.Error("IsEqual: non-nil hash matches nil hash")
	}

	// Invalid size for SetBytes.
	err = hash.SetBytes([]byte{0x00})
	if err == nil {
		t.Errorf("SetBytes: failed to received expected err - got: nil")
	}

	// Invalid size for NewHash.
	invalidHash := make([]byte, HashSize+1)
	_, err = NewHash(invalidHash)
	if err == nil {
		t.Errorf("NewHash: failed to received expected err - got: nil")
	}
}

// TestHashString  tests the stringized output for hashes.
func TestHashString(t *testing.T) {
	// Block 100000 hash.
	wantStr := "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"
	hash := Hash([HashSize]byte{
		0x06, 0xe5, 0x33, 0xfd, 0x1a, 0xda, 0x86, 0x39,
		0x1f, 0x3f, 0x6c, 0x34, 0x32, 0x04, 0xb0, 0xd2,
		0x78, 0xd4, 0xaa, 0xec, 0x1c, 0x0b, 0x20, 0xaa,
		0x27, 0xba, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
	})

	hashStr := hash.String()
	if hashStr != wantStr {
		t.Errorf("String: wrong hash string - got %v, want %v",
			hashStr, wantStr)
	}
}

func TestHashesStrings(t *testing.T) {
	first := &Hash{
		0x06, 0xe5, 0x33, 0xfd, 0x1a, 0xda, 0x86, 0x39,
		0x1f, 0x3f, 0x6c, 0x34, 0x32, 0x04, 0xb0, 0xd2,
		0x78, 0xd4, 0xaa, 0xec, 0x1c, 0x0b, 0x20, 0xaa,
		0x27, 0xba, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	firstStr := "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"

	second := &Hash{}
	secondStr := "0000000000000000000000000000000000000000000000000000000000000000"

	tests := []struct {
		name            string
		hashes          []*Hash
		expectedStrings []string
	}{
		{"empty", []*Hash{}, []string{}},
		{"two hashes", []*Hash{first, second}, []string{firstStr, secondStr}},
		{"two hashes inversed", []*Hash{second, first}, []string{secondStr, firstStr}},
	}

	for _, test := range tests {
		strings := Strings(test.hashes)
		if !reflect.DeepEqual(strings, test.expectedStrings) {
			t.Errorf("HashesStrings: %s: expected: %v, got: %v",
				test.name, test.expectedStrings, strings)
		}
	}
}

// TestNewHashFromStr executes tests against the NewHashFromStr function.
func TestNewHashFromStr(t *testing.T) {
	tests := []struct {
		in   string
		want Hash
		err  error
	}{
		// Genesis hash.
		{
			"000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
			mainNetGenesisHash,
			nil,
		},

		// Genesis hash with stripped leading zeros.
		{
			"19d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
			mainNetGenesisHash,
			nil,
		},

		// Empty string.
		{
			"",
			Hash{},
			nil,
		},

		// Single digit hash.
		{
			"1",
			Hash([HashSize]byte{
				0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			}),
			nil,
		},

		// Block 203707 with stripped leading zeros.
		{
			"3264bc2ac36a60840790ba1d475d01367e7c723da941069e9dc",
			Hash([HashSize]byte{
				0xdc, 0xe9, 0x69, 0x10, 0x94, 0xda, 0x23, 0xc7,
				0xe7, 0x67, 0x13, 0xd0, 0x75, 0xd4, 0xa1, 0x0b,
				0x79, 0x40, 0x08, 0xa6, 0x36, 0xac, 0xc2, 0x4b,
				0x26, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			}),
			nil,
		},

		// Hash string that is too long.
		{
			"01234567890123456789012345678901234567890123456789012345678912345",
			Hash{},
			ErrHashStrSize,
		},

		// Hash string that is contains non-hex chars.
		{
			"abcdefg",
			Hash{},
			hex.InvalidByteError('g'),
		},
	}

	unexpectedErrStr := "NewHashFromStr #%d failed to detect expected error - got: %v want: %v"
	unexpectedResultStr := "NewHashFromStr #%d got: %v want: %v"
	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result, err := NewHashFromStr(test.in)
		if err != test.err {
			t.Errorf(unexpectedErrStr, i, err, test.err)
			continue
		} else if err != nil {
			// Got expected error. Move on to the next test.
			continue
		}
		if !test.want.IsEqual(result) {
			t.Errorf(unexpectedResultStr, i, result, &test.want)
			continue
		}
	}
}

// TestAreEqual executes tests against the AreEqual function.
func TestAreEqual(t *testing.T) {
	hash0, _ := NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	hash2, _ := NewHashFromStr("2222222222222222222222222222222222222222222222222222222222222222")
	hash3, _ := NewHashFromStr("3333333333333333333333333333333333333333333333333333333333333333")
	hashes0To2 := []*Hash{hash0, hash1, hash2}
	hashes1To3 := []*Hash{hash1, hash2, hash3}
	hashes0To3 := []*Hash{hash0, hash1, hash2, hash3}

	tests := []struct {
		name     string
		first    []*Hash
		second   []*Hash
		expected bool
	}{
		{
			name:     "self-equality",
			first:    hashes0To2,
			second:   hashes0To2,
			expected: true,
		},
		{
			name:     "same slice length but only some members are equal",
			first:    hashes0To2,
			second:   hashes1To3,
			expected: false,
		},
		{
			name:     "different slice lengths, one slice containing all the other's members",
			first:    hashes0To3,
			second:   hashes0To2,
			expected: false,
		},
	}

	for _, test := range tests {
		result := AreEqual(test.first, test.second)
		if result != test.expected {
			t.Errorf("unexpected AreEqual result for"+
				" test \"%s\". Expected: %t, got: %t.", test.name, test.expected, result)
		}
	}
}

func TestHashToBig(t *testing.T) {
	hash0, _ := NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	big0 := big.NewInt(0)
	hash1, _ := NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	big1 := big.NewInt(0)
	big1.SetString("1111111111111111111111111111111111111111111111111111111111111111", 16)
	hash2, _ := NewHashFromStr("2222222222222222222222222222222222222222222222222222222222222222")
	big2 := big.NewInt(0)
	big2.SetString("2222222222222222222222222222222222222222222222222222222222222222", 16)
	hash3, _ := NewHashFromStr("3333333333333333333333333333333333333333333333333333333333333333")
	big3 := big.NewInt(0)
	big3.SetString("3333333333333333333333333333333333333333333333333333333333333333", 16)

	tests := []struct {
		hash     *Hash
		expected *big.Int
	}{
		{hash0, big0},
		{hash1, big1},
		{hash2, big2},
		{hash3, big3},
	}

	for _, test := range tests {
		result := HashToBig(test.hash)

		if result.Cmp(test.expected) != 0 {
			t.Errorf("unexpected HashToBig result for"+
				" test \"%s\". Expected: %s, got: %s.", test.hash, test.expected, result)
		}
	}
}

func TestHashCmp(t *testing.T) {
	hash0, _ := NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	hash2, _ := NewHashFromStr("2222222222222222222222222222222222222222222222222222222222222222")

	tests := []struct {
		name     string
		first    *Hash
		second   *Hash
		expected int
	}{
		{"equal 0", hash0, hash0, 0},
		{"equal 2", hash2, hash2, 0},
		{"1 vs 0", hash1, hash0, 1},
		{"0 vs 1", hash0, hash1, -1},
		{"2 vs 1", hash2, hash1, 1},
		{"2 vs 0", hash2, hash0, 1},
		{"0 vs 2", hash0, hash2, -1},
	}

	for _, test := range tests {
		result := test.first.Cmp(test.second)

		if result != test.expected {
			t.Errorf("unexpected Hash.Cmp result for"+
				" test \"%s\". Expected: %d, got: %d.", test.name, test.expected, result)
		}
	}
}

func TestHashLess(t *testing.T) {
	hash0, _ := NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	hash2, _ := NewHashFromStr("2222222222222222222222222222222222222222222222222222222222222222")

	tests := []struct {
		name     string
		first    *Hash
		second   *Hash
		expected bool
	}{
		{"equal 0", hash0, hash0, false},
		{"equal 2", hash2, hash2, false},
		{"1 vs 0", hash1, hash0, false},
		{"0 vs 1", hash0, hash1, true},
		{"2 vs 1", hash2, hash1, false},
		{"2 vs 0", hash2, hash0, false},
		{"0 vs 2", hash0, hash2, true},
	}

	for _, test := range tests {
		result := Less(test.first, test.second)

		if result != test.expected {
			t.Errorf("unexpected Hash.Less result for"+
				" test \"%s\". Expected: %t, got: %t.", test.name, test.expected, result)
		}
	}
}

func TestJoinHashesStrings(t *testing.T) {
	hash0, _ := NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")

	tests := []struct {
		name      string
		hashes    []*Hash
		separator string
		expected  string
	}{
		{"no separator", []*Hash{hash0, hash1}, "",
			"00000000000000000000000000000000000000000000000000000000000000001111111111111111111111111111111111111111111111111111111111111111"},
		{", separator", []*Hash{hash0, hash1}, ",",
			"0000000000000000000000000000000000000000000000000000000000000000,1111111111111111111111111111111111111111111111111111111111111111"},
		{"blabla separator", []*Hash{hash0, hash1}, "blabla",
			"0000000000000000000000000000000000000000000000000000000000000000blabla1111111111111111111111111111111111111111111111111111111111111111"},
		{"1 hash", []*Hash{hash0}, ",", "0000000000000000000000000000000000000000000000000000000000000000"},
		{"0 hashes", []*Hash{}, ",", ""},
	}

	for _, test := range tests {
		result := JoinHashesStrings(test.hashes, test.separator)

		if result != test.expected {
			t.Errorf("unexpected JoinHashesStrings result for"+
				" test \"%s\". Expected: %s, got: %s.", test.name, test.expected, result)
		}
	}
}

func TestSort(t *testing.T) {
	hash0, _ := NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	hash2, _ := NewHashFromStr("2222222222222222222222222222222222222222222222222222222222222222")
	hash3, _ := NewHashFromStr("3333333333333333333333333333333333333333333333333333333333333333")

	tests := []struct {
		name     string
		hashes   []*Hash
		expected []*Hash
	}{
		{"empty", []*Hash{}, []*Hash{}},
		{"single item", []*Hash{hash0}, []*Hash{hash0}},
		{"already sorted", []*Hash{hash0, hash1, hash2, hash3}, []*Hash{hash0, hash1, hash2, hash3}},
		{"inverted", []*Hash{hash3, hash2, hash1, hash0}, []*Hash{hash0, hash1, hash2, hash3}},
		{"shuffled", []*Hash{hash2, hash3, hash0, hash1}, []*Hash{hash0, hash1, hash2, hash3}},
		{"with duplicates", []*Hash{hash2, hash3, hash0, hash1, hash1}, []*Hash{hash0, hash1, hash1, hash2, hash3}},
	}

	for _, test := range tests {
		Sort(test.hashes)

		if !reflect.DeepEqual(test.hashes, test.expected) {
			t.Errorf("unexpected Sort result for"+
				" test \"%s\". Expected: %v, got: %v.", test.name, test.expected, test.hashes)
		}
	}
}
