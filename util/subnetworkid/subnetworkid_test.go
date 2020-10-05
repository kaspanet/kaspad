// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package subnetworkid

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

// TestSubnetworkID tests the SubnetworkID API.
func TestSubnetworkID(t *testing.T) {
	subnetworkIDStr := "a3eb3f82edc878cea25ec41d6b790744e5daeef"
	subnetworkID, err := NewFromStr(subnetworkIDStr)
	if err != nil {
		t.Errorf("NewFromStr: %v", err)
	}
	subnetworkID2Bytes := []byte{
		0x79, 0xa6, 0x1a, 0xdb, 0xc6,
		0xe5, 0xa2, 0xe1, 0x39, 0xd2,
		0x71, 0x3a, 0x54, 0x6e, 0xc7,
		0xc8, 0x75, 0x63, 0x2e, 0x75,
	}

	subnetworkID2, err := New(subnetworkID2Bytes)
	if err != nil {
		t.Errorf("New: unexpected error %v", err)
	}

	// Ensure proper size.
	if len(subnetworkID2) != IDLength {
		t.Errorf("New: SubnetworkID length mismatch - got: %v, want: %v",
			len(subnetworkID2), IDLength)
	}

	// Ensure contents match.
	if !bytes.Equal(subnetworkID2[:], subnetworkID2Bytes) {
		t.Errorf("New: SubnetworkID contents mismatch - got: %v, want: %v",
			subnetworkID2[:], subnetworkID2Bytes)
	}

	if subnetworkID2.IsEqual(subnetworkID) {
		t.Errorf("IsEqual: SubnetworkID contents should not match - got: %v, want: %v",
			subnetworkID2, subnetworkID)
	}

	// Set SubnetworkID from byte slice and ensure contents match.
	err = subnetworkID2.SetBytes(subnetworkID.CloneBytes())
	if err != nil {
		t.Errorf("SetBytes: %v", err)
	}
	if !subnetworkID2.IsEqual(subnetworkID) {
		t.Errorf("IsEqual: SubnetworkID contents mismatch - got: %v, want: %v",
			subnetworkID2, subnetworkID)
	}

	// Ensure nil SubnetworkIDs are handled properly.
	if !(*SubnetworkID)(nil).IsEqual(nil) {
		t.Error("IsEqual: nil SubnetworkIDs should match")
	}
	if subnetworkID2.IsEqual(nil) {
		t.Error("IsEqual: non-nil SubnetworkID matches nil SubnetworkID")
	}

	// Invalid size for SetBytes.
	err = subnetworkID2.SetBytes([]byte{0x00})
	if err == nil {
		t.Errorf("SetBytes: failed to received expected err - got: nil")
	}

	// Invalid size for New.
	invalidSubnetworkID := make([]byte, IDLength+1)
	_, err = New(invalidSubnetworkID)
	if err == nil {
		t.Errorf("New: failed to received expected err - got: nil")
	}
}

// TestSubnetworkIDString  tests the stringized output for SubnetworkIDs.
func TestSubnetworkIDString(t *testing.T) {
	wantStr := "ecaad478d2b00432346c3f1f3986da1afd33e506"
	SubnetworkID := SubnetworkID([IDLength]byte{
		0x06, 0xe5, 0x33, 0xfd, 0x1a,
		0xda, 0x86, 0x39, 0x1f, 0x3f,
		0x6c, 0x34, 0x32, 0x04, 0xb0,
		0xd2, 0x78, 0xd4, 0xaa, 0xec,
	})

	SubnetworkIDStr := SubnetworkID.String()
	if SubnetworkIDStr != wantStr {
		t.Errorf("String: wrong SubnetworkID string - got %v, want %v",
			SubnetworkIDStr, wantStr)
	}
}

func TestSubnetworkIDsStrings(t *testing.T) {
	first := SubnetworkID{
		0x06, 0xe5, 0x33, 0xfd, 0x1a,
		0xda, 0x86, 0x39, 0x1f, 0x3f,
		0x6c, 0x34, 0x32, 0x04, 0xb0,
		0xd2, 0x78, 0xd4, 0xaa, 0xec,
	}
	firstStr := "ecaad478d2b00432346c3f1f3986da1afd33e506"

	second := SubnetworkID{}
	secondStr := "0000000000000000000000000000000000000000"

	tests := []struct {
		name            string
		SubnetworkIDs   []SubnetworkID
		expectedStrings []string
	}{
		{"empty", []SubnetworkID{}, []string{}},
		{"two SubnetworkIDs", []SubnetworkID{first, second}, []string{firstStr, secondStr}},
		{"two SubnetworkIDs inversed", []SubnetworkID{second, first}, []string{secondStr, firstStr}},
	}

	for _, test := range tests {
		strings := Strings(test.SubnetworkIDs)
		if !reflect.DeepEqual(strings, test.expectedStrings) {
			t.Errorf("SubnetworkIDsStrings: %s: expected: %v, got: %v",
				test.name, test.expectedStrings, strings)
		}
	}
}

