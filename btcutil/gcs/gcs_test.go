// Copyright (c) 2016-2017 The btcsuite developers
// Copyright (c) 2016-2017 The Lightning Network Developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gcs_test

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
	"github.com/daglabs/btcd/btcutil/gcs"
)

var (
	// No need to allocate an err variable in every test
	err error

	// Collision probability for the tests (1/2**20)
	P = uint8(20)

	// Filters are conserved between tests but we must define with an
	// interface which functions we're testing because the gcsFilter type
	// isn't exported
	filter, filter2, filter3, filter4, filter5 *gcs.Filter

	// We need to use the same key for building and querying the filters
	key [gcs.KeySize]byte

	// List of values for building a filter
	contents = [][]byte{
		[]byte("Alex"),
		[]byte("Bob"),
		[]byte("Charlie"),
		[]byte("Dick"),
		[]byte("Ed"),
		[]byte("Frank"),
		[]byte("George"),
		[]byte("Harry"),
		[]byte("Ilya"),
		[]byte("John"),
		[]byte("Kevin"),
		[]byte("Larry"),
		[]byte("Michael"),
		[]byte("Nate"),
		[]byte("Owen"),
		[]byte("Paul"),
		[]byte("Quentin"),
	}

	// List of values for querying a filter using MatchAny()
	contents2 = [][]byte{
		[]byte("Alice"),
		[]byte("Betty"),
		[]byte("Charmaine"),
		[]byte("Donna"),
		[]byte("Edith"),
		[]byte("Faina"),
		[]byte("Georgia"),
		[]byte("Hannah"),
		[]byte("Ilsbeth"),
		[]byte("Jennifer"),
		[]byte("Kayla"),
		[]byte("Lena"),
		[]byte("Michelle"),
		[]byte("Natalie"),
		[]byte("Ophelia"),
		[]byte("Peggy"),
		[]byte("Queenie"),
	}
)

// TestGCSFilterBuild builds a test filter with a randomized key. For Bitcoin
// use, deterministic filter generation is desired. Therefore, a key that's
// derived deterministically would be required.
func TestGCSFilterBuild(t *testing.T) {
	for i := 0; i < gcs.KeySize; i += 4 {
		binary.BigEndian.PutUint32(key[i:], rand.Uint32())
	}
	filter, err = gcs.BuildGCSFilter(P, key, contents)
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
}

// TestGCSFilterCopy deserializes and serializes a filter to create a copy.
func TestGCSFilterCopy(t *testing.T) {
	serialized2, err := filter.Bytes()
	if err != nil {
		t.Fatalf("Filter Bytes() failed: %v", err)
	}
	filter2, err = gcs.FromBytes(filter.N(), P, serialized2)
	if err != nil {
		t.Fatalf("Filter copy failed: %s", err.Error())
	}
	serialized3, err := filter.NBytes()
	if err != nil {
		t.Fatalf("Filter NBytes() failed: %v", err)
	}
	filter3, err = gcs.FromNBytes(filter.P(), serialized3)
	if err != nil {
		t.Fatalf("Filter copy failed: %s", err.Error())
	}
	serialized4, err := filter.PBytes()
	if err != nil {
		t.Fatalf("Filter PBytes() failed: %v", err)
	}
	filter4, err = gcs.FromPBytes(filter.N(), serialized4)
	if err != nil {
		t.Fatalf("Filter copy failed: %s", err.Error())
	}
	serialized5, err := filter.NPBytes()
	if err != nil {
		t.Fatalf("Filter NPBytes() failed: %v", err)
	}
	filter5, err = gcs.FromNPBytes(serialized5)
	if err != nil {
		t.Fatalf("Filter copy failed: %s", err.Error())
	}
}

