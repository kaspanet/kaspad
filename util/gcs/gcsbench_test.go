// Copyright (c) 2016-2017 The btcsuite developers
// Copyright (c) 2016-2017 The Lightning Network Developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gcs_test

import (
	"encoding/binary"
	"math/rand"
	"testing"
	"github.com/daglabs/btcd/util/gcs"
)

func genRandFilterElements(numElements uint) ([][]byte, error) {
	testContents := make([][]byte, numElements)
	for i := range contents {
		randElem := make([]byte, 32)
		if _, err := rand.Read(randElem); err != nil {
			return nil, err
		}
		testContents[i] = randElem
	}

	return testContents, nil
}

var (
	generatedFilter *gcs.Filter
	filterErr       error
)

// BenchmarkGCSFilterBuild benchmarks building a filter.
func BenchmarkGCSFilterBuild50000(b *testing.B) {
	b.StopTimer()
	var testKey [gcs.KeySize]byte
	for i := 0; i < gcs.KeySize; i += 4 {
		binary.BigEndian.PutUint32(testKey[i:], rand.Uint32())
	}
	randFilterElems, genErr := genRandFilterElements(50000)
	if err != nil {
		b.Fatalf("unable to generate random item: %v", genErr)
	}
	b.StartTimer()

	var localFilter *gcs.Filter
	for i := 0; i < b.N; i++ {
		localFilter, err = gcs.BuildGCSFilter(P, key,
			randFilterElems)
		if err != nil {
			b.Fatalf("unable to generate filter: %v", err)
		}
	}
	generatedFilter = localFilter
}

// BenchmarkGCSFilterBuild benchmarks building a filter.
func BenchmarkGCSFilterBuild100000(b *testing.B) {
	b.StopTimer()
	var testKey [gcs.KeySize]byte
	for i := 0; i < gcs.KeySize; i += 4 {
		binary.BigEndian.PutUint32(testKey[i:], rand.Uint32())
	}
	randFilterElems, genErr := genRandFilterElements(100000)
	if err != nil {
		b.Fatalf("unable to generate random item: %v", genErr)
	}
	b.StartTimer()

	var localFilter *gcs.Filter
	for i := 0; i < b.N; i++ {
		localFilter, err = gcs.BuildGCSFilter(P, key,
			randFilterElems)
		if err != nil {
			b.Fatalf("unable to generate filter: %v", err)
		}
	}
	generatedFilter = localFilter
}

var (
	match bool
)

// BenchmarkGCSFilterMatch benchmarks querying a filter for a single value.
func BenchmarkGCSFilterMatch(b *testing.B) {
	b.StopTimer()
	filter, err := gcs.BuildGCSFilter(P, key, contents)
	if err != nil {
		b.Fatalf("Failed to build filter")
	}
	b.StartTimer()

	var (
		localMatch bool
	)
	for i := 0; i < b.N; i++ {
		localMatch, err = filter.Match(key, []byte("Nate"))
		if err != nil {
			b.Fatalf("unable to match filter: %v", err)
		}

		localMatch, err = filter.Match(key, []byte("Nates"))
		if err != nil {
			b.Fatalf("unable to match filter: %v", err)
		}
	}
	match = localMatch
}

// BenchmarkGCSFilterMatchAny benchmarks querying a filter for a list of
// values.
func BenchmarkGCSFilterMatchAny(b *testing.B) {
	b.StopTimer()
	filter, err := gcs.BuildGCSFilter(P, key, contents)
	if err != nil {
		b.Fatalf("Failed to build filter")
	}
	b.StartTimer()

	var (
		localMatch bool
	)
	for i := 0; i < b.N; i++ {
		localMatch, err = filter.MatchAny(key, contents2)
		if err != nil {
			b.Fatalf("unable to match filter: %v", err)
		}
	}
	match = localMatch
}