// TestNewFromStr executes tests against the NewFromStr function.
func TestNewFromStr(t *testing.T) {
	tests := []struct {
		in          string
		expected    SubnetworkID
		expectedErr error
	}{

		// Empty string.
		{
			"",
			SubnetworkID{},
			nil,
		},

		// Single digit SubnetworkID.
		{
			"1",
			SubnetworkID([IDLength]byte{
				0x01, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00,
			}),
			nil,
		},

		{
			"a60840790ba1d475d01367e7c723da941069e9dc",
			SubnetworkID([IDLength]byte{
				0xdc, 0xe9, 0x69, 0x10, 0x94,
				0xda, 0x23, 0xc7, 0xe7, 0x67,
				0x13, 0xd0, 0x75, 0xd4, 0xa1,
				0x0b, 0x79, 0x40, 0x08, 0xa6,
			}),
			nil,
		},

		// SubnetworkID string that is too long.
		{
			"01234567890123456789012345678901234567890123456789012345678912345",
			SubnetworkID{},
			ErrIDStrSize,
		},

		// SubnetworkID string that is contains non-hex chars.
		{
			"abcdefg",
			SubnetworkID{},
			hex.InvalidByteError('g'),
		},
	}

	unexpectedErrStr := "NewFromStr #%d failed to detect expected error - got: %v want: %v"
	unexpectedResultStr := "NewFromStr #%d got: %v want: %v"
	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result, err := NewFromStr(test.in)
		if !errors.Is(err, test.expectedErr) {
			t.Errorf(unexpectedErrStr, i, err, test.expectedErr)
			continue
		} else if err != nil {
			// Got expected error. Move on to the next test.
			continue
		}
		if !test.expected.IsEqual(result) {
			t.Errorf(unexpectedResultStr, i, result, &test.expected)
			continue
		}
	}
}