// TestGCSFilterMetadata checks that the filter metadata is built and copied
// correctly.
func TestGCSFilterMetadata(t *testing.T) {
	if filter.P() != P {
		t.Fatal("P not correctly stored in filter metadata")
	}
	if filter.N() != uint32(len(contents)) {
		t.Fatal("N not correctly stored in filter metadata")
	}
	if filter.P() != filter2.P() {
		t.Fatal("P doesn't match between copied filters")
	}
	if filter.P() != filter3.P() {
		t.Fatal("P doesn't match between copied filters")
	}
	if filter.P() != filter4.P() {
		t.Fatal("P doesn't match between copied filters")
	}
	if filter.P() != filter5.P() {
		t.Fatal("P doesn't match between copied filters")
	}
	if filter.N() != filter2.N() {
		t.Fatal("N doesn't match between copied filters")
	}
	if filter.N() != filter3.N() {
		t.Fatal("N doesn't match between copied filters")
	}
	if filter.N() != filter4.N() {
		t.Fatal("N doesn't match between copied filters")
	}
	if filter.N() != filter5.N() {
		t.Fatal("N doesn't match between copied filters")
	}
	serialized, err := filter.Bytes()
	if err != nil {
		t.Fatalf("Filter Bytes() failed: %v", err)
	}
	serialized2, err := filter2.Bytes()
	if err != nil {
		t.Fatalf("Filter Bytes() failed: %v", err)
	}
	if !bytes.Equal(serialized, serialized2) {
		t.Fatal("Bytes don't match between copied filters")
	}
	serialized3, err := filter3.Bytes()
	if err != nil {
		t.Fatalf("Filter Bytes() failed: %v", err)
	}
	if !bytes.Equal(serialized, serialized3) {
		t.Fatal("Bytes don't match between copied filters")
	}
	serialized4, err := filter3.Bytes()
	if err != nil {
		t.Fatalf("Filter Bytes() failed: %v", err)
	}
	if !bytes.Equal(serialized, serialized4) {
		t.Fatal("Bytes don't match between copied filters")
	}
	serialized5, err := filter5.Bytes()
	if err != nil {
		t.Fatalf("Filter Bytes() failed: %v", err)
	}
	if !bytes.Equal(serialized, serialized5) {
		t.Fatal("Bytes don't match between copied filters")
	}
}

// TestGCSFilterMatch checks that both the built and copied filters match
// correctly, logging any false positives without failing on them.
func TestGCSFilterMatch(t *testing.T) {
	match, err := filter.Match(key, []byte("Nate"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err.Error())
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}
	match, err = filter2.Match(key, []byte("Nate"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err.Error())
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}
	match, err = filter.Match(key, []byte("Quentin"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err.Error())
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}
	match, err = filter2.Match(key, []byte("Quentin"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err.Error())
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}
	match, err = filter.Match(key, []byte("Nates"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err.Error())
	}
	if match {
		t.Logf("False positive match, should be 1 in 2**%d!", P)
	}
	match, err = filter2.Match(key, []byte("Nates"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err.Error())
	}
	if match {
		t.Logf("False positive match, should be 1 in 2**%d!", P)
	}
	match, err = filter.Match(key, []byte("Quentins"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err.Error())
	}
	if match {
		t.Logf("False positive match, should be 1 in 2**%d!", P)
	}
	match, err = filter2.Match(key, []byte("Quentins"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err.Error())
	}
	if match {
		t.Logf("False positive match, should be 1 in 2**%d!", P)
	}
}

// TestGCSFilterMatchAny checks that both the built and copied filters match a
// list correctly, logging any false positives without failing on them.
func TestGCSFilterMatchAny(t *testing.T) {
	match, err := filter.MatchAny(key, contents2)
	if err != nil {
		t.Fatalf("Filter match any failed: %s", err.Error())
	}
	if match {
		t.Logf("False positive match, should be 1 in 2**%d!", P)
	}
	match, err = filter2.MatchAny(key, contents2)
	if err != nil {
		t.Fatalf("Filter match any failed: %s", err.Error())
	}
	if match {
		t.Logf("False positive match, should be 1 in 2**%d!", P)
	}
	contents2 = append(contents2, []byte("Nate"))
	match, err = filter.MatchAny(key, contents2)
	if err != nil {
		t.Fatalf("Filter match any failed: %s", err.Error())
	}
	if !match {
		t.Fatal("Filter didn't match any when it should have!")
	}
	match, err = filter2.MatchAny(key, contents2)
	if err != nil {
		t.Fatalf("Filter match any failed: %s", err.Error())
	}
	if !match {
		t.Fatal("Filter didn't match any when it should have!")
	}
}