// TestAreEqual executes tests against the AreEqual function.
func TestAreEqual(t *testing.T) {
	subnetworkIDAll0, _ := NewFromStr("0000000000000000000000000000000000000000")
	subnetworkIDAll1, _ := NewFromStr("1111111111111111111111111111111111111111")
	subnetworkIDAll2, _ := NewFromStr("2222222222222222222222222222222222222222")
	subnetworkIDAll3, _ := NewFromStr("3333333333333333333333333333333333333333")
	subnetworkIDs0To2 := []SubnetworkID{*subnetworkIDAll0, *subnetworkIDAll1, *subnetworkIDAll2}
	subnetworkIDs1To3 := []SubnetworkID{*subnetworkIDAll1, *subnetworkIDAll2, *subnetworkIDAll3}
	subnetworkIDs0To3 := []SubnetworkID{*subnetworkIDAll0, *subnetworkIDAll1, *subnetworkIDAll2, *subnetworkIDAll3}

	tests := []struct {
		name     string
		first    []SubnetworkID
		second   []SubnetworkID
		expected bool
	}{
		{
			name:     "self-equality",
			first:    subnetworkIDs0To2,
			second:   subnetworkIDs0To2,
			expected: true,
		},
		{
			name:     "same slice length but only some members are equal",
			first:    subnetworkIDs0To2,
			second:   subnetworkIDs1To3,
			expected: false,
		},
		{
			name:     "different slice lengths, one slice containing all the other's members",
			first:    subnetworkIDs0To3,
			second:   subnetworkIDs0To2,
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

func TestToBig(t *testing.T) {
	subnetworkIDAll0, _ := NewFromStr("0000000000000000000000000000000000000000")
	big0 := big.NewInt(0)
	subnetworkIDAll1, _ := NewFromStr("1111111111111111111111111111111111111111")
	big1 := big.NewInt(0)
	big1.SetString("1111111111111111111111111111111111111111", 16)
	subnetworkIDAll2, _ := NewFromStr("2222222222222222222222222222222222222222")
	big2 := big.NewInt(0)
	big2.SetString("2222222222222222222222222222222222222222", 16)
	subnetworkIDAll3, _ := NewFromStr("3333333333333333333333333333333333333333")
	big3 := big.NewInt(0)
	big3.SetString("3333333333333333333333333333333333333333", 16)

	tests := []struct {
		SubnetworkID *SubnetworkID
		expected     *big.Int
	}{
		{subnetworkIDAll0, big0},
		{subnetworkIDAll1, big1},
		{subnetworkIDAll2, big2},
		{subnetworkIDAll3, big3},
	}

	for _, test := range tests {
		result := ToBig(test.SubnetworkID)

		if result.Cmp(test.expected) != 0 {
			t.Errorf("unexpected ToBig result for"+
				" test \"%s\". Expected: %s, got: %s.", test.SubnetworkID, test.expected, result)
		}
	}
}

func TestSubnetworkIDCmp(t *testing.T) {
	subnetworkIDAll0, _ := NewFromStr("0000000000000000000000000000000000000000")
	subnetworkIDAll1, _ := NewFromStr("1111111111111111111111111111111111111111")
	subnetworkIDAll2, _ := NewFromStr("2222222222222222222222222222222222222222")
	SubnetworkIDEndFF, _ := NewFromStr("00000000000000000000000000000000000000FF")
	SubnetworkIDStartFF, _ := NewFromStr("FF00000000000000000000000000000000000000")

	tests := []struct {
		name     string
		first    *SubnetworkID
		second   *SubnetworkID
		expected int
	}{
		{"equal 0", subnetworkIDAll0, subnetworkIDAll0, 0},
		{"equal 2", subnetworkIDAll2, subnetworkIDAll2, 0},
		{"1 vs 0", subnetworkIDAll1, subnetworkIDAll0, 1},
		{"0 vs 1", subnetworkIDAll0, subnetworkIDAll1, -1},
		{"2 vs 1", subnetworkIDAll2, subnetworkIDAll1, 1},
		{"2 vs 0", subnetworkIDAll2, subnetworkIDAll0, 1},
		{"0 vs 2", subnetworkIDAll0, subnetworkIDAll2, -1},
		{"SubnetworkIDEndFF vs SubnetworkIDStartFF", SubnetworkIDEndFF, SubnetworkIDStartFF, -1},
		{"SubnetworkIDStartFF vs SubnetworkIDEndFF", SubnetworkIDStartFF, SubnetworkIDEndFF, 1},
	}

	for _, test := range tests {
		result := test.first.Cmp(test.second)

		if result != test.expected {
			t.Errorf("unexpected SubnetworkID.Cmp result for"+
				" test \"%s\". Expected: %d, got: %d.", test.name, test.expected, result)
		}
	}
}

func TestSubnetworkIDLess(t *testing.T) {
	subnetworkIDAll0, _ := NewFromStr("0000000000000000000000000000000000000000")
	subnetworkIDAll1, _ := NewFromStr("1111111111111111111111111111111111111111")
	subnetworkIDAll2, _ := NewFromStr("2222222222222222222222222222222222222222")
	SubnetworkIDEndFF, _ := NewFromStr("00000000000000000000000000000000000000FF")
	SubnetworkIDStartFF, _ := NewFromStr("FF00000000000000000000000000000000000000")

	tests := []struct {
		name     string
		first    *SubnetworkID
		second   *SubnetworkID
		expected bool
	}{
		{"equal 0", subnetworkIDAll0, subnetworkIDAll0, false},
		{"equal 2", subnetworkIDAll2, subnetworkIDAll2, false},
		{"1 vs 0", subnetworkIDAll1, subnetworkIDAll0, false},
		{"0 vs 1", subnetworkIDAll0, subnetworkIDAll1, true},
		{"2 vs 1", subnetworkIDAll2, subnetworkIDAll1, false},
		{"2 vs 0", subnetworkIDAll2, subnetworkIDAll0, false},
		{"0 vs 2", subnetworkIDAll0, subnetworkIDAll2, true},
		{"SubnetworkIDEndFF vs SubnetworkIDStartFF", SubnetworkIDEndFF, SubnetworkIDStartFF, true},
		{"SubnetworkIDStartFF vs SubnetworkIDEndFF", SubnetworkIDStartFF, SubnetworkIDEndFF, false},
	}

	for _, test := range tests {
		result := Less(test.first, test.second)

		if result != test.expected {
			t.Errorf("unexpected SubnetworkID.Less result for"+
				" test \"%s\". Expected: %t, got: %t.", test.name, test.expected, result)
		}
	}
}

func TestSort(t *testing.T) {
	subnetworkIDAll0, _ := NewFromStr("0000000000000000000000000000000000000000")
	subnetworkIDAll1, _ := NewFromStr("1111111111111111111111111111111111111111")
	subnetworkIDAll2, _ := NewFromStr("2222222222222222222222222222222222222222")
	subnetworkIDAll3, _ := NewFromStr("3333333333333333333333333333333333333333")
	SubnetworkIDEndFF, _ := NewFromStr("00000000000000000000000000000000000000FF")
	SubnetworkIDStartFF, _ := NewFromStr("FF00000000000000000000000000000000000000")

	tests := []struct {
		name          string
		SubnetworkIDs []*SubnetworkID
		expected      []*SubnetworkID
	}{
		{"empty", []*SubnetworkID{}, []*SubnetworkID{}},
		{"single item", []*SubnetworkID{subnetworkIDAll0}, []*SubnetworkID{subnetworkIDAll0}},
		{"already sorted", []*SubnetworkID{subnetworkIDAll0, SubnetworkIDEndFF, subnetworkIDAll1, subnetworkIDAll2, subnetworkIDAll3, SubnetworkIDStartFF}, []*SubnetworkID{subnetworkIDAll0, SubnetworkIDEndFF, subnetworkIDAll1, subnetworkIDAll2, subnetworkIDAll3, SubnetworkIDStartFF}},
		{"inverted", []*SubnetworkID{SubnetworkIDStartFF, subnetworkIDAll3, subnetworkIDAll2, subnetworkIDAll1, SubnetworkIDEndFF, subnetworkIDAll0}, []*SubnetworkID{subnetworkIDAll0, SubnetworkIDEndFF, subnetworkIDAll1, subnetworkIDAll2, subnetworkIDAll3, SubnetworkIDStartFF}},
		{"shuffled", []*SubnetworkID{subnetworkIDAll2, SubnetworkIDEndFF, subnetworkIDAll3, subnetworkIDAll0, SubnetworkIDStartFF, subnetworkIDAll1}, []*SubnetworkID{subnetworkIDAll0, SubnetworkIDEndFF, subnetworkIDAll1, subnetworkIDAll2, subnetworkIDAll3, SubnetworkIDStartFF}},
		{"with duplicates", []*SubnetworkID{SubnetworkIDEndFF, subnetworkIDAll2, subnetworkIDAll3, SubnetworkIDEndFF, subnetworkIDAll0, subnetworkIDAll1, subnetworkIDAll1, SubnetworkIDStartFF}, []*SubnetworkID{subnetworkIDAll0, SubnetworkIDEndFF, SubnetworkIDEndFF, subnetworkIDAll1, subnetworkIDAll1, subnetworkIDAll2, subnetworkIDAll3, SubnetworkIDStartFF}},
	}

	for _, test := range tests {
		sort.Slice(test.SubnetworkIDs, func(i, j int) bool {
			return Less(test.SubnetworkIDs[i], test.SubnetworkIDs[j])
		})

		if !reflect.DeepEqual(test.SubnetworkIDs, test.expected) {
			t.Errorf("unexpected Sort result for"+
				" test \"%s\". Expected: %v, got: %v.", test.name, test.expected, test.SubnetworkIDs)
		}
	}
}

func SubnetworkIDFlipBit(subnetworkID SubnetworkID, bit int) SubnetworkID {
	word := bit / 8
	bit = bit % 8
	subnetworkID[word] ^= 1 << bit
	return subnetworkID
}

func TestSubnetworkID_Cmp(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	for i := 0; i < 100; i++ {
		SubnetworkID := SubnetworkID{}
		n, err := r.Read(SubnetworkID[:])
		if err != nil {
			t.Fatalf("Failed generating a random SubnetworkID '%s'", err)
		} else if n != len(SubnetworkID) {
			t.Fatalf("Failed generating a random SubnetworkID, expected reading: %d. instead read: %d.", len(SubnetworkID), n)
		}
		SubnetworkIDBig := ToBig(&SubnetworkID)
		// Iterate bit by bit, flip it and compare.
		for bit := 0; bit < IDLength*8; bit++ {
			New := SubnetworkIDFlipBit(SubnetworkID, bit)
			if SubnetworkID.Cmp(&New) != SubnetworkIDBig.Cmp(ToBig(&New)) {
				t.Errorf("SubnetworkID.Cmp disagrees with big.Int.Cmp New: %s, SubnetworkID: %s", New, SubnetworkID)
			}
		}
		for bit := 0; bit < IDLength*8; bit++ {
			flippedFromLeft := SubnetworkIDFlipBit(SubnetworkID, bit)
			flippedFromRight := SubnetworkIDFlipBit(SubnetworkID, IDLength*8-bit-1)
			if flippedFromLeft.Cmp(&flippedFromRight) != ToBig(&flippedFromLeft).Cmp(ToBig(&flippedFromRight)) {
				t.Errorf("SubnetworkID.Cmp disagrees with big.Int.Cmp flippedFromLeft: %s, flippedFromRight: %s", flippedFromLeft, flippedFromRight)
			}
		}
	}
}

func BenchmarkSubnetworkID_Cmp(b *testing.B) {
	subnetworkIDAll0, err := NewFromStr("3333333333333333333333333333333333333333")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		subnetworkIDAll0.Cmp(subnetworkIDAll0)
	}
}

func BenchmarkSubnetworkID_Equal(b *testing.B) {
	subnetworkIDAll0, err := NewFromStr("3333333333333333333333333333333333333333")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		subnetworkIDAll0.IsEqual(subnetworkIDAll0)
	}
}
